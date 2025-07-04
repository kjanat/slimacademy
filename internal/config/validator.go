package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   string
	Issue   string
	Suggest string
}

func (e ValidationError) Error() string {
	if e.Suggest != "" {
		return fmt.Sprintf("config validation failed for %s='%s': %s (suggestion: %s)",
			e.Field, e.Value, e.Issue, e.Suggest)
	}
	return fmt.Sprintf("config validation failed for %s='%s': %s",
		e.Field, e.Value, e.Issue)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
}

// Validator provides configuration validation functionality
type Validator struct{}

// NewValidator returns a new Validator instance for configuration validation.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMarkdownConfig validates markdown configuration
func (v *Validator) ValidateMarkdownConfig(cfg *MarkdownConfig) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate emphasis markers for conflicts
	if err := v.validateEmphasisMarkers(cfg.BoldFormat, cfg.ItalicFormat); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate code markers using actual configuration values
	if err := v.validateCodeMarkers(cfg.CodeBlockMarker, cfg.InlineCodeMarker); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate list markers using actual configuration values
	if err := v.validateListMarkers(cfg.UnorderedListMarker, cfg.OrderedListMarker); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Basic format validation
	if cfg.BoldFormat == "" || cfg.ItalicFormat == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "Formats",
			Value:   fmt.Sprintf("bold='%s', italic='%s'", cfg.BoldFormat, cfg.ItalicFormat),
			Issue:   "empty format markers",
			Suggest: "provide non-empty format markers",
		})
		result.Valid = false
	}

	return result
}

// ValidateHTMLConfig validates HTML configuration
func (v *Validator) ValidateHTMLConfig(cfg *HTMLConfig) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate CSS classes
	if err := v.validateCSSClass(cfg.TableClass); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate element names
	elements := []string{cfg.BoldElement, cfg.ItalicElement, cfg.StrikeElement}
	for _, element := range elements {
		if element == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "HTMLElement",
				Value:   element,
				Issue:   "empty HTML element name",
				Suggest: "provide valid HTML element names",
			})
			result.Valid = false
		}
	}

	return result
}

// ValidateLaTeXConfig validates LaTeX configuration
func (v *Validator) ValidateLaTeXConfig(cfg *LaTeXConfig) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate document class
	if err := v.validateLaTeXDocumentClass(cfg.DocumentClass); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate packages
	for _, pkg := range cfg.Packages {
		if err := v.validateLaTeXPackage(pkg); err != nil {
			result.Errors = append(result.Errors, *err)
			result.Valid = false
		}
	}

	// Validate generated preamble for security
	generatedPreamble := cfg.GetDocumentPreamble()
	if err := v.validateLaTeXPreamble(generatedPreamble); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	return result
}

// ValidateEPUBConfig validates EPUB configuration
func (v *Validator) ValidateEPUBConfig(cfg *EPUBConfig) ValidationResult {
	result := ValidationResult{Valid: true}

	// Validate creator (author)
	if cfg.Creator == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "Creator",
			Value:   "",
			Issue:   "EPUB creator is recommended",
			Suggest: "provide a creator name",
		})
	}

	// Validate chapter prefix
	if err := v.validateHeadingPrefix(cfg.ChapterPrefix); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate language code
	if err := v.validateLanguageCode(cfg.Language); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate version
	if cfg.Version != "2.0" && cfg.Version != "3.0" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "Version",
			Value:   cfg.Version,
			Issue:   "non-standard EPUB version",
			Suggest: "use '2.0' or '3.0'",
		})
	}

	return result
}

// ValidateUUID validates a UUID string format - useful for external UUID sources
func (v *Validator) ValidateUUID(uuid string) *ValidationError {
	return v.validateUUID(uuid)
}

// Specific validation methods

func (v *Validator) validateHeadingPrefix(prefix string) *ValidationError {
	if len(prefix) == 0 {
		return &ValidationError{
			Field:   "HeadingPrefix",
			Value:   prefix,
			Issue:   "empty heading prefix",
			Suggest: "use '#' for standard markdown",
		}
	}

	if len(prefix) > 6 {
		return &ValidationError{
			Field:   "HeadingPrefix",
			Value:   prefix,
			Issue:   "heading prefix too long",
			Suggest: "use 1-6 characters maximum",
		}
	}

	// Check for conflicting characters
	if strings.ContainsAny(prefix, " \t\n\r") {
		return &ValidationError{
			Field:   "HeadingPrefix",
			Value:   prefix,
			Issue:   "heading prefix contains whitespace",
			Suggest: "use non-whitespace characters only",
		}
	}

	return nil
}

