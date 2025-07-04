package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewTokenStore(t *testing.T) {
	outputDir := "/tmp/test-output"
	ts := NewTokenStore(outputDir)

	if ts == nil {
		t.Fatal("NewTokenStore returned nil")
	}

	expectedPath := filepath.Join(outputDir, ".token_info")
	if ts.tokenFile != expectedPath {
		t.Errorf("Expected token file path %q, got %q", expectedPath, ts.tokenFile)
	}
}

func TestTokenStore_SaveAndLoadToken(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	now := time.Now()
	expiresAt := now.Add(time.Hour)

	originalToken := &TokenInfo{
		Token:        "test-access-token",
		TokenType:    "Bearer",
		CreatedAt:    now,
		ExpiresIn:    3600,
		ExpiresAt:    expiresAt,
		Username:     "test@example.com",
		RefreshToken: "test-refresh-token",
	}

	// Test SaveToken
	err := ts.SaveToken(originalToken)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Verify token file was created
	if _, err := os.Stat(ts.tokenFile); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}

	// Test LoadToken
	loadedToken, err := ts.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() failed: %v", err)
	}

	// Compare all fields
	if loadedToken.Token != originalToken.Token {
		t.Errorf("Expected token %q, got %q", originalToken.Token, loadedToken.Token)
	}
	if loadedToken.TokenType != originalToken.TokenType {
		t.Errorf("Expected token type %q, got %q", originalToken.TokenType, loadedToken.TokenType)
	}
	if loadedToken.ExpiresIn != originalToken.ExpiresIn {
		t.Errorf("Expected expires in %d, got %d", originalToken.ExpiresIn, loadedToken.ExpiresIn)
	}
	if loadedToken.Username != originalToken.Username {
		t.Errorf("Expected username %q, got %q", originalToken.Username, loadedToken.Username)
	}
	if loadedToken.RefreshToken != originalToken.RefreshToken {
		t.Errorf("Expected refresh token %q, got %q", originalToken.RefreshToken, loadedToken.RefreshToken)
	}

	// Check time fields (with some tolerance for serialization)
	if timeDiff := loadedToken.CreatedAt.Sub(originalToken.CreatedAt); timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("CreatedAt time mismatch: expected %v, got %v", originalToken.CreatedAt, loadedToken.CreatedAt)
	}
	if timeDiff := loadedToken.ExpiresAt.Sub(originalToken.ExpiresAt); timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("ExpiresAt time mismatch: expected %v, got %v", originalToken.ExpiresAt, loadedToken.ExpiresAt)
	}
}

func TestTokenStore_LoadToken_NoFile(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	_, err := ts.LoadToken()
	if err == nil {
		t.Error("Expected error when token file doesn't exist")
	}

	if !strings.Contains(err.Error(), "no token found") {
		t.Errorf("Expected 'no token found' error, got %v", err)
	}
}

func TestTokenStore_LoadToken_CorruptFile(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	// Create corrupt token file
	corruptData := []byte("invalid json data")
	if err := os.WriteFile(ts.tokenFile, corruptData, 0600); err != nil {
		t.Fatalf("Failed to create corrupt token file: %v", err)
	}

	_, err := ts.LoadToken()
	if err == nil {
		t.Error("Expected error when token file is corrupt")
	}

	if !strings.Contains(err.Error(), "failed to parse token file") {
		t.Errorf("Expected parse error, got %v", err)
	}
}

