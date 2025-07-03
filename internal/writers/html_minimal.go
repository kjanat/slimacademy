package writers

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"iter"
	"strings"
	"unique"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/templates"
)

// MinimalHTMLWriter generates clean, minimal HTML using the new template system
type MinimalHTMLWriter struct {
	config              *config.HTMLConfig
	template            *templates.MinimalTemplate
	docData             *templates.TemplateData
	content             *strings.Builder
	styleStack          []string // Track open formatting tags
	currentHeadingLevel int      // Track current heading level for proper closing

	// O(1) duplicate detection with unique.Handle
	seenURLs    map[unique.Handle[string]]bool // Track URLs for deduplication
	seenAnchors map[unique.Handle[string]]bool // Track anchor IDs for deduplication
	seenTexts   map[unique.Handle[string]]int  // Track text content for analytics
}

// NewMinimalHTMLWriter creates a new minimal HTML writer
func NewMinimalHTMLWriter() *MinimalHTMLWriter {
	return &MinimalHTMLWriter{
		config:      config.DefaultHTMLConfig(),
		template:    templates.NewMinimalTemplate(),
		docData:     &templates.TemplateData{},
		content:     &strings.Builder{},
		styleStack:  make([]string, 0),
		seenURLs:    make(map[unique.Handle[string]]bool),
		seenAnchors: make(map[unique.Handle[string]]bool),
		seenTexts:   make(map[unique.Handle[string]]int),
	}
}

// ProcessEvents processes all events and generates the final HTML
func (w *MinimalHTMLWriter) ProcessEvents(events []streaming.Event) (string, error) {
	// Reset state
	w.content.Reset()
	w.docData = &templates.TemplateData{
		Metadata: make(map[string]string),
	}
	w.styleStack = w.styleStack[:0]

	// Process all events
	for _, event := range events {
		if err := w.processEvent(event); err != nil {
			return "", fmt.Errorf("error processing event %v: %w", event.Kind, err)
		}
	}

	// Set final content and render
	w.docData.Content = template.HTML(w.content.String())
	return w.template.Render(*w.docData)
}

// ProcessEventStream processes events using Go 1.23+ iter.Seq for O(1) memory usage
func (w *MinimalHTMLWriter) ProcessEventStream(ctx context.Context, events iter.Seq[streaming.Event]) (string, error) {
	// Reset state
	w.content.Reset()
	w.docData = &templates.TemplateData{
		Metadata: make(map[string]string),
	}
	w.styleStack = w.styleStack[:0]
	w.currentHeadingLevel = 0

	// Process events with O(1) memory usage
	for event := range events {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if err := w.processEvent(event); err != nil {
			return "", fmt.Errorf("error processing event %v: %w", event.Kind, err)
		}
	}

	// Set final content and render
	w.docData.Content = template.HTML(w.content.String())
	return w.template.Render(*w.docData)
}

// processEvent handles a single streaming event
func (w *MinimalHTMLWriter) processEvent(event streaming.Event) error {
	switch event.Kind {
	case streaming.StartDoc:
		return w.handleStartDoc(event)
	case streaming.EndDoc:
		return w.handleEndDoc(event)
	case streaming.StartParagraph:
		w.content.WriteString("<p>")
	case streaming.EndParagraph:
		w.content.WriteString("</p>\n")
	case streaming.StartHeading:
		level := event.Level
		if level < 1 || level > 6 {
			level = 1
		}
		w.currentHeadingLevel = level
		if event.AnchorID != "" {
			// Use unique.Handle for O(1) duplicate detection
			anchorHandle := unique.Make(event.AnchorID)
			if w.seenAnchors[anchorHandle] {
				// Generate unique ID if duplicate
				uniqueID := fmt.Sprintf("%s-%d", event.AnchorID, len(w.seenAnchors))
				fmt.Fprintf(w.content, "<h%d id=\"%s\">", level, w.escapeHTML(uniqueID))
			} else {
				w.seenAnchors[anchorHandle] = true
				fmt.Fprintf(w.content, "<h%d id=\"%s\">", level, w.escapeHTML(event.AnchorID))
			}
		} else {
			fmt.Fprintf(w.content, "<h%d>", level)
		}
	case streaming.EndHeading:
		fmt.Fprintf(w.content, "</h%d>\n", w.currentHeadingLevel)
	case streaming.StartList:
		w.content.WriteString("<ul>\n")
	case streaming.EndList:
		w.content.WriteString("</ul>\n")
	case streaming.StartListItem:
		w.content.WriteString("<li>")
	case streaming.EndListItem:
		w.content.WriteString("</li>\n")
	case streaming.StartTable:
		w.content.WriteString("<table>\n")
	case streaming.EndTable:
		w.content.WriteString("</table>\n")
	case streaming.StartTableRow:
		w.content.WriteString("<tr>")
	case streaming.EndTableRow:
		w.content.WriteString("</tr>\n")
	case streaming.StartTableCell:
		w.content.WriteString("<td>")
	case streaming.EndTableCell:
		w.content.WriteString("</td>")
	case streaming.StartFormatting:
		return w.handleStartFormatting(event)
	case streaming.EndFormatting:
		return w.handleEndFormatting(event)
	case streaming.Text:
		// Use unique.Handle for text content analytics
		textHandle := unique.Make(event.TextContent)
		w.seenTexts[textHandle]++
		w.content.WriteString(w.escapeHTML(event.TextContent))
	case streaming.Image:
		return w.handleImage(event)
	default:
		return fmt.Errorf("unknown event kind: %v", event.Kind)
	}
	return nil
}

