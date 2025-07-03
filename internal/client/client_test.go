package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test helper to create temporary directory
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "client-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// Test helper to create credentials file
func createCredentialsFile(t *testing.T, dir, username, password string) string {
	envFile := filepath.Join(dir, ".env")
	content := fmt.Sprintf("USERNAME=%s\nPASSWORD=%s\n", username, password)
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create credentials file: %v", err)
	}
	return envFile
}

func TestNewSlimClient(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)

	// Test client initialization
	if client == nil {
		t.Fatal("NewSlimClient returned nil")
	}

	// Test default values
	if client.baseURL != "https://api.slimacademy.nl" {
		t.Errorf("Expected baseURL to be 'https://api.slimacademy.nl', got %q", client.baseURL)
	}

	if client.clientID != "slim_api" {
		t.Errorf("Expected clientID to be 'slim_api', got %q", client.clientID)
	}

	if client.deviceID != "4b3c1096-114f-423b-822f-22785cc3f05e" {
		t.Errorf("Expected deviceID to be set, got %q", client.deviceID)
	}

	if client.origin != "https://app.slimacademy.nl" {
		t.Errorf("Expected origin to be 'https://app.slimacademy.nl', got %q", client.origin)
	}

	if client.referer != "https://app.slimacademy.nl/" {
		t.Errorf("Expected referer to be 'https://app.slimacademy.nl/', got %q", client.referer)
	}

	// Test HTTP client timeout
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected HTTP client timeout to be 30s, got %v", client.httpClient.Timeout)
	}

	// Test components are initialized
	if client.tokenStore == nil {
		t.Error("Expected tokenStore to be initialized")
	}

	if client.credManager == nil {
		t.Error("Expected credManager to be initialized")
	}
}

func TestSlimClient_newRequest(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)
	ctx := context.Background()

	tests := []struct {
		name     string
		method   string
		endpoint string
		wantURL  string
		wantErr  bool
	}{
		{
			name:     "GET request",
			method:   "GET",
			endpoint: "/api/test",
			wantURL:  "https://api.slimacademy.nl/api/test",
			wantErr:  false,
		},
		{
			name:     "POST request",
			method:   "POST",
			endpoint: "/api/auth/login",
			wantURL:  "https://api.slimacademy.nl/api/auth/login",
			wantErr:  false,
		},
		{
			name:     "endpoint without leading slash",
			method:   "GET",
			endpoint: "api/profile",
			wantURL:  "https://api.slimacademy.nlapi/profile",
			wantErr:  false,
		},
		{
			name:     "invalid method",
			method:   "\x00",
			endpoint: "/api/test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.newRequest(ctx, tt.method, tt.endpoint, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("newRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check URL
			if req.URL.String() != tt.wantURL {
				t.Errorf("Expected URL %q, got %q", tt.wantURL, req.URL.String())
			}

			// Check method
			if req.Method != tt.method {
				t.Errorf("Expected method %q, got %q", tt.method, req.Method)
			}

			// Check required headers
			expectedHeaders := map[string]string{
				"accept":             "application/json, text/plain, */*",
				"accept-language":    "en",
				"content-type":       "application/json",
				"dnt":                "1",
				"origin":             client.origin,
				"referer":            client.referer,
				"sec-ch-ua-mobile":   "?0",
				"sec-ch-ua-platform": `"Linux"`,
				"sec-fetch-dest":     "empty",
				"sec-fetch-mode":     "cors",
				"sec-fetch-site":     "same-site",
			}

			for key, expectedValue := range expectedHeaders {
				if actualValue := req.Header.Get(key); actualValue != expectedValue {
					t.Errorf("Expected header %q to be %q, got %q", key, expectedValue, actualValue)
				}
			}

			// Check that authorization header is NOT set (for newRequest)
			if auth := req.Header.Get("authorization"); auth != "" {
				t.Errorf("Expected no authorization header in basic request, got %q", auth)
			}
		})
	}
}

