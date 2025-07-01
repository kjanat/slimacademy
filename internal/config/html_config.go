package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// HTMLConfig holds configuration for HTML output
type HTMLConfig struct {
	// Element configurations
	BoldElement        string `json:"boldElement"`
	ItalicElement      string `json:"italicElement"`
	StrikeElement      string `json:"strikeElement"`
	UnderlineElement   string `json:"underlineElement"`
	SubscriptElement   string `json:"subscriptElement"`
	SuperscriptElement string `json:"superscriptElement"`
	HighlightElement   string `json:"highlightElement"`
	
	// Document structure
	UseHTML5           bool   `json:"useHTML5"`
	IncludeCSS         bool   `json:"includeCSS"`
	CSSStylesheet      string `json:"cssStylesheet"`
	DocType            string `json:"docType"`
	
	// Formatting options
	PrettyPrint        bool   `json:"prettyPrint"`
	Charset            string `json:"charset"`
	Language           string `json:"language"`
	
	// Table configuration
	TableClass         string `json:"tableClass"`
	TableBorder        bool   `json:"tableBorder"`
	
	// Code configuration
	UseCodeElement     bool   `json:"useCodeElement"`
	CodeClass          string `json:"codeClass"`
}

// DefaultHTMLConfig returns the default HTML configuration
func DefaultHTMLConfig() *HTMLConfig {
	return &HTMLConfig{
		BoldElement:        "strong",
		ItalicElement:      "em",
		StrikeElement:      "del",
		UnderlineElement:   "u",
		SubscriptElement:   "sub",
		SuperscriptElement: "sup",
		HighlightElement:   "mark",
		
		UseHTML5:           true,
		IncludeCSS:         true,
		CSSStylesheet:      "",
		DocType:            "<!DOCTYPE html>",
		
		PrettyPrint:        false,
		Charset:            "UTF-8",
		Language:           "en",
		
		TableClass:         "",
		TableBorder:        true,
		
		UseCodeElement:     true,
		CodeClass:          "",
	}
}

// LoadHTMLConfig loads configuration from a JSON file
func LoadHTMLConfig(filename string) (*HTMLConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config HTMLConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// GetBoldTags returns the opening and closing tags for bold text
func (c *HTMLConfig) GetBoldTags() (string, string) {
	return fmt.Sprintf("<%s>", c.BoldElement), fmt.Sprintf("</%s>", c.BoldElement)
}

// GetItalicTags returns the opening and closing tags for italic text
func (c *HTMLConfig) GetItalicTags() (string, string) {
	return fmt.Sprintf("<%s>", c.ItalicElement), fmt.Sprintf("</%s>", c.ItalicElement)
}

// GetStrikeTags returns the opening and closing tags for strikethrough text
func (c *HTMLConfig) GetStrikeTags() (string, string) {
	return fmt.Sprintf("<%s>", c.StrikeElement), fmt.Sprintf("</%s>", c.StrikeElement)
}

// GetUnderlineTags returns the opening and closing tags for underlined text
func (c *HTMLConfig) GetUnderlineTags() (string, string) {
	return fmt.Sprintf("<%s>", c.UnderlineElement), fmt.Sprintf("</%s>", c.UnderlineElement)
}

// GetSubscriptTags returns the opening and closing tags for subscript text
func (c *HTMLConfig) GetSubscriptTags() (string, string) {
	return fmt.Sprintf("<%s>", c.SubscriptElement), fmt.Sprintf("</%s>", c.SubscriptElement)
}

// GetSuperscriptTags returns the opening and closing tags for superscript text
func (c *HTMLConfig) GetSuperscriptTags() (string, string) {
	return fmt.Sprintf("<%s>", c.SuperscriptElement), fmt.Sprintf("</%s>", c.SuperscriptElement)
}

// GetHighlightTags returns the opening and closing tags for highlighted text
func (c *HTMLConfig) GetHighlightTags() (string, string) {
	return fmt.Sprintf("<%s>", c.HighlightElement), fmt.Sprintf("</%s>", c.HighlightElement)
}

// GetTableAttributes returns table-specific attributes
func (c *HTMLConfig) GetTableAttributes() string {
	attrs := ""
	if c.TableClass != "" {
		attrs += fmt.Sprintf(` class="%s"`, c.TableClass)
	}
	if c.TableBorder {
		attrs += ` border="1"`
	}
	return attrs
}