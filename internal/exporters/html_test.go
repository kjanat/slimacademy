package exporters

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
	testutils "github.com/kjanat/slimacademy/test/utils"
)

func TestHTMLExporter_Export(t *testing.T) {
	exporter := NewHTMLExporter()

	book := createTestBook()
	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "test.html")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	testutils.AssertFileExists(t, outputPath)

	content := testutils.ReadFileString(t, outputPath)

	// Verify basic structure
	testutils.AssertStringContains(t, content, "<h1>Test Book</h1>")
	testutils.AssertStringContains(t, content, "Test book description")
	testutils.AssertStringContains(t, content, "<h2>Table of Contents</h2>")
	testutils.AssertStringContains(t, content, "<h2 id=\"introduction\">Introduction</h2>")
	testutils.AssertStringContains(t, content, "<strong>bold text</strong>")
}

func TestHTMLExporter_GetExtension(t *testing.T) {
	exporter := NewHTMLExporter()

	ext := exporter.GetExtension()
	if ext != "html" {
		t.Errorf("Expected extension 'html', got '%s'", ext)
	}
}

func TestHTMLExporter_GetName(t *testing.T) {
	exporter := NewHTMLExporter()

	name := exporter.GetName()
	if name != "HTML" {
		t.Errorf("Expected name 'HTML', got '%s'", name)
	}
}

func TestHTMLExporter_GenerateTableOfContents(t *testing.T) {
	exporter := &HTMLExporter{}

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

	var html strings.Builder
	exporter.generateHTMLTableOfContents(&html, chapters)

	toc := html.String()

	// Should contain visible chapters
	testutils.AssertStringContains(t, toc, "<a href=\"#introduction\">Introduction</a>")
	testutils.AssertStringContains(t, toc, "<a href=\"#chapter-1\">Chapter 1</a>")
	testutils.AssertStringContains(t, toc, "<a href=\"#section-1.1\">Section 1.1</a>")

	// Should not contain hidden chapters
	testutils.AssertStringNotContains(t, toc, "Hidden Chapter")
	testutils.AssertStringNotContains(t, toc, "Section 1.2")
}

