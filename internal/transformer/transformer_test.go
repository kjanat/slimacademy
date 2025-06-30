package transformer

import (
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
	testutils "github.com/kjanat/slimacademy/test/utils"
)

func TestTransformer_Transform(t *testing.T) {
	transformer := NewTransformer()

	// Create a test book with content
	book := &models.Book{
		ID:    123,
		Title: "Test Book",
		Content: models.Content{
			DocumentID: "test-doc",
			Body: models.Body{
				Content: []models.ContentElement{
					{
						EndIndex:   10,
						StartIndex: intPtr(1),
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									EndIndex:   10,
									StartIndex: 1,
									TextRun: &models.TextRun{
										Content: "  Test content  \r\n",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Chapters: []models.Chapter{
			{
				ID:             1,
				Title:          "Test Chapter",
				GDocsChapterID: "h.test",
			},
		},
	}

	transformedBook, err := transformer.Transform(book)
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
	}

	if transformedBook == nil {
		t.Fatalf("Expected transformed book to be non-nil")
	}

	// Verify that text was cleaned (whitespace trimmed, \r removed)
	textRun := transformedBook.Content.Body.Content[0].Paragraph.Elements[0].TextRun
	if textRun.Content != "Test content" {
		t.Errorf("Expected cleaned text to be 'Test content', got '%s'", textRun.Content)
	}
}

func TestTransformer_ProcessContent(t *testing.T) {
	transformer := NewTransformer()

	book := &models.Book{
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "\r\n  Whitespace test  \r\n",
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "Another\rtext\n",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	transformer.processContent(book)

	elements := book.Content.Body.Content[0].Paragraph.Elements

	// Check first element
	if elements[0].TextRun.Content != "Whitespace test" {
		t.Errorf("Expected first element to be 'Whitespace test', got '%s'", elements[0].TextRun.Content)
	}

	// Check second element (cleanText removes \r and trims, so \n at end is removed)
	if elements[1].TextRun.Content != "Anothertext" {
		t.Errorf("Expected second element to be 'Anothertext', got '%s'", elements[1].TextRun.Content)
	}
}

func TestTransformer_CleanText(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove carriage returns",
			input:    "test\r\ncontent",
			expected: "test\ncontent",
		},
		{
			name:     "trim whitespace",
			input:    "  \t  test content  \t  ",
			expected: "test content",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \r\n\t   ",
			expected: "",
		},
		{
			name:     "no changes needed",
			input:    "clean text",
			expected: "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestTransformer_BuildChapterMapping(t *testing.T) {
	transformer := NewTransformer()

	book := &models.Book{
		Chapters: []models.Chapter{
			{
				ID:             1,
				Title:          "Chapter 1",
				GDocsChapterID: "h.chapter1",
				SubChapters: []models.Chapter{
					{
						ID:             2,
						Title:          "Sub Chapter 1.1",
						GDocsChapterID: "h.subchapter1_1",
					},
				},
			},
			{
				ID:             3,
				Title:          "Chapter 2",
				GDocsChapterID: "h.chapter2",
			},
		},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							ParagraphStyle: models.ParagraphStyle{
								HeadingID: stringPtr("h.chapter1"),
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							ParagraphStyle: models.ParagraphStyle{
								HeadingID: stringPtr("h.subchapter1_1"),
							},
						},
					},
				},
			},
		},
	}

	transformer.buildChapterMapping(book)

	// This test mainly ensures no panics occur during chapter mapping
	// The actual mapping logic is internal and harder to test directly
	// In a real implementation, you might expose the mapping for testing
}

func TestTransformer_GetPlainText(t *testing.T) {
	transformer := NewTransformer()

	book := &models.Book{
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "First paragraph. ",
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "More text.\n",
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Second paragraph.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	plainText := transformer.GetPlainText(book)
	expected := "First paragraph. More text.\nSecond paragraph."

	if plainText != expected {
		t.Errorf("GetPlainText() = '%s', expected '%s'", plainText, expected)
	}
}

func TestTransformer_GetPlainTextWithNonTextElements(t *testing.T) {
	transformer := NewTransformer()

	book := &models.Book{
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Text before image ",
									},
								},
								{
									// InlineObjectElement (no text content)
									InlineObjectElement: &models.InlineObjectElement{
										InlineObjectID: "image-1",
									},
								},
								{
									TextRun: &models.TextRun{
										Content: " text after image.",
									},
								},
							},
						},
					},
					{
						// SectionBreak (no text content)
						SectionBreak: &models.SectionBreak{},
					},
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Another paragraph.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	plainText := transformer.GetPlainText(book)
	expected := "Text before image  text after image.Another paragraph."

	if plainText != expected {
		t.Errorf("GetPlainText() = '%s', expected '%s'", plainText, expected)
	}
}

func TestTransformer_GetChapterText(t *testing.T) {
	transformer := NewTransformer()

	book := &models.Book{
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							ParagraphStyle: models.ParagraphStyle{
								HeadingID: stringPtr("h.chapter1"),
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Chapter 1 Title\n",
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							ParagraphStyle: models.ParagraphStyle{
								HeadingID: stringPtr("h.chapter2"),
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Chapter 2 Title\n",
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Regular paragraph without heading ID.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	chapterText := transformer.GetChapterText(book, "h.chapter1")
	expected := "Chapter 1 Title\n"

	if chapterText != expected {
		t.Errorf("GetChapterText() = '%s', expected '%s'", chapterText, expected)
	}

	// Test non-existent chapter
	nonExistentText := transformer.GetChapterText(book, "h.nonexistent")
	if nonExistentText != "" {
		t.Errorf("Expected empty string for non-existent chapter, got '%s'", nonExistentText)
	}
}

func TestTransformer_WithRealTestData(t *testing.T) {
	transformer := NewTransformer()

	// Load our test book fixture
	book := testutils.LoadTestBook(t, "simple_book")

	transformedBook, err := transformer.Transform(book)
	if err != nil {
		t.Fatalf("Transform() failed with real data: %v", err)
	}

	// Verify basic structure is preserved
	if transformedBook.ID != book.ID {
		t.Errorf("Expected ID to be preserved, got %d instead of %d", transformedBook.ID, book.ID)
	}

	if transformedBook.Title != book.Title {
		t.Errorf("Expected title to be preserved, got '%s' instead of '%s'", transformedBook.Title, book.Title)
	}

	// Test plain text extraction
	plainText := transformer.GetPlainText(transformedBook)
	if len(plainText) == 0 {
		t.Errorf("Expected plain text to be extracted from real data")
	}

	// Verify content contains expected text from our fixture
	if !strings.Contains(plainText, "Introduction") {
		t.Errorf("Expected plain text to contain 'Introduction' from test fixture")
	}

	if !strings.Contains(plainText, "This is a simple test paragraph") {
		t.Errorf("Expected plain text to contain test paragraph content")
	}
}

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