func TestSlimClient_newAuthenticatedRequest(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)
	ctx := context.Background()

	t.Run("with valid token", func(t *testing.T) {
		// Create a valid token
		tokenInfo := &TokenInfo{
			Token:     "test-token-123",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour),
			Username:  "test@example.com",
		}

		if err := client.tokenStore.SaveToken(tokenInfo); err != nil {
			t.Fatalf("Failed to save test token: %v", err)
		}

		req, err := client.newAuthenticatedRequest(ctx, "GET", "/api/profile", nil)
		if err != nil {
			t.Fatalf("newAuthenticatedRequest() failed: %v", err)
		}

		// Check authorization header
		expectedAuth := "Bearer test-token-123"
		if auth := req.Header.Get("authorization"); auth != expectedAuth {
			t.Errorf("Expected authorization header %q, got %q", expectedAuth, auth)
		}

		// Check that other headers are still set
		if contentType := req.Header.Get("content-type"); contentType != "application/json" {
			t.Errorf("Expected content-type header to be preserved")
		}
	})

	t.Run("without valid token", func(t *testing.T) {
		// Clear any existing token
		client.tokenStore.ClearToken()

		_, err := client.newAuthenticatedRequest(ctx, "GET", "/api/profile", nil)
		if err == nil {
			t.Error("Expected error when no valid token available")
		}

		if !strings.Contains(err.Error(), "authentication required") {
			t.Errorf("Expected authentication error, got %v", err)
		}
	})

	t.Run("with expired token", func(t *testing.T) {
		// Create an expired token
		expiredTokenInfo := &TokenInfo{
			Token:     "expired-token",
			TokenType: "Bearer",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
			Username:  "test@example.com",
		}

		if err := client.tokenStore.SaveToken(expiredTokenInfo); err != nil {
			t.Fatalf("Failed to save expired token: %v", err)
		}

		_, err := client.newAuthenticatedRequest(ctx, "GET", "/api/profile", nil)
		if err == nil {
			t.Error("Expected error when token is expired")
		}

		if !strings.Contains(err.Error(), "authentication required") {
			t.Errorf("Expected authentication error, got %v", err)
		}
	})
}

func TestSlimClient_doRequest(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  bool
		expectedResult string
	}{
		{
			name:           "successful request",
			statusCode:     http.StatusOK,
			responseBody:   `{"success": true}`,
			expectedError:  false,
			expectedResult: `{"success": true}`,
		},
		{
			name:          "client error",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error": "unauthorized"}`,
			expectedError: true,
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": "internal server error"}`,
			expectedError: true,
		},
		{
			name:          "not found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "not found"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				writeResponse(t, w, []byte(tt.responseBody))
			}))
			defer server.Close()

			tempDir := createTempDir(t)
			client := NewSlimClient(tempDir)

			// Create request to test server
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create test request: %v", err)
			}

			// Execute request
			result, err := client.doRequest(req)

			if (err != nil) != tt.expectedError {
				t.Errorf("doRequest() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				// Check that error is of correct type
				if apiErr, ok := err.(*APIError); ok {
					if apiErr.StatusCode != tt.statusCode {
						t.Errorf("Expected status code %d, got %d", tt.statusCode, apiErr.StatusCode)
					}
					if apiErr.Response != tt.responseBody {
						t.Errorf("Expected response body %q, got %q", tt.responseBody, apiErr.Response)
					}
				} else {
					t.Errorf("Expected APIError, got %T", err)
				}
			} else {
				if string(result) != tt.expectedResult {
					t.Errorf("Expected result %q, got %q", tt.expectedResult, string(result))
				}
			}
		})
	}
}

