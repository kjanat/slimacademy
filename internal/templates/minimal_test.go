package templates

import (
	"html/template"
	"strings"
	"testing"
)

func TestMinimalTemplate(t *testing.T) {
	t.Run("BasicRendering", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		data := TemplateData{
			Title:       "Test Document",
			Description: "A test document for minimal template",
			Content:     template.HTML("<p>Hello, World!</p>"),
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
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
			t.Error("Expected content to be rendered")
		}
	})

	t.Run("WithMetadata", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		data := TemplateData{
			Title:   "Academic Document",
			Content: template.HTML("<p>Content with metadata</p>"),
			Metadata: map[string]string{
				"Author": "Test Author",
				"Year":   "2024",
				"Course": "Test Course",
			},
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
		}

		// Check for metadata rendering
		if !strings.Contains(html, "document-metadata") {
			t.Error("Expected metadata container")
		}

		if !strings.Contains(html, "Author") || !strings.Contains(html, "Test Author") {
			t.Error("Expected author metadata")
		}

		if !strings.Contains(html, "Year") || !strings.Contains(html, "2024") {
			t.Error("Expected year metadata")
		}
	})

	t.Run("WithoutDescription", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		data := TemplateData{
			Title:   "Simple Document",
			Content: template.HTML("<p>No description</p>"),
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
		}

		// Should not contain description elements when description is empty
		if strings.Contains(html, `<p class="document-description">`) {
			t.Error("Should not render description when empty")
		}

		if strings.Contains(html, `<meta name="description"`) {
			t.Error("Should not include description meta tag when empty")
		}
	})

	t.Run("CustomCSS", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		customCSS := template.CSS("body { background: red; }")
		data := TemplateData{
			Title:   "Custom Style Document",
			Content: template.HTML("<p>Custom styles</p>"),
			CSS:     customCSS,
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
		}

		if !strings.Contains(html, "body { background: red; }") {
			t.Error("Expected custom CSS to be included")
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		data := TemplateData{
			Title:   "Default Test",
			Content: template.HTML("<p>Testing defaults</p>"),
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
		}

		// Should include default generator
		if !strings.Contains(html, "SlimAcademy") {
			t.Error("Expected default generator to be SlimAcademy")
		}

		// Should include default CSS
		if !strings.Contains(html, "font-family: -apple-system") {
			t.Error("Expected default CSS to be included")
		}
	})

	t.Run("HTMLEscaping", func(t *testing.T) {
		tmpl := NewMinimalTemplate()

		data := TemplateData{
			Title:       "Test <script>alert('xss')</script>",
			Description: "Description with <dangerous> content",
			Content:     template.HTML("<p>Safe content</p>"), // template.HTML is trusted
			Metadata: map[string]string{
				"Key": "Value with <script>",
			},
		}

		html, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Error rendering template: %v", err)
		}

		// Title and description should be escaped
		if strings.Contains(html, "<script>alert('xss')</script>") {
			t.Error("Script tag should be escaped in title")
		}

		if strings.Contains(html, "<dangerous>") {
			t.Error("Dangerous tag should be escaped in description")
		}

		// But should contain escaped versions
		if !strings.Contains(html, "&lt;script&gt;") {
			t.Error("Expected escaped script tag in title")
		}

		// Content should not be escaped (it's template.HTML)
		if !strings.Contains(html, "<p>Safe content</p>") {
			t.Error("Content should not be escaped when using template.HTML")
		}
	})
}

func TestDefaultCSS(t *testing.T) {
	css := DefaultCSS()

	cssString := string(css)

	// Check for essential styles
	if !strings.Contains(cssString, "font-family:") {
		t.Error("Expected font-family in default CSS")
	}

	if !strings.Contains(cssString, ".document") {
		t.Error("Expected document class in default CSS")
	}

	if !strings.Contains(cssString, "@media (max-width: 768px)") {
		t.Error("Expected responsive styles in default CSS")
	}

	if !strings.Contains(cssString, "@media print") {
		t.Error("Expected print styles in default CSS")
	}
}

func BenchmarkTemplateRender(b *testing.B) {
	tmpl := NewMinimalTemplate()

	data := TemplateData{
		Title:       "Benchmark Document",
		Description: "Testing template performance",
		Content:     template.HTML("<p>Benchmark content with <strong>formatting</strong> and <a href='#'>links</a>.</p>"),
		Metadata: map[string]string{
			"Author": "Benchmark Author",
			"Year":   "2024",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tmpl.Render(data)
		if err != nil {
			b.Fatalf("Error rendering template: %v", err)
		}
	}
}
