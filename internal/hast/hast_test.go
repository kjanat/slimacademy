package hast

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHASTNodeCreation(t *testing.T) {
	t.Run("NewRoot", func(t *testing.T) {
		root := NewRoot()
		if root.Type() != "root" {
			t.Errorf("Expected type 'root', got %q", root.Type())
		}
		if len(root.Children) != 0 {
			t.Errorf("Expected empty children, got %d", len(root.Children))
		}
	})

	t.Run("NewElement", func(t *testing.T) {
		element := NewElement("div")
		if element.Type() != "element" {
			t.Errorf("Expected type 'element', got %q", element.Type())
		}
		if element.TagName != "div" {
			t.Errorf("Expected tagName 'div', got %q", element.TagName)
		}
		if len(element.Children) != 0 {
			t.Errorf("Expected empty children, got %d", len(element.Children))
		}
		if len(element.Properties) != 0 {
			t.Errorf("Expected empty properties, got %d", len(element.Properties))
		}
	})

	t.Run("NewText", func(t *testing.T) {
		text := NewText("Hello, World!")
		if text.Type() != "text" {
			t.Errorf("Expected type 'text', got %q", text.Type())
		}
		if text.Value != "Hello, World!" {
			t.Errorf("Expected value 'Hello, World!', got %q", text.Value)
		}
	})

	t.Run("NewComment", func(t *testing.T) {
		comment := NewComment("This is a comment")
		if comment.Type() != "comment" {
			t.Errorf("Expected type 'comment', got %q", comment.Type())
		}
		if comment.Value != "This is a comment" {
			t.Errorf("Expected value 'This is a comment', got %q", comment.Value)
		}
	})
}

func TestHASTTreeBuilding(t *testing.T) {
	t.Run("AddChild", func(t *testing.T) {
		root := NewRoot()
		element := NewElement("div")
		text := NewText("Hello")

		err := AddChild(root, element)
		if err != nil {
			t.Errorf("Unexpected error adding element to root: %v", err)
		}

		err = AddChild(element, text)
		if err != nil {
			t.Errorf("Unexpected error adding text to element: %v", err)
		}

		if len(root.Children) != 1 {
			t.Errorf("Expected 1 child in root, got %d", len(root.Children))
		}

		if len(element.Children) != 1 {
			t.Errorf("Expected 1 child in element, got %d", len(element.Children))
		}

		// Test error case
		err = AddChild(text, element)
		if err == nil {
			t.Error("Expected error when adding child to text node")
		}
	})

	t.Run("Element Properties", func(t *testing.T) {
		element := NewElement("a")

		// Test SetProperty
		element.SetProperty("href", "https://example.com")
		element.SetProperty("class", "link")

		// Test GetProperty
		href, exists := element.GetProperty("href")
		if !exists || href != "https://example.com" {
			t.Errorf("Expected href property 'https://example.com', got %v (exists: %v)", href, exists)
		}

		// Test HasProperty
		if !element.HasProperty("class") {
			t.Error("Expected element to have 'class' property")
		}

		if element.HasProperty("nonexistent") {
			t.Error("Expected element not to have 'nonexistent' property")
		}
	})
}

