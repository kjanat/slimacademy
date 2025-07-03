package writers

import (
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/streaming"
)

func TestNewHTMLWriter(t *testing.T) {
	writer := NewHTMLWriter()

	if writer == nil {
		t.Error("NewHTMLWriter should return non-nil writer")
	}
	if writer.config == nil {
		t.Error("HTMLWriter should have default config")
	}
	if writer.out == nil {
		t.Error("HTMLWriter should have string builder")
	}
	if writer.eventHandlers == nil {
		t.Error("HTMLWriter should have initialized event handlers")
	}
}

func TestNewHTMLWriterWithConfig(t *testing.T) {
	customConfig := &config.HTMLConfig{}
	writer := NewHTMLWriterWithConfig(customConfig)

	if writer.config != customConfig {
		t.Error("HTMLWriter should use provided config")
	}

	// Test with nil config
	writer2 := NewHTMLWriterWithConfig(nil)
	if writer2.config == nil {
		t.Error("HTMLWriter should use default config when nil provided")
	}
}

func TestHTMLWriter_BasicDocument(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{
			Kind:        streaming.StartDoc,
			Title:       "Test Document",
			Description: "A test document",
		},
		{
			Kind: streaming.StartParagraph,
		},
		{
			Kind:        streaming.Text,
			TextContent: "Hello, World!",
		},
		{
			Kind: streaming.EndParagraph,
		},
		{
			Kind: streaming.EndDoc,
		},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify HTML structure
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("Result should contain DOCTYPE declaration")
	}
	if !strings.Contains(result, "<html lang=\"en\">") {
		t.Error("Result should contain HTML tag with lang attribute")
	}
	if !strings.Contains(result, "<title>Test Document</title>") {
		t.Error("Result should contain title")
	}
	if !strings.Contains(result, "Hello, World!") {
		t.Error("Result should contain text content")
	}
	if !strings.Contains(result, "<p>Hello, World!</p>") {
		t.Error("Result should contain paragraph with text")
	}
}

func TestHTMLWriter_AcademicHeader(t *testing.T) {
	writer := NewHTMLWriter()

	readProgress := int64(50)
	event := streaming.Event{
		Kind:               streaming.StartDoc,
		Title:              "Academic Book",
		Description:        "An academic textbook",
		AvailableDate:      "2024-01-15",
		ExamDate:           "2024-06-30",
		BachelorYearNumber: "Year 2",
		CollegeStartYear:   2022,
		ReadProgress:       &readProgress,
		PageCount:          100,
		HasFreeChapters:    5,
		Periods:            []string{"Q1", "Q2"},
	}

	writer.Handle(event)
	result := writer.Result()

	// Verify academic header elements
	if !strings.Contains(result, `class="academic-header"`) {
		t.Error("Result should contain academic header")
	}
	if !strings.Contains(result, `class="book-title">Academic Book</h1>`) {
		t.Error("Result should contain book title")
	}
	if !strings.Contains(result, `class="book-description">An academic textbook</p>`) {
		t.Error("Result should contain book description")
	}
	if !strings.Contains(result, `class="metadata-grid"`) {
		t.Error("Result should contain metadata grid")
	}

	// Verify metadata items
	if !strings.Contains(result, "Academic Year:</span> Year 2") {
		t.Error("Result should contain academic year")
	}
	if !strings.Contains(result, "Available:</span> 2024-01-15") {
		t.Error("Result should contain available date")
	}
	if !strings.Contains(result, "Exam Date:</span> 2024-06-30") {
		t.Error("Result should contain exam date")
	}
	if !strings.Contains(result, "College Start:</span> 2022") {
		t.Error("Result should contain college start year")
	}
	if !strings.Contains(result, "Pages:</span> 100") {
		t.Error("Result should contain page count")
	}
	if !strings.Contains(result, "Periods:</span> Q1, Q2") {
		t.Error("Result should contain periods")
	}

	// Verify progress indicator
	if !strings.Contains(result, "Progress:</span> 50/100 pages (50.0%)") {
		t.Error("Result should contain progress information")
	}
	if !strings.Contains(result, `class="progress-bar"`) {
		t.Error("Result should contain progress bar")
	}
	if !strings.Contains(result, `style="width: 50.0%"`) {
		t.Error("Result should contain progress fill with correct percentage")
	}
}