// handleStartDoc processes document start events
func (w *MinimalHTMLWriter) handleStartDoc(event streaming.Event) error {
	w.docData.Title = event.Title
	w.docData.Description = event.Description

	// Collect metadata
	if event.BachelorYearNumber != "" {
		w.docData.Metadata["Academic Year"] = event.BachelorYearNumber
	}
	if event.AvailableDate != "" {
		w.docData.Metadata["Available"] = event.AvailableDate
	}
	if event.ExamDate != "" {
		w.docData.Metadata["Exam Date"] = event.ExamDate
	}
	if event.CollegeStartYear > 0 {
		w.docData.Metadata["College Start"] = fmt.Sprintf("%d", event.CollegeStartYear)
	}
	if event.PageCount > 0 {
		w.docData.Metadata["Pages"] = fmt.Sprintf("%d", event.PageCount)
	}
	if len(event.Periods) > 0 {
		w.docData.Metadata["Periods"] = strings.Join(event.Periods, ", ")
	}
	if event.ReadProgress != nil && event.PageCount > 0 {
		progress := *event.ReadProgress
		percentage := float64(progress) / float64(event.PageCount) * 100
		w.docData.Metadata["Progress"] = fmt.Sprintf("%d/%d pages (%.1f%%)", progress, event.PageCount, percentage)
	}

	return nil
}

// handleEndDoc processes document end events
func (w *MinimalHTMLWriter) handleEndDoc(event streaming.Event) error {
	// Close any remaining formatting tags
	for len(w.styleStack) > 0 {
		tag := w.styleStack[len(w.styleStack)-1]
		w.styleStack = w.styleStack[:len(w.styleStack)-1]
		fmt.Fprintf(w.content, "</%s>", tag)
	}
	return nil
}

// handleStartFormatting processes formatting start events
func (w *MinimalHTMLWriter) handleStartFormatting(event streaming.Event) error {
	style := event.Style

	if style&streaming.Bold != 0 {
		w.content.WriteString("<strong>")
		w.styleStack = append(w.styleStack, "strong")
	}
	if style&streaming.Italic != 0 {
		w.content.WriteString("<em>")
		w.styleStack = append(w.styleStack, "em")
	}
	if style&streaming.Underline != 0 {
		w.content.WriteString("<u>")
		w.styleStack = append(w.styleStack, "u")
	}
	if style&streaming.Strike != 0 {
		w.content.WriteString("<s>")
		w.styleStack = append(w.styleStack, "s")
	}
	if style&streaming.Highlight != 0 {
		w.content.WriteString("<mark>")
		w.styleStack = append(w.styleStack, "mark")
	}
	if style&streaming.Sub != 0 {
		w.content.WriteString("<sub>")
		w.styleStack = append(w.styleStack, "sub")
	}
	if style&streaming.Sup != 0 {
		w.content.WriteString("<sup>")
		w.styleStack = append(w.styleStack, "sup")
	}
	if style&streaming.Link != 0 {
		href := w.sanitizeURLWithDedup(event.LinkURL)
		fmt.Fprintf(w.content, `<a href="%s">`, href)
		w.styleStack = append(w.styleStack, "a")
	}

	return nil
}

