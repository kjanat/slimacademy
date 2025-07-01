package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LaTeXConfig holds configuration for LaTeX output
type LaTeXConfig struct {
	// Command configurations
	BoldCommand        string `json:"boldCommand"`
	ItalicCommand      string `json:"italicCommand"`
	StrikeCommand      string `json:"strikeCommand"`
	UnderlineCommand   string `json:"underlineCommand"`
	SubscriptCommand   string `json:"subscriptCommand"`
	SuperscriptCommand string `json:"superscriptCommand"`
	HighlightCommand   string `json:"highlightCommand"`
	
	// Document structure
	DocumentClass      string   `json:"documentClass"`
	DocumentOptions    []string `json:"documentOptions"`
	Packages           []string `json:"packages"`
	
	// Formatting options
	UseUTF8            bool     `json:"useUTF8"`
	UseGeometry        bool     `json:"useGeometry"`
	GeometryOptions    []string `json:"geometryOptions"`
	
	// Section formatting
	SectionCommand     string   `json:"sectionCommand"`
	SubsectionCommand  string   `json:"subsectionCommand"`
	SubsubsectionCommand string `json:"subsubsectionCommand"`
	
	// Table configuration
	TableEnvironment   string   `json:"tableEnvironment"`
	TableAlignment     string   `json:"tableAlignment"`
	UseBooktabs        bool     `json:"useBooktabs"`
	
	// Math configuration
	MathEnvironment    string   `json:"mathEnvironment"`
	InlineMathDelim    string   `json:"inlineMathDelim"`
	
	// Bibliography
	BibliographyStyle  string   `json:"bibliographyStyle"`
	UseBiblatex        bool     `json:"useBiblatex"`
}

// DefaultLaTeXConfig returns the default LaTeX configuration
func DefaultLaTeXConfig() *LaTeXConfig {
	return &LaTeXConfig{
		BoldCommand:        "textbf",
		ItalicCommand:      "emph",
		StrikeCommand:      "sout",
		UnderlineCommand:   "underline",
		SubscriptCommand:   "textsubscript",
		SuperscriptCommand: "textsuperscript",
		HighlightCommand:   "hl",
		
		DocumentClass:      "article",
		DocumentOptions:    []string{"11pt", "a4paper"},
		Packages:           []string{"inputenc", "fontenc", "geometry", "ulem", "soul", "amsmath", "amsfonts", "amssymb", "hyperref"},
		
		UseUTF8:            true,
		UseGeometry:        true,
		GeometryOptions:    []string{"margin=1in"},
		
		SectionCommand:     "section",
		SubsectionCommand:  "subsection",
		SubsubsectionCommand: "subsubsection",
		
		TableEnvironment:   "tabular",
		TableAlignment:     "l",
		UseBooktabs:        true,
		
		MathEnvironment:    "equation",
		InlineMathDelim:    "$",
		
		BibliographyStyle:  "plain",
		UseBiblatex:        false,
	}
}

// LoadLaTeXConfig loads configuration from a JSON file
func LoadLaTeXConfig(filename string) (*LaTeXConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config LaTeXConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// GetBoldCommand returns the LaTeX command for bold text
func (c *LaTeXConfig) GetBoldCommand() string {
	return fmt.Sprintf("\\%s{", c.BoldCommand)
}

// GetItalicCommand returns the LaTeX command for italic text
func (c *LaTeXConfig) GetItalicCommand() string {
	return fmt.Sprintf("\\%s{", c.ItalicCommand)
}

// GetStrikeCommand returns the LaTeX command for strikethrough text
func (c *LaTeXConfig) GetStrikeCommand() string {
	return fmt.Sprintf("\\%s{", c.StrikeCommand)
}

// GetUnderlineCommand returns the LaTeX command for underlined text
func (c *LaTeXConfig) GetUnderlineCommand() string {
	return fmt.Sprintf("\\%s{", c.UnderlineCommand)
}

// GetSubscriptCommand returns the LaTeX command for subscript text
func (c *LaTeXConfig) GetSubscriptCommand() string {
	return fmt.Sprintf("\\%s{", c.SubscriptCommand)
}

// GetSuperscriptCommand returns the LaTeX command for superscript text
func (c *LaTeXConfig) GetSuperscriptCommand() string {
	return fmt.Sprintf("\\%s{", c.SuperscriptCommand)
}

// GetHighlightCommand returns the LaTeX command for highlighted text
func (c *LaTeXConfig) GetHighlightCommand() string {
	return fmt.Sprintf("\\%s{", c.HighlightCommand)
}

// GetDocumentPreamble returns the document preamble
func (c *LaTeXConfig) GetDocumentPreamble() string {
	preamble := fmt.Sprintf("\\documentclass[%s]{%s}\n", 
		joinStrings(c.DocumentOptions, ","), c.DocumentClass)
	
	// Add packages
	for _, pkg := range c.Packages {
		if pkg == "inputenc" && c.UseUTF8 {
			preamble += "\\usepackage[utf8]{inputenc}\n"
		} else if pkg == "fontenc" {
			preamble += "\\usepackage[T1]{fontenc}\n"
		} else if pkg == "geometry" && c.UseGeometry {
			preamble += fmt.Sprintf("\\usepackage[%s]{geometry}\n", 
				joinStrings(c.GeometryOptions, ","))
		} else {
			preamble += fmt.Sprintf("\\usepackage{%s}\n", pkg)
		}
	}
	
	return preamble
}

// GetHeadingCommand returns the heading command for the given level
func (c *LaTeXConfig) GetHeadingCommand(level int) string {
	switch level {
	case 1:
		return fmt.Sprintf("\\%s{", c.SectionCommand)
	case 2:
		return fmt.Sprintf("\\%s{", c.SubsectionCommand)
	case 3:
		return fmt.Sprintf("\\%s{", c.SubsubsectionCommand)
	default:
		return fmt.Sprintf("\\%s{", c.SubsubsectionCommand)
	}
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}