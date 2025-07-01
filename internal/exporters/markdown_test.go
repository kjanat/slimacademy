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

func TestMarkdownExporter_FormattingSpacing(t *testing.T) {
	exporter := NewMarkdownExporter().(*MarkdownExporter)

	tests := []struct {
		name     string
		paragraph *models.Paragraph
		expected string
		reason   string
	}{
		{
			name: "trailing space should move outside bold formatting",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "bold text ",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "normal text",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "**bold text** normal text",
			reason:   "Trailing space should be outside formatting markers",
		},
		{
			name: "leading space should move outside italic formatting",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "normal text",
							TextStyle: models.TextStyle{},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: " italic text",
							TextStyle: models.TextStyle{
								Italic: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "normal text *italic text*",
			reason:   "Leading space should be outside formatting markers",
		},
		{
			name: "multiple spaces should be handled correctly",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "before  ",
							TextStyle: models.TextStyle{},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "  bold text  ",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "  after",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "before    **bold text**    after",
			reason:   "Multiple spaces should be preserved but moved outside formatting",
		},
		{
			name: "nested formatting should handle spaces correctly",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "bold and italic ",
							TextStyle: models.TextStyle{
								Bold:   boolPtr(true),
								Italic: boolPtr(true),
							},
						},
					},
					{
						TextRun: &models.TextRun{
							Content: "normal",
							TextStyle: models.TextStyle{},
						},
					},
				},
			},
			expected: "***bold and italic*** normal",
			reason:   "Nested formatting with trailing space should move space outside",
		},
		{
			name: "underline with spaces",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: " underlined text ",
							TextStyle: models.TextStyle{
								Underline: boolPtr(true),
							},
						},
					},
				},
			},
			expected: " __underlined text__ ",
			reason:   "Underline formatting should handle spaces correctly",
		},
		{
			name: "edge case: line ending with double space (markdown line break)",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "line break  ",
							TextStyle: models.TextStyle{
								Bold: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "**line break**  ",
			reason:   "Double space line breaks should be preserved outside formatting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatText(tt.paragraph)
			if result != tt.expected {
				t.Errorf("formatText() = '%s', expected '%s'. Reason: %s", result, tt.expected, tt.reason)
			}
		})
	}
}

