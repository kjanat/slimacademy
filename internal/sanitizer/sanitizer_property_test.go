package sanitizer

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/kjanat/slimacademy/internal/models"
)

// Property-based tests for the sanitizer

func TestSanitizer_TextSanitizationProperties(t *testing.T) {
	s := NewSanitizer()

	// Property: Sanitized text should always be valid UTF-8
	t.Run("sanitized_text_is_valid_utf8", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := generateRandomText(rand.Intn(1000) + 1)
			output := s.sanitizeText(input)

			if !utf8.ValidString(output) {
				t.Errorf("Sanitized text is not valid UTF-8: input=%q, output=%q", input, output)
			}
		}
	})

	// Property: Sanitization should be idempotent
	t.Run("sanitization_is_idempotent", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := generateRandomText(rand.Intn(500) + 1)
			first := s.sanitizeText(input)
			second := s.sanitizeText(first)

			if first != second {
				t.Errorf("Sanitization not idempotent: first=%q, second=%q", first, second)
			}
		}
	})

	// Property: No control characters in output (except allowed whitespace)
	t.Run("no_control_characters_in_output", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			input := generateTextWithControlChars(rand.Intn(200) + 1)
			output := s.sanitizeText(input)

			for _, r := range output {
				if unicode.IsControl(r) && r != '\n' && r != '\t' && r != ' ' {
					t.Errorf("Control character found in output: %U in %q", r, output)
				}
			}
		}
	})

	// Property: Consecutive spaces are normalized to single spaces within lines
	t.Run("consecutive_spaces_normalized", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			input := generateTextWithExcessiveSpaces(rand.Intn(100) + 1)
			output := s.sanitizeText(input)

			// Check each line individually
			lines := strings.Split(output, "\n")
			for lineNum, line := range lines {
				if strings.Contains(line, "  ") { // Two or more consecutive spaces
					t.Errorf("Found consecutive spaces in line %d: %q (full output: %q)",
						lineNum, line, output)
				}
			}
		}
	})

	// Property: Newlines are preserved during normalization
	t.Run("newlines_preserved", func(t *testing.T) {
		inputs := []string{
			"Line1\nLine2\nLine3",
			"Line1\n\nLine3", // Empty line
			"\nStart with newline",
			"End with newline\n",
			"Multiple\n\n\nNewlines",
		}

		for _, input := range inputs {
			output := s.sanitizeText(input)
			inputNewlines := strings.Count(input, "\n")
			outputNewlines := strings.Count(output, "\n")

			if inputNewlines != outputNewlines {
				t.Errorf("Newline count changed: input had %d, output has %d (input=%q, output=%q)",
					inputNewlines, outputNewlines, input, output)
			}
		}
	})
}

func TestSanitizer_BookCopyProperties(t *testing.T) {
	s := NewSanitizer()

	// Property: Deep copy should preserve all scalar values
	t.Run("deep_copy_preserves_scalars", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			original := generateRandomBook()
			copy := s.deepCopyBook(original)

			if copy.ID != original.ID {
				t.Errorf("ID not preserved: original=%d, copy=%d", original.ID, copy.ID)
			}
			if copy.Title != original.Title {
				t.Errorf("Title not preserved: original=%q, copy=%q", original.Title, copy.Title)
			}
			if copy.CollegeStartYear != original.CollegeStartYear {
				t.Errorf("CollegeStartYear not preserved: original=%d, copy=%d",
					original.CollegeStartYear, copy.CollegeStartYear)
			}
		}
	})

	// Property: Deep copy should create independent objects
	t.Run("deep_copy_creates_independent_objects", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			original := generateRandomBook()
			copy := s.deepCopyBook(original)

			// Modify copy and ensure original is unchanged
			originalTitle := original.Title
			copy.Title = "Modified Title"

			if original.Title != originalTitle {
				t.Error("Modifying copy affected original title")
			}

			// Test slice independence
			if len(copy.Images) > 0 {
				originalImageURL := ""
				if len(original.Images) > 0 {
					originalImageURL = original.Images[0].ImageURL
				}

				copy.Images[0].ImageURL = "modified-url.png"

				if len(original.Images) > 0 && original.Images[0].ImageURL != originalImageURL {
					t.Error("Modifying copy affected original image URL")
				}
			}
		}
	})

	// Property: Deep copy should handle nil values correctly
	t.Run("deep_copy_handles_nil_values", func(t *testing.T) {
		// Test nil book
		nilCopy := s.deepCopyBook(nil)
		if nilCopy != nil {
			t.Error("Deep copy of nil should return nil")
		}

		// Test book with nil fields
		book := &models.Book{
			ID:           123,
			Title:        "Test",
			Content:      nil,
			ReadProgress: nil,
		}

		copy := s.deepCopyBook(book)
		if copy.Content != nil {
			t.Error("Nil content should remain nil")
		}
		if copy.ReadProgress != nil {
			t.Error("Nil read progress should remain nil")
		}
	})
}

