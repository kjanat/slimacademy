package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// EPUBConfig holds configuration for EPUB output
type EPUBConfig struct {
	// Metadata
	Creator        string            `json:"creator" yaml:"creator"`
	Publisher      string            `json:"publisher" yaml:"publisher"`
	Language       string            `json:"language" yaml:"language"`
	Subject        string            `json:"subject" yaml:"subject"`
	Description    string            `json:"description" yaml:"description"`
	Rights         string            `json:"rights" yaml:"rights"`
	CustomMetadata map[string]string `json:"customMetadata" yaml:"customMetadata"`

	// EPUB structure
	Version       string `json:"version" yaml:"version"`
	ChapterSplit  bool   `json:"chapterSplit" yaml:"chapterSplit"`
	ChapterPrefix string `json:"chapterPrefix" yaml:"chapterPrefix"`
	GenerateTOC   bool   `json:"generateTOC" yaml:"generateTOC"`
	TOCDepth      int    `json:"tocDepth"`

	// CSS and styling
	IncludeCSS bool   `json:"includeCSS"`
	CSSFile    string `json:"cssFile"`
	CustomCSS  string `json:"customCSS"`

	// HTML configuration (embedded)
	HTMLConfig *HTMLConfig `json:"htmlConfig"`

	// Content options
	IncludeImages    bool `json:"includeImages"`
	ImageCompression bool `json:"imageCompression"`
	ImageQuality     int  `json:"imageQuality"`

	// File naming
	FilenameSanitize  bool `json:"filenameSanitize"`
	MaxFilenameLength int  `json:"maxFilenameLength"`

	// Navigation
	UseLinearReading bool `json:"useLinearReading"`
	GuideEnabled     bool `json:"guideEnabled"`
	LandmarkNav      bool `json:"landmarkNav"`
}

// DefaultEPUBConfig returns a pointer to an EPUBConfig struct initialized with default values for all configuration fields, including metadata, EPUB structure, styling, content options, file naming, and navigation settings.
func DefaultEPUBConfig() *EPUBConfig {
	return &EPUBConfig{
		Creator:        "SlimAcademy Transformer",
		Publisher:      "SlimAcademy",
		Language:       "en",
		Subject:        "",
		Description:    "",
		Rights:         "",
		CustomMetadata: make(map[string]string),

		Version:       "2.0",
		ChapterSplit:  true,
		ChapterPrefix: "chapter_",
		GenerateTOC:   true,
		TOCDepth:      3,

		IncludeCSS: true,
		CSSFile:    "styles.css",
		CustomCSS:  "",

		HTMLConfig: DefaultHTMLConfig(),

		IncludeImages:    true,
		ImageCompression: false,
		ImageQuality:     80,

		FilenameSanitize:  true,
		MaxFilenameLength: 100,

		UseLinearReading: true,
		GuideEnabled:     true,
		LandmarkNav:      true,
	}
}

// LoadEPUBConfig reads an EPUB configuration from a JSON file and returns an EPUBConfig instance.
// Returns an error if the file cannot be read or the JSON is invalid. If the embedded HTMLConfig is missing, it is set to a default value.
func LoadEPUBConfig(filename string) (*EPUBConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config EPUBConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Ensure HTMLConfig is set
	if config.HTMLConfig == nil {
		config.HTMLConfig = DefaultHTMLConfig()
	}

	return &config, nil
}

// GetMetadataElement returns a metadata element for OPF
func (c *EPUBConfig) GetMetadataElement(key, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("    <dc:%s>%s</dc:%s>\n", key, escapeXML(value), key)
}

// GetCustomMetadataElements returns custom metadata elements
func (c *EPUBConfig) GetCustomMetadataElements() string {
	if len(c.CustomMetadata) == 0 {
		return ""
	}

	var elements string
	for key, value := range c.CustomMetadata {
		if value != "" {
			elements += fmt.Sprintf("    <meta name=\"%s\" content=\"%s\"/>\n",
				escapeXML(key), escapeXML(value))
		}
	}
	return elements
}

