package core

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() error = %v", err)
	}

	// Verify verifier is 32 bytes base64 rawurl encoded
	verifierBytes, err := base64.RawURLEncoding.DecodeString(verifier)
	if err != nil {
		t.Fatalf("verifier is not valid base64: %v", err)
	}
	if len(verifierBytes) != 32 {
		t.Errorf("verifier length = %d, want 32", len(verifierBytes))
	}

	// Verify challenge is S256 hash of verifier
	hash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	if challenge != expectedChallenge {
		t.Errorf("challenge = %s, want %s", challenge, expectedChallenge)
	}

	// Generate another pair to ensure they're different (random)
	verifier2, challenge2, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() second call error = %v", err)
	}
	if verifier == verifier2 {
		t.Error("two calls to GeneratePKCE produced same verifier")
	}
	if challenge == challenge2 {
		t.Error("two calls to GeneratePKCE produced same challenge")
	}
}

func TestBuildAuthURL(t *testing.T) {
	appID := "test-app-id"
	state := "test-state"
	challenge := "test-challenge"

	urlStr := BuildAuthURL(appID, state, challenge)

	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("BuildAuthURL returned invalid URL: %v", err)
	}

	if parsed.Scheme != "https" {
		t.Errorf("scheme = %s, want https", parsed.Scheme)
	}
	if parsed.Host != "open.feishu.cn" {
		t.Errorf("host = %s, want open.feishu.cn", parsed.Host)
	}
	if parsed.Path != "/open-apis/authen/v1/authorize" {
		t.Errorf("path = %s, want /open-apis/authen/v1/authorize", parsed.Path)
	}

	params := parsed.Query()
	if params.Get("app_id") != appID {
		t.Errorf("app_id = %s, want %s", params.Get("app_id"), appID)
	}
	if params.Get("redirect_uri") != redirectURI {
		t.Errorf("redirect_uri = %s, want %s", params.Get("redirect_uri"), redirectURI)
	}
	if params.Get("scope") != scope {
		t.Errorf("scope = %s, want %s", params.Get("scope"), scope)
	}
	if params.Get("response_type") != "code" {
		t.Errorf("response_type = %s, want code", params.Get("response_type"))
	}
	if params.Get("state") != state {
		t.Errorf("state = %s, want %s", params.Get("state"), state)
	}
	if params.Get("code_challenge") != challenge {
		t.Errorf("code_challenge = %s, want %s", params.Get("code_challenge"), challenge)
	}
	if params.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method = %s, want S256", params.Get("code_challenge_method"))
	}
}

func TestOAuthTokenStruct(t *testing.T) {
	token := OAuthToken{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    7200,
		TokenType:    "Bearer",
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %s, want test-access-token", token.AccessToken)
	}
	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("RefreshToken = %s, want test-refresh-token", token.RefreshToken)
	}
	if token.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn = %d, want 7200", token.ExpiresIn)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %s, want Bearer", token.TokenType)
	}
}

func TestConstants(t *testing.T) {
	if redirectURI != "http://127.0.0.1:8088/callback" {
		t.Errorf("redirectURI = %s, want http://127.0.0.1:8088/callback", redirectURI)
	}
	if scope != "docx:document:readonly drive:file:readonly wiki:wiki:readonly" {
		t.Errorf("scope = %s, want docx:document:readonly drive:file:readonly wiki:wiki:readonly", scope)
	}
	if !strings.Contains(authEndpoint, "open.feishu.cn") {
		t.Errorf("authEndpoint should contain open.feishu.cn")
	}
	if !strings.Contains(tokenEndpoint, "open.feishu.cn") {
		t.Errorf("tokenEndpoint should contain open.feishu.cn")
	}
}

func TestExchangeCodeForToken(t *testing.T) {
	expectedToken := &OAuthToken{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    7200,
		TokenType:    "Bearer",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		resp := struct {
			Code int        `json:"code"`
			Msg  string     `json:"msg"`
			Data OAuthToken `json:"data"`
		}{
			Code: 0,
			Msg:  "success",
			Data: *expectedToken,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the token endpoint for testing
	originalEndpoint := tokenEndpoint
	tokenEndpoint = server.URL
	defer func() { tokenEndpoint = originalEndpoint }()

	token, err := ExchangeCodeForToken("test-client-id", "test-client-secret", "test-code", "test-verifier")
	if err != nil {
		t.Fatalf("ExchangeCodeForToken() error = %v", err)
	}
	if token.AccessToken != expectedToken.AccessToken {
		t.Errorf("AccessToken = %s, want %s", token.AccessToken, expectedToken.AccessToken)
	}
	if token.RefreshToken != expectedToken.RefreshToken {
		t.Errorf("RefreshToken = %s, want %s", token.RefreshToken, expectedToken.RefreshToken)
	}
	if token.ExpiresIn != expectedToken.ExpiresIn {
		t.Errorf("ExpiresIn = %d, want %d", token.ExpiresIn, expectedToken.ExpiresIn)
	}
}

func TestExchangeCodeForTokenValidation(t *testing.T) {
	_, err := ExchangeCodeForToken("test-client-id", "test-client-secret", "", "test-verifier")
	if err == nil {
		t.Error("ExchangeCodeForToken() expected error for empty code, got nil")
	}
	if err.Error() != "code is required" {
		t.Errorf("error message = %s, want 'code is required'", err.Error())
	}
}

func TestRefreshUserToken(t *testing.T) {
	expectedToken := &OAuthToken{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    7200,
		TokenType:    "Bearer",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		resp := struct {
			Code int        `json:"code"`
			Msg  string     `json:"msg"`
			Data OAuthToken `json:"data"`
		}{
			Code: 0,
			Msg:  "success",
			Data: *expectedToken,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalEndpoint := tokenEndpoint
	tokenEndpoint = server.URL
	defer func() { tokenEndpoint = originalEndpoint }()

	token, err := RefreshUserToken("test-client-id", "test-client-secret", "test-refresh-token")
	if err != nil {
		t.Fatalf("RefreshUserToken() error = %v", err)
	}
	if token.AccessToken != expectedToken.AccessToken {
		t.Errorf("AccessToken = %s, want %s", token.AccessToken, expectedToken.AccessToken)
	}
	if token.RefreshToken != expectedToken.RefreshToken {
		t.Errorf("RefreshToken = %s, want %s", token.RefreshToken, expectedToken.RefreshToken)
	}
}

func TestRefreshUserTokenValidation(t *testing.T) {
	_, err := RefreshUserToken("test-client-id", "test-client-secret", "")
	if err == nil {
		t.Error("RefreshUserToken() expected error for empty refreshToken, got nil")
	}
	if err.Error() != "refreshToken is required" {
		t.Errorf("error message = %s, want 'refreshToken is required'", err.Error())
	}
}