func TestTokenStore_SaveToken_DirectoryCreation(t *testing.T) {
	tempDir := createTempDir(t)

	// Use nested directory that doesn't exist
	nestedDir := filepath.Join(tempDir, "nested", "path")
	ts := NewTokenStore(nestedDir)

	token := &TokenInfo{
		Token:     "test-token",
		TokenType: "Bearer",
		CreatedAt: time.Now(),
		ExpiresIn: 3600,
		ExpiresAt: time.Now().Add(time.Hour),
		Username:  "test@example.com",
	}

	err := ts.SaveToken(token)
	if err != nil {
		t.Fatalf("SaveToken() failed with nested directory: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(ts.tokenFile)); os.IsNotExist(err) {
		t.Error("Expected nested directory to be created")
	}

	// Verify token file exists
	if _, err := os.Stat(ts.tokenFile); os.IsNotExist(err) {
		t.Error("Token file was not created in nested directory")
	}
}

func TestTokenStore_IsTokenValid(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	t.Run("no token file", func(t *testing.T) {
		if ts.IsTokenValid() {
			t.Error("Expected IsTokenValid to return false when no token file exists")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		validToken := &TokenInfo{
			Token:     "valid-token",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour), // Expires in 1 hour
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(validToken); err != nil {
			t.Fatalf("Failed to save valid token: %v", err)
		}

		if !ts.IsTokenValid() {
			t.Error("Expected IsTokenValid to return true for valid token")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken := &TokenInfo{
			Token:     "expired-token",
			TokenType: "Bearer",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(expiredToken); err != nil {
			t.Fatalf("Failed to save expired token: %v", err)
		}

		if ts.IsTokenValid() {
			t.Error("Expected IsTokenValid to return false for expired token")
		}
	})

	t.Run("token expiring soon", func(t *testing.T) {
		// Token expires in 3 minutes (less than 5 minute buffer)
		soonToExpireToken := &TokenInfo{
			Token:     "soon-to-expire-token",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 180,
			ExpiresAt: time.Now().Add(3 * time.Minute),
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(soonToExpireToken); err != nil {
			t.Fatalf("Failed to save soon-to-expire token: %v", err)
		}

		if ts.IsTokenValid() {
			t.Error("Expected IsTokenValid to return false for token expiring within 5 minutes")
		}
	})

	t.Run("token with enough buffer", func(t *testing.T) {
		// Token expires in 10 minutes (more than 5 minute buffer)
		tokenWithBuffer := &TokenInfo{
			Token:     "token-with-buffer",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 600,
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(tokenWithBuffer); err != nil {
			t.Fatalf("Failed to save token with buffer: %v", err)
		}

		if !ts.IsTokenValid() {
			t.Error("Expected IsTokenValid to return true for token with sufficient buffer")
		}
	})
}

func TestTokenStore_ClearToken(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	t.Run("clear existing token", func(t *testing.T) {
		// Create a token first
		token := &TokenInfo{
			Token:     "test-token",
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour),
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(token); err != nil {
			t.Fatalf("Failed to save token: %v", err)
		}

		// Verify token exists
		if _, err := os.Stat(ts.tokenFile); os.IsNotExist(err) {
			t.Error("Token file should exist before clearing")
		}

		// Clear token
		if err := ts.ClearToken(); err != nil {
			t.Errorf("ClearToken() failed: %v", err)
		}

		// Verify token file is gone
		if _, err := os.Stat(ts.tokenFile); !os.IsNotExist(err) {
			t.Error("Token file should not exist after clearing")
		}
	})

	t.Run("clear non-existent token", func(t *testing.T) {
		// Clear token when no file exists (should not error)
		if err := ts.ClearToken(); err != nil {
			t.Errorf("ClearToken() should not error when file doesn't exist: %v", err)
		}
	})
}

func TestTokenStore_GetValidToken(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	t.Run("no token", func(t *testing.T) {
		_, err := ts.GetValidToken()
		if err == nil {
			t.Error("Expected error when no token exists")
		}

		if !strings.Contains(err.Error(), "no valid token found") {
			t.Errorf("Expected 'no valid token found' error, got %v", err)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		expectedToken := "valid-access-token"
		tokenInfo := &TokenInfo{
			Token:     expectedToken,
			TokenType: "Bearer",
			CreatedAt: time.Now(),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(time.Hour),
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(tokenInfo); err != nil {
			t.Fatalf("Failed to save token: %v", err)
		}

		token, err := ts.GetValidToken()
		if err != nil {
			t.Errorf("GetValidToken() failed: %v", err)
		}

		if token != expectedToken {
			t.Errorf("Expected token %q, got %q", expectedToken, token)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiredToken := &TokenInfo{
			Token:     "expired-token",
			TokenType: "Bearer",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresIn: 3600,
			ExpiresAt: time.Now().Add(-time.Hour),
			Username:  "test@example.com",
		}

		if err := ts.SaveToken(expiredToken); err != nil {
			t.Fatalf("Failed to save expired token: %v", err)
		}

		_, err := ts.GetValidToken()
		if err == nil {
			t.Error("Expected error for expired token")
		}

		if !strings.Contains(err.Error(), "no valid token found") {
			t.Errorf("Expected 'no valid token found' error, got %v", err)
		}
	})
}

func TestNewCredentialManager(t *testing.T) {
	t.Run("with env file", func(t *testing.T) {
		envFile := "/path/to/.env"
		cm := NewCredentialManager(envFile)

		if cm == nil {
			t.Fatal("NewCredentialManager returned nil")
		}

		if cm.envFile != envFile {
			t.Errorf("Expected env file %q, got %q", envFile, cm.envFile)
		}
	})

	t.Run("with empty env file", func(t *testing.T) {
		cm := NewCredentialManager("")

		if cm.envFile != ".env" {
			t.Errorf("Expected default env file '.env', got %q", cm.envFile)
		}
	})
}

func TestCredentialManager_LoadCredentials(t *testing.T) {
	tempDir := createTempDir(t)

	tests := []struct {
		name          string
		fileContent   string
		expectedUser  string
		expectedPass  string
		expectedError bool
		errorContains string
	}{
		{
			name:         "valid credentials",
			fileContent:  "USERNAME=test@example.com\nPASSWORD=mypassword123\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:         "credentials with quotes",
			fileContent:  `USERNAME="test@example.com"` + "\n" + `PASSWORD='mypassword123'` + "\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:         "credentials with spaces",
			fileContent:  "USERNAME = test@example.com \nPASSWORD = mypassword123 \n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:         "credentials with comments",
			fileContent:  "# This is a comment\nUSERNAME=test@example.com\n# Another comment\nPASSWORD=mypassword123\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:         "credentials with empty lines",
			fileContent:  "\nUSERNAME=test@example.com\n\nPASSWORD=mypassword123\n\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:         "credentials with other variables",
			fileContent:  "OTHER_VAR=something\nUSERNAME=test@example.com\nPASSWORD=mypassword123\nANOTHER_VAR=else\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
		{
			name:          "missing username",
			fileContent:   "PASSWORD=mypassword123\n",
			expectedError: true,
			errorContains: "not found in environment variables or .env file",
		},
		{
			name:          "missing password",
			fileContent:   "USERNAME=test@example.com\n",
			expectedError: true,
			errorContains: "not found in environment variables or .env file",
		},
		{
			name:          "empty username",
			fileContent:   "USERNAME=\nPASSWORD=mypassword123\n",
			expectedError: true,
			errorContains: "not found in environment variables or .env file",
		},
		{
			name:          "empty password",
			fileContent:   "USERNAME=test@example.com\nPASSWORD=\n",
			expectedError: true,
			errorContains: "not found in environment variables or .env file",
		},
		{
			name:         "malformed line (ignored)",
			fileContent:  "USERNAME=test@example.com\nMALFORMED LINE\nPASSWORD=mypassword123\n",
			expectedUser: "test@example.com",
			expectedPass: "mypassword123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := filepath.Join(tempDir, tt.name+".env")
			if err := os.WriteFile(envFile, []byte(tt.fileContent), 0600); err != nil {
				t.Fatalf("Failed to create test env file: %v", err)
			}

			cm := NewCredentialManager(envFile)
			creds, err := cm.LoadCredentials()

			if (err != nil) != tt.expectedError {
				t.Errorf("LoadCredentials() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %v", tt.errorContains, err)
				}
				return
			}

			if creds == nil {
				t.Error("Expected credentials to be non-nil")
				return
			}

			if creds.Username != tt.expectedUser {
				t.Errorf("Expected username %q, got %q", tt.expectedUser, creds.Username)
			}

			if creds.Password != tt.expectedPass {
				t.Errorf("Expected password %q, got %q", tt.expectedPass, creds.Password)
			}
		})
	}
}

func TestCredentialManager_LoadCredentials_FileNotFound(t *testing.T) {
	cm := NewCredentialManager("/path/that/does/not/exist/.env")

	_, err := cm.LoadCredentials()
	if err == nil {
		t.Error("Expected error when .env file does not exist")
	}

	if !strings.Contains(err.Error(), "credentials not found") {
		t.Errorf("Expected credentials not found error, got %v", err)
	}

	// Check that error message contains helpful instructions
	if !strings.Contains(err.Error(), "Create a .env file with:") {
		t.Error("Expected error message to contain setup instructions")
	}
	if !strings.Contains(err.Error(), "USERNAME=") {
		t.Error("Expected error message to show USERNAME format")
	}
	if !strings.Contains(err.Error(), "PASSWORD=") {
		t.Error("Expected error message to show PASSWORD format")
	}
}

func TestCredentialManager_LoadCredentials_ReadError(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a directory with same name as the expected file to cause read error
	envPath := filepath.Join(tempDir, ".env")
	if err := os.Mkdir(envPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	cm := NewCredentialManager(envPath)

	_, err := cm.LoadCredentials()
	if err == nil {
		t.Error("Expected error when trying to read directory as file")
	}

	// Should get an error about reading the file
	if !strings.Contains(err.Error(), "error reading .env file") {
		t.Errorf("Expected file read error, got %v", err)
	}
}

func TestCredentialManager_ValidateCredentials(t *testing.T) {
	tempDir := createTempDir(t)

	t.Run("valid credentials", func(t *testing.T) {
		envFile := filepath.Join(tempDir, "valid.env")
		content := "USERNAME=test@example.com\nPASSWORD=mypassword123\n"
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		if err := cm.ValidateCredentials(); err != nil {
			t.Errorf("ValidateCredentials() failed: %v", err)
		}
	})

	t.Run("missing credentials file", func(t *testing.T) {
		cm := NewCredentialManager("/path/that/does/not/exist/.env")

		err := cm.ValidateCredentials()
		if err == nil {
			t.Error("Expected error when credentials file is missing")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		envFile := filepath.Join(tempDir, "invalid.env")
		content := "USERNAME=\nPASSWORD=mypassword123\n"
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		cm := NewCredentialManager(envFile)

		err := cm.ValidateCredentials()
		if err == nil {
			t.Error("Expected error with empty username")
		}

		if !strings.Contains(err.Error(), "not found in environment variables or .env file") {
			t.Errorf("Expected validation error, got %v", err)
		}
	})
}

func TestTokenStore_FilePermissions(t *testing.T) {
	tempDir := createTempDir(t)
	ts := NewTokenStore(tempDir)

	token := &TokenInfo{
		Token:     "sensitive-token",
		TokenType: "Bearer",
		CreatedAt: time.Now(),
		ExpiresIn: 3600,
		ExpiresAt: time.Now().Add(time.Hour),
		Username:  "test@example.com",
	}

	if err := ts.SaveToken(token); err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Check file permissions (should be 0600 - readable/writable by owner only)
	info, err := os.Stat(ts.tokenFile)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	perm := info.Mode().Perm()
	expected := os.FileMode(0600)
	if perm != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, perm)
	}
}

func TestCredentialManager_LoadCredentials_EdgeCases(t *testing.T) {
	tempDir := createTempDir(t)

	t.Run("equals sign in password", func(t *testing.T) {
		envFile := filepath.Join(tempDir, "equals.env")
		content := "USERNAME=test@example.com\nPASSWORD=pass=with=equals\n"
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials() failed: %v", err)
		}

		if creds.Password != "pass=with=equals" {
			t.Errorf("Expected password 'pass=with=equals', got %q", creds.Password)
		}
	})

	t.Run("mixed quote types", func(t *testing.T) {
		envFile := filepath.Join(tempDir, "mixed_quotes.env")
		content := `USERNAME="test@example.com'` + "\n" + `PASSWORD='mypassword"123'` + "\n"
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials() failed: %v", err)
		}

		if creds.Username != "test@example.com" {
			t.Errorf("Expected username 'test@example.com', got %q", creds.Username)
		}
		if creds.Password != `mypassword"123` {
			t.Errorf("Expected password 'mypassword\"123', got %q", creds.Password)
		}
	})

	t.Run("unicode in credentials", func(t *testing.T) {
		envFile := filepath.Join(tempDir, "unicode.env")
		content := "USERNAME=test@例え.com\nPASSWORD=pássw0rd\n"
		if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials() failed: %v", err)
		}

		if creds.Username != "test@例え.com" {
			t.Errorf("Expected unicode username, got %q", creds.Username)
		}
		if creds.Password != "pássw0rd" {
			t.Errorf("Expected unicode password, got %q", creds.Password)
		}
	})
}
