// Package client provides HTTP client functionality for accessing the SlimAcademy API.
// It includes authentication, request/response handling, and API endpoint management
// for retrieving book data and metadata.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SlimClient represents the Slim Academy API client
type SlimClient struct {
	httpClient  *http.Client
	baseURL     string
	tokenStore  *TokenStore
	credManager *CredentialManager

	// API Configuration (matching bash script)
	clientID     string
	clientSecret string
	deviceID     string
	origin       string
	referer      string
}

// NewSlimClient creates a new Slim Academy API client
func NewSlimClient(outputDir string) *SlimClient {
	return &SlimClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://api.slimacademy.nl",
		tokenStore:  NewTokenStore(outputDir),
		credManager: NewCredentialManager(".env"),

		// Match bash script configuration
		clientID:     "slim_api",
		clientSecret: "",
		deviceID:     "4b3c1096-114f-423b-822f-22785cc3f05e",
		origin:       "https://app.slimacademy.nl",
		referer:      "https://app.slimacademy.nl/",
	}
}

// newRequest creates a new HTTP request with proper headers (matching bash script)
func (c *SlimClient) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Set headers matching the bash script exactly
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", c.origin)
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", c.referer)
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Linux"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")

	return req, nil
}

// newAuthenticatedRequest creates a request with authentication header
func (c *SlimClient) newAuthenticatedRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := c.newRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	// Get valid token
	token, err := c.tokenStore.GetValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	req.Header.Set("authorization", "Bearer "+token)
	return req, nil
}

// doRequest executes an HTTP request and handles the response
func (c *SlimClient) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API request failed with status %d", resp.StatusCode),
			Response:   string(body),
		}
	}

	return body, nil
}

// Login authenticates with the Slim Academy API
func (c *SlimClient) Login(ctx context.Context) error {
	// Load credentials
	creds, err := c.credManager.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Build login request (matching bash script)
	loginReq := LoginRequest{
		GrantType:    "external_password",
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		DeviceID:     c.deviceID,
		Username:     creds.Username,
		Password:     creds.Password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create request
	req, err := c.newRequest(ctx, "POST", "/api/auth/login", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	// Execute request
	respBody, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	// Parse response
	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)

	// Save token info
	tokenInfo := &TokenInfo{
		Token:        loginResp.AccessToken,
		TokenType:    loginResp.TokenType,
		CreatedAt:    time.Now(),
		ExpiresIn:    loginResp.ExpiresIn,
		ExpiresAt:    expiresAt,
		Username:     creds.Username,
		RefreshToken: loginResp.RefreshToken,
	}

	if err := c.tokenStore.SaveToken(tokenInfo); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// IsLoggedIn checks if the client has a valid authentication token
func (c *SlimClient) IsLoggedIn() bool {
	return c.tokenStore.IsTokenValid()
}

// GetTokenInfo returns information about the current token
func (c *SlimClient) GetTokenInfo() (*TokenInfo, error) {
	return c.tokenStore.LoadToken()
}

// Logout clears the stored authentication token
func (c *SlimClient) Logout() error {
	return c.tokenStore.ClearToken()
}

// EnsureAuthenticated ensures the client is authenticated, logging in if necessary
func (c *SlimClient) EnsureAuthenticated(ctx context.Context) error {
	if c.IsLoggedIn() {
		return nil
	}

	fmt.Println("Authentication required, logging in...")
	return c.Login(ctx)
}
