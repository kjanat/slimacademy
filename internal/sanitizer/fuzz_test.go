package sanitizer

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/kjanat/slimacademy/internal/models"
)

// FuzzSanitize tests the URL sanitization function with random inputs
func FuzzSanitize(f *testing.F) {
	// Seed with interesting test cases
	testCases := []string{
		"",
		"http://example.com",
		"https://safe.example.com/path?query=value",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
		"vbscript:msgbox('xss')",
		"mailto:test@example.com",
		"tel:+1234567890",
		"//example.com/path",
		"../../../etc/passwd",
		"<script>alert('xss')</script>",
		"http://example.com<script>alert('xss')</script>",
		"   http://example.com   ",
		"http://user:pass@example.com:8080/path?query=value#fragment",
		"ftp://example.com/file.txt",
		string([]byte{0x00, 0x01, 0x02, 0xFF}), // Binary data
		strings.Repeat("a", 10000),             // Very long string
		"🚀💻🔒",                                  // Unicode emojis
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	s := NewSanitizer()

	f.Fuzz(func(t *testing.T, input string) {
		// Test URL sanitization
		result := s.sanitizeURL(input, "fuzz-test")

		// Verify the result doesn't crash and produces valid output
		if !utf8.ValidString(result) {
			t.Errorf("sanitizeURL produced invalid UTF-8: %q", result)
		}

		// Verify dangerous patterns are removed
		lowerResult := strings.ToLower(result)
		dangerousPatterns := []string{
			"javascript:",
			"data:",
			"vbscript:",
			"<script",
			"</script>",
		}

		for _, pattern := range dangerousPatterns {
			if strings.Contains(lowerResult, pattern) {
				t.Errorf("sanitizeURL failed to remove dangerous pattern %q from input %q, got %q",
					pattern, input, result)
			}
		}

		// Verify result is either a safe URL or safe default
		if result != "#" {
			// Should be a valid, safe URL or relative path
			if strings.HasPrefix(lowerResult, "javascript:") ||
				strings.HasPrefix(lowerResult, "data:") ||
				strings.HasPrefix(lowerResult, "vbscript:") {
				t.Errorf("sanitizeURL returned unsafe URL: %q", result)
			}
		}
	})
}

// FuzzTextSanitization tests text sanitization with random inputs
func FuzzTextSanitization(f *testing.F) {
	// Seed with test cases
	testCases := []string{
		"",
		"Hello, World!",
		"Text with\nnewlines\nand\ttabs",
		"Text with \x00 null \x01 bytes",
		"Mixed\r\nline\nendings\r",
		"Unicode: 你好世界 🌍",
		strings.Repeat("a", 1000),
		"\u202E\u202D", // Unicode direction overrides
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	s := NewSanitizer()

	f.Fuzz(func(t *testing.T, input string) {
		result := s.sanitizeText(input)

		// Verify the result is valid UTF-8
		if !utf8.ValidString(result) {
			t.Errorf("sanitizeText produced invalid UTF-8: %q", result)
		}

		// Verify no control characters except allowed whitespace
		for _, r := range result {
			if r < 32 && r != '\n' && r != '\t' && r != '\r' && r != ' ' {
				t.Errorf("sanitizeText failed to remove control character U+%04X", r)
			}
		}

		// Verify result length is reasonable (shouldn't grow excessively)
		if len(result) > len(input)*2 {
			t.Errorf("sanitizeText result grew too much: input %d bytes, output %d bytes",
				len(input), len(result))
		}
	})
}

// FuzzFullSanitization tests the complete sanitization pipeline
func FuzzFullSanitization(f *testing.F) {
	// Create a test book with various content types
	f.Add("http://example.com", "Test content", "Test title")
	f.Add("javascript:alert('xss')", "Content with <script>", "Title")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, url, content, title string) {
		// Create a test book
		book := &models.Book{
			Title:       title,
			Description: content,
			Content: &models.Content{
				Document: &models.Document{
					Body: models.Body{
						Content: []models.StructuralElement{
							{
								Paragraph: &models.Paragraph{
									Elements: []models.ParagraphElement{
										{
											TextRun: &models.TextRun{
												Content: content,
												TextStyle: models.TextStyle{
													Link: &models.Link{
														URL: &url,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		s := NewSanitizer()
		result := s.Sanitize(book)

		// Verify sanitization doesn't crash and produces valid output
		if result == nil {
			t.Error("Sanitize returned nil")
			return
		}

		if result.Book == nil {
			t.Error("Sanitized book is nil")
			return
		}

		// Verify the original book wasn't modified (deep copy worked)
		if book.Title != title {
			t.Error("Original book was modified during sanitization")
		}

		// Verify warnings are reasonable
		if len(result.Warnings) > 100 {
			t.Errorf("Too many warnings generated: %d", len(result.Warnings))
		}

		// Verify sanitized content doesn't contain dangerous patterns
		if result.Book.Content != nil &&
			result.Book.Content.Document != nil &&
			len(result.Book.Content.Document.Body.Content) > 0 {

			para := result.Book.Content.Document.Body.Content[0].Paragraph
			if para != nil && len(para.Elements) > 0 {
				textRun := para.Elements[0].TextRun
				if textRun != nil && textRun.TextStyle.Link != nil && textRun.TextStyle.Link.URL != nil {
					sanitizedURL := *textRun.TextStyle.Link.URL
					if strings.HasPrefix(strings.ToLower(sanitizedURL), "javascript:") {
						t.Errorf("Dangerous URL not sanitized: %q", sanitizedURL)
					}
				}
			}
		}
	})
}