func TestSanitizer_WarningProperties(t *testing.T) {
	// Property: Warnings should accumulate correctly
	t.Run("warnings_accumulate", func(t *testing.T) {
		s := NewSanitizer()

		expectedCount := rand.Intn(10) + 1
		for i := 0; i < expectedCount; i++ {
			s.addWarning("loc", "issue", "orig", "fixed")
		}

		if len(s.warnings) != expectedCount {
			t.Errorf("Expected %d warnings, got %d", expectedCount, len(s.warnings))
		}
	})

	// Property: Each warning should have all required fields
	t.Run("warnings_have_required_fields", func(t *testing.T) {
		s := NewSanitizer()

		s.addWarning("test_location", "test_issue", "original_text", "fixed_text")

		if len(s.warnings) != 1 {
			t.Fatal("Expected exactly one warning")
		}

		w := s.warnings[0]
		if w.Location == "" || w.Issue == "" || w.Original == "" || w.Fixed == "" {
			t.Errorf("Warning missing required fields: %+v", w)
		}
	})

	// Property: Sanitize should reset warnings
	t.Run("sanitize_resets_warnings", func(t *testing.T) {
		s := NewSanitizer()

		// Add some warnings
		s.addWarning("loc1", "issue1", "orig1", "fixed1")
		s.addWarning("loc2", "issue2", "orig2", "fixed2")

		if len(s.warnings) != 2 {
			t.Logf("Actual warnings: %+v", s.warnings)
			t.Fatalf("Expected 2 warnings before sanitize, got %d", len(s.warnings))
		}

		// Sanitize should reset warnings and start fresh
		book := &models.Book{ID: 123, Title: "Test"}
		result := s.Sanitize(book)

		// The result should contain warnings from this sanitization only (0 in this case)
		// not the 2 warnings we added manually before
		if len(result.Warnings) != 0 {
			t.Errorf("Expected 0 warnings from clean book, got %d", len(result.Warnings))
		}

		// Verify internal warnings were reset
		if len(s.warnings) != 0 {
			t.Errorf("Internal warnings should be reset after sanitize, got %d", len(s.warnings))
		}
	})
}

func TestSanitizer_StructuralProperties(t *testing.T) {
	s := NewSanitizer()

	// Property: Sanitization preserves document structure
	t.Run("preserves_document_structure", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			book := generateBookWithRandomContent()
			result := s.Sanitize(book)

			// Check that major structures are preserved
			if book.Content != nil && result.Book.Content == nil {
				t.Error("Content structure was lost during sanitization")
			}

			if book.Content != nil && book.Content.Document != nil {
				if result.Book.Content.Document == nil {
					t.Error("Document structure was lost during sanitization")
				} else {
					originalElementCount := len(book.Content.Document.Body.Content)
					sanitizedElementCount := len(result.Book.Content.Document.Body.Content)

					if originalElementCount != sanitizedElementCount {
						t.Errorf("Element count changed: original=%d, sanitized=%d",
							originalElementCount, sanitizedElementCount)
					}
				}
			}
		}
	})

	// Property: Heading detection should be consistent
	t.Run("heading_detection_consistent", func(t *testing.T) {
		headingTypes := []string{"HEADING_1", "HEADING_2", "HEADING_3", "HEADING_4", "HEADING_5", "HEADING_6"}
		nonHeadingTypes := []string{"NORMAL_TEXT", "SUBTITLE", "TITLE"}

		for _, headingType := range headingTypes {
			p := &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: headingType,
				},
			}

			if !s.isHeading(p) {
				t.Errorf("Should detect %s as heading", headingType)
			}
		}

		for _, nonHeadingType := range nonHeadingTypes {
			p := &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: nonHeadingType,
				},
			}

			if s.isHeading(p) {
				t.Errorf("Should not detect %s as heading", nonHeadingType)
			}
		}
	})
}