func TestSlimClient_Login(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		serverResponse LoginResponse
		serverStatus   int
		expectedError  bool
		errorContains  string
	}{
		{
			name:     "successful login",
			username: "test@example.com",
			password: "password123",
			serverResponse: LoginResponse{
				AccessToken:  "access-token-123",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "refresh-token-123",
			},
			serverStatus:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "invalid credentials",
			username:      "test@example.com",
			password:      "wrongpassword",
			serverStatus:  http.StatusUnauthorized,
			expectedError: true,
			errorContains: "login request failed",
		},
		{
			name:          "server error",
			username:      "test@example.com",
			password:      "password123",
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
			errorContains: "login request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method and endpoint
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if r.URL.Path != "/api/auth/login" {
					t.Errorf("Expected path /api/auth/login, got %s", r.URL.Path)
				}

				// Check headers
				if contentType := r.Header.Get("content-type"); contentType != "application/json" {
					t.Errorf("Expected content-type application/json, got %s", contentType)
				}

				// Check request body
				var loginReq LoginRequest
				if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
					t.Errorf("Failed to decode login request: %v", err)
				}

				if loginReq.Username != tt.username {
					t.Errorf("Expected username %q, got %q", tt.username, loginReq.Username)
				}
				if loginReq.Password != tt.password {
					t.Errorf("Expected password %q, got %q", tt.password, loginReq.Password)
				}
				if loginReq.GrantType != "external_password" {
					t.Errorf("Expected grant_type 'external_password', got %q", loginReq.GrantType)
				}
				if loginReq.ClientID != "slim_api" {
					t.Errorf("Expected client_id 'slim_api', got %q", loginReq.ClientID)
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					encodeJSON(t, w, tt.serverResponse)
				} else {
					writeResponse(t, w, []byte(`{"error": "authentication failed"}`))
				}
			}))
			defer server.Close()

			tempDir := createTempDir(t)

			// Create credentials file
			createCredentialsFile(t, tempDir, tt.username, tt.password)

			client := NewSlimClient(tempDir)
			// Override base URL to use test server
			client.baseURL = server.URL
			// Set credential manager to use temp dir
			client.credManager = NewCredentialManager(filepath.Join(tempDir, ".env"))

			ctx := context.Background()
			err := client.Login(ctx)

			if (err != nil) != tt.expectedError {
				t.Errorf("Login() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %v", tt.errorContains, err)
				}
			} else {
				// Check that token was saved
				tokenInfo, err := client.tokenStore.LoadToken()
				if err != nil {
					t.Errorf("Expected token to be saved after successful login: %v", err)
				} else {
					if tokenInfo.Token != tt.serverResponse.AccessToken {
						t.Errorf("Expected saved token %q, got %q", tt.serverResponse.AccessToken, tokenInfo.Token)
					}
					if tokenInfo.TokenType != tt.serverResponse.TokenType {
						t.Errorf("Expected saved token type %q, got %q", tt.serverResponse.TokenType, tokenInfo.TokenType)
					}
					if tokenInfo.Username != tt.username {
						t.Errorf("Expected saved username %q, got %q", tt.username, tokenInfo.Username)
					}
				}
			}
		})
	}
}

func TestSlimClient_Login_CredentialErrors(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)
	ctx := context.Background()

	t.Run("missing credentials file", func(t *testing.T) {
		err := client.Login(ctx)
		if err == nil {
			t.Error("Expected error when credentials file is missing")
		}
		if !strings.Contains(err.Error(), "failed to load credentials") {
			t.Errorf("Expected credentials error, got %v", err)
		}
	})

	t.Run("invalid credentials file", func(t *testing.T) {
		// Create invalid credentials file
		envFile := filepath.Join(tempDir, ".env")
		if err := os.WriteFile(envFile, []byte("INVALID=content\n"), 0600); err != nil {
			t.Fatalf("Failed to create invalid credentials file: %v", err)
		}

		client.credManager = NewCredentialManager(envFile)

		err := client.Login(ctx)
		if err == nil {
			t.Error("Expected error with invalid credentials")
		}
		if !strings.Contains(err.Error(), "failed to load credentials") {
			t.Errorf("Expected credentials error, got %v", err)
		}
	})
}

func TestSlimClient_IsLoggedIn(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)

	t.Run("no token", func(t *testing.T) {
		if client.IsLoggedIn() {
			t.Error("Expected IsLoggedIn to return false when no token exists")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		tokenInfo := &TokenInfo{
			Token:     "valid-token",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour),
			Username:  "test@example.com",
		}

		if err := client.tokenStore.SaveToken(tokenInfo); err != nil {
			t.Fatalf("Failed to save test token: %v", err)
		}

		if !client.IsLoggedIn() {
			t.Error("Expected IsLoggedIn to return true with valid token")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiredTokenInfo := &TokenInfo{
			Token:     "expired-token",
			TokenType: "Bearer",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(-time.Hour),
			Username:  "test@example.com",
		}

		if err := client.tokenStore.SaveToken(expiredTokenInfo); err != nil {
			t.Fatalf("Failed to save expired token: %v", err)
		}

		if client.IsLoggedIn() {
			t.Error("Expected IsLoggedIn to return false with expired token")
		}
	})
}

