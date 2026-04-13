package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authEndpoint = "https://open.feishu.cn/open-apis/authen/v1/authorize"
	redirectURI  = "http://127.0.0.1:8088/callback"
	scope        = "docx:document:readonly drive:drive:readonly wiki:wiki:readonly"
)

var (
	tokenEndpoint = "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
	defaultClient = &http.Client{Timeout: 30 * time.Second}
)

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func GeneratePKCE() (verifier string, challenge string, err error) {
	// Generate 32 random bytes for code verifier
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes)

	// Generate S256 code challenge using SHA256 hash
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

func BuildAuthURL(appID, state, codeChallenge string) string {
	params := url.Values{}
	params.Set("app_id", appID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", scope)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", authEndpoint, params.Encode())
}

func ExchangeCodeForToken(clientID, clientSecret, code, codeVerifier string) (*OAuthToken, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("app_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	return doTokenRequest(data)
}

func RefreshUserToken(clientID, clientSecret, refreshToken string) (*OAuthToken, error) {
	if refreshToken == "" {
		return nil, errors.New("refreshToken is required")
	}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("app_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)

	return doTokenRequest(data)
}

func doTokenRequest(data url.Values) (*OAuthToken, error) {
	// Use HTTP Basic Auth for client_secret only, keep app_id in body
	var clientSecret string
	appID := data.Get("app_id")
	if appID != "" {
		clientSecret = data.Get("client_secret")
		data.Del("client_secret") // Only remove secret from body, keep app_id
	}

	bodyReader := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if clientSecret != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(appID + ":" + clientSecret))
		req.Header.Set("Authorization", "Basic "+creds)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB limit
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Code    int         `json:"code"`
		Msg     string      `json:"msg"`
		Data    OAuthToken  `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("token request failed: code=%d, msg=%s", result.Code, result.Msg)
	}

	return &result.Data, nil
}
