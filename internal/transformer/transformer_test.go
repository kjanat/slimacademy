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

func TestTransformer_ConsolidateTextRuns(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  []models.Element
	}{
		{
			name: "consolidate fragmented bold text - D-dimeer case",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Text before ",
							TextStyle: models.TextStyle{
								Bold: nil,
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "D-",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "d",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
								// Missing font properties - should inherit from adjacent
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "imeer",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: " after.",
							TextStyle: models.TextStyle{
								Bold: nil,
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Text before ",
						TextStyle: models.TextStyle{
							Bold: nil,
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "D-dimeer",
						TextStyle: models.TextStyle{
							Bold:     boolPtr(true),
							FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
							WeightedFontFamily: &models.WeightedFontFamily{
								FontFamily: "Open Sans",
								Weight:     400,
							},
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: " after.",
						TextStyle: models.TextStyle{
							Bold: nil,
						},
					},
				},
			},
		},
		{
			name: "consolidate β1-receptoren pattern",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "β",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "1",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "-receptoren",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: " bevinden zich",
							TextStyle: models.TextStyle{
								Bold: nil,
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "β1-receptoren",
						TextStyle: models.TextStyle{
							Bold: boolPtr(true),
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: " bevinden zich",
						TextStyle: models.TextStyle{
							Bold: nil,
						},
					},
				},
			},
		},
		{
			name: "do not consolidate different formatting",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "normal",
							TextStyle: models.TextStyle{
								Bold: nil,
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "bold",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "italic",
							TextStyle: models.TextStyle{
								Italic: boolPtr(true),
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "normal",
						TextStyle: models.TextStyle{
							Bold: nil,
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "bold",
						TextStyle: models.TextStyle{
							Bold: boolPtr(true),
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "italic",
						TextStyle: models.TextStyle{
							Italic: boolPtr(true),
						},
					},
				},
			},
		},
		{
			name: "consolidate same links",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Click ",
							TextStyle: models.TextStyle{
								Link: &models.Link{URL: "https://example.com"},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "here",
							TextStyle: models.TextStyle{
								Link: &models.Link{URL: "https://example.com"},
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Click here",
						TextStyle: models.TextStyle{
							Link: &models.Link{URL: "https://example.com"},
						},
					},
				},
			},
		},
		{
			name: "do not consolidate different links",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Link1",
							TextStyle: models.TextStyle{
								Link: &models.Link{URL: "https://example.com"},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "Link2",
							TextStyle: models.TextStyle{
								Link: &models.Link{URL: "https://different.com"},
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Link1",
						TextStyle: models.TextStyle{
							Link: &models.Link{URL: "https://example.com"},
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "Link2",
						TextStyle: models.TextStyle{
							Link: &models.Link{URL: "https://different.com"},
						},
					},
				},
			},
		},
		{
			name: "preserve non-TextRun elements",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Before",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "image",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "After",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "image",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Beforeimage",
						TextStyle: models.TextStyle{
							Bold: boolPtr(true),
						},
					},
				},
				{
					InlineObjectElement: &models.InlineObjectElement{
						InlineObjectID: "img1",
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "Afterimage",
						TextStyle: models.TextStyle{
							Bold: boolPtr(true),
						},
					},
				},
			},
		},
		{
			name: "consolidate mixed formatting with same properties",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Bold",
							TextStyle: models.TextStyle{
								Bold:   boolPtr(true),
								Italic: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "Italic",
							TextStyle: models.TextStyle{
								Bold:   boolPtr(true),
								Italic: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "Text",
							TextStyle: models.TextStyle{
								Bold:   boolPtr(true),
								Italic: boolPtr(true),
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "BoldItalicText",
						TextStyle: models.TextStyle{
							Bold:   boolPtr(true),
							Italic: boolPtr(true),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to avoid modifying the original
			paragraphCopy := *tt.paragraph
			elementsCopy := make([]models.Element, len(tt.paragraph.Elements))
			copy(elementsCopy, tt.paragraph.Elements)
			paragraphCopy.Elements = elementsCopy

			transformer.consolidateTextRuns(&paragraphCopy)

			if len(paragraphCopy.Elements) != len(tt.expected) {
				t.Errorf("Expected %d elements, got %d", len(tt.expected), len(paragraphCopy.Elements))
			}

			for i, expected := range tt.expected {
				if i >= len(paragraphCopy.Elements) {
					t.Errorf("Missing element at index %d", i)
					continue
				}

				actual := paragraphCopy.Elements[i]

				// Compare TextRun elements
				if expected.TextRun != nil {
					if actual.TextRun == nil {
						t.Errorf("Element %d: expected TextRun, got nil", i)
						continue
					}

					if actual.TextRun.Content != expected.TextRun.Content {
						t.Errorf("Element %d: expected content '%s', got '%s'",
							i, expected.TextRun.Content, actual.TextRun.Content)
					}

					// Compare bold formatting
					if expected.TextRun.TextStyle.Bold != nil && actual.TextRun.TextStyle.Bold != nil {
						if *expected.TextRun.TextStyle.Bold != *actual.TextRun.TextStyle.Bold {
							t.Errorf("Element %d: expected bold %v, got %v",
								i, *expected.TextRun.TextStyle.Bold, *actual.TextRun.TextStyle.Bold)
						}
					} else if expected.TextRun.TextStyle.Bold != actual.TextRun.TextStyle.Bold {
						t.Errorf("Element %d: bold mismatch - expected %v, got %v",
							i, expected.TextRun.TextStyle.Bold, actual.TextRun.TextStyle.Bold)
					}

					// Compare italic formatting
					if expected.TextRun.TextStyle.Italic != nil && actual.TextRun.TextStyle.Italic != nil {
						if *expected.TextRun.TextStyle.Italic != *actual.TextRun.TextStyle.Italic {
							t.Errorf("Element %d: expected italic %v, got %v",
								i, *expected.TextRun.TextStyle.Italic, *actual.TextRun.TextStyle.Italic)
						}
					} else if expected.TextRun.TextStyle.Italic != actual.TextRun.TextStyle.Italic {
						t.Errorf("Element %d: italic mismatch - expected %v, got %v",
							i, expected.TextRun.TextStyle.Italic, actual.TextRun.TextStyle.Italic)
					}

					// Compare links
					if expected.TextRun.TextStyle.Link != nil && actual.TextRun.TextStyle.Link != nil {
						if expected.TextRun.TextStyle.Link.URL != actual.TextRun.TextStyle.Link.URL {
							t.Errorf("Element %d: expected link URL '%s', got '%s'",
								i, expected.TextRun.TextStyle.Link.URL, actual.TextRun.TextStyle.Link.URL)
						}
					} else if expected.TextRun.TextStyle.Link != actual.TextRun.TextStyle.Link {
						t.Errorf("Element %d: link mismatch - expected %v, got %v",
							i, expected.TextRun.TextStyle.Link, actual.TextRun.TextStyle.Link)
					}
				}

				// Compare InlineObjectElement
				if expected.InlineObjectElement != nil {
					if actual.InlineObjectElement == nil {
						t.Errorf("Element %d: expected InlineObjectElement, got nil", i)
						continue
					}

					if expected.InlineObjectElement.InlineObjectID != actual.InlineObjectElement.InlineObjectID {
						t.Errorf("Element %d: expected InlineObjectID '%s', got '%s'",
							i, expected.InlineObjectElement.InlineObjectID, actual.InlineObjectElement.InlineObjectID)
					}
				}
			}
		})
	}
}

func TestTransformer_ConsolidateTextRuns_Integration(t *testing.T) {
	transformer := NewTransformer()

	// Test integration with the Transform method
	book := &models.Book{
		ID:    123,
		Title: "Test Book",
		Content: models.Content{
			DocumentID: "test-doc",
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "β",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "1",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "-receptoren",
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
	}

	transformedBook, err := transformer.Transform(book)
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
	}

	// After transformation, the fragmented TextRuns should be consolidated
	elements := transformedBook.Content.Body.Content[0].Paragraph.Elements
	if len(elements) != 1 {
		t.Errorf("Expected 1 consolidated element, got %d", len(elements))
	}

	if elements[0].TextRun == nil {
		t.Fatalf("Expected TextRun element")
	}

	if elements[0].TextRun.Content != "β1-receptoren" {
		t.Errorf("Expected consolidated content 'β1-receptoren', got '%s'", elements[0].TextRun.Content)
	}

	if elements[0].TextRun.TextStyle.Bold == nil || !*elements[0].TextRun.TextStyle.Bold {
		t.Errorf("Expected consolidated element to maintain bold formatting")
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

func TestTransformer_HasFormattingWithFontProperties(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name     string
		style    models.TextStyle
		expected bool
	}{
		{
			name: "FontSize alone should be considered formatting",
			style: models.TextStyle{
				FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
			},
			expected: true,
		},
		{
			name: "WeightedFontFamily alone should be considered formatting",
			style: models.TextStyle{
				WeightedFontFamily: &models.WeightedFontFamily{
					FontFamily: "Arial",
					Weight:     400,
				},
			},
			expected: true,
		},
		{
			name: "Both font properties should be considered formatting",
			style: models.TextStyle{
				FontSize: &models.FontSize{Magnitude: 14, Unit: "PT"},
				WeightedFontFamily: &models.WeightedFontFamily{
					FontFamily: "Open Sans",
					Weight:     600,
				},
			},
			expected: true,
		},
		{
			name: "Font properties combined with other formatting",
			style: models.TextStyle{
				Bold:     boolPtr(true),
				FontSize: &models.FontSize{Magnitude: 16, Unit: "PT"},
			},
			expected: true,
		},
		{
			name: "No formatting properties",
			style: models.TextStyle{
				Bold:   nil,
				Italic: nil,
			},
			expected: false,
		},
		{
			name: "Only bold formatting (existing behavior)",
			style: models.TextStyle{
				Bold: boolPtr(true),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.hasFormatting(&tt.style)
			if result != tt.expected {
				t.Errorf("hasFormatting() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTransformer_FontCompatibilityChecks(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name     string
		textRun1 *models.TextRun
		textRun2 *models.TextRun
		expected bool
		reason   string
	}{
		{
			name: "Different FontSize should prevent consolidation",
			textRun1: &models.TextRun{
				Content: "text1",
				TextStyle: models.TextStyle{
					Bold:     boolPtr(true),
					FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
				},
			},
			textRun2: &models.TextRun{
				Content: "text2",
				TextStyle: models.TextStyle{
					Bold:     boolPtr(true),
					FontSize: &models.FontSize{Magnitude: 14, Unit: "PT"},
				},
			},
			expected: false,
			reason:   "Different font sizes should not be compatible",
		},
		{
			name: "Different WeightedFontFamily should prevent consolidation",
			textRun1: &models.TextRun{
				Content: "text1",
				TextStyle: models.TextStyle{
					Bold: boolPtr(true),
					WeightedFontFamily: &models.WeightedFontFamily{
						FontFamily: "Arial",
						Weight:     400,
					},
				},
			},
			textRun2: &models.TextRun{
				Content: "text2",
				TextStyle: models.TextStyle{
					Bold: boolPtr(true),
					WeightedFontFamily: &models.WeightedFontFamily{
						FontFamily: "Open Sans",
						Weight:     400,
					},
				},
			},
			expected: false,
			reason:   "Different font families should not be compatible",
		},
		{
			name: "Same font properties should allow consolidation",
			textRun1: &models.TextRun{
				Content: "text1",
				TextStyle: models.TextStyle{
					Bold:     boolPtr(true),
					FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
					WeightedFontFamily: &models.WeightedFontFamily{
						FontFamily: "Open Sans",
						Weight:     400,
					},
				},
			},
			textRun2: &models.TextRun{
				Content: "text2",
				TextStyle: models.TextStyle{
					Bold:     boolPtr(true),
					FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
					WeightedFontFamily: &models.WeightedFontFamily{
						FontFamily: "Open Sans",
						Weight:     400,
					},
				},
			},
			expected: true,
			reason:   "Identical font properties should be compatible",
		},
		{
			name: "Nil font properties should be compatible",
			textRun1: &models.TextRun{
				Content: "text1",
				TextStyle: models.TextStyle{
					Bold: boolPtr(true),
				},
			},
			textRun2: &models.TextRun{
				Content: "text2",
				TextStyle: models.TextStyle{
					Bold: boolPtr(true),
				},
			},
			expected: true,
			reason:   "Both nil font properties should be compatible",
		},
		{
			name: "One nil, one non-nil font property should allow consolidation (inheritance)",
			textRun1: &models.TextRun{
				Content: "text1",
				TextStyle: models.TextStyle{
					Bold:     boolPtr(true),
					FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
				},
			},
			textRun2: &models.TextRun{
				Content: "text2",
				TextStyle: models.TextStyle{
					Bold: boolPtr(true),
					// No FontSize - should inherit from adjacent
				},
			},
			expected: true,
			reason:   "Nil font properties can inherit from non-nil properties",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformer.areTextRunsCompatible(tt.textRun1, tt.textRun2)
			if result != tt.expected {
				t.Errorf("areTextRunsCompatible() = %v, expected %v. Reason: %s", result, tt.expected, tt.reason)
			}
		})
	}
}

func TestTransformer_ConsolidateTextRunsWithFonts(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  []models.Element
	}{
		{
			name: "consolidate TextRuns with same font properties",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Same",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "Font",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "SameFont",
						TextStyle: models.TextStyle{
							FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
							WeightedFontFamily: &models.WeightedFontFamily{
								FontFamily: "Open Sans",
								Weight:     400,
							},
						},
					},
				},
			},
		},
		{
			name: "do not consolidate TextRuns with different font sizes",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Size12",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "Size14",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 14, Unit: "PT"},
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Size12",
						TextStyle: models.TextStyle{
							FontSize: &models.FontSize{Magnitude: 12, Unit: "PT"},
						},
					},
				},
				{
					TextRun: &models.TextRun{
						Content: "Size14",
						TextStyle: models.TextStyle{
							FontSize: &models.FontSize{Magnitude: 14, Unit: "PT"},
						},
					},
				},
			},
		},
		{
			name: "real-world font fragmentation pattern",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Term",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 11, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "inology",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 11, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: " in ",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 11, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "medicine",
							TextStyle: models.TextStyle{
								FontSize: &models.FontSize{Magnitude: 11, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
				},
			},
			expected: []models.Element{
				{
					TextRun: &models.TextRun{
						Content: "Terminology in medicine",
						TextStyle: models.TextStyle{
							FontSize: &models.FontSize{Magnitude: 11, Unit: "PT"},
							WeightedFontFamily: &models.WeightedFontFamily{
								FontFamily: "Open Sans",
								Weight:     400,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to avoid modifying the original
			paragraphCopy := *tt.paragraph
			elementsCopy := make([]models.Element, len(tt.paragraph.Elements))
			copy(elementsCopy, tt.paragraph.Elements)
			paragraphCopy.Elements = elementsCopy

			transformer.consolidateTextRuns(&paragraphCopy)

			if len(paragraphCopy.Elements) != len(tt.expected) {
				t.Errorf("Expected %d elements, got %d", len(tt.expected), len(paragraphCopy.Elements))
			}

			for i, expected := range tt.expected {
				if i >= len(paragraphCopy.Elements) {
					t.Errorf("Missing element at index %d", i)
					continue
				}

				actual := paragraphCopy.Elements[i]

				if expected.TextRun != nil {
					if actual.TextRun == nil {
						t.Errorf("Element %d: expected TextRun, got nil", i)
						continue
					}

					if actual.TextRun.Content != expected.TextRun.Content {
						t.Errorf("Element %d: expected content '%s', got '%s'",
							i, expected.TextRun.Content, actual.TextRun.Content)
					}

					// Check font size
					if expected.TextRun.TextStyle.FontSize != nil && actual.TextRun.TextStyle.FontSize != nil {
						if expected.TextRun.TextStyle.FontSize.Magnitude != actual.TextRun.TextStyle.FontSize.Magnitude {
							t.Errorf("Element %d: expected font size %v, got %v",
								i, expected.TextRun.TextStyle.FontSize.Magnitude, actual.TextRun.TextStyle.FontSize.Magnitude)
						}
					} else if expected.TextRun.TextStyle.FontSize != actual.TextRun.TextStyle.FontSize {
						t.Errorf("Element %d: font size mismatch - expected %v, got %v",
							i, expected.TextRun.TextStyle.FontSize, actual.TextRun.TextStyle.FontSize)
					}

					// Check font family
					if expected.TextRun.TextStyle.WeightedFontFamily != nil && actual.TextRun.TextStyle.WeightedFontFamily != nil {
						if expected.TextRun.TextStyle.WeightedFontFamily.FontFamily != actual.TextRun.TextStyle.WeightedFontFamily.FontFamily {
							t.Errorf("Element %d: expected font family '%s', got '%s'",
								i, expected.TextRun.TextStyle.WeightedFontFamily.FontFamily, actual.TextRun.TextStyle.WeightedFontFamily.FontFamily)
						}
					} else if expected.TextRun.TextStyle.WeightedFontFamily != actual.TextRun.TextStyle.WeightedFontFamily {
						t.Errorf("Element %d: font family mismatch - expected %v, got %v",
							i, expected.TextRun.TextStyle.WeightedFontFamily, actual.TextRun.TextStyle.WeightedFontFamily)
					}
				}
			}
		})
	}
}

func TestTransformer_RealGoogleDocsPatterns(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name         string
		paragraph    *models.Paragraph
		expectedText string
		description  string
	}{
		{
			name: "real Google Docs fragmentation - beta1-receptor pattern",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "β",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "1",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "-receptoren ",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
				},
			},
			expectedText: "β1-receptoren ",
			description:  "Google Docs fragments identical formatting into separate TextRuns - should consolidate",
		},
		{
			name: "medical term fragmentation - alpha2-agonist pattern",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "α",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "2",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "-agonist",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
				},
			},
			expectedText: "α2-agonist",
			description:  "Medical terminology fragmentation should be consolidated",
		},
		{
			name: "chemical formula fragmentation - H2O pattern",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "H",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "2",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "O ",
							TextStyle: models.TextStyle{
								Bold:     boolPtr(true),
								FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
								WeightedFontFamily: &models.WeightedFontFamily{
									FontFamily: "Open Sans",
									Weight:     400,
								},
							},
						},
					},
				},
			},
			expectedText: "H2O ",
			description:  "Chemical formulas should consolidate when identically formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to avoid modifying the original
			paragraphCopy := *tt.paragraph
			elementsCopy := make([]models.Element, len(tt.paragraph.Elements))
			copy(elementsCopy, tt.paragraph.Elements)
			paragraphCopy.Elements = elementsCopy

			transformer.consolidateTextRuns(&paragraphCopy)

			// Extract the text content to verify spaces are preserved
			var actualText strings.Builder
			for _, element := range paragraphCopy.Elements {
				if element.TextRun != nil {
					actualText.WriteString(element.TextRun.Content)
				}
			}

			result := actualText.String()
			if result != tt.expectedText {
				t.Errorf("Space preservation failed for %s.\nExpected: %q\nActual:   %q\nDescription: %s",
					tt.name, tt.expectedText, result, tt.description)
			}
		})
	}
}