func TestMarkdownExporter_ListBlockSpacing(t *testing.T) {
	exporter := NewMarkdownExporter()

	book := &models.Book{
		ID:          1,
		Title:       "Test Book",
		Description: "Test description",
		Chapters:    []models.Chapter{},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Paragraph before list",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Bullet: &models.Bullet{
								ListID:       "list1",
								NestingLevel: intPtr(0),
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "First list item",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Bullet: &models.Bullet{
								ListID:       "list1",
								NestingLevel: intPtr(0),
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Second list item",
										TextStyle: models.TextStyle{},
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
										Content: "Paragraph after list",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "list_spacing.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)
	lines := strings.Split(content, "\n")

	// Find the positions of key elements
	var beforeListLine, firstItemLine, secondItemLine, afterListLine int
	for i, line := range lines {
		if strings.Contains(line, "Paragraph before list") {
			beforeListLine = i
		} else if strings.Contains(line, "- First list item") {
			firstItemLine = i
		} else if strings.Contains(line, "- Second list item") {
			secondItemLine = i
		} else if strings.Contains(line, "Paragraph after list") {
			afterListLine = i
		}
	}

	// Verify blank line before list block
	if firstItemLine-beforeListLine != 2 { // Should have one blank line between
		t.Errorf("Expected blank line before list block. Lines between: %d", firstItemLine-beforeListLine)
		t.Logf("Before list line %d: '%s'", beforeListLine, lines[beforeListLine])
		if beforeListLine+1 < len(lines) {
			t.Logf("Line %d: '%s'", beforeListLine+1, lines[beforeListLine+1])
		}
		t.Logf("First item line %d: '%s'", firstItemLine, lines[firstItemLine])
	}

	// Verify no extra spacing between list items
	if secondItemLine-firstItemLine != 1 { // Should be consecutive
		t.Errorf("Expected consecutive list items. Lines between: %d", secondItemLine-firstItemLine)
	}

	// Verify blank line after list block
	if afterListLine-secondItemLine != 2 { // Should have one blank line between
		t.Errorf("Expected blank line after list block. Lines between: %d", afterListLine-secondItemLine)
		t.Logf("Second item line %d: '%s'", secondItemLine, lines[secondItemLine])
		if secondItemLine+1 < len(lines) {
			t.Logf("Line %d: '%s'", secondItemLine+1, lines[secondItemLine+1])
		}
		t.Logf("After list line %d: '%s'", afterListLine, lines[afterListLine])
	}
}

func TestMarkdownExporter_NestedListSpacing(t *testing.T) {
	exporter := NewMarkdownExporter()

	book := &models.Book{
		ID:          1,
		Title:       "Test Book",
		Description: "Test description",
		Chapters:    []models.Chapter{},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Before nested list",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Bullet: &models.Bullet{
								ListID:       "list1",
								NestingLevel: intPtr(0),
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Top level item",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Bullet: &models.Bullet{
								ListID:       "list1",
								NestingLevel: intPtr(1), // Nested
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Nested item",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						Paragraph: &models.Paragraph{
							Bullet: &models.Bullet{
								ListID:       "list1",
								NestingLevel: intPtr(0), // Back to top level
							},
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "Another top level",
										TextStyle: models.TextStyle{},
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
										Content: "After nested list",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "nested_list.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)

	// Verify nested list structure
	testutils.AssertStringContains(t, content, "- Top level item")
	testutils.AssertStringContains(t, content, "  - Nested item") // Should have indentation
	testutils.AssertStringContains(t, content, "- Another top level")

	// Verify proper spacing around the entire nested list block
	lines := strings.Split(content, "\n")
	var beforeLine, afterLine int
	for i, line := range lines {
		if strings.Contains(line, "Before nested list") {
			beforeLine = i
		} else if strings.Contains(line, "After nested list") {
			afterLine = i
		}
	}

	// Should have blank lines around the entire list block
	// Find first and last list items
	var firstListLine, lastListLine int
	for i, line := range lines {
		if strings.Contains(line, "- Top level item") && firstListLine == 0 {
			firstListLine = i
		} else if strings.Contains(line, "- Another top level") {
			lastListLine = i
		}
	}

	if firstListLine-beforeLine != 2 {
		t.Errorf("Expected blank line before nested list block")
	}
	if afterLine-lastListLine != 2 {
		t.Errorf("Expected blank line after nested list block")
	}
}

func TestMarkdownExporter_SpacingBetweenTextRuns_Integration(t *testing.T) {
	exporter := NewMarkdownExporter()

	// Test cases based on real issues found in output
	book := &models.Book{
		ID:          1,
		Title:       "Spacing Integration Test",
		Description: "Test description",
		Chapters:    []models.Chapter{},
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						// Test case: "studieGeneeskunde.Slim" should be "studie Geneeskunde. Slim"
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "studie",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: " ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "Geneeskunde",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: ". ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "Slim",
										TextStyle: models.TextStyle{
											Bold: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: " Academy",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						// Test case: "op[__klantenservice" should be "op [__klantenservice"
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "mail ons dan op",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: " ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "klantenservice@slimacademy.nl",
										TextStyle: models.TextStyle{
											Underline: boolPtr(true),
											Link: &models.Link{
												URL: "mailto:klantenservice@slimacademy.nl",
											},
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: ", ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "dan gaan wij",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
					{
						// Test case: "vanaf**19-06-2025**zijn" should be "vanaf **19-06-2025** zijn"
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content: "vanaf",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: " ",
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
										Content: " ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "zijn",
										TextStyle: models.TextStyle{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "spacing_integration.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)
	lines := strings.Split(content, "\n")

	// Find content lines (skip headers and TOC)
	var contentLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "|") {
			contentLines = append(contentLines, line)
		}
	}

	// Test case 1: Spaces around bold text should be preserved
	found1 := false
	for _, line := range contentLines {
		if strings.Contains(line, "studie **Geneeskunde**. **Slim** Academy") {
			found1 = true
			break
		}
	}
	if !found1 {
		t.Errorf("Expected 'studie **Geneeskunde**. **Slim** Academy' with proper spacing")
		t.Logf("Content lines: %v", contentLines)
	}

	// Test case 2: Spaces around links should be preserved 
	found2 := false
	for _, line := range contentLines {
		if strings.Contains(line, "mail ons dan op __klantenservice@slimacademy.nl__, dan gaan wij") ||
		   strings.Contains(line, "mail ons dan op [__klantenservice@slimacademy.nl__]") {
			found2 = true
			break
		}
	}
	if !found2 {
		t.Errorf("Expected proper spacing around email link")
		t.Logf("Content lines: %v", contentLines)
	}

	// Test case 3: Spaces around date formatting should be preserved
	found3 := false
	for _, line := range contentLines {
		if strings.Contains(line, "vanaf **19-06-2025** zijn") {
			found3 = true
			break
		}
	}
	if !found3 {
		t.Errorf("Expected 'vanaf **19-06-2025** zijn' with proper spacing")
		t.Logf("Content lines: %v", contentLines)
	}
}

func TestMarkdownExporter_RealDataIntegration(t *testing.T) {
	// This test processes real JSON data and validates the output
	exporter := NewMarkdownExporter()
	
	// Load real Station B3 data
	book := testutils.LoadTestBook(t, "simple_book")
	
	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "real_data_test.md")
	
	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed with real data: %v", err)
	}
	
	content := testutils.ReadFileString(t, outputPath)
	
	// Test for common spacing anti-patterns that should NOT appear
	antiPatterns := []string{
		".**", // Missing space before bold
		")**", // Missing space after parenthesis before bold  
		"*,",  // Missing space before comma after formatting
		"*.",  // Missing space before period after formatting
		"]*",  // Missing space after link before formatting
		"**http", // Space inside bold before link
	}
	
	for _, pattern := range antiPatterns {
		if strings.Contains(content, pattern) {
			t.Errorf("Found anti-pattern '%s' in generated markdown", pattern)
		}
	}
	
	// Test for correct spacing patterns that SHOULD appear
	correctPatterns := []string{
		". **", // Space before bold after sentence
		") **", // Space before bold after parenthesis
		"** ",  // Space after bold (when followed by text)
		"__](", // Proper link formatting
	}
	
	foundPatterns := 0
	for _, pattern := range correctPatterns {
		if strings.Contains(content, pattern) {
			foundPatterns++
		}
	}
	
	if foundPatterns == 0 {
		t.Errorf("No correct spacing patterns found - this suggests spacing is broken")
	}
}