func TestHTMLWriter_TableOfContents(t *testing.T) {
	writer := NewHTMLWriter()

	chapters := []models.Chapter{
		{
			ID:       1,
			Title:    "Introduction",
			IsFree:   models.BoolInt(true),
			IsLocked: models.BoolInt(false),
			SubChapters: []models.Chapter{
				{ID: 2, Title: "Getting Started", IsFree: models.BoolInt(true)},
				{ID: 3, Title: "Basic Concepts", IsLocked: models.BoolInt(true)},
			},
		},
		{
			ID:       4,
			Title:    "Advanced Topics",
			IsLocked: models.BoolInt(true),
		},
	}

	event := streaming.Event{
		Kind:     streaming.StartDoc,
		Title:    "Book with TOC",
		Chapters: chapters,
	}

	writer.Handle(event)
	result := writer.Result()

	// Verify TOC structure
	if !strings.Contains(result, `class="table-of-contents"`) {
		t.Error("Result should contain table of contents")
	}
	if !strings.Contains(result, `class="toc-title">Table of Contents</h2>`) {
		t.Error("Result should contain TOC title")
	}
	if !strings.Contains(result, `class="toc-list"`) {
		t.Error("Result should contain TOC list")
	}

	// Verify chapter entries
	if !strings.Contains(result, `href="#introduction"`) {
		t.Error("Result should contain link to introduction")
	}
	if !strings.Contains(result, "🆓 Introduction") {
		t.Error("Result should show free chapter indicator")
	}
	if !strings.Contains(result, "🔒 Advanced Topics") {
		t.Error("Result should show locked chapter indicator")
	}

	// Verify subchapter structure
	if !strings.Contains(result, `class="toc-sublist"`) {
		t.Error("Result should contain subchapter list")
	}
	if !strings.Contains(result, "Getting Started") && !strings.Contains(result, "Basic Concepts") {
		t.Error("Result should contain subchapter titles")
	}
}

func TestHTMLWriter_Headings(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Heading Test"},
		{Kind: streaming.StartHeading, Level: 2, AnchorID: "test-heading"},
		{Kind: streaming.Text, TextContent: "Test Heading"},
		{Kind: streaming.EndHeading},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify heading structure
	if !strings.Contains(result, `<h2 id="test-heading">Test Heading</h2>`) {
		t.Error("Result should contain heading with ID and text")
	}
}

func TestHTMLWriter_Lists(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "List Test"},
		{Kind: streaming.StartList, ListLevel: 0, ListOrdered: false},
		{Kind: streaming.Text, TextContent: "First item"},
		{Kind: streaming.Text, TextContent: "Second item"},
		{Kind: streaming.EndList},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify list structure
	if !strings.Contains(result, "<ul>") {
		t.Error("Result should contain unordered list")
	}
	if !strings.Contains(result, "<li>") {
		t.Error("Result should contain list items")
	}
	if !strings.Contains(result, "First item") {
		t.Error("Result should contain first item text")
	}
	if !strings.Contains(result, "Second item") {
		t.Error("Result should contain second item text")
	}
	if !strings.Contains(result, "</ul>") {
		t.Error("Result should close unordered list")
	}
}

func TestHTMLWriter_Tables(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Table Test"},
		{Kind: streaming.StartTable, TableColumns: 2, TableRows: 2},
		{Kind: streaming.StartTableRow},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Header 1"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Header 2"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.EndTableRow},
		{Kind: streaming.StartTableRow},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Cell 1"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Cell 2"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.EndTableRow},
		{Kind: streaming.EndTable},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify table structure
	if !strings.Contains(result, "<table") {
		t.Error("Result should contain table")
	}
	if !strings.Contains(result, "<th") {
		t.Error("Result should contain table headers")
	}
	if !strings.Contains(result, "<td") {
		t.Error("Result should contain table cells")
	}
	if !strings.Contains(result, "Header 1") && !strings.Contains(result, "Header 2") {
		t.Error("Result should contain header text")
	}
	if !strings.Contains(result, "Cell 1") && !strings.Contains(result, "Cell 2") {
		t.Error("Result should contain cell text")
	}

	// Verify header styling
	if !strings.Contains(result, "background-color: #f2f2f2") {
		t.Error("Result should style table headers")
	}
}

