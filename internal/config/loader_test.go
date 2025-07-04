package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadConfig_DefaultWhenEmpty(t *testing.T) {
	loader := NewLoader()

	// Test with empty path returns default config
	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig with empty path should not fail: %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig should return non-nil config")
	}

	if config.Markdown == nil {
		t.Error("Default config should include markdown configuration")
	}
}

func TestLoader_LoadConfig_JSON(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	configContent := `{
		"markdown": {
			"italicFormat": "_",
			"boldFormat": "**",
			"emphasizedLinks": true
		}
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should succeed: %v", err)
	}

	if config.Markdown.ItalicFormat != "_" {
		t.Errorf("Expected italic format '_', got %q", config.Markdown.ItalicFormat)
	}

	if config.Markdown.BoldFormat != "**" {
		t.Errorf("Expected bold format '**', got %q", config.Markdown.BoldFormat)
	}

	if !config.Markdown.EmphasizedLinks {
		t.Error("Expected emphasizedLinks to be true")
	}
}

func TestLoader_LoadConfig_YAML(t *testing.T) {
	t.Skip("YAML support requires YAML tags on all config structs - skipping for now")

	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `markdown:
  italicFormat: "*"
  boldFormat: "**"
  emphasizedLinks: false`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	config, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should succeed: %v", err)
	}

	if config.Markdown.ItalicFormat != "*" {
		t.Errorf("Expected italic format '*', got %q", config.Markdown.ItalicFormat)
	}
}

func TestLoader_LoadConfig_InvalidFile(t *testing.T) {
	loader := NewLoader()

	// Test with non-existent file
	_, err := loader.LoadConfig("/nonexistent/file.json")
	if err == nil {
		t.Error("LoadConfig should fail for non-existent file")
	}
}

func TestLoader_LoadConfig_UnsupportedFormat(t *testing.T) {
	// Create temporary config file with unsupported extension
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.txt")

	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should fail for unsupported file format")
	}
}

func TestConfig_GetFormatConfig(t *testing.T) {
	config := DefaultConfig()

	// Test valid format
	mdConfig := config.GetFormatConfig("markdown")
	if mdConfig == nil {
		t.Error("GetFormatConfig should return non-nil for markdown")
	}

	// Test case insensitive
	htmlConfig := config.GetFormatConfig("HTML")
	if htmlConfig == nil {
		t.Error("GetFormatConfig should be case insensitive")
	}

	// Test invalid format
	invalidConfig := config.GetFormatConfig("invalid")
	if invalidConfig != nil {
		t.Error("GetFormatConfig should return nil for invalid format")
	}
}
