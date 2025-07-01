package sanitizer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kjanat/slimacademy/internal/models"
)

// Warning represents a sanitization warning with location info
type Warning struct {
	Location string
	Issue    string
	Original string
	Fixed    string
}

// Result contains the sanitized book and any warnings
type Result struct {
	Book     *models.Book
	Warnings []Warning
}

// Sanitizer cleans and validates document content before event generation
type Sanitizer struct {
	warnings []Warning
}

// NewSanitizer creates a new document sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		warnings: make([]Warning, 0),
	}
}

// Sanitize processes a book and returns cleaned content with warnings
func (s *Sanitizer) Sanitize(book *models.Book) *Result {
	s.warnings = nil // Reset warnings

	// Create a copy to avoid mutating the original
	sanitized := s.deepCopyBook(book)

	// Sanitize document content
	s.sanitizeContent(sanitized)
	s.sanitizeChapters(sanitized)
	s.sanitizeImages(sanitized)

	return &Result{
		Book:     sanitized,
		Warnings: s.warnings,
	}
}

// sanitizeContent cleans the document body content
func (s *Sanitizer) sanitizeContent(book *models.Book) {
	if book.Content.Body.Content == nil {
		return
	}

	for i := range book.Content.Body.Content {
		element := &book.Content.Body.Content[i]
		location := fmt.Sprintf("content[%d]", i)

		if element.Paragraph != nil {
			s.sanitizeParagraph(element.Paragraph, location)
		} else if element.Table != nil {
			s.sanitizeTable(element.Table, location)
		}
	}
}

// sanitizeParagraph cleans paragraph content
func (s *Sanitizer) sanitizeParagraph(paragraph *models.Paragraph, location string) {
	for i := range paragraph.Elements {
		element := &paragraph.Elements[i]
		elemLocation := fmt.Sprintf("%s.elements[%d]", location, i)

		if element.TextRun != nil {
			s.sanitizeTextRun(element.TextRun, elemLocation)
		}
	}

	// Check for empty heading payloads
	if s.isHeading(paragraph) {
		text := s.extractText(paragraph)
		if strings.TrimSpace(text) == "" {
			s.addWarning(location, "empty heading payload", text, "[EMPTY HEADING REMOVED]")
		}
	}
}

// sanitizeTextRun cleans text run content
func (s *Sanitizer) sanitizeTextRun(textRun *models.TextRun, location string) {
	original := textRun.Content
	cleaned := s.sanitizeText(original)

	if cleaned != original {
		s.addWarning(location, "text content sanitized", original, cleaned)
		textRun.Content = cleaned
	}

	// Validate link URLs if present
	if textRun.TextStyle.Link != nil {
		s.validateLinkURL(textRun.TextStyle.Link, location)
	}
}

// sanitizeTable cleans table content
func (s *Sanitizer) sanitizeTable(table *models.Table, location string) {
	for rowIdx, row := range table.TableRows {
		for cellIdx, cell := range row.TableCells {
			cellLocation := fmt.Sprintf("%s.row[%d].cell[%d]", location, rowIdx, cellIdx)
			for contentIdx := range cell.Content {
				contentLocation := fmt.Sprintf("%s.content[%d]", cellLocation, contentIdx)
				if cell.Content[contentIdx].Paragraph != nil {
					s.sanitizeParagraph(cell.Content[contentIdx].Paragraph, contentLocation)
				}
			}
		}
	}
}

// sanitizeChapters validates chapter structure
func (s *Sanitizer) sanitizeChapters(book *models.Book) {
	for i := range book.Chapters {
		chapter := &book.Chapters[i]
		location := fmt.Sprintf("chapters[%d]", i)

		if strings.TrimSpace(chapter.Title) == "" {
			s.addWarning(location, "empty chapter title", chapter.Title, "[EMPTY CHAPTER]")
		}

		for j := range chapter.SubChapters {
			subChapter := &chapter.SubChapters[j]
			subLocation := fmt.Sprintf("%s.subchapters[%d]", location, j)

			if strings.TrimSpace(subChapter.Title) == "" {
				s.addWarning(subLocation, "empty subchapter title", subChapter.Title, "[EMPTY SUBCHAPTER]")
			}
		}
	}
}

// sanitizeImages validates image references
func (s *Sanitizer) sanitizeImages(book *models.Book) {
	for i := range book.Images {
		image := &book.Images[i]
		location := fmt.Sprintf("images[%d]", i)

		if image.ImageURL == "" {
			s.addWarning(location, "empty image URL", "", "[MISSING IMAGE]")
		}
	}
}

// sanitizeText removes control characters and normalizes whitespace
func (s *Sanitizer) sanitizeText(text string) string {
	if text == "" {
		return text
	}

	// Remove non-printable control characters except common whitespace
	var cleaned strings.Builder
	cleaned.Grow(len(text))

	for _, r := range text {
		switch {
		case r == '\n' || r == '\t' || r == '\r' || r == ' ':
			// Keep common whitespace
			cleaned.WriteRune(r)
		case unicode.IsControl(r):
			// Remove other control characters
			continue
		case !utf8.ValidRune(r):
			// Remove invalid UTF-8 runes
			continue
		default:
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()

	// Normalize excessive whitespace
	result = s.normalizeWhitespace(result)

	return result
}

// normalizeWhitespace removes excessive whitespace while preserving structure
func (s *Sanitizer) normalizeWhitespace(text string) string {
	// Remove carriage returns
	text = strings.ReplaceAll(text, "\r", "")

	// Replace multiple spaces with single space (except leading/trailing)
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	// Reconstruct with single spaces between words
	return strings.Join(words, " ")
}

// validateLinkURL checks if a link URL is well-formed
func (s *Sanitizer) validateLinkURL(link *models.Link, location string) {
	if link.URL == "" {
		s.addWarning(location, "empty link URL", "", "[EMPTY LINK]")
		return
	}

	// Basic URL validation - check for common patterns
	url := strings.TrimSpace(link.URL)
	if url != link.URL {
		s.addWarning(location, "link URL has whitespace", link.URL, url)
		link.URL = url
	}

	// Check for unbalanced tags in URL (common issue)
	if strings.Contains(url, "<") && !strings.Contains(url, ">") {
		s.addWarning(location, "unbalanced tags in URL", url, strings.ReplaceAll(url, "<", ""))
		link.URL = strings.ReplaceAll(url, "<", "")
	}
}

// Helper methods

func (s *Sanitizer) isHeading(paragraph *models.Paragraph) bool {
	return strings.HasPrefix(paragraph.ParagraphStyle.NamedStyleType, "HEADING_")
}

func (s *Sanitizer) extractText(paragraph *models.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return text.String()
}

func (s *Sanitizer) addWarning(location, issue, original, fixed string) {
	s.warnings = append(s.warnings, Warning{
		Location: location,
		Issue:    issue,
		Original: original,
		Fixed:    fixed,
	})
}

// deepCopyBook creates a deep copy of the book to avoid mutations
func (s *Sanitizer) deepCopyBook(book *models.Book) *models.Book {
	// For now, return the original book
	// In a production system, implement proper deep copy
	// This requires careful handling of all nested structures
	return book
}
