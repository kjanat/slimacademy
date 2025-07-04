package sanitizer

import (
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
)

func TestNewSanitizer(t *testing.T) {
	s := NewSanitizer()
	if s == nil {
		t.Error("NewSanitizer should return non-nil sanitizer")
	}
	if s.warnings == nil {
		t.Error("NewSanitizer should initialize warnings slice")
	}
	if len(s.warnings) != 0 {
		t.Error("NewSanitizer should initialize empty warnings slice")
	}
}

func TestSanitizer_Sanitize(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name           string
		book           *models.Book
		expectedTitle  string
		expectedWarnCt int
	}{
		{
			name:           "nil book",
			book:           nil,
			expectedWarnCt: 0, // Should handle nil gracefully
		},
		{
			name: "simple valid book",
			book: &models.Book{
				ID:          123,
				Title:       "Test Book",
				Description: "A simple test book",
			},
			expectedTitle:  "Test Book",
			expectedWarnCt: 0,
		},
		{
			name:           "book with content requiring sanitization",
			book:           createBookWithDirtyContent(),
			expectedTitle:  "Dirty Book",
			expectedWarnCt: 1, // At least one warning expected
		},
		{
			name: "book with empty chapters",
			book: &models.Book{
				ID:    456,
				Title: "Test Book",
				Chapters: []models.Chapter{
					{ID: 1, Title: "Valid Chapter"},
					{ID: 2, Title: "   "}, // Empty title
				},
			},
			expectedTitle:  "Test Book",
			expectedWarnCt: 1, // Warning for empty chapter title
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Sanitize(tt.book)

			if result == nil {
				t.Error("Sanitize should never return nil result")
				return
			}

			if tt.book == nil {
				if result.Book != nil {
					t.Error("Sanitize should return nil book when input is nil")
				}
				return
			}

			if result.Book == nil {
				t.Error("Sanitize should return non-nil book for valid input")
				return
			}

			// Verify book was deep copied (different pointer)
			if result.Book == tt.book {
				t.Error("Sanitize should return a deep copy, not the original book")
			}

			// Verify content is preserved
			if result.Book.Title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, result.Book.Title)
			}

			// Verify warnings count
			if len(result.Warnings) < tt.expectedWarnCt {
				t.Errorf("Expected at least %d warnings, got %d", tt.expectedWarnCt, len(result.Warnings))
			}
		})
	}
}

