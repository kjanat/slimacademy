package hast

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
)

// HTMLRenderer provides methods for converting HAST to HTML
type HTMLRenderer struct {
	// Configuration options
	escapeText  bool
	validateURL bool
}

// NewHTMLRenderer creates a new HTML renderer with default options
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{
		escapeText:  true,
		validateURL: true,
	}
}

// RenderToHTML converts a HAST tree to HTML string
func (r *HTMLRenderer) RenderToHTML(node Node) (string, error) {
	var builder strings.Builder
	err := r.renderNode(node, &builder)
	return builder.String(), err
}

// renderNode recursively renders a HAST node to HTML
func (r *HTMLRenderer) renderNode(node Node, builder *strings.Builder) error {
	switch n := node.(type) {
	case *Root:
		return r.renderRoot(n, builder)
	case *Element:
		return r.renderElement(n, builder)
	case *Text:
		return r.renderText(n, builder)
	case *Comment:
		return r.renderComment(n, builder)
	default:
		return fmt.Errorf("unknown HAST node type: %T", node)
	}
}

// renderRoot renders a root node (just renders children)
func (r *HTMLRenderer) renderRoot(root *Root, builder *strings.Builder) error {
	for _, child := range root.Children {
		if err := r.renderNode(child, builder); err != nil {
			return err
		}
	}
	return nil
}

// renderElement renders an HTML element
func (r *HTMLRenderer) renderElement(element *Element, builder *strings.Builder) error {
	// Handle void elements (self-closing tags)
	isVoid := r.isVoidElement(element.TagName)

	// Open tag
	builder.WriteString("<")
	builder.WriteString(element.TagName)

	// Render attributes
	if err := r.renderProperties(element.Properties, builder); err != nil {
		return err
	}

	if isVoid {
		builder.WriteString(" />")
		return nil
	}

	builder.WriteString(">")

	// Render children
	for _, child := range element.Children {
		if err := r.renderNode(child, builder); err != nil {
			return err
		}
	}

	// Close tag
	builder.WriteString("</")
	builder.WriteString(element.TagName)
	builder.WriteString(">")

	return nil
}

// renderText renders a text node
func (r *HTMLRenderer) renderText(text *Text, builder *strings.Builder) error {
	if r.escapeText {
		escaped := template.HTMLEscapeString(text.Value)
		builder.WriteString(escaped)
	} else {
		builder.WriteString(text.Value)
	}
	return nil
}

// renderComment renders an HTML comment
func (r *HTMLRenderer) renderComment(comment *Comment, builder *strings.Builder) error {
	builder.WriteString("<!-- ")
	// Comments should be escaped to prevent injection
	escaped := template.HTMLEscapeString(comment.Value)
	builder.WriteString(escaped)
	builder.WriteString(" -->")
	return nil
}

// renderProperties renders HTML attributes from HAST properties
func (r *HTMLRenderer) renderProperties(properties map[string]any, builder *strings.Builder) error {
	for key, value := range properties {
		builder.WriteString(" ")

		// Handle special properties
		switch key {
		case "className":
			builder.WriteString("class")
		case "htmlFor":
			builder.WriteString("for")
		default:
			builder.WriteString(key)
		}

		// Convert value to string and escape it
		strValue := r.propertyValueToString(value)

		// Special handling for URLs in href and src attributes
		if (key == "href" || key == "src") && r.validateURL {
			strValue = r.sanitizeURL(strValue)
		}

		builder.WriteString("=\"")
		escaped := template.HTMLEscapeString(strValue)
		builder.WriteString(escaped)
		builder.WriteString("\"")
	}
	return nil
}