func (v *Validator) validateListMarkers(unordered, ordered string) *ValidationError {
	// Check prefix-free property
	if strings.HasPrefix(unordered, ordered) || strings.HasPrefix(ordered, unordered) {
		return &ValidationError{
			Field:   "ListMarkers",
			Value:   fmt.Sprintf("unordered='%s', ordered='%s'", unordered, ordered),
			Issue:   "list markers are not prefix-free",
			Suggest: "use distinct markers like '-' and '1.'",
		}
	}

	// Validate individual markers
	if len(unordered) == 0 || len(ordered) == 0 {
		return &ValidationError{
			Field:   "ListMarkers",
			Value:   fmt.Sprintf("unordered='%s', ordered='%s'", unordered, ordered),
			Issue:   "empty list marker",
			Suggest: "provide non-empty markers",
		}
	}

	return nil
}

func (v *Validator) validateEmphasisMarkers(bold, italic string) *ValidationError {
	// Check for conflicts
	if bold == italic {
		return &ValidationError{
			Field:   "EmphasisMarkers",
			Value:   fmt.Sprintf("bold='%s', italic='%s'", bold, italic),
			Issue:   "bold and italic markers are identical",
			Suggest: "use different markers like '**' and '*'",
		}
	}

	// Check prefix-free property
	if strings.HasPrefix(bold, italic) || strings.HasPrefix(italic, bold) {
		return &ValidationError{
			Field:   "EmphasisMarkers",
			Value:   fmt.Sprintf("bold='%s', italic='%s'", bold, italic),
			Issue:   "emphasis markers are not prefix-free",
			Suggest: "use distinct markers",
		}
	}

	return nil
}

func (v *Validator) validateCodeMarkers(block, inline string) *ValidationError {
	if strings.Contains(block, inline) || strings.Contains(inline, block) {
		return &ValidationError{
			Field:   "CodeMarkers",
			Value:   fmt.Sprintf("block='%s', inline='%s'", block, inline),
			Issue:   "code markers conflict",
			Suggest: "use distinct markers like '```' and '`'",
		}
	}
	return nil
}

func (v *Validator) validateCSSClass(class string) *ValidationError {
	if class == "" {
		return nil // Empty class is valid
	}

	// CSS class name validation
	validClass := regexp.MustCompile(`^[a-zA-Z][-_a-zA-Z0-9]*$`)
	if !validClass.MatchString(class) {
		return &ValidationError{
			Field:   "CSSClass",
			Value:   class,
			Issue:   "invalid CSS class name",
			Suggest: "use alphanumeric characters, hyphens, and underscores only",
		}
	}

	return nil
}

func (v *Validator) validateLaTeXDocumentClass(class string) *ValidationError {
	validClasses := []string{"article", "book", "report", "letter", "memoir", "scrbook", "scrreprt", "scrartcl"}

	if slices.Contains(validClasses, class) {
		return nil
	}

	return &ValidationError{
		Field:   "DocumentClass",
		Value:   class,
		Issue:   "unknown LaTeX document class",
		Suggest: "use standard classes like 'article' or 'book'",
	}
}

func (v *Validator) validateLaTeXPackage(pkg string) *ValidationError {
	if pkg == "" {
		return &ValidationError{
			Field:   "Package",
			Value:   pkg,
			Issue:   "empty package name",
			Suggest: "provide valid package name",
		}
	}

	// Basic LaTeX package name validation
	validPackage := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
	if !validPackage.MatchString(pkg) {
		return &ValidationError{
			Field:   "Package",
			Value:   pkg,
			Issue:   "invalid LaTeX package name",
			Suggest: "use alphanumeric characters and hyphens only",
		}
	}

	return nil
}

func (v *Validator) validateLaTeXPreamble(preamble string) *ValidationError {
	// Check for potentially dangerous commands
	dangerous := []string{"\\write", "\\input", "\\include", "\\openout", "\\openin"}

	for _, cmd := range dangerous {
		if strings.Contains(preamble, cmd) {
			return &ValidationError{
				Field:   "Preamble",
				Value:   preamble,
				Issue:   fmt.Sprintf("potentially dangerous command: %s", cmd),
				Suggest: "remove file system commands from preamble",
			}
		}
	}

	// Check for valid UTF-8
	if !utf8.ValidString(preamble) {
		return &ValidationError{
			Field:   "Preamble",
			Value:   preamble,
			Issue:   "invalid UTF-8 encoding",
			Suggest: "ensure preamble is valid UTF-8",
		}
	}

	return nil
}

func (v *Validator) validateLanguageCode(lang string) *ValidationError {
	if lang == "" {
		return &ValidationError{
			Field:   "Language",
			Value:   lang,
			Issue:   "empty language code",
			Suggest: "use ISO 639-1 code like 'en' or 'nl'",
		}
	}

	// Basic language code validation (ISO 639-1)
	validLang := regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	if !validLang.MatchString(lang) {
		return &ValidationError{
			Field:   "Language",
			Value:   lang,
			Issue:   "invalid language code format",
			Suggest: "use ISO 639-1 format like 'en', 'nl', or 'en-US'",
		}
	}

	return nil
}

func (v *Validator) validateUUID(uuid string) *ValidationError {
	validUUID := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !validUUID.MatchString(uuid) {
		return &ValidationError{
			Field:   "UUID",
			Value:   uuid,
			Issue:   "invalid UUID format",
			Suggest: "use standard UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		}
	}

	return nil
}