func TestTransformer_EndToEndWithRealGoogleDocsPattern(t *testing.T) {
	transformer := NewTransformer()

	// Real Google Docs fragmentation pattern based on actual JSON data
	book := &models.Book{
		ID:    1,
		Title: "Test Integration",
		Content: models.Content{
			DocumentID: "test-doc",
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "β",
										TextStyle: models.TextStyle{
											Bold:     boolPtr(true),
											FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
											WeightedFontFamily: &models.WeightedFontFamily{
												FontFamily: "Open Sans",
												Weight:     400,
											},
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "1",
										TextStyle: models.TextStyle{
											Bold:     boolPtr(true),
											FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
											WeightedFontFamily: &models.WeightedFontFamily{
												FontFamily: "Open Sans",
												Weight:     400,
											},
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "-receptoren ",
										TextStyle: models.TextStyle{
											Bold:     boolPtr(true),
											FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
											WeightedFontFamily: &models.WeightedFontFamily{
												FontFamily: "Open Sans",
												Weight:     400,
											},
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "bevinden zich in het hart",
										TextStyle: models.TextStyle{
											FontSize: &models.FontSize{Magnitude: 10, Unit: "PT"},
											WeightedFontFamily: &models.WeightedFontFamily{
												FontFamily: "Open Sans",
												Weight:     400,
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

	transformedBook, err := transformer.Transform(book)
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
	}

	// Extract all text content to verify the final result
	var finalText strings.Builder
	for _, element := range transformedBook.Content.Body.Content[0].Paragraph.Elements {
		if element.TextRun != nil {
			finalText.WriteString(element.TextRun.Content)
		}
	}

	result := finalText.String()

	// Expected: "β1-receptoren bevinden zich in het hart"
	// The first 3 TextRuns should consolidate to "β1-receptoren " (with trailing space)
	// Followed by the unformatted "bevinden zich in het hart"

	t.Logf("Final transformed text: %q", result)

	// The key test: consolidation should work and preserve spacing
	expectedResult := "β1-receptoren bevinden zich in het hart"
	if result != expectedResult {
		t.Errorf("Expected consolidated text: %q, got: %q", expectedResult, result)
	}

	// Verify structure: should be 2 elements after consolidation
	elements := transformedBook.Content.Body.Content[0].Paragraph.Elements
	t.Logf("Number of elements after transformation: %d", len(elements))
	for i, element := range elements {
		if element.TextRun != nil {
			hasFormatting := transformer.hasFormatting(&element.TextRun.TextStyle)
			t.Logf("Element %d: %q (formatted: %v)", i, element.TextRun.Content, hasFormatting)
		}
	}

	// Should have exactly 2 elements:
	// 1. "β1-receptoren " (bold, formatted)
	// 2. "bevinden zich in het hart" (unformatted)
	if len(elements) != 2 {
		t.Errorf("Expected 2 elements after consolidation, got %d", len(elements))
	}

	if len(elements) >= 1 && elements[0].TextRun != nil {
		if elements[0].TextRun.Content != "β1-receptoren " {
			t.Errorf("Expected first element to be 'β1-receptoren ', got %q", elements[0].TextRun.Content)
		}
		if elements[0].TextRun.TextStyle.Bold == nil || !*elements[0].TextRun.TextStyle.Bold {
			t.Errorf("Expected first element to be bold")
		}
	}

	if len(elements) >= 2 && elements[1].TextRun != nil {
		if elements[1].TextRun.Content != "bevinden zich in het hart" {
			t.Errorf("Expected second element to be 'bevinden zich in het hart', got %q", elements[1].TextRun.Content)
		}
		if elements[1].TextRun.TextStyle.Bold != nil && *elements[1].TextRun.TextStyle.Bold {
			t.Errorf("Expected second element to NOT be bold")
		}
	}
}
