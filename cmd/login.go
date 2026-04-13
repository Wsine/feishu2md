package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/Wsine/feishu2md/core"
)

type LoginOpts struct {
	port int
}

var loginOpts = LoginOpts{port: 8088}

func handleLoginCommand() error {
	// 1. Load config to get app_id and app_secret
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := core.ReadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if cfg.Feishu.AppId == "" || cfg.Feishu.AppSecret == "" {
		return fmt.Errorf("app_id and app_secret are required in config, please run 'feishu2md config --appId <id> --appSecret <secret>' first")
	}

	// 2. Generate PKCE
	codeVerifier, codeChallenge, err := core.GeneratePKCE()
	if err != nil {
		return fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// 3. Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// 4. Build auth URL
	authURL := core.BuildAuthURL(cfg.Feishu.AppId, state, codeChallenge)

	// 5. Display URL in terminal
	fmt.Println("Please visit the following URL to login:")
	fmt.Println(authURL)
	fmt.Println()

	// 6. Open browser (platform-specific)
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Println("Please manually open the URL above in your browser.")
	}

	// 7. Start callback server
	fmt.Printf("Waiting for callback on http://127.0.0.1:%d/callback...\n", loginOpts.port)

	done := make(chan error, 1)
	go func() {
		done <- startCallbackServer(loginOpts.port, codeVerifier, configPath, cfg, state)
	}()

	// Wait for callback or timeout
	select {
	case err := <-done:
		if err != nil {
			return err
		}
		fmt.Println("Login successful! Tokens have been saved to config.")
		return nil
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("login timed out after 5 minutes")
	}
}

func startCallbackServer(port int, codeVerifier string, configPath string, cfg *core.Config, expectedState string) error {
	mux := http.NewServeMux()

	// Use a channel to signal callback completion with error
	callbackDone := make(chan error, 1)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Check for error in callback
		if errMsg := query.Get("error"); errMsg != "" {
			callbackDone <- fmt.Errorf("oauth error: %s - %s", errMsg, query.Get("error_description"))
			return
		}

		// Validate state
		state := query.Get("state")
		if state != expectedState {
			callbackDone <- fmt.Errorf("state mismatch: expected %s, got %s", expectedState, state)
			return
		}

		code := query.Get("code")
		if code == "" {
			callbackDone <- fmt.Errorf("no code received in callback")
			return
		}

		// Exchange code for tokens
		token, err := core.ExchangeCodeForToken(cfg.Feishu.AppId, cfg.Feishu.AppSecret, code, codeVerifier)
		if err != nil {
			callbackDone <- fmt.Errorf("failed to exchange code for token: %w", err)
			return
		}

		// Update config with tokens
		cfg.Feishu.UserAccessToken = token.AccessToken
		cfg.Feishu.RefreshToken = token.RefreshToken
		cfg.Feishu.TokenExpireTime = time.Now().Unix() + int64(token.ExpiresIn)

		// Save config
		if err := cfg.WriteConfig2File(configPath); err != nil {
			callbackDone <- fmt.Errorf("failed to save config: %w", err)
			return
		}

		// Signal success
		callbackDone <- nil

		// Return success HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Login Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 40px; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #52c41a; margin-bottom: 10px; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Login Successful!</h1>
        <p>You can now close this window and use feishu2md commands.</p>
    </div>
</body>
</html>`))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for either callback completion or error
	select {
	case callbackErr := <-callbackDone:
		server.Close()
		return callbackErr
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

func openBrowser(url string) error {
	var cmd string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, url).Start()
}

func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