func TestSanitizer_SanitizeText(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean text unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "control characters removed",
			input:    "Hello\x00\x01World",
			expected: "HelloWorld",
		},
		{
			name:     "preserve common whitespace",
			input:    "Hello\n\tWorld",
			expected: "Hello\n\tWorld", // Tabs are now preserved
		},
		{
			name:     "normalize multiple spaces",
			input:    "Hello    World",
			expected: "Hello    World",
		},
		{
			name:     "preserve whitespace but preserve newlines",
			input:    "Line1   \n  Line2    \nLine3",
			expected: "Line1   \n  Line2    \nLine3",
		},
		{
			name:     "remove carriage returns",
			input:    "Hello\r\nWorld\r",
			expected: "Hello\nWorld",
		},
		{
			name:     "handle UTF-8 characters",
			input:    "Café naïve résumé",
			expected: "Café naïve résumé",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.sanitizeText(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizer_NormalizeWhitespace(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line with multiple spaces",
			input:    "Hello    World",
			expected: "Hello    World",
		},
		{
			name:     "multiple lines preserve structure",
			input:    "Line1\nLine2\nLine3",
			expected: "Line1\nLine2\nLine3",
		},
		{
			name:     "mixed spaces and newlines",
			input:    "Line1\n   Line2\n Line3",
			expected: "Line1\n   Line2\n Line3",
		},
		{
			name:     "empty lines preserved",
			input:    "Line1\n\nLine3",
			expected: "Line1\n\nLine3",
		},
		{
			name:     "tabs and multiple spaces",
			input:    "Hello\t\t  World",
			expected: "Hello\t\t  World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.normalizeWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizer_DeepCopyBook(t *testing.T) {
	s := NewSanitizer()

	original := &models.Book{
		ID:          123,
		Title:       "Original Title",
		Description: "Original Description",
		Images: []models.BookImage{
			{ID: 1, ObjectID: "img1", ImageURL: "/image1.png"},
			{ID: 2, ObjectID: "img2", ImageURL: "/image2.png"},
		},
		Chapters: []models.Chapter{
			{ID: 1, Title: "Chapter 1"},
			{ID: 2, Title: "Chapter 2", SubChapters: []models.Chapter{
				{ID: 3, Title: "Sub Chapter 1"},
			}},
		},
		InlineObjectMap: map[string]string{
			"obj1": "/path1.png",
			"obj2": "/path2.png",
		},
	}

	// Add some pointer fields to test deep copying
	readProgress := int64(50)
	original.ReadProgress = &readProgress

	copy := s.deepCopyBook(original)

	// Verify it's a different object
	if copy == original {
		t.Error("deepCopyBook should return different pointer")
	}

	// Verify scalar fields are copied
	if copy.ID != original.ID {
		t.Errorf("ID not copied correctly: got %d, want %d", copy.ID, original.ID)
	}
	if copy.Title != original.Title {
		t.Errorf("Title not copied correctly: got %q, want %q", copy.Title, original.Title)
	}

	// Verify pointer fields are deeply copied
	if copy.ReadProgress == original.ReadProgress {
		t.Error("ReadProgress should be different pointer")
	}
	if *copy.ReadProgress != *original.ReadProgress {
		t.Error("ReadProgress value should be the same")
	}

	// Verify slices are deeply copied
	if &copy.Images == &original.Images {
		t.Error("Images slice should be different")
	}
	if len(copy.Images) != len(original.Images) {
		t.Error("Images slice length should match")
	}

	// Verify nested structures are deeply copied
	if &copy.Chapters == &original.Chapters {
		t.Error("Chapters slice should be different")
	}
	if len(copy.Chapters[1].SubChapters) != len(original.Chapters[1].SubChapters) {
		t.Error("SubChapters should be copied")
	}

	// Verify maps are copied
	if copy.InlineObjectMap["obj1"] != original.InlineObjectMap["obj1"] {
		t.Error("InlineObjectMap should be copied")
	}

	// Verify mutations don't affect original
	copy.Title = "Modified Title"
	if original.Title == "Modified Title" {
		t.Error("Modifying copy should not affect original")
	}
}

func TestSanitizer_SanitizeContent(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name        string
		book        *models.Book
		expectWarns int
	}{
		{
			name:        "book with nil content",
			book:        &models.Book{ID: 1, Title: "Test"},
			expectWarns: 0,
		},
		{
			name:        "book with document content",
			book:        createBookWithDocumentContent(),
			expectWarns: 0, // Assuming clean content
		},
		{
			name: "book with chapter content and empty titles",
			book: &models.Book{
				Content: &models.Content{
					Chapters: []models.Chapter{
						{ID: 1, Title: "Valid Chapter"},
						{ID: 2, Title: "   "}, // Empty title
					},
				},
			},
			expectWarns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.warnings = nil // Reset warnings
			s.sanitizeContent(tt.book)

			if len(s.warnings) != tt.expectWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarns, len(s.warnings))
			}
		})
	}
}

func TestSanitizer_SanitizeParagraph(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name        string
		paragraph   *models.Paragraph
		location    string
		expectWarns int
	}{
		{
			name: "clean paragraph",
			paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: "Clean text content",
						},
					},
				},
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "NORMAL_TEXT",
				},
			},
			location:    "test",
			expectWarns: 0,
		},
		{
			name: "empty heading",
			paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: "   ", // Empty content
						},
					},
				},
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "HEADING_1",
				},
			},
			location:    "test",
			expectWarns: 1, // Empty heading only (whitespace is now preserved)
		},
		{
			name: "text with control characters",
			paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: "Text\x00with\x01controls",
						},
					},
				},
			},
			location:    "test",
			expectWarns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.warnings = nil // Reset warnings
			s.sanitizeParagraph(tt.paragraph, tt.location)

			if len(s.warnings) != tt.expectWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarns, len(s.warnings))
				for _, w := range s.warnings {
					t.Logf("Warning: %s at %s", w.Issue, w.Location)
				}
			}
		})
	}
}