func TestSanitizer_LinkValidationProperties(t *testing.T) {
	s := NewSanitizer()

	// Property: Valid URLs should pass through unchanged
	t.Run("valid_urls_unchanged", func(t *testing.T) {
		validURLs := []string{
			"https://example.com",
			"http://example.com/path",
			"https://subdomain.example.com/path?query=value",
			"mailto:test@example.com",
			"#internal-link",
			"/relative/path",
		}

		for _, url := range validURLs {
			link := &models.Link{URL: &url}
			original := *link.URL

			s.warnings = nil
			s.validateLinkURL(link, "test")

			if *link.URL != original {
				t.Errorf("Valid URL was modified: %q -> %q", original, *link.URL)
			}
			if len(s.warnings) > 0 {
				t.Errorf("Valid URL generated warnings: %q", original)
			}
		}
	})

	// Property: URLs with HTML should be blocked for security (enhanced security)
	t.Run("html_in_urls_cleaned", func(t *testing.T) {
		htmlURLs := map[string]string{
			"https://example.com<script>":         "#",
			"<a href='evil'>https://example.com":  "#",
			"https://example.com<img src='evil'>": "#",
		}

		for input, expected := range htmlURLs {
			link := &models.Link{URL: &input}

			s.warnings = nil
			s.validateLinkURL(link, "test")

			if *link.URL != expected {
				t.Errorf("HTML cleaning failed: input=%q, expected=%q, got=%q",
					input, expected, *link.URL)
			}
			if len(s.warnings) == 0 {
				t.Errorf("HTML in URL should generate warning: %q", input)
			}
		}
	})
}

// Helper functions for property-based testing

func generateRandomText(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 \n\t"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func generateTextWithControlChars(length int) string {
	result := make([]rune, length)
	for i := range result {
		switch rand.Intn(4) {
		case 0:
			result[i] = rune(rand.Intn(32)) // Control characters 0-31
		case 1:
			result[i] = rune(127 + rand.Intn(32)) // Control characters 127-159
		default:
			result[i] = rune(32 + rand.Intn(95)) // Printable ASCII
		}
	}
	return string(result)
}

func generateTextWithExcessiveSpaces(length int) string {
	var result strings.Builder
	for i := 0; i < length; i++ {
		switch rand.Intn(5) {
		case 0:
			// Add multiple spaces
			spaces := rand.Intn(5) + 2
			result.WriteString(strings.Repeat(" ", spaces))
		case 1:
			// Add newline
			result.WriteString("\n")
		default:
			// Add regular character
			result.WriteString("a")
		}
	}
	return result.String()
}

func generateRandomBook() *models.Book {
	book := &models.Book{
		ID:               int64(rand.Intn(10000)),
		Title:            "Random Book " + randomString(10),
		Description:      "Description " + randomString(20),
		CollegeStartYear: int64(2020 + rand.Intn(10)),
	}

	// Add some random images
	imageCount := rand.Intn(5)
	book.Images = make([]models.BookImage, imageCount)
	for i := range book.Images {
		book.Images[i] = models.BookImage{
			ID:       int64(i + 1),
			ObjectID: "img" + randomString(5),
			ImageURL: "/images/" + randomString(10) + ".png",
		}
	}

	// Add random chapters
	chapterCount := rand.Intn(10) + 1
	book.Chapters = make([]models.Chapter, chapterCount)
	for i := range book.Chapters {
		book.Chapters[i] = models.Chapter{
			ID:    int64(i + 1),
			Title: "Chapter " + randomString(8),
		}
	}

	return book
}

func generateBookWithRandomContent() *models.Book {
	book := generateRandomBook()

	// Add content with random structure
	if rand.Intn(2) == 0 {
		// Document content
		book.Content = &models.Content{
			Document: &models.Document{
				DocumentID: "doc-" + randomString(8),
				Body: models.Body{
					Content: generateRandomStructuralElements(rand.Intn(10) + 1),
				},
			},
		}
	} else {
		// Chapter content
		chapterCount := rand.Intn(5) + 1
		chapters := make([]models.Chapter, chapterCount)
		for i := range chapters {
			chapters[i] = models.Chapter{
				ID:    int64(i + 1),
				Title: "Content Chapter " + randomString(5),
			}
		}
		book.Content = &models.Content{
			Chapters: chapters,
		}
	}

	return book
}

func generateRandomStructuralElements(count int) []models.StructuralElement {
	elements := make([]models.StructuralElement, count)
	for i := range elements {
		elements[i] = models.StructuralElement{
			Paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: "Random content " + randomString(20),
						},
					},
				},
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: randomParagraphStyle(),
				},
			},
		}
	}
	return elements
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func randomParagraphStyle() string {
	styles := []string{
		"NORMAL_TEXT", "HEADING_1", "HEADING_2", "HEADING_3",
		"SUBTITLE", "TITLE",
	}
	return styles[rand.Intn(len(styles))]
}
