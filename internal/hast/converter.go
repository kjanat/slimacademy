package hast

import (
	"fmt"
	"strings"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// EventToHASTConverter converts SlimAcademy streaming events to HAST
type EventToHASTConverter struct {
	// Stack for managing nested elements
	elementStack []*Element
	// Current root
	root *Root
	// Current element being built
	current Node
	// Options for conversion
	options ConversionOptions
}

// ConversionOptions provides configuration for the conversion process
type ConversionOptions struct {
	// IncludeMetadata includes document metadata in the output
	IncludeMetadata bool
	// GenerateIDs automatically generates IDs for headings
	GenerateIDs bool
	// SanitizeURLs enables URL sanitization
	SanitizeURLs bool
	// IncludeStyles includes style information
	IncludeStyles bool
}

// DefaultConversionOptions returns sensible defaults
func DefaultConversionOptions() ConversionOptions {
	return ConversionOptions{
		IncludeMetadata: true,
		GenerateIDs:     true,
		SanitizeURLs:    true,
		IncludeStyles:   true,
	}
}

// NewEventToHASTConverter creates a new converter
func NewEventToHASTConverter(options ConversionOptions) *EventToHASTConverter {
	return &EventToHASTConverter{
		elementStack: make([]*Element, 0),
		root:         NewRoot(),
		current:      nil,
		options:      options,
	}
}

// Convert processes a stream of events and returns a HAST tree
func (c *EventToHASTConverter) Convert(events []streaming.Event) (*Root, error) {
	c.reset()

	for _, event := range events {
		if err := c.processEvent(event); err != nil {
			return nil, fmt.Errorf("error processing event %v: %w", event.Kind, err)
		}
	}

	return c.root, nil
}

// reset prepares the converter for a new document
func (c *EventToHASTConverter) reset() {
	c.elementStack = c.elementStack[:0]
	c.root = NewRoot()
	c.current = c.root
}

// processEvent handles a single streaming event
func (c *EventToHASTConverter) processEvent(event streaming.Event) error {
	switch event.Kind {
	case streaming.StartDoc:
		return c.handleStartDoc(event)
	case streaming.EndDoc:
		return c.handleEndDoc(event)
	case streaming.StartParagraph:
		return c.handleStartParagraph(event)
	case streaming.EndParagraph:
		return c.handleEndParagraph(event)
	case streaming.StartHeading:
		return c.handleStartHeading(event)
	case streaming.EndHeading:
		return c.handleEndHeading(event)
	case streaming.StartList:
		return c.handleStartList(event)
	case streaming.EndList:
		return c.handleEndList(event)
	case streaming.StartListItem:
		return c.handleStartListItem(event)
	case streaming.EndListItem:
		return c.handleEndListItem(event)
	case streaming.StartTable:
		return c.handleStartTable(event)
	case streaming.EndTable:
		return c.handleEndTable(event)
	case streaming.StartTableRow:
		return c.handleStartTableRow(event)
	case streaming.EndTableRow:
		return c.handleEndTableRow(event)
	case streaming.StartTableCell:
		return c.handleStartTableCell(event)
	case streaming.EndTableCell:
		return c.handleEndTableCell(event)
	case streaming.StartFormatting:
		return c.handleStartFormatting(event)
	case streaming.EndFormatting:
		return c.handleEndFormatting(event)
	case streaming.Text:
		return c.handleText(event)
	case streaming.Image:
		return c.handleImage(event)
	default:
		return fmt.Errorf("unknown event kind: %v", event.Kind)
	}
}

// Document-level handlers

func (c *EventToHASTConverter) handleStartDoc(event streaming.Event) error {
	if c.options.IncludeMetadata {
		// Add document metadata as meta elements or data attributes
		if event.Title != "" {
			title := NewElement("title")
			AddChild(title, NewText(event.Title))
			AddChild(c.root, title)
		}

		// Add metadata as data attributes to a wrapper div
		if event.Description != "" || len(event.Periods) > 0 {
			wrapper := NewElement("div")
			wrapper.SetProperty("class", "document-metadata")

			if event.Description != "" {
				wrapper.SetProperty("data-description", event.Description)
			}

			if len(event.Periods) > 0 {
				wrapper.SetProperty("data-periods", strings.Join(event.Periods, ","))
			}

			if event.BachelorYearNumber != "" {
				wrapper.SetProperty("data-bachelor-year", event.BachelorYearNumber)
			}

			AddChild(c.root, wrapper)
			c.pushElement(wrapper)
		}
	}
	return nil
}

func (c *EventToHASTConverter) handleEndDoc(event streaming.Event) error {
	// Close any remaining open elements
	for len(c.elementStack) > 0 {
		c.popElement()
	}
	return nil
}

// Block-level handlers

func (c *EventToHASTConverter) handleStartParagraph(event streaming.Event) error {
	p := NewElement("p")
	c.addToCurrentParent(p)
	c.pushElement(p)
	return nil
}

func (c *EventToHASTConverter) handleEndParagraph(event streaming.Event) error {
	c.popElement()
	return nil
}

func (c *EventToHASTConverter) handleStartHeading(event streaming.Event) error {
	level := event.Level
	if level < 1 || level > 6 {
		level = 1
	}

	h := NewElement(fmt.Sprintf("h%d", level))

	// Add ID if specified or auto-generate
	if event.AnchorID != "" {
		h.SetProperty("id", event.AnchorID)
	} else if c.options.GenerateIDs {
		// Auto-generate ID from heading content (will be set after text is processed)
		h.SetProperty("data-auto-id", "true")
	}

	c.addToCurrentParent(h)
	c.pushElement(h)
	return nil
}

func (c *EventToHASTConverter) handleEndHeading(event streaming.Event) error {
	// If auto-ID was requested, generate it from the text content
	if current := c.getCurrentElement(); current != nil {
		if current.HasProperty("data-auto-id") {
			current.Properties["data-auto-id"] = nil // Remove the marker
			delete(current.Properties, "data-auto-id")

			// Extract text content and generate ID
			textContent := c.extractTextFromElement(current)
			if textContent != "" {
				id := c.generateSlug(textContent)
				current.SetProperty("id", id)
			}
		}
	}

	c.popElement()
	return nil
}

// List handlers

func (c *EventToHASTConverter) handleStartList(event streaming.Event) error {
	// For now, assume unordered lists (could be enhanced with list type detection)
	ul := NewElement("ul")
	c.addToCurrentParent(ul)
	c.pushElement(ul)
	return nil
}

func (c *EventToHASTConverter) handleEndList(event streaming.Event) error {
	c.popElement()
	return nil
}

func (c *EventToHASTConverter) handleStartListItem(event streaming.Event) error {
	li := NewElement("li")
	c.addToCurrentParent(li)
	c.pushElement(li)
	return nil
}

func (c *EventToHASTConverter) handleEndListItem(event streaming.Event) error {
	c.popElement()
	return nil
}

// Table handlers

func (c *EventToHASTConverter) handleStartTable(event streaming.Event) error {
	table := NewElement("table")
	if c.options.IncludeStyles {
		table.SetProperty("style", "border-collapse: collapse; width: 100%; margin: 20px 0;")
	}
	c.addToCurrentParent(table)
	c.pushElement(table)
	return nil
}

func (c *EventToHASTConverter) handleEndTable(event streaming.Event) error {
	c.popElement()
	return nil
}

func (c *EventToHASTConverter) handleStartTableRow(event streaming.Event) error {
	tr := NewElement("tr")
	c.addToCurrentParent(tr)
	c.pushElement(tr)
	return nil
}

func (c *EventToHASTConverter) handleEndTableRow(event streaming.Event) error {
	c.popElement()
	return nil
}

func (c *EventToHASTConverter) handleStartTableCell(event streaming.Event) error {
	// Determine if this is a header cell (could be enhanced with better detection)
	tagName := "td"

	// Simple heuristic: if this is the first row in the table, use th
	if parent := c.getCurrentElement(); parent != nil && parent.TagName == "tr" {
		if grandparent := c.getParentElement(); grandparent != nil && grandparent.TagName == "table" {
			if len(grandparent.Children) == 1 { // First row
				tagName = "th"
			}
		}
	}

	cell := NewElement(tagName)

	if c.options.IncludeStyles {
		style := "border: 1px solid #ddd; padding: 8px;"
		if tagName == "th" {
			style += " background-color: #f2f2f2; font-weight: bold;"
		}
		cell.SetProperty("style", style)
	}

	c.addToCurrentParent(cell)
	c.pushElement(cell)
	return nil
}

func (c *EventToHASTConverter) handleEndTableCell(event streaming.Event) error {
	c.popElement()
	return nil
}

// Inline formatting handlers

func (c *EventToHASTConverter) handleStartFormatting(event streaming.Event) error {
	style := event.Style

	// Handle each style flag
	if style&streaming.Bold != 0 {
		strong := NewElement("strong")
		c.addToCurrentParent(strong)
		c.pushElement(strong)
	}

	if style&streaming.Italic != 0 {
		em := NewElement("em")
		c.addToCurrentParent(em)
		c.pushElement(em)
	}

	if style&streaming.Underline != 0 {
		u := NewElement("u")
		c.addToCurrentParent(u)
		c.pushElement(u)
	}

	if style&streaming.Strike != 0 {
		s := NewElement("s")
		c.addToCurrentParent(s)
		c.pushElement(s)
	}

	if style&streaming.Highlight != 0 {
		mark := NewElement("mark")
		c.addToCurrentParent(mark)
		c.pushElement(mark)
	}

	if style&streaming.Sub != 0 {
		sub := NewElement("sub")
		c.addToCurrentParent(sub)
		c.pushElement(sub)
	}

	if style&streaming.Sup != 0 {
		sup := NewElement("sup")
		c.addToCurrentParent(sup)
		c.pushElement(sup)
	}

	if style&streaming.Link != 0 {
		a := NewElement("a")
		href := event.LinkURL
		if c.options.SanitizeURLs {
			href = c.sanitizeURL(href)
		}
		a.SetProperty("href", href)
		c.addToCurrentParent(a)
		c.pushElement(a)
	}

	return nil
}

func (c *EventToHASTConverter) handleEndFormatting(event streaming.Event) error {
	style := event.Style

	// Pop elements in reverse order of how they were pushed
	var elementsToClose []streaming.StyleFlags

	if style&streaming.Link != 0 {
		elementsToClose = append(elementsToClose, streaming.Link)
	}
	if style&streaming.Sup != 0 {
		elementsToClose = append(elementsToClose, streaming.Sup)
	}
	if style&streaming.Sub != 0 {
		elementsToClose = append(elementsToClose, streaming.Sub)
	}
	if style&streaming.Highlight != 0 {
		elementsToClose = append(elementsToClose, streaming.Highlight)
	}
	if style&streaming.Strike != 0 {
		elementsToClose = append(elementsToClose, streaming.Strike)
	}
	if style&streaming.Underline != 0 {
		elementsToClose = append(elementsToClose, streaming.Underline)
	}
	if style&streaming.Italic != 0 {
		elementsToClose = append(elementsToClose, streaming.Italic)
	}
	if style&streaming.Bold != 0 {
		elementsToClose = append(elementsToClose, streaming.Bold)
	}

	// Close elements in reverse order
	for i := len(elementsToClose) - 1; i >= 0; i-- {
		c.popElement()
	}

	return nil
}

// Content handlers

func (c *EventToHASTConverter) handleText(event streaming.Event) error {
	text := NewText(event.TextContent)
	c.addToCurrentParent(text)
	return nil
}

func (c *EventToHASTConverter) handleImage(event streaming.Event) error {
	img := NewElement("img")

	src := event.ImageURL
	if c.options.SanitizeURLs {
		src = c.sanitizeURL(src)
	}

	img.SetProperty("src", src)
	img.SetProperty("alt", event.ImageAlt)

	if c.options.IncludeStyles {
		img.SetProperty("style", "max-width: 100%; height: auto;")
	}

	c.addToCurrentParent(img)
	return nil
}

// Helper methods for element stack management

func (c *EventToHASTConverter) pushElement(element *Element) {
	c.elementStack = append(c.elementStack, element)
}

func (c *EventToHASTConverter) popElement() *Element {
	if len(c.elementStack) == 0 {
		return nil
	}

	element := c.elementStack[len(c.elementStack)-1]
	c.elementStack = c.elementStack[:len(c.elementStack)-1]
	return element
}

func (c *EventToHASTConverter) getCurrentElement() *Element {
	if len(c.elementStack) == 0 {
		return nil
	}
	return c.elementStack[len(c.elementStack)-1]
}

func (c *EventToHASTConverter) getParentElement() *Element {
	if len(c.elementStack) < 2 {
		return nil
	}
	return c.elementStack[len(c.elementStack)-2]
}

func (c *EventToHASTConverter) addToCurrentParent(node Node) {
	if len(c.elementStack) > 0 {
		current := c.getCurrentElement()
		AddChild(current, node)
	} else {
		AddChild(c.root, node)
	}
}

// Utility methods

func (c *EventToHASTConverter) sanitizeURL(url string) string {
	// Use the same logic as the HTML renderer
	renderer := NewHTMLRenderer()
	return renderer.sanitizeURL(url)
}

func (c *EventToHASTConverter) generateSlug(text string) string {
	// Simple slug generation (could be enhanced)
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (c *EventToHASTConverter) extractTextFromElement(element *Element) string {
	var result strings.Builder

	for _, child := range element.Children {
		switch n := child.(type) {
		case *Text:
			result.WriteString(n.Value)
		case *Element:
			result.WriteString(c.extractTextFromElement(n))
		}
	}

	return result.String()
}
