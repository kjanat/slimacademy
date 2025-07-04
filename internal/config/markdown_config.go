package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// MarkdownConfig holds configuration for markdown formatting
type MarkdownConfig struct {
	ItalicFormat        string `json:"italicFormat" yaml:"italicFormat"`
	BoldFormat          string `json:"boldFormat" yaml:"boldFormat"`
	StrikethroughFormat string `json:"strikethroughFormat" yaml:"strikethroughFormat"`
	UnderlineFormat     string `json:"underlineFormat" yaml:"underlineFormat"`
	SubscriptFormat     string `json:"subscriptFormat" yaml:"subscriptFormat"`
	SuperscriptFormat   string `json:"superscriptFormat" yaml:"superscriptFormat"`
	HighlightFormat     string `json:"highlightFormat" yaml:"highlightFormat"`
	EmphasizedLinks     bool   `json:"emphasizedLinks" yaml:"emphasizedLinks"`
	CodeLinks           bool   `json:"codeLinks" yaml:"codeLinks"`

	// Code formatting
	InlineCodeMarker string `json:"inlineCodeMarker" yaml:"inlineCodeMarker"`
	CodeBlockMarker  string `json:"codeBlockMarker" yaml:"codeBlockMarker"`

	// List formatting
	UnorderedListMarker string `json:"unorderedListMarker" yaml:"unorderedListMarker"`
	OrderedListMarker   string `json:"orderedListMarker" yaml:"orderedListMarker"`
}

// DefaultMarkdownConfig returns a MarkdownConfig instance initialized with standard markdown and HTML formatting markers for various text styles.
func DefaultMarkdownConfig() *MarkdownConfig {
	return &MarkdownConfig{
		ItalicFormat:        "_",
		BoldFormat:          "**",
		StrikethroughFormat: "~~",
		UnderlineFormat:     "<ins></ins>",
		SubscriptFormat:     "<sub></sub>",
		SuperscriptFormat:   "<sup></sup>",
		HighlightFormat:     "==",
		EmphasizedLinks:     false,
		CodeLinks:           true,

		// Code formatting
		InlineCodeMarker: "`",
		CodeBlockMarker:  "```",

		// List formatting
		UnorderedListMarker: "-",
		OrderedListMarker:   "1.",
	}
}

// LoadMarkdownConfig reads a JSON file and unmarshals its contents into a MarkdownConfig struct.
// Returns a pointer to the loaded configuration and an error if file reading or JSON parsing fails.
func LoadMarkdownConfig(filename string) (*MarkdownConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MarkdownConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// GetBoldMarkers returns the opening and closing markers for bold text
func (c *MarkdownConfig) GetBoldMarkers() (string, string) {
	return c.BoldFormat, c.BoldFormat
}

// GetItalicMarkers returns the opening and closing markers for italic text
func (c *MarkdownConfig) GetItalicMarkers() (string, string) {
	return c.ItalicFormat, c.ItalicFormat
}

// GetStrikethroughMarkers returns the opening and closing markers for strikethrough text
func (c *MarkdownConfig) GetStrikethroughMarkers() (string, string) {
	return c.StrikethroughFormat, c.StrikethroughFormat
}

// GetUnderlineMarkers returns the opening and closing markers for underlined text
func (c *MarkdownConfig) GetUnderlineMarkers() (string, string) {
	if c.UnderlineFormat == "<ins></ins>" {
		return "<ins>", "</ins>"
	}
	return c.UnderlineFormat, c.UnderlineFormat
}

// GetSubscriptMarkers returns the opening and closing markers for subscript text
func (c *MarkdownConfig) GetSubscriptMarkers() (string, string) {
	if c.SubscriptFormat == "<sub></sub>" {
		return "<sub>", "</sub>"
	}
	return c.SubscriptFormat, c.SubscriptFormat
}

// GetSuperscriptMarkers returns the opening and closing markers for superscript text
func (c *MarkdownConfig) GetSuperscriptMarkers() (string, string) {
	if c.SuperscriptFormat == "<sup></sup>" {
		return "<sup>", "</sup>"
	}
	return c.SuperscriptFormat, c.SuperscriptFormat
}

// GetHighlightMarkers returns the opening and closing markers for highlighted text
func (c *MarkdownConfig) GetHighlightMarkers() (string, string) {
	return c.HighlightFormat, c.HighlightFormat
}
