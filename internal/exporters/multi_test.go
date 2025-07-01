package exporters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
)

func TestMultiExporter_ExportFormats(t *testing.T) {
	// Create a test book
	book := &models.Book{
		Title:       "Test Book",
		Description: "A test book for the event stream system",
		Chapters: []models.Chapter{
			{
				GDocsChapterID: "h.intro",
				Title:          "Introduction",
				IsVisible:      1,
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
										Content:   "Introduction",
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
										Content:   "This is ",
										TextStyle: models.TextStyle{},
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
										Content:   " and ",
										TextStyle: models.TextStyle{},
									},
								},
								{
									TextRun: &models.TextRun{
										Content: "italic text",
										TextStyle: models.TextStyle{
											Italic: boolPtr(true),
										},
									},
								},
								{
									TextRun: &models.TextRun{
										Content:   ".",
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

	// Create temporary directory
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test")

	// Create exporter
	exporter := NewMultiExporter(nil)

	// Test single format (Markdown)
	err := exporter.ExportFormats(book, outputPath, []ExportFormat{FormatMarkdown})
	if err != nil {
		t.Fatalf("Failed to export Markdown: %v", err)
	}

	// Check Markdown output
	mdFile := filepath.Join(tempDir, "test.md")
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		t.Error("Markdown file was not created")
	}

	mdContent, err := os.ReadFile(mdFile)
	if err != nil {
		t.Fatalf("Failed to read Markdown file: %v", err)
	}

	mdStr := string(mdContent)
	if !strings.Contains(mdStr, "# Test Book") {
		t.Error("Markdown doesn't contain book title")
	}
	if !strings.Contains(mdStr, "## Introduction") {
		t.Error("Markdown doesn't contain heading")
	}
	if !strings.Contains(mdStr, "**bold text**") {
		t.Error("Markdown doesn't contain bold formatting")
	}
	if !strings.Contains(mdStr, "_italic text_") {
		t.Error("Markdown doesn't contain italic formatting")
	}

	// Test multi-format export
	err = exporter.ExportFormats(book, outputPath, []ExportFormat{FormatMarkdown, FormatHTML, FormatLaTeX})
	if err != nil {
		t.Fatalf("Failed to export multiple formats: %v", err)
	}

	// Check that all files were created
	formats := []string{"md", "html", "tex"}
	for _, format := range formats {
		file := filepath.Join(tempDir, "test."+format)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("%s file was not created", format)
		}
	}

	// Check HTML content
	htmlFile := filepath.Join(tempDir, "test.html")
	htmlContent, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	htmlStr := string(htmlContent)
	if !strings.Contains(htmlStr, "<h1>Test Book</h1>") {
		t.Error("HTML doesn't contain book title")
	}
	if !strings.Contains(htmlStr, "<h2 id=\"introduction\">Introduction</h2>") {
		t.Error("HTML doesn't contain heading")
	}
	if !strings.Contains(htmlStr, "<strong>bold text</strong>") {
		t.Error("HTML doesn't contain bold formatting")
	}
	if !strings.Contains(htmlStr, "<em>italic text</em>") {
		t.Error("HTML doesn't contain italic formatting")
	}

	// Check LaTeX content
	texFile := filepath.Join(tempDir, "test.tex")
	texContent, err := os.ReadFile(texFile)
	if err != nil {
		t.Fatalf("Failed to read LaTeX file: %v", err)
	}

	texStr := string(texContent)
	if !strings.Contains(texStr, "\\title{Test Book}") {
		t.Error("LaTeX doesn't contain book title")
	}
	if !strings.Contains(texStr, "\\subsection{Introduction}") {
		t.Error("LaTeX doesn't contain heading")
	}
	if !strings.Contains(texStr, "\\textbf{bold text}") {
		t.Error("LaTeX doesn't contain bold formatting")
	}
	if !strings.Contains(texStr, "\\emph{italic text}") {
		t.Error("LaTeX doesn't contain italic formatting")
	}
}

func TestMarkdownExporter_Legacy(t *testing.T) {
	// Test that legacy interface still works
	exporter := NewMarkdownExporter()

	book := &models.Book{
		Title: "Legacy Test",
		Content: models.Content{
			Body: models.Body{
				Content: []models.ContentElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.Element{
								{
									TextRun: &models.TextRun{
										Content:   "Simple test",
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

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "legacy.md")

	err := exporter.Export(book, outputPath)
	if err != nil {
		t.Fatalf("Legacy export failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Legacy Markdown file was not created")
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read legacy file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Legacy Test") {
		t.Error("Legacy output doesn't contain book title")
	}
	if !strings.Contains(contentStr, "Simple test") {
		t.Error("Legacy output doesn't contain paragraph text")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