func TestHTMLWriter_Formatting(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Formatting Test"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "Bold text"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: " normal "},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "italic text"},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: " and "},
		{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
		{Kind: streaming.Text, TextContent: "link text"},
		{Kind: streaming.EndFormatting, Style: streaming.Link},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify formatting
	if !strings.Contains(result, "<strong>Bold text</strong>") {
		t.Error("Result should contain bold formatting")
	}
	if !strings.Contains(result, "<em>italic text</em>") {
		t.Error("Result should contain italic formatting")
	}
	if !strings.Contains(result, `<a href="https://example.com">link text</a>`) {
		t.Error("Result should contain link formatting")
	}
}

func TestHTMLWriter_Images(t *testing.T) {
	writer := NewHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Image Test"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "Text before image"},
		{Kind: streaming.Image, ImageURL: "/path/to/image.jpg", ImageAlt: "Test Image"},
		{Kind: streaming.Text, TextContent: "Text after image"},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify image
	if !strings.Contains(result, `<img src="/path/to/image.jpg"`) {
		t.Error("Result should contain image with correct src")
	}
	if !strings.Contains(result, `alt="Test Image"`) {
		t.Error("Result should contain image with correct alt text")
	}
	if !strings.Contains(result, `style="max-width: 100%; height: auto;"`) {
		t.Error("Result should contain responsive image styling")
	}
}

func TestHTMLWriter_EscapeHTML(t *testing.T) {
	writer := NewHTMLWriter()

	tests := []struct {
		input    string
		expected string
	}{
		{"Simple text", "Simple text"},
		{"Text with & ampersand", "Text with &amp; ampersand"},
		{"Text with < and >", "Text with &lt; and &gt;"},
		{"Text with \"quotes\"", "Text with &quot;quotes&quot;"},
		{"Text with 'apostrophes'", "Text with &#39;apostrophes&#39;"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := writer.escapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHTMLWriter_SlugGeneration(t *testing.T) {
	writer := NewHTMLWriter()

	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "simple-title"},
		{"Title with Numbers 123", "title-with-numbers-123"},
		{"Title with Special!@# Characters", "title-with-special-characters"},
		{"   Whitespace   ", "---whitespace---"},
		{"UPPERCASE Title", "uppercase-title"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := writer.slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHTMLWriter_Reset(t *testing.T) {
	writer := NewHTMLWriter()

	// Add some content and state
	writer.Handle(streaming.Event{Kind: streaming.StartDoc, Title: "Test"})
	writer.Handle(streaming.Event{Kind: streaming.StartList})
	writer.inList = true
	writer.activeStyle = streaming.Bold

	// Reset
	writer.Reset()

	// Verify state is cleared
	if writer.out.Len() != 0 {
		t.Error("Output should be empty after reset")
	}
	if writer.activeStyle != 0 {
		t.Error("Active style should be cleared after reset")
	}
	if writer.inList {
		t.Error("List state should be cleared after reset")
	}
	if writer.inTable {
		t.Error("Table state should be cleared after reset")
	}
}

func TestHTMLWriter_CSS(t *testing.T) {
	writer := NewHTMLWriter()

	css := writer.getCSS()
	if css == "" {
		t.Error("getCSS should return non-empty CSS")
	}

	enhancedCSS := writer.getEnhancedCSS()
	if enhancedCSS == "" {
		t.Error("getEnhancedCSS should return non-empty CSS")
	}
	if len(enhancedCSS) <= len(css) {
		t.Error("Enhanced CSS should be longer than basic CSS")
	}

	// Verify enhanced CSS contains academic styles
	if !strings.Contains(enhancedCSS, "academic-header") {
		t.Error("Enhanced CSS should contain academic header styles")
	}
	if !strings.Contains(enhancedCSS, "table-of-contents") {
		t.Error("Enhanced CSS should contain TOC styles")
	}
	if !strings.Contains(enhancedCSS, "progress-bar") {
		t.Error("Enhanced CSS should contain progress bar styles")
	}
	if !strings.Contains(enhancedCSS, "@media") {
		t.Error("Enhanced CSS should contain responsive media queries")
	}
}

func TestHTMLWriterV2_Interface(t *testing.T) {
	writer := &HTMLWriterV2{
		HTMLWriter: NewHTMLWriter(),
	}

	// Test that it implements WriterV2 interface
	var _ WriterV2 = writer

	// Test Handle with error handling
	event := streaming.Event{Kind: streaming.StartDoc, Title: "Test"}
	err := writer.Handle(event)
	if err != nil {
		t.Errorf("Handle should not return error for valid event: %v", err)
	}

	// Verify stats are tracked
	stats := writer.Stats()
	if stats.EventsProcessed != 1 {
		t.Errorf("Expected 1 event processed, got %d", stats.EventsProcessed)
	}
}

func TestHTMLWriterV2_Statistics(t *testing.T) {
	writer := &HTMLWriterV2{
		HTMLWriter: NewHTMLWriter(),
	}

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Stats Test"},
		{Kind: streaming.Text, TextContent: "Hello World"},
		{Kind: streaming.Image, ImageURL: "/image.jpg"},
		{Kind: streaming.StartTable},
		{Kind: streaming.StartHeading, Level: 2},
		{Kind: streaming.StartList},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		err := writer.Handle(event)
		if err != nil {
			t.Errorf("Handle should not return error: %v", err)
		}
	}

	stats := writer.Stats()

	if stats.EventsProcessed != len(events) {
		t.Errorf("Expected %d events processed, got %d", len(events), stats.EventsProcessed)
	}
	if stats.TextChars != 11 { // "Hello World" = 11 chars
		t.Errorf("Expected 11 text chars, got %d", stats.TextChars)
	}
	if stats.Images != 1 {
		t.Errorf("Expected 1 image, got %d", stats.Images)
	}
	if stats.Tables != 1 {
		t.Errorf("Expected 1 table, got %d", stats.Tables)
	}
	if stats.Headings != 1 {
		t.Errorf("Expected 1 heading, got %d", stats.Headings)
	}
	if stats.Lists != 1 {
		t.Errorf("Expected 1 list, got %d", stats.Lists)
	}
}

func TestHTMLWriterV2_Flush(t *testing.T) {
	writer := &HTMLWriterV2{
		HTMLWriter: NewHTMLWriter(),
	}

	writer.Handle(streaming.Event{Kind: streaming.StartDoc, Title: "Flush Test"})
	writer.Handle(streaming.Event{Kind: streaming.EndDoc})

	result, err := writer.Flush()
	if err != nil {
		t.Errorf("Flush should not return error: %v", err)
	}

	if result == "" {
		t.Error("Flush should return non-empty result")
	}
	if !strings.Contains(result, "Flush Test") {
		t.Error("Flush result should contain document content")
	}
}

func TestHTMLWriter_ComplexDocument(t *testing.T) {
	writer := NewHTMLWriter()

	// Simulate a complex document with various elements
	events := []streaming.Event{
		{
			Kind:               streaming.StartDoc,
			Title:              "Complex Academic Document",
			Description:        "A comprehensive test document",
			BachelorYearNumber: "Year 3",
			PageCount:          200,
			Chapters: []models.Chapter{
				{ID: 1, Title: "Introduction", IsFree: models.BoolInt(true)},
				{ID: 2, Title: "Methodology", IsLocked: models.BoolInt(true)},
			},
		},
		{Kind: streaming.StartHeading, Level: 2, AnchorID: "introduction"},
		{Kind: streaming.Text, TextContent: "Introduction"},
		{Kind: streaming.EndHeading},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "This is the introduction with "},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "bold"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: " and "},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "italic"},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: " text."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.StartList},
		{Kind: streaming.Text, TextContent: "First point"},
		{Kind: streaming.Text, TextContent: "Second point"},
		{Kind: streaming.EndList},
		{Kind: streaming.StartTable, TableColumns: 2, TableRows: 2},
		{Kind: streaming.StartTableRow},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Data 1"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.StartTableCell},
		{Kind: streaming.Text, TextContent: "Data 2"},
		{Kind: streaming.EndTableCell},
		{Kind: streaming.EndTableRow},
		{Kind: streaming.EndTable},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Verify complex document structure
	if !strings.Contains(result, "Complex Academic Document") {
		t.Error("Result should contain document title")
	}
	if !strings.Contains(result, "table-of-contents") {
		t.Error("Result should contain table of contents")
	}
	if !strings.Contains(result, `<h2 id="introduction">Introduction</h2>`) {
		t.Error("Result should contain heading")
	}
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Error("Result should contain bold formatting")
	}
	if !strings.Contains(result, "<em>italic</em>") {
		t.Error("Result should contain italic formatting")
	}
	if !strings.Contains(result, "<ul>") && !strings.Contains(result, "<li>") {
		t.Error("Result should contain list elements")
	}
	if !strings.Contains(result, "<table") {
		t.Error("Result should contain table")
	}

	// Verify document is well-formed HTML
	if !strings.HasPrefix(result, "<!DOCTYPE html>") {
		t.Error("Result should start with DOCTYPE")
	}
	if !strings.HasSuffix(strings.TrimSpace(result), "</html>") {
		t.Error("Result should end with closing html tag")
	}
}