func TestHASTJSONSerialization(t *testing.T) {
	t.Run("Root JSON", func(t *testing.T) {
		root := NewRoot()
		element := NewElement("p")
		text := NewText("Hello")

		AddChild(element, text)
		AddChild(root, element)

		data, err := json.Marshal(root)
		if err != nil {
			t.Fatalf("Error marshaling root: %v", err)
		}

		// Verify the JSON contains expected structure
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"type":"root"`) {
			t.Error("Expected JSON to contain root type")
		}
		if !strings.Contains(jsonStr, `"tagName":"p"`) {
			t.Error("Expected JSON to contain paragraph tagName")
		}
		if !strings.Contains(jsonStr, `"value":"Hello"`) {
			t.Error("Expected JSON to contain text value")
		}
	})
}

func TestHTMLRenderer(t *testing.T) {
	renderer := NewHTMLRenderer()

	t.Run("RenderText", func(t *testing.T) {
		text := NewText("Hello & <World>")
		html, err := renderer.RenderToHTML(text)
		if err != nil {
			t.Fatalf("Error rendering text: %v", err)
		}

		expected := "Hello &amp; &lt;World&gt;"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("RenderElement", func(t *testing.T) {
		element := NewElement("div")
		element.SetProperty("class", "container")
		text := NewText("Content")
		AddChild(element, text)

		html, err := renderer.RenderToHTML(element)
		if err != nil {
			t.Fatalf("Error rendering element: %v", err)
		}

		expected := `<div class="container">Content</div>`
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("RenderVoidElement", func(t *testing.T) {
		img := NewElement("img")
		img.SetProperty("src", "test.jpg")
		img.SetProperty("alt", "Test image")

		html, err := renderer.RenderToHTML(img)
		if err != nil {
			t.Fatalf("Error rendering void element: %v", err)
		}

		expected := `<img src="test.jpg" alt="Test image" />`
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("RenderComment", func(t *testing.T) {
		comment := NewComment("This is a comment with <script>")
		html, err := renderer.RenderToHTML(comment)
		if err != nil {
			t.Fatalf("Error rendering comment: %v", err)
		}

		expected := `<!-- This is a comment with &lt;script&gt; -->`
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("URLSanitization", func(t *testing.T) {
		tests := []struct {
			name     string
			href     string
			expected string
		}{
			{"safe HTTP", "http://example.com", "http://example.com"},
			{"safe HTTPS", "https://example.com", "https://example.com"},
			{"safe mailto", "mailto:test@example.com", "mailto:test@example.com"},
			{"dangerous javascript", "javascript:alert('xss')", "#"},
			{"dangerous data", "data:text/html,<script>", "#"},
			{"relative safe", "/path/to/page", "/path/to/page"},
			{"fragment safe", "#section", "#section"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				link := NewElement("a")
				link.SetProperty("href", tt.href)
				AddChild(link, NewText("Link"))

				html, err := renderer.RenderToHTML(link)
				if err != nil {
					t.Fatalf("Error rendering link: %v", err)
				}

				expectedHTML := `<a href="` + tt.expected + `">Link</a>`
				if html != expectedHTML {
					t.Errorf("Expected %q, got %q", expectedHTML, html)
				}
			})
		}
	})
}

func TestHASTHelperFunctions(t *testing.T) {
	renderer := NewHTMLRenderer()

	t.Run("Paragraph", func(t *testing.T) {
		p := Paragraph("This is a paragraph.")
		html, err := renderer.RenderToHTML(p)
		if err != nil {
			t.Fatalf("Error rendering paragraph: %v", err)
		}

		expected := "<p>This is a paragraph.</p>"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("Heading", func(t *testing.T) {
		h := Heading(2, "Chapter Title")
		html, err := renderer.RenderToHTML(h)
		if err != nil {
			t.Fatalf("Error rendering heading: %v", err)
		}

		expected := "<h2>Chapter Title</h2>"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("Link", func(t *testing.T) {
		link := Link("https://example.com", "Example")
		html, err := renderer.RenderToHTML(link)
		if err != nil {
			t.Fatalf("Error rendering link: %v", err)
		}

		expected := `<a href="https://example.com">Example</a>`
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("Image", func(t *testing.T) {
		img := Image("test.jpg", "Test Image")
		html, err := renderer.RenderToHTML(img)
		if err != nil {
			t.Fatalf("Error rendering image: %v", err)
		}

		expected := `<img src="test.jpg" alt="Test Image" />`
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("List", func(t *testing.T) {
		items := []string{"Item 1", "Item 2", "Item 3"}
		ul := List(false, items)
		html, err := renderer.RenderToHTML(ul)
		if err != nil {
			t.Fatalf("Error rendering unordered list: %v", err)
		}

		expected := "<ul><li>Item 1</li><li>Item 2</li><li>Item 3</li></ul>"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}

		ol := List(true, items)
		html, err = renderer.RenderToHTML(ol)
		if err != nil {
			t.Fatalf("Error rendering ordered list: %v", err)
		}

		expected = "<ol><li>Item 1</li><li>Item 2</li><li>Item 3</li></ol>"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})

	t.Run("Table", func(t *testing.T) {
		headers := []string{"Name", "Age"}
		rows := [][]string{
			{"Alice", "30"},
			{"Bob", "25"},
		}

		table := Table(headers, rows)
		html, err := renderer.RenderToHTML(table)
		if err != nil {
			t.Fatalf("Error rendering table: %v", err)
		}

		expected := "<table><thead><tr><th>Name</th><th>Age</th></tr></thead><tbody><tr><td>Alice</td><td>30</td></tr><tr><td>Bob</td><td>25</td></tr></tbody></table>"
		if html != expected {
			t.Errorf("Expected %q, got %q", expected, html)
		}
	})
}

// MockVisitor for testing the visitor pattern
type MockVisitor struct {
	visitedNodes []string
}

func (mv *MockVisitor) VisitRoot(*Root) error {
	mv.visitedNodes = append(mv.visitedNodes, "root")
	return nil
}

func (mv *MockVisitor) VisitElement(e *Element) error {
	mv.visitedNodes = append(mv.visitedNodes, "element:"+e.TagName)
	return nil
}

func (mv *MockVisitor) VisitText(t *Text) error {
	mv.visitedNodes = append(mv.visitedNodes, "text:"+t.Value)
	return nil
}

func (mv *MockVisitor) VisitComment(c *Comment) error {
	mv.visitedNodes = append(mv.visitedNodes, "comment:"+c.Value)
	return nil
}

func TestTreeWalker(t *testing.T) {
	// Build a simple tree
	root := NewRoot()
	div := NewElement("div")
	p := NewElement("p")
	text := NewText("Hello")
	comment := NewComment("Test comment")

	AddChild(p, text)
	AddChild(div, p)
	AddChild(div, comment)
	AddChild(root, div)

	// Create walker with mock visitor
	walker := NewTreeWalker()
	visitor := &MockVisitor{}
	walker.AddVisitor(visitor)

	// Walk the tree
	err := walker.Walk(root)
	if err != nil {
		t.Fatalf("Error walking tree: %v", err)
	}

	// Check that all nodes were visited
	expected := []string{
		"root",
		"element:div",
		"element:p",
		"text:Hello",
		"comment:Test comment",
	}

	if len(visitor.visitedNodes) != len(expected) {
		t.Errorf("Expected %d visited nodes, got %d", len(expected), len(visitor.visitedNodes))
	}

	for i, expectedNode := range expected {
		if i >= len(visitor.visitedNodes) || visitor.visitedNodes[i] != expectedNode {
			t.Errorf("Expected node %d to be %q, got %q", i, expectedNode, visitor.visitedNodes[i])
		}
	}
}
