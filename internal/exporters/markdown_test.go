package exporters

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
	testutils "github.com/kjanat/slimacademy/test/utils"
)

func TestMarkdownExporter_Export(t *testing.T) {
	exporter := NewMarkdownExporter()

	book := createTestBook()
	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "test.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	testutils.AssertFileExists(t, outputPath)

	content := testutils.ReadFileString(t, outputPath)

	// Verify basic structure
	testutils.AssertStringContains(t, content, "# Test Book")
	testutils.AssertStringContains(t, content, "Test book description")
	testutils.AssertStringContains(t, content, "## Table of Contents")
	testutils.AssertStringContains(t, content, "## Introduction")
	testutils.AssertStringContains(t, content, "**bold text**")
}

func TestMarkdownExporter_GetExtension(t *testing.T) {
	exporter := NewMarkdownExporter()

	ext := exporter.GetExtension()
	if ext != "md" {
		t.Errorf("Expected extension 'md', got '%s'", ext)
	}
}

func TestMarkdownExporter_GetName(t *testing.T) {
	exporter := NewMarkdownExporter()

	name := exporter.GetName()
	if name != "Markdown" {
		t.Errorf("Expected name 'Markdown', got '%s'", name)
	}
}

func TestMarkdownExporter_GenerateTableOfContents(t *testing.T) {
	exporter := &MarkdownExporter{}

	chapters := []models.Chapter{
		{
			ID:        1,
			Title:     "Introduction",
			IsVisible: 1,
		},
		{
			ID:        2,
			Title:     "Chapter 1",
			IsVisible: 1,
			SubChapters: []models.Chapter{
				{
					ID:              3,
					Title:           "Section 1.1",
					IsVisible:       1,
					ParentChapterID: intPtr(2),
				},
				{
					ID:              4,
					Title:           "Section 1.2",
					IsVisible:       0, // Hidden chapter
					ParentChapterID: intPtr(2),
				},
			},
		},
		{
			ID:        5,
			Title:     "Hidden Chapter",
			IsVisible: 0, // Should not appear in TOC
		},
	}

	var md strings.Builder
	exporter.generateTableOfContents(&md, chapters)

	toc := md.String()

	// Should contain visible chapters
	testutils.AssertStringContains(t, toc, "[Introduction](#introduction)")
	testutils.AssertStringContains(t, toc, "[Chapter 1](#chapter-1)")
	testutils.AssertStringContains(t, toc, "[Section 1.1](#section-1.1)")

	// Should not contain hidden chapters
	testutils.AssertStringNotContains(t, toc, "Hidden Chapter")
	testutils.AssertStringNotContains(t, toc, "Section 1.2")
}