func TestSlimClient_GetTokenInfo(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)

	t.Run("no token", func(t *testing.T) {
		_, err := client.GetTokenInfo()
		if err == nil {
			t.Error("Expected error when no token exists")
		}
	})

	t.Run("with token", func(t *testing.T) {
		originalToken := &TokenInfo{
			Token:        "test-token",
			TokenType:    "Bearer",
			CreatedAt:    time.Now(),
			ExpiresIn:    3600,
			ExpiresAt:    time.Now().Add(time.Hour),
			Username:     "test@example.com",
			RefreshToken: "refresh-token",
		}

		if err := client.tokenStore.SaveToken(originalToken); err != nil {
			t.Fatalf("Failed to save test token: %v", err)
		}

		tokenInfo, err := client.GetTokenInfo()
		if err != nil {
			t.Fatalf("GetTokenInfo() failed: %v", err)
		}

		if tokenInfo.Token != originalToken.Token {
			t.Errorf("Expected token %q, got %q", originalToken.Token, tokenInfo.Token)
		}
		if tokenInfo.Username != originalToken.Username {
			t.Errorf("Expected username %q, got %q", originalToken.Username, tokenInfo.Username)
		}
	})
}

func TestSlimClient_Logout(t *testing.T) {
	tempDir := createTempDir(t)
	client := NewSlimClient(tempDir)

	// Create a token first
	tokenInfo := &TokenInfo{
		Token:     "test-token",
		TokenType: "Bearer",
		CreatedAt: time.Now(),
		ExpiresIn: 3600,
		ExpiresAt: time.Now().Add(time.Hour),
		Username:  "test@example.com",
	}

	if err := client.tokenStore.SaveToken(tokenInfo); err != nil {
		t.Fatalf("Failed to save test token: %v", err)
	}

	// Verify token exists
	if !client.IsLoggedIn() {
		t.Error("Expected to be logged in before logout")
	}

	// Logout
	if err := client.Logout(); err != nil {
		t.Errorf("Logout() failed: %v", err)
	}

	// Verify token is cleared
	if client.IsLoggedIn() {
		t.Error("Expected to be logged out after Logout()")
	}

	// Logout again should not error
	if err := client.Logout(); err != nil {
		t.Errorf("Second Logout() should not error: %v", err)
	}
}

func TestSlimClient_EnsureAuthenticated(t *testing.T) {
	ctx := context.Background()

	t.Run("already authenticated", func(t *testing.T) {
		tempDir := createTempDir(t)
		client := NewSlimClient(tempDir)

		// Create valid token
		tokenInfo := &TokenInfo{
			Token:     "valid-token",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour),
			Username:  "test@example.com",
		}

		if err := client.tokenStore.SaveToken(tokenInfo); err != nil {
			t.Fatalf("Failed to save test token: %v", err)
		}

		err := client.EnsureAuthenticated(ctx)
		if err != nil {
			t.Errorf("EnsureAuthenticated() failed with valid token: %v", err)
		}
	})

	t.Run("needs authentication", func(t *testing.T) {
		tempDir := createTempDir(t)

		// Create test server for login
		loginCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/login" {
				loginCalled = true
				resp := LoginResponse{
					AccessToken: "new-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				}
				encodeJSON(t, w, resp)
			}
		}))
		defer server.Close()

		client := NewSlimClient(tempDir)
		client.baseURL = server.URL

		// Create credentials file
		createCredentialsFile(t, tempDir, "test@example.com", "password123")
		client.credManager = NewCredentialManager(filepath.Join(tempDir, ".env"))

		err := client.EnsureAuthenticated(ctx)
		if err != nil {
			t.Errorf("EnsureAuthenticated() failed: %v", err)
		}

		if !loginCalled {
			t.Error("Expected login to be called when not authenticated")
		}

		// Verify token was saved
		if !client.IsLoggedIn() {
			t.Error("Expected to be logged in after EnsureAuthenticated")
		}
	})

	t.Run("authentication fails", func(t *testing.T) {
		tempDir := createTempDir(t)
		client := NewSlimClient(tempDir)

		// Don't create credentials file to force failure
		err := client.EnsureAuthenticated(ctx)
		if err == nil {
			t.Error("Expected EnsureAuthenticated to fail without credentials")
		}
	})
}
