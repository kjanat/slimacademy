package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialManager_EnvironmentVariableFallback(t *testing.T) {
	// Save original environment variables
	originalUsername := os.Getenv("USERNAME")
	originalPassword := os.Getenv("PASSWORD")
	defer func() {
		if originalUsername != "" {
			os.Setenv("USERNAME", originalUsername)
		} else {
			os.Unsetenv("USERNAME")
		}
		if originalPassword != "" {
			os.Setenv("PASSWORD", originalPassword)
		} else {
			os.Unsetenv("PASSWORD")
		}
	}()

	t.Run("environment variables only", func(t *testing.T) {
		// Clear env vars first
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		// Set test credentials in environment
		os.Setenv("USERNAME", "env@example.com")
		os.Setenv("PASSWORD", "envpassword")

		// Use non-existent .env file
		cm := NewCredentialManager("/nonexistent/.env")
		creds, err := cm.LoadCredentials()

		if err != nil {
			t.Errorf("Expected no error with env vars, got: %v", err)
		}
		if creds == nil {
			t.Fatal("Expected credentials to be loaded from environment")
		}
		if creds.Username != "env@example.com" {
			t.Errorf("Expected username 'env@example.com', got '%s'", creds.Username)
		}
		if creds.Password != "envpassword" {
			t.Errorf("Expected password 'envpassword', got '%s'", creds.Password)
		}
	})

	t.Run("partial environment variables", func(t *testing.T) {
		// Clear env vars first
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		// Set only username in environment
		os.Setenv("USERNAME", "partial@example.com")

		cm := NewCredentialManager("/nonexistent/.env")
		_, err := cm.LoadCredentials()

		if err == nil {
			t.Error("Expected error with partial credentials")
		}
		if err.Error() != "PASSWORD not found in environment variables or .env file" {
			t.Errorf("Unexpected error message: %v", err.Error())
		}
	})

	t.Run("no credentials anywhere", func(t *testing.T) {
		// Clear env vars
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		cm := NewCredentialManager("/nonexistent/.env")
		_, err := cm.LoadCredentials()

		if err == nil {
			t.Error("Expected error with no credentials")
		}
		expectedError := "credentials not found. Either:\n1. Set environment variables: USERNAME=your@email.com PASSWORD=yourpassword\n2. Create a .env file with:\n   USERNAME=your@email.com\n   PASSWORD=yourpassword"
		if err.Error() != expectedError {
			t.Errorf("Expected helpful error message, got: %v", err.Error())
		}
	})

	t.Run("env file overrides environment variables", func(t *testing.T) {
		// Clear environment variables first
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		// Set environment variables
		os.Setenv("USERNAME", "env@example.com")
		os.Setenv("PASSWORD", "envpassword")

		// Create temporary .env file
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envContent := `USERNAME=file@example.com
PASSWORD=filepassword`

		if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if creds == nil {
			t.Fatal("Expected credentials to be loaded")
		}

		// .env file should override environment variables
		if creds.Username != "file@example.com" {
			t.Errorf("Expected username from file 'file@example.com', got '%s'", creds.Username)
		}
		if creds.Password != "filepassword" {
			t.Errorf("Expected password from file 'filepassword', got '%s'", creds.Password)
		}
	})

	t.Run("partial env file with env var fallback", func(t *testing.T) {
		// Clear environment variables first
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		// Set environment variables
		os.Setenv("USERNAME", "env@example.com")
		os.Setenv("PASSWORD", "envpassword")

		// Create temporary .env file with only PASSWORD
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envContent := `PASSWORD=filepassword`

		if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if creds == nil {
			t.Fatal("Expected credentials to be loaded")
		}

		// Should use USERNAME from env and PASSWORD from file
		if creds.Username != "env@example.com" {
			t.Errorf("Expected username from env 'env@example.com', got '%s'", creds.Username)
		}
		if creds.Password != "filepassword" {
			t.Errorf("Expected password from file 'filepassword', got '%s'", creds.Password)
		}
	})

	t.Run("quoted values in env file", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		// Create temporary .env file with quoted values
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envContent := `USERNAME="quoted@example.com"
PASSWORD='singlequoted'`

		if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		cm := NewCredentialManager(envFile)
		creds, err := cm.LoadCredentials()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if creds == nil {
			t.Fatal("Expected credentials to be loaded")
		}

		// Quotes should be stripped
		if creds.Username != "quoted@example.com" {
			t.Errorf("Expected unquoted username 'quoted@example.com', got '%s'", creds.Username)
		}
		if creds.Password != "singlequoted" {
			t.Errorf("Expected unquoted password 'singlequoted', got '%s'", creds.Password)
		}
	})
}

func TestCredentialManager_ErrorMessages(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("USERNAME")
	os.Unsetenv("PASSWORD")

	tests := []struct {
		name        string
		envContent  string
		expectedErr string
	}{
		{
			name:        "missing USERNAME in file",
			envContent:  "PASSWORD=test",
			expectedErr: "USERNAME not found in environment variables or .env file",
		},
		{
			name:        "missing PASSWORD in file",
			envContent:  "USERNAME=test@example.com",
			expectedErr: "PASSWORD not found in environment variables or .env file",
		},
		{
			name:        "empty file",
			envContent:  "",
			expectedErr: "USERNAME and PASSWORD not found in environment variables or .env file",
		},
		{
			name:        "comments only",
			envContent:  "# This is a comment\n# Another comment",
			expectedErr: "USERNAME and PASSWORD not found in environment variables or .env file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary .env file
			tempDir := t.TempDir()
			envFile := filepath.Join(tempDir, ".env")

			if err := os.WriteFile(envFile, []byte(tt.envContent), 0600); err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			cm := NewCredentialManager(envFile)
			_, err := cm.LoadCredentials()

			if err == nil {
				t.Error("Expected error but got none")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestCredentialManager_ValidateCredentialsEnhanced(t *testing.T) {
	t.Run("valid credentials from environment", func(t *testing.T) {
		// Save original environment variables
		originalUsername := os.Getenv("USERNAME")
		originalPassword := os.Getenv("PASSWORD")
		defer func() {
			if originalUsername != "" {
				os.Setenv("USERNAME", originalUsername)
			} else {
				os.Unsetenv("USERNAME")
			}
			if originalPassword != "" {
				os.Setenv("PASSWORD", originalPassword)
			} else {
				os.Unsetenv("PASSWORD")
			}
		}()

		os.Setenv("USERNAME", "valid@example.com")
		os.Setenv("PASSWORD", "validpassword")

		cm := NewCredentialManager("/nonexistent/.env")
		err := cm.ValidateCredentials()

		if err != nil {
			t.Errorf("Expected validation to pass, got error: %v", err)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("USERNAME")
		os.Unsetenv("PASSWORD")

		cm := NewCredentialManager("/nonexistent/.env")
		err := cm.ValidateCredentials()

		if err == nil {
			t.Error("Expected validation to fail")
		}
	})
}
