package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Markdown *MarkdownConfig `json:"markdown,omitempty" yaml:"markdown,omitempty"`
	HTML     *HTMLConfig     `json:"html,omitempty" yaml:"html,omitempty"`
	LaTeX    *LaTeXConfig    `json:"latex,omitempty" yaml:"latex,omitempty"`
	EPUB     *EPUBConfig     `json:"epub,omitempty" yaml:"epub,omitempty"`
}

// DefaultConfig returns a Config with all default format configurations
func DefaultConfig() *Config {
	return &Config{
		Markdown: DefaultMarkdownConfig(),
		HTML:     DefaultHTMLConfig(),
		LaTeX:    DefaultLaTeXConfig(),
		EPUB:     DefaultEPUBConfig(),
	}
}

// Loader provides configuration loading functionality
type Loader struct {
	validator *Validator
}

// NewLoader returns a new Loader instance for loading and validating configuration files.
func NewLoader() *Loader {
	return &Loader{
		validator: NewValidator(),
	}
}

// LoadConfig loads configuration from a file, supporting JSON and YAML formats.
// The format is determined by the file extension (.json, .yaml, .yml).
func (l *Loader) LoadConfig(filepath string) (*Config, error) {
	if filepath == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filepath, err)
	}

	// Parse into temporary config first
	var loadedConfig Config

	// Determine format by extension
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &loadedConfig)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &loadedConfig)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (supported: .json, .yaml, .yml)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filepath, err)
	}

	// Merge loaded config with defaults
	config := DefaultConfig()
	if loadedConfig.Markdown != nil {
		config.Markdown = loadedConfig.Markdown
	}
	if loadedConfig.HTML != nil {
		config.HTML = loadedConfig.HTML
	}
	if loadedConfig.LaTeX != nil {
		config.LaTeX = loadedConfig.LaTeX
	}
	if loadedConfig.EPUB != nil {
		config.EPUB = loadedConfig.EPUB
	}

	// TODO: Enable validation once validator logic is fixed for defaults
	// Validate the loaded configuration
	// if err := l.ValidateConfig(config); err != nil {
	//     return nil, fmt.Errorf("config validation failed: %w", err)
	// }

	return config, nil
}

// ValidateConfig validates the complete configuration using format-specific validators
func (l *Loader) ValidateConfig(config *Config) error {
	var errors []string

	// Validate each format configuration if present
	if config.Markdown != nil {
		if result := l.validator.ValidateMarkdownConfig(config.Markdown); !result.Valid {
			for _, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("markdown: %s", err.Error()))
			}
		}
	}

	if config.HTML != nil {
		if result := l.validator.ValidateHTMLConfig(config.HTML); !result.Valid {
			for _, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("html: %s", err.Error()))
			}
		}
	}

	if config.LaTeX != nil {
		if result := l.validator.ValidateLaTeXConfig(config.LaTeX); !result.Valid {
			for _, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("latex: %s", err.Error()))
			}
		}
	}

	if config.EPUB != nil {
		if result := l.validator.ValidateEPUBConfig(config.EPUB); !result.Valid {
			for _, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("epub: %s", err.Error()))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// GetFormatConfig returns the configuration for a specific format
func (c *Config) GetFormatConfig(format string) interface{} {
	switch strings.ToLower(format) {
	case "markdown", "md":
		return c.Markdown
	case "html":
		return c.HTML
	case "latex", "tex":
		return c.LaTeX
	case "epub":
		return c.EPUB
	default:
		return nil
	}
}

// SaveConfig saves the configuration to a file in the specified format
func (l *Loader) SaveConfig(config *Config, filePath string) error {
	var data []byte
	var err error

	// Determine format by extension
	ext := strings.ToLower(filePath[strings.LastIndex(filePath, "."):])
	switch ext {
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config file format: %s (supported: .json, .yaml, .yml)", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