// propertyValueToString converts a property value to string
func (r *HTMLRenderer) propertyValueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []string:
		return strings.Join(v, " ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// sanitizeURL validates and sanitizes URLs for security
func (r *HTMLRenderer) sanitizeURL(urlStr string) string {
	if urlStr == "" {
		return "#"
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "#"
	}

	// Allow only safe schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	switch scheme {
	case "http", "https", "mailto", "tel":
		return parsedURL.String()
	case "":
		// Relative URLs - check for dangerous prefixes
		lowerURL := strings.ToLower(urlStr)
		if strings.HasPrefix(lowerURL, "javascript:") ||
			strings.HasPrefix(lowerURL, "data:") ||
			strings.HasPrefix(lowerURL, "vbscript:") {
			return "#"
		}
		return urlStr
	default:
		return "#"
	}
}

// isVoidElement checks if an HTML tag is a void (self-closing) element
func (r *HTMLRenderer) isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area":   true,
		"base":   true,
		"br":     true,
		"col":    true,
		"embed":  true,
		"hr":     true,
		"img":    true,
		"input":  true,
		"link":   true,
		"meta":   true,
		"param":  true,
		"source": true,
		"track":  true,
		"wbr":    true,
	}
	return voidElements[strings.ToLower(tagName)]
}

// TreeWalker provides utilities for traversing HAST trees
type TreeWalker struct {
	visitors []Visitor
}

// NewTreeWalker creates a new tree walker
func NewTreeWalker() *TreeWalker {
	return &TreeWalker{
		visitors: make([]Visitor, 0),
	}
}

// AddVisitor adds a visitor to the walker
func (tw *TreeWalker) AddVisitor(visitor Visitor) {
	tw.visitors = append(tw.visitors, visitor)
}

// Walk traverses the HAST tree and applies all visitors
func (tw *TreeWalker) Walk(node Node) error {
	// Apply all visitors to current node
	for _, visitor := range tw.visitors {
		if err := node.Accept(visitor); err != nil {
			return err
		}
	}

	// Recursively walk children
	switch n := node.(type) {
	case *Root:
		for _, child := range n.Children {
			if err := tw.Walk(child); err != nil {
				return err
			}
		}
	case *Element:
		for _, child := range n.Children {
			if err := tw.Walk(child); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper functions for building common HAST structures

// Paragraph creates a paragraph element with text content
func Paragraph(text string) *Element {
	p := NewElement("p")
	AddChild(p, NewText(text))
	return p
}

// Heading creates a heading element (h1-h6) with text content
func Heading(level int, text string) *Element {
	if level < 1 || level > 6 {
		level = 1
	}
	h := NewElement(fmt.Sprintf("h%d", level))
	AddChild(h, NewText(text))
	return h
}

// Link creates an anchor element with href and text
func Link(href, text string) *Element {
	a := NewElement("a")
	a.SetProperty("href", href)
	AddChild(a, NewText(text))
	return a
}

// Image creates an img element with src and alt
func Image(src, alt string) *Element {
	img := NewElement("img")
	img.SetProperty("src", src)
	img.SetProperty("alt", alt)
	return img
}

// List creates a ul or ol element with list items
func List(ordered bool, items []string) *Element {
	var listElement *Element
	if ordered {
		listElement = NewElement("ol")
	} else {
		listElement = NewElement("ul")
	}

	for _, item := range items {
		li := NewElement("li")
		AddChild(li, NewText(item))
		AddChild(listElement, li)
	}

	return listElement
}

// Table creates a table element with headers and rows
func Table(headers []string, rows [][]string) *Element {
	table := NewElement("table")

	// Create header
	if len(headers) > 0 {
		thead := NewElement("thead")
		tr := NewElement("tr")

		for _, header := range headers {
			th := NewElement("th")
			AddChild(th, NewText(header))
			AddChild(tr, th)
		}

		AddChild(thead, tr)
		AddChild(table, thead)
	}

	// Create body
	if len(rows) > 0 {
		tbody := NewElement("tbody")

		for _, row := range rows {
			tr := NewElement("tr")

			for _, cell := range row {
				td := NewElement("td")
				AddChild(td, NewText(cell))
				AddChild(tr, td)
			}

			AddChild(tbody, tr)
		}

		AddChild(table, tbody)
	}

	return table
}
