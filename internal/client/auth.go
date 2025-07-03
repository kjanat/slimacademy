package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TokenStore handles persistence of authentication tokens
type TokenStore struct {
	tokenFile string
}

// NewTokenStore creates a new token store
func NewTokenStore(outputDir string) *TokenStore {
	return &TokenStore{
		tokenFile: filepath.Join(outputDir, ".token_info"),
	}
}

// SaveToken persists the token information to disk
func (ts *TokenStore) SaveToken(info *TokenInfo) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ts.tokenFile), 0755); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	if err := os.WriteFile(ts.tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadToken loads the token information from disk
func (ts *TokenStore) LoadToken() (*TokenInfo, error) {
	data, err := os.ReadFile(ts.tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no token found, please login first")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var info TokenInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &info, nil
}

// IsTokenValid checks if the stored token is still valid
func (ts *TokenStore) IsTokenValid() bool {
	info, err := ts.LoadToken()
	if err != nil {
		return false
	}

	// Check if token has expired (with 5 minute buffer)
	return time.Now().Add(5 * time.Minute).Before(info.ExpiresAt)
}

// ClearToken removes the stored token
func (ts *TokenStore) ClearToken() error {
	if err := os.Remove(ts.tokenFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}

// GetValidToken returns a valid token or an error if none exists
func (ts *TokenStore) GetValidToken() (string, error) {
	if !ts.IsTokenValid() {
		return "", fmt.Errorf("no valid token found, please login first")
	}

	info, err := ts.LoadToken()
	if err != nil {
		return "", err
	}

	return info.Token, nil
}

// CredentialManager handles loading credentials from .env file
type CredentialManager struct {
	envFile string
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(envFile string) *CredentialManager {
	if envFile == "" {
		envFile = ".env"
	}
	return &CredentialManager{envFile: envFile}
}

// LoadCredentials loads credentials from .env file with environment variable fallback
func (cm *CredentialManager) LoadCredentials() (*Credentials, error) {
	// Get environment variables as fallback
	envUsername := os.Getenv("USERNAME")
	envPassword := os.Getenv("PASSWORD")

	// Try to read .env file first (it takes precedence)
	file, err := os.Open(cm.envFile)
	if err != nil {
		// If no .env file, use environment variables if available
		if envUsername != "" && envPassword != "" {
			return &Credentials{
				Username: envUsername,
				Password: envPassword,
			}, nil
		}

		// If no .env file and no complete env vars, provide helpful error
		if envUsername == "" && envPassword == "" {
			return nil, fmt.Errorf("credentials not found. Either:\n1. Set environment variables: USERNAME=your@email.com PASSWORD=yourpassword\n2. Create a .env file with:\n   USERNAME=your@email.com\n   PASSWORD=yourpassword")
		}
		// Partial credentials from env vars
		if envUsername == "" {
			return nil, fmt.Errorf("USERNAME not found in environment variables or .env file")
		}
		return nil, fmt.Errorf("PASSWORD not found in environment variables or .env file")
	}
	defer file.Close()

	// Start with environment variables as defaults, .env file will override
	username := envUsername
	password := envPassword

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// .env file values override environment variables
		switch key {
		case "USERNAME":
			username = value
		case "PASSWORD":
			password = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}

	if username == "" || password == "" {
		missing := []string{}
		if username == "" {
			missing = append(missing, "USERNAME")
		}
		if password == "" {
			missing = append(missing, "PASSWORD")
		}
		return nil, fmt.Errorf("%s not found in environment variables or .env file", strings.Join(missing, " and "))
	}

	return &Credentials{
		Username: username,
		Password: password,
	}, nil
}

// ValidateCredentials checks if credentials are properly loaded
func (cm *CredentialManager) ValidateCredentials() error {
	creds, err := cm.LoadCredentials()
	if err != nil {
		return err
	}

	if creds.Username == "" || creds.Password == "" {
		return fmt.Errorf("invalid credentials: username or password is empty")
	}

	return nil
}