func TestHTMLExporter_FormatHTMLTextWithBook(t *testing.T) {
	exporter := &HTMLExporter{}

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		book      *models.Book
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
			expected: "<strong>Bold text</strong>",
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
			expected: "<em>Italic text</em>",
		},
		{
			name: "underline text",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Underline text",
							TextStyle: models.TextStyle{
								Underline: boolPtr(true),
							},
						},
					},
				},
			},
			expected: "<u>Underline text</u>",
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
			expected: "<a href=\"https://example.com\">Link text</a>",
		},
		{
			name: "inline image",
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
					"img1": "https://example.com/image.png",
				},
			},
			expected: "<img src=\"https://example.com/image.png\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" />",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatHTMLTextWithBook(tt.paragraph, tt.book)
			if result != tt.expected {
				t.Errorf("formatHTMLTextWithBook() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestHTMLExporter_BoldTextSpacing(t *testing.T) {
	exporter := &HTMLExporter{}

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
			expected: "vanaf <strong>19-06-2025</strong> zijn",
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
			expected: "<strong>Viscerale pijn</strong> is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatHTMLTextWithBook(tt.paragraph, nil)
			if result != tt.expected {
				t.Errorf("formatHTMLTextWithBook() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}

func TestHTMLExporter_GetHeadingLevel(t *testing.T) {
	exporter := &HTMLExporter{}

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

func TestHTMLExporter_Slugify(t *testing.T) {
	exporter := &HTMLExporter{}

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

func TestHTMLExporter_EscapeHTML(t *testing.T) {
	exporter := &HTMLExporter{}

	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Quotes: \"double\" & 'single'", "Quotes: &quot;double&quot; &amp; &#39;single&#39;"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := exporter.escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHTMLExporter_BuildChapterMap(t *testing.T) {
	exporter := &HTMLExporter{}

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

func TestHTMLExporter_WithRealTestData(t *testing.T) {
	exporter := NewHTMLExporter()

	// Load real test data
	book := testutils.LoadTestBook(t, "simple_book")

	tempDir := testutils.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "real_test.html")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed with real data: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)

	// Verify structure based on our test fixture
	testutils.AssertStringContains(t, content, "<h1>Test Book</h1>")
	testutils.AssertStringContains(t, content, "<h2>Table of Contents</h2>")
	testutils.AssertStringContains(t, content, "<a href=\"#introduction\">Introduction</a>")
	testutils.AssertStringContains(t, content, "<h2 id=\"introduction\">Introduction</h2>")
	testutils.AssertStringContains(t, content, "This is a simple test paragraph")
	testutils.AssertStringContains(t, content, "<strong>")

	// Verify valid HTML structure
	testutils.AssertStringContains(t, content, "<!DOCTYPE html>")
	testutils.AssertStringContains(t, content, "<html lang=\"en\">")
	testutils.AssertStringContains(t, content, "</html>")
	testutils.AssertStringContains(t, content, "<head>")
	testutils.AssertStringContains(t, content, "</head>")
	testutils.AssertStringContains(t, content, "<body>")
	testutils.AssertStringContains(t, content, "</body>")
}

func TestHTMLExporter_EmptyContent(t *testing.T) {
	exporter := NewHTMLExporter()

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
	outputPath := filepath.Join(tempDir, "empty.html")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Export() failed with empty content: %v", err)
	}

	content := testutils.ReadFileString(t, outputPath)

	// Should still have basic structure
	testutils.AssertStringContains(t, content, "<h1>Empty Book</h1>")
	testutils.AssertStringContains(t, content, "<h2>Table of Contents</h2>")
	testutils.AssertStringContains(t, content, "<!DOCTYPE html>")
}

func TestHTMLExporter_WriteError(t *testing.T) {
	exporter := NewHTMLExporter()
	book := createTestBook()

	// Try to write to an invalid path
	invalidPath := "/root/cannot_write_here.html"

	err := exporter.Export(book, invalidPath)
	if err == nil {
		t.Errorf("Expected error when writing to invalid path")
	}
}

func TestHTMLExporter_HasInlineObjects(t *testing.T) {
	exporter := &HTMLExporter{}

	tests := []struct {
		name      string
		paragraph *models.Paragraph
		expected  bool
	}{
		{
			name: "no inline objects",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Just text",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "has inline objects",
			paragraph: &models.Paragraph{
				Elements: []models.Element{
					{
						TextRun: &models.TextRun{
							Content: "Text before image",
						},
					},
					{
						InlineObjectElement: &models.InlineObjectElement{
							InlineObjectID: "img1",
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.hasInlineObjects(tt.paragraph)
			if result != tt.expected {
				t.Errorf("hasInlineObjects() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHTMLExporter_InlineImageRendering(t *testing.T) {
	exporter := &HTMLExporter{}

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
			expected: "Before image: <img src=\"https://api.slimacademy.nl/image1.png\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" /> after image.",
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
			expected: "<img src=\"https://api.slimacademy.nl/image1.png\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" /> and <img src=\"https://api.slimacademy.nl/image2.png\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" />",
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
			expected: "<img src=\"https://api.slimacademy.nl/image1.png\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" />",
		},
		{
			name: "HTML escaping in image URL",
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
					"img1": "https://example.com/image?param=<script>",
				},
			},
			expected: "<img src=\"https://example.com/image?param=&lt;script&gt;\" alt=\"Image\" style=\"max-width: 100%; height: auto; margin: 0 5px;\" />",
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
			result := exporter.formatHTMLTextWithBook(tt.paragraph, tt.book)
			if result != tt.expected {
				t.Errorf("formatHTMLTextWithBook() = '%s', expected '%s'", result, tt.expected)
			}
		})
	}
}