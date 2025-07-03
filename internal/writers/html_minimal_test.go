package writers

import (
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/streaming"
)

func TestMinimalHTMLWriter(t *testing.T) {
	t.Run("BasicDocument", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Test Document", Description: "A test document"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Hello, World!"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		// Check for basic HTML structure
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Error("Expected HTML5 doctype")
		}

		if !strings.Contains(html, "<title>Test Document</title>") {
			t.Error("Expected title in head")
		}

		if !strings.Contains(html, "<h1 class=\"document-title\">Test Document</h1>") {
			t.Error("Expected document title in body")
		}

		if !strings.Contains(html, "<p>Hello, World!</p>") {
			t.Error("Expected paragraph content")
		}

		if !strings.Contains(html, "A test document") {
			t.Error("Expected description in output")
		}
	})

	t.Run("WithMetadata", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{
				Kind:               streaming.StartDoc,
				Title:              "Academic Document",
				Description:        "Course material",
				BachelorYearNumber: "2024",
				PageCount:          100,
			},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Content with metadata"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		// Check for metadata rendering
		if !strings.Contains(html, "document-metadata") {
			t.Error("Expected metadata container")
		}

		if !strings.Contains(html, "2024") {
			t.Error("Expected academic year in metadata")
		}

		if !strings.Contains(html, "100") {
			t.Error("Expected page count in metadata")
		}
	})

	t.Run("TextFormatting", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Formatting Test"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Plain text "},
			{Kind: streaming.StartFormatting, Style: streaming.Bold},
			{Kind: streaming.Text, TextContent: "bold"},
			{Kind: streaming.EndFormatting, Style: streaming.Bold},
			{Kind: streaming.Text, TextContent: " and "},
			{Kind: streaming.StartFormatting, Style: streaming.Italic},
			{Kind: streaming.Text, TextContent: "italic"},
			{Kind: streaming.EndFormatting, Style: streaming.Italic},
			{Kind: streaming.Text, TextContent: " text."},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		expected := "<p>Plain text <strong>bold</strong> and <em>italic</em> text.</p>"
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %q in output", expected)
		}
	})

	t.Run("Links", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Link Test"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Visit "},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
			{Kind: streaming.Text, TextContent: "our website"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.Text, TextContent: " for more info."},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		expected := `<p>Visit <a href="https://example.com">our website</a> for more info.</p>`
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %q in output", expected)
		}
	})

	t.Run("URLSanitization", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "XSS Test"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "javascript:alert('xss')"},
			{Kind: streaming.Text, TextContent: "malicious link"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		// Should sanitize dangerous URL to safe default
		expected := `<a href="#">malicious link</a>`
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %q in output", expected)
		}

		// Should not contain the dangerous URL
		if strings.Contains(html, "javascript:alert") {
			t.Error("Dangerous URL should be sanitized")
		}
	})

	t.Run("Images", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Image Test"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Image, ImageURL: "test.jpg", ImageAlt: "Test Image"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		expected := `<img src="test.jpg" alt="Test Image" />`
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %q in output", expected)
		}
	})

	t.Run("Lists", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "List Test"},
			{Kind: streaming.StartList},
			{Kind: streaming.StartListItem},
			{Kind: streaming.Text, TextContent: "First item"},
			{Kind: streaming.EndListItem},
			{Kind: streaming.StartListItem},
			{Kind: streaming.Text, TextContent: "Second item"},
			{Kind: streaming.EndListItem},
			{Kind: streaming.EndList},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		expected := "<ul>\n<li>First item</li>\n<li>Second item</li>\n</ul>"
		if !strings.Contains(html, expected) {
			t.Errorf("Expected %q in output", expected)
		}
	})

	t.Run("Headings", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Heading Test"},
			{Kind: streaming.StartHeading, Level: 2, AnchorID: "test-heading"},
			{Kind: streaming.Text, TextContent: "Test Heading"},
			{Kind: streaming.EndHeading},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		if !strings.Contains(html, `<h2 id="test-heading">`) {
			t.Error("Expected heading with ID")
		}

		if !strings.Contains(html, "Test Heading") {
			t.Error("Expected heading text")
		}
	})

	t.Run("HTMLEscaping", func(t *testing.T) {
		writer := NewMinimalHTMLWriter()

		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Test <script>alert('xss')</script>"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Content with <dangerous> & special chars"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		html, err := writer.ProcessEvents(events)
		if err != nil {
			t.Fatalf("Error processing events: %v", err)
		}

		// Check that dangerous content is escaped
		if strings.Contains(html, "<script>alert('xss')</script>") {
			t.Error("Script tags should be escaped")
		}

		if strings.Contains(html, "<dangerous>") {
			t.Error("Dangerous tags should be escaped")
		}

		// Should contain escaped versions
		if !strings.Contains(html, "&lt;script&gt;") {
			t.Error("Expected escaped script tag")
		}

		if !strings.Contains(html, "&lt;dangerous&gt;") {
			t.Error("Expected escaped dangerous tag")
		}

		if !strings.Contains(html, "&amp;") {
			t.Error("Expected escaped ampersand")
		}
	})
}

func BenchmarkMinimalHTMLWriter(b *testing.B) {
	writer := NewMinimalHTMLWriter()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Benchmark Document", Description: "Performance test"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "This is a benchmark test with "},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "bold"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: " and "},
		{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
		{Kind: streaming.Text, TextContent: "link"},
		{Kind: streaming.EndFormatting, Style: streaming.Link},
		{Kind: streaming.Text, TextContent: " content."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := writer.ProcessEvents(events)
		if err != nil {
			b.Fatalf("Error processing events: %v", err)
		}
		writer.Reset()
	}
}