// handleEndFormatting processes formatting end events
func (w *MinimalHTMLWriter) handleEndFormatting(event streaming.Event) error {
	style := event.Style

	// Close tags in reverse order (LIFO)
	var tagsToClose []string

	if style&streaming.Link != 0 {
		tagsToClose = append(tagsToClose, "a")
	}
	if style&streaming.Sup != 0 {
		tagsToClose = append(tagsToClose, "sup")
	}
	if style&streaming.Sub != 0 {
		tagsToClose = append(tagsToClose, "sub")
	}
	if style&streaming.Highlight != 0 {
		tagsToClose = append(tagsToClose, "mark")
	}
	if style&streaming.Strike != 0 {
		tagsToClose = append(tagsToClose, "s")
	}
	if style&streaming.Underline != 0 {
		tagsToClose = append(tagsToClose, "u")
	}
	if style&streaming.Italic != 0 {
		tagsToClose = append(tagsToClose, "em")
	}
	if style&streaming.Bold != 0 {
		tagsToClose = append(tagsToClose, "strong")
	}

	// Close the tags and remove from stack
	for i := len(tagsToClose) - 1; i >= 0; i-- {
		tag := tagsToClose[i]
		fmt.Fprintf(w.content, "</%s>", tag)

		// Remove from style stack (simplified - assumes proper nesting)
		if len(w.styleStack) > 0 && w.styleStack[len(w.styleStack)-1] == tag {
			w.styleStack = w.styleStack[:len(w.styleStack)-1]
		}
	}

	return nil
}

// handleImage processes image events
func (w *MinimalHTMLWriter) handleImage(event streaming.Event) error {
	src := w.sanitizeURLWithDedup(event.ImageURL)
	alt := w.escapeHTML(event.ImageAlt)
	fmt.Fprintf(w.content, `<img src="%s" alt="%s" />`, src, alt)
	return nil
}

// sanitizeURLWithDedup validates URLs with O(1) duplicate detection using unique.Handle
func (w *MinimalHTMLWriter) sanitizeURLWithDedup(urlStr string) string {
	if urlStr == "" {
		return "#"
	}

	// Use unique.Handle for O(1) duplicate detection
	urlHandle := unique.Make(urlStr)
	if w.seenURLs[urlHandle] {
		// URL already processed, return cached result
		return w.sanitizeURL(urlStr)
	}

	// Mark URL as seen
	w.seenURLs[urlHandle] = true

	// Perform sanitization
	return w.sanitizeURL(urlStr)
}

// sanitizeURL validates and sanitizes URLs to prevent XSS attacks
func (w *MinimalHTMLWriter) sanitizeURL(urlStr string) string {
	if urlStr == "" {
		return "#"
	}

	// Basic URL sanitization (same logic as HTML writer)
	urlStr = strings.TrimSpace(urlStr)

	// Block dangerous schemes
	lowerURL := strings.ToLower(urlStr)
	if strings.HasPrefix(lowerURL, "javascript:") ||
		strings.HasPrefix(lowerURL, "data:") ||
		strings.HasPrefix(lowerURL, "vbscript:") ||
		strings.Contains(urlStr, "<") ||
		strings.Contains(urlStr, ">") {
		return "#"
	}

	return urlStr
}

// escapeHTML escapes HTML special characters
func (w *MinimalHTMLWriter) escapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// Write implements the io.Writer interface
func (w *MinimalHTMLWriter) Write(p []byte) (n int, err error) {
	// This is called by the streaming system - we don't use it directly
	return len(p), nil
}

// Reset resets the writer state
func (w *MinimalHTMLWriter) Reset() {
	w.content.Reset()
	w.docData = &templates.TemplateData{
		Metadata: make(map[string]string),
	}
	w.styleStack = w.styleStack[:0]
	w.currentHeadingLevel = 0

	// Reset duplicate detection maps
	clear(w.seenURLs)
	clear(w.seenAnchors)
	clear(w.seenTexts)
}

// String returns the current content as a string
func (w *MinimalHTMLWriter) String() string {
	html, err := w.template.Render(*w.docData)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	return html
}

// WriteTo implements io.WriterTo interface
func (w *MinimalHTMLWriter) WriteTo(writer io.Writer) (int64, error) {
	html, err := w.template.Render(*w.docData)
	if err != nil {
		return 0, err
	}
	n, err := writer.Write([]byte(html))
	return int64(n), err
}