// Benchmark tests

func BenchmarkHTMLWriter_SimpleDocument(b *testing.B) {
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Benchmark Test"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "Simple paragraph content for benchmarking."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := NewHTMLWriter()
		for _, event := range events {
			writer.Handle(event)
		}
		_ = writer.Result()
	}
}

func BenchmarkHTMLWriter_ComplexFormatting(b *testing.B) {
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Complex Formatting"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
		{Kind: streaming.Text, TextContent: "Complex formatted text"},
		{Kind: streaming.EndFormatting, Style: streaming.Link},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := NewHTMLWriter()
		for _, event := range events {
			writer.Handle(event)
		}
		_ = writer.Result()
	}
}

func BenchmarkHTMLWriter_HTMLEscaping(b *testing.B) {
	writer := NewHTMLWriter()
	text := "<script>alert('This needs escaping & formatting');</script>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = writer.escapeHTML(text)
	}
}

func BenchmarkHTMLWriter_SlugGeneration(b *testing.B) {
	writer := NewHTMLWriter()
	titles := []string{
		"Simple Title",
		"Complex Title with Many Words and Numbers 123",
		"Title with Special Characters!@#$%",
		"Very Long Title That Should Be Processed Efficiently Without Performance Issues",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		title := titles[i%len(titles)]
		_ = writer.slugify(title)
	}
}