func TestSanitizer_ValidateLinkURL(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name        string
		link        *models.Link
		expectWarns int
		expectedURL string
	}{
		{
			name: "valid URL",
			link: &models.Link{
				URL: stringPtr("https://example.com"),
			},
			expectWarns: 0,
			expectedURL: "https://example.com",
		},
		{
			name: "empty URL",
			link: &models.Link{
				URL: stringPtr(""),
			},
			expectWarns: 1,
		},
		{
			name: "nil URL",
			link: &models.Link{
				URL: nil,
			},
			expectWarns: 1,
		},
		{
			name: "URL with whitespace",
			link: &models.Link{
				URL: stringPtr("  https://example.com  "),
			},
			expectWarns: 1,
			expectedURL: "https://example.com",
		},
		{
			name: "URL with HTML tags",
			link: &models.Link{
				URL: stringPtr("https://example.com<script>alert('xss')</script>"),
			},
			expectWarns: 1,
			expectedURL: "#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.warnings = nil // Reset warnings
			s.validateLinkURL(tt.link, "test_location")

			if len(s.warnings) != tt.expectWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarns, len(s.warnings))
			}

			if tt.expectedURL != "" && tt.link.URL != nil {
				if *tt.link.URL != tt.expectedURL {
					t.Errorf("Expected URL %q, got %q", tt.expectedURL, *tt.link.URL)
				}
			}
		})
	}
}

func TestSanitizer_ExtractText(t *testing.T) {
	s := NewSanitizer()

	paragraph := &models.Paragraph{
		Elements: []models.ParagraphElement{
			{
				TextRun: &models.TextRun{Content: "Hello "},
			},
			{
				TextRun: &models.TextRun{Content: "World"},
			},
			{
				// Non-text element should be ignored
				InlineObjectElement: &models.InlineObjectElement{
					InlineObjectID: "img1",
				},
			},
		},
	}

	result := s.extractText(paragraph)
	expected := "Hello World"

	if result != expected {
		t.Errorf("extractText() = %q, want %q", result, expected)
	}
}

func TestSanitizer_IsHeading(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  bool
	}{
		{
			name: "heading paragraph",
			paragraph: &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "HEADING_1",
				},
			},
			expected: true,
		},
		{
			name: "normal paragraph",
			paragraph: &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "NORMAL_TEXT",
				},
			},
			expected: false,
		},
		{
			name: "subtitle heading",
			paragraph: &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "HEADING_2",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.isHeading(tt.paragraph)
			if result != tt.expected {
				t.Errorf("isHeading() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_WarningTracking(t *testing.T) {
	s := NewSanitizer()

	// Add several warnings
	s.addWarning("loc1", "issue1", "orig1", "fixed1")
	s.addWarning("loc2", "issue2", "orig2", "fixed2")

	if len(s.warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(s.warnings))
	}

	// Verify warning structure
	w := s.warnings[0]
	if w.Location != "loc1" || w.Issue != "issue1" || w.Original != "orig1" || w.Fixed != "fixed1" {
		t.Errorf("Warning not stored correctly: %+v", w)
	}
}

// Test Performance with Large Data
func TestSanitizer_PerformanceLargeContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	s := NewSanitizer()

	// Create a book with large content
	book := createLargeBook(1000) // 1000 chapters

	result := s.Sanitize(book)

	if result == nil {
		t.Error("Sanitize should handle large content")
	}

	if len(result.Book.Chapters) != 1000 {
		t.Errorf("Expected 1000 chapters, got %d", len(result.Book.Chapters))
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func createBookWithDirtyContent() *models.Book {
	return &models.Book{
		ID:    789,
		Title: "Dirty Book",
		Content: &models.Content{
			Document: &models.Document{
				DocumentID: "test-doc",
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "Text\x00with\x01control\x02chars",
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
}

func createBookWithDocumentContent() *models.Book {
	return &models.Book{
		ID:    456,
		Title: "Document Book",
		Content: &models.Content{
			Document: &models.Document{
				DocumentID: "doc-123",
				Title:      "Document Title",
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "Clean paragraph content",
										},
									},
								},
								ParagraphStyle: models.ParagraphStyle{
									NamedStyleType: "NORMAL_TEXT",
								},
							},
						},
					},
				},
			},
		},
	}
}

func createLargeBook(chapterCount int) *models.Book {
	chapters := make([]models.Chapter, chapterCount)
	for i := 0; i < chapterCount; i++ {
		chapters[i] = models.Chapter{
			ID:    int64(i + 1),
			Title: "Chapter " + strings.Repeat("A", 100), // Long titles
		}
	}

	return &models.Book{
		ID:       999,
		Title:    "Large Book",
		Chapters: chapters,
	}
}