func TestMarkdownExporter_FormatText(t *testing.T) {
	exporter := &MarkdownExporter{}

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  string
	}{
		{
			name: "plain text",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "Plain text content",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "Plain text content",
		},
		{
			name: "bold text",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Bold text",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "**Bold text**",
		},
		{
			name: "italic text",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Italic text",
							TextStyle: models.TextStyle{
								Italic: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "*Italic text*",
		},
		{
			name: "link",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Link text",
							TextStyle: models.TextStyle{
								Link: &models.Link{
									URL: "https://example.com",
								},
							},
						},
					},
				},
			},
			expected: "[Link text](https://example.com)",
		},
		{
			name: "combined formatting",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "Regular ",
							TextStyle: models.TextStyle{},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "bold ",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "and italic",
							TextStyle: models.TextStyle{
								Italic: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "Regular **bold ***and italic*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatText(tt.paragraph)
			if result != tt.expected {
				t.Errorf("formatText() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestMarkdownExporter_GetHeadingLevel(t *testing.T) {
	exporter := &MarkdownExporter{}

	tests := []struct {
		namedStyle string
		expected   int
	}{
		{"HEADING_1", 2},
		{"HEADING_2", 3},
		{"HEADING_3", 4},
		{"HEADING_4", 5},
		{"HEADING_5", 6},
		{"HEADING_6", 6},
		{"NORMAL_TEXT", 2}, // Default
		{"UNKNOWN", 2},     // Default
	}

	for _, tt := range tests {
		t.Run(tt.namedStyle, func(t *testing.T) {
			level := exporter.getHeadingLevel(tt.namedStyle)
			if level != tt.expected {
				t.Errorf("getHeadingLevel(%s) = %d, expected %d", tt.namedStyle, level, tt.expected)
			}
		})
	}
}

func TestMarkdownExporter_Slugify(t *testing.T) {
	exporter := &MarkdownExporter{}

	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "simple-title"},
		{"Title with: Special Characters!", "title-with-special-characters"},
		{"Multiple   Spaces", "multiple---spaces"},
		{"Title/With\\Slashes", "title-with-slashes"},
		{"Question? And & Ampersand", "question-and-and-ampersand"},
		{"(Parentheses) Test", "parentheses-test"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := exporter.slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownExporter_BuildChapterMap(t *testing.T) {
	exporter := &MarkdownExporter{}

	chapters := []models.Chapter{
		{
			ID:             1,
			Title:          "Chapter 1",
			GDocsChapterID: "h.chapter1",
			SubChapters: []models.Chapter{
				{
					ID:             2,
					Title:          "Sub Chapter",
					GDocsChapterID: "h.subchapter",
				},
			},
		},
	}

	chapterMap := exporter.buildChapterMap(chapters)

	if len(chapterMap) != 2 {
		t.Errorf("Expected chapter map to have 2 entries, got %d", len(chapterMap))
	}

	if chapter, exists := chapterMap["h.chapter1"]; !exists || chapter.Title != "Chapter 1" {
		t.Errorf("Expected to find chapter 'Chapter 1' in map")
	}

	if subChapter, exists := chapterMap["h.subchapter"]; !exists || subChapter.Title != "Sub Chapter" {
		t.Errorf("Expected to find sub-chapter 'Sub Chapter' in map")
	}
}

func TestMarkdownExporter_WithRealTestData(t *testing.T) {
	exporter := NewMarkdownExporter()

	// Load real test data
	book := testutils.LoadTestBook(t, "simple_book")

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "real_test.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed with real data: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)

	// Verify structure based on our test fixture
	testutils.AssertStringContains(t, content, "# Test Book")
	testutils.AssertStringContains(t, content, "## Table of Contents")
	testutils.AssertStringContains(t, content, "[Introduction](#introduction)")
	testutils.AssertStringContains(t, content, "## Introduction")
	testutils.AssertStringContains(t, content, "This is a simple test paragraph")
	testutils.AssertStringContains(t, content, "**bold text**")

	// Verify no empty sections
	testutils.AssertStringNotContains(t, content, "##  \n")   // Empty heading
	testutils.AssertStringNotContains(t, content, "\n\n\n\n") // Too many newlines
}

func TestMarkdownExporter_EmptyContent(t *testing.T) {
	exporter := NewMarkdownExporter()

	book := &models.Book{
		ID:          1,
		Title:       "Empty Book",
		Description: "",
		Chapters:    []models.Chapter{},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{},
			},
		},
	}

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "empty.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed with empty content: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)

	// Should still have basic structure
	testutils.AssertStringContains(t, content, "# Empty Book")
	testutils.AssertStringContains(t, content, "## Table of Contents")
}

func TestMarkdownExporter_WriteError(t *testing.T) {
	exporter := NewMarkdownExporter()
	book := createTestBook()

	// Try to write to an invalid path
	invalidPath := "/root/cannot_write_here.md"

	err := exporter.Export(book, invalidPath)
	if err == nil {
		t.Errorf("Expected error when writing to invalid path")
	}
}

func TestMarkdownExporter_BoldTextSpacing(t *testing.T) {
	exporter := &MarkdownExporter{}

	// Test the spacing issue from the real data where bold text lacks proper spacing
	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  string
	}{
		{
			name: "bold text with spacing before",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "vanaf",
							TextStyle: models.TextStyle{},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "19-06-2025",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content:   "zijn",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "vanaf **19-06-2025** zijn",
		},
		{
			name: "bold text with spacing after",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Viscerale pijn",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content:   "is",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "**Viscerale pijn** is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatText(tt.paragraph)
			if result != tt.expected {
				t.Errorf("formatText() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestMarkdownExporter_InlineImageRendering(t *testing.T) {
	exporter := &MarkdownExporter{}

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		book      *models.Book
		expected  string
	}{
		{
			name: "text with inline image",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "Before image: ",
							TextStyle: models.TextStyle{},
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
					{
						TextRun: &models.TextRun{
							Content:   " after image.",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			book: &models.Book{
				InlineObjectMap: map[string]string{
					"img1": "https://api.slimacademy.nl/image1.png",
				},
			},
			expected: "Before image: ![Image](https://api.slimacademy.nl/image1.png) after image.",
		},
		{
			name: "multiple inline images",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
					{
						TextRun: &models.TextRun{
							Content:   " and ",
							TextStyle: models.TextStyle{},
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img2",
						},
					},
				},
			},
			book: &models.Book{
				InlineObjectMap: map[string]string{
					"img1": "https://api.slimacademy.nl/image1.png",
					"img2": "https://api.slimacademy.nl/image2.png",
				},
			},
			expected: "![Image](https://api.slimacademy.nl/image1.png) and ![Image](https://api.slimacademy.nl/image2.png)",
		},
		{
			name: "image only paragraph",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
				},
			},
			book: &models.Book{
				InlineObjectMap: map[string]string{
					"img1": "https://api.slimacademy.nl/image1.png",
				},
			},
			expected: "![Image](https://api.slimacademy.nl/image1.png)",
		},
		{
			name: "missing image in map",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "Text with missing image: ",
							TextStyle: models.TextStyle{},
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "missing_img",
						},
					},
				},
			},
			book: &models.Book{
				InlineObjectMap: map[string]string{
					"img1": "https://api.slimacademy.nl/image1.png",
				},
			},
			expected: "Text with missing image:",
		},
		{
			name: "no book provided",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content:   "Text only",
							TextStyle: models.TextStyle{},
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
				},
			},
			book:     nil,
			expected: "Text only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatTextWithBook(tt.paragraph, tt.book)
			if result != tt.expected {
				t.Errorf("formatTextWithBook() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

// Helper functions
func createTestBook() *models.Book {
	return &models.Book{
		ID:          123,
		Title:       "Test Book",
		Description: "Test book description",
		Chapters: []models.Chapter{
			{
				ID:             1,
				Title:          "Introduction",
				IsVisible:      1,
				GDocsChapterID: "h.intro",
			},
			{
				ID:             2,
				Title:          "Chapter 1",
				IsVisible:      1,
				GDocsChapterID: "h.chapter1",
			},
		},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							ParagraphStyle: models.ParagraphStyle{
								HeadingID:      stringPtr("h.intro"),
								NamedStyleType: "HEADING_1",
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Introduction\n",
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
										Content: "Regular text and ",
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "bold text",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: ".\n",
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

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
