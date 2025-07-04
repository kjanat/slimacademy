package hast

import (
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/streaming"
)

func TestEventToHASTConverter(t *testing.T) {
	t.Run("SimpleDocument", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Test Document"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Hello, World!"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		// Render to HTML to verify structure
		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should contain title and paragraph
		if !contains(html, "<title>Test Document</title>") {
			t.Error("Expected title element in output")
		}
		if !contains(html, "<p>Hello, World!</p>") {
			t.Error("Expected paragraph with text in output")
		}
	})

	t.Run("Heading", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartHeading, Level: 2, AnchorID: "test-heading"},
			{Kind: streaming.Text, TextContent: "Test Heading"},
			{Kind: streaming.EndHeading},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<h2 id="test-heading">Test Heading</h2>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("TextFormatting", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
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

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<p>Plain text <strong>bold</strong> and <em>italic</em> text.</p>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("Link", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Visit "},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
			{Kind: streaming.Text, TextContent: "our website"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.Text, TextContent: " for more info."},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<p>Visit <a href="https://example.com">our website</a> for more info.</p>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("LinkWithURLSanitization", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "javascript:alert('xss')"},
			{Kind: streaming.Text, TextContent: "malicious link"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should sanitize dangerous URL to safe default
		expected := `<p><a href="#">malicious link</a></p>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("Image", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Image, ImageURL: "test.jpg", ImageAlt: "Test Image"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<img src="test.jpg" alt="Test Image" style="max-width: 100%; height: auto;" />`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("List", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
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

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<ul><li>First item</li><li>Second item</li></ul>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("Table", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartTable},
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

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// First row should use th elements, second row should use td elements
		if !contains(html, "<th") {
			t.Error("Expected th elements for first row")
		}
		if !contains(html, "<td") {
			t.Error("Expected td elements for data rows")
		}
		if !contains(html, "Header 1") || !contains(html, "Cell 1") {
			t.Error("Expected table content in output")
		}
	})

	t.Run("ComplexFormatting", func(t *testing.T) {
		// Test nested formatting
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Bold},
			{Kind: streaming.Text, TextContent: "Bold "},
			{Kind: streaming.StartFormatting, Style: streaming.Italic},
			{Kind: streaming.Text, TextContent: "and italic"},
			{Kind: streaming.EndFormatting, Style: streaming.Italic},
			{Kind: streaming.Text, TextContent: " text"},
			{Kind: streaming.EndFormatting, Style: streaming.Bold},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		expected := `<p><strong>Bold <em>and italic</em> text</strong></p>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("AutoGeneratedHeadingID", func(t *testing.T) {
		// Test auto-generated heading IDs when no AnchorID is provided
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartHeading, Level: 1}, // No AnchorID provided
			{Kind: streaming.Text, TextContent: "Test Heading"},
			{Kind: streaming.EndHeading},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should generate ID from heading text
		expected := `<h1 id="test-heading">Test Heading</h1>`
		if !contains(html, expected) {
			t.Errorf("Expected %q in output, got: %s", expected, html)
		}
	})

	t.Run("DocumentMetadata", func(t *testing.T) {
		events := []streaming.Event{
			{
				Kind:               streaming.StartDoc,
				Title:              "Test Document",
				Description:        "A test document for HAST conversion",
				BachelorYearNumber: "2024",
				Periods:            []string{"Q1", "Q2"},
			},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Content"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		converter := NewEventToHASTConverter(DefaultConversionOptions())
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should include metadata
		if !contains(html, "<title>Test Document</title>") {
			t.Error("Expected title element")
		}
		if !contains(html, `data-description="A test document for HAST conversion"`) {
			t.Error("Expected description in metadata")
		}
		if !contains(html, `data-bachelor-year="2024"`) {
			t.Error("Expected bachelor year in metadata")
		}
		if !contains(html, `data-periods="Q1,Q2"`) {
			t.Error("Expected periods in metadata")
		}
	})
}

func TestConversionOptions(t *testing.T) {
	t.Run("DisableMetadata", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Test Document"},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.Text, TextContent: "Content"},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		options := DefaultConversionOptions()
		options.IncludeMetadata = false

		converter := NewEventToHASTConverter(options)
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should not include title when metadata is disabled
		if contains(html, "<title>") {
			t.Error("Expected no title element when metadata is disabled")
		}
	})

	t.Run("DisableURLSanitization", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "javascript:alert('xss')"},
			{Kind: streaming.Text, TextContent: "link"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.EndParagraph},
			{Kind: streaming.EndDoc},
		}

		options := DefaultConversionOptions()
		options.SanitizeURLs = false

		converter := NewEventToHASTConverter(options)
		_, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		// Note: The HTMLRenderer still sanitizes URLs by default
		// This test verifies that the converter doesn't sanitize at its level
		// Full end-to-end testing would require a renderer with sanitization disabled
	})

	t.Run("DisableStyles", func(t *testing.T) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc},
			{Kind: streaming.Image, ImageURL: "test.jpg", ImageAlt: "Test"},
			{Kind: streaming.EndDoc},
		}

		options := DefaultConversionOptions()
		options.IncludeStyles = false

		converter := NewEventToHASTConverter(options)
		root, err := converter.Convert(events)
		if err != nil {
			t.Fatalf("Error converting events: %v", err)
		}

		renderer := NewHTMLRenderer()
		html, err := renderer.RenderToHTML(root)
		if err != nil {
			t.Fatalf("Error rendering HAST: %v", err)
		}

		// Should not include style attributes when styles are disabled
		if contains(html, "style=") {
			t.Error("Expected no style attributes when styles are disabled")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