// GetChapterFilename returns a sanitized chapter filename
func (c *EPUBConfig) GetChapterFilename(title, id string) string {
	filename := c.ChapterPrefix + id

	if c.FilenameSanitize {
		filename = sanitizeFilename(filename)
	}

	if c.MaxFilenameLength > 0 && len(filename) > c.MaxFilenameLength {
		filename = filename[:c.MaxFilenameLength]
	}

	return filename + ".xhtml"
}

// GetDefaultCSS returns the default CSS for EPUB
func (c *EPUBConfig) GetDefaultCSS() string {
	if c.CustomCSS != "" {
		return c.CustomCSS
	}

	return `
body {
    font-family: 'Georgia', 'Times New Roman', serif;
    line-height: 1.6;
    margin: 1em;
    color: #000;
    background-color: #fff;
}

h1, h2, h3, h4, h5, h6 {
    color: #333;
    margin-top: 1.5em;
    margin-bottom: 0.5em;
    page-break-after: avoid;
    font-weight: bold;
}

h1 { font-size: 2em; }
h2 { font-size: 1.5em; }
h3 { font-size: 1.3em; }
h4 { font-size: 1.1em; }
h5 { font-size: 1em; }
h6 { font-size: 0.9em; }

p {
    margin: 0.5em 0;
    text-align: justify;
    text-indent: 1em;
}

p:first-child, h1 + p, h2 + p, h3 + p, h4 + p, h5 + p, h6 + p {
    text-indent: 0;
}

a {
    color: #0066cc;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

strong, b {
    font-weight: bold;
}

em, i {
    font-style: italic;
}

u {
    text-decoration: underline;
}

del, s {
    text-decoration: line-through;
}

sub {
    vertical-align: sub;
    font-size: smaller;
}

sup {
    vertical-align: super;
    font-size: smaller;
}

mark {
    background-color: #fff3cd;
    padding: 0.1em 0.2em;
}

ul, ol {
    margin: 0.5em 0;
    padding-left: 2em;
}

li {
    margin: 0.2em 0;
}

table {
    border-collapse: collapse;
    width: 100%;
    margin: 1em 0;
}

th, td {
    border: 1px solid #ccc;
    padding: 0.5em;
    text-align: left;
    vertical-align: top;
}

th {
    background-color: #f5f5f5;
    font-weight: bold;
}

blockquote {
    margin: 1em 2em;
    padding: 0.5em 1em;
    border-left: 3px solid #ccc;
    font-style: italic;
}

code {
    font-family: 'Courier New', monospace;
    background-color: #f5f5f5;
    padding: 0.1em 0.3em;
    border-radius: 3px;
}

pre {
    font-family: 'Courier New', monospace;
    background-color: #f5f5f5;
    padding: 1em;
    margin: 1em 0;
    border-radius: 5px;
    overflow-x: auto;
}

img {
    max-width: 100%;
    height: auto;
    display: block;
    margin: 1em auto;
}

.page-break {
    page-break-before: always;
}`
}

// sanitizeFilename returns a sanitized version of the input filename by replacing invalid or problematic characters with underscores and removing non-printable ASCII characters.
func sanitizeFilename(filename string) string {
	// Replace common problematic characters
	replacements := map[rune]string{
		'/':  "_",
		'\\': "_",
		':':  "_",
		'*':  "_",
		'?':  "_",
		'"':  "_",
		'<':  "_",
		'>':  "_",
		'|':  "_",
		' ':  "_",
	}

	var result []rune
	for _, r := range filename {
		if replacement, exists := replacements[r]; exists {
			for _, newR := range replacement {
				result = append(result, newR)
			}
		} else if r >= 32 && r < 127 { // printable ASCII
			result = append(result, r)
		}
	}

	return string(result)
}

// escapeXML returns a copy of the input string with XML special characters replaced by their corresponding escape sequences.
func escapeXML(text string) string {
	replacements := map[rune]string{
		'&':  "&amp;",
		'<':  "&lt;",
		'>':  "&gt;",
		'"':  "&quot;",
		'\'': "&apos;",
	}

	var result []rune
	for _, r := range text {
		if replacement, exists := replacements[r]; exists {
			for _, newR := range replacement {
				result = append(result, newR)
			}
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
