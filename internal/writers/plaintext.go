package writers

import (
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/events"
)

// PlainTextWriter generates raw plain text from events for debugging
type PlainTextWriter struct {
	out    *strings.Builder
	inList bool
}

// NewPlainTextWriter creates a new plain text writer
func NewPlainTextWriter() *PlainTextWriter {
	return &PlainTextWriter{
		out: &strings.Builder{},
	}
}

// Handle processes a single event
func (w *PlainTextWriter) Handle(event events.Event) {
	switch event.Kind {
	case events.StartDoc:
		title := event.Arg.(string)
		w.out.WriteString("=== DOCUMENT START ===\n")
		w.out.WriteString("TITLE: ")
		w.out.WriteString(title)
		w.out.WriteString("\n\n")

	case events.EndDoc:
		w.out.WriteString("\n=== DOCUMENT END ===\n")

	case events.StartParagraph:
		w.out.WriteString("[PARA_START]")

	case events.EndParagraph:
		w.out.WriteString("[PARA_END]\n")

	case events.StartHeading:
		info := event.Arg.(events.HeadingInfo)
		w.out.WriteString("[HEADING_START:")
		w.out.WriteString(string(rune('0' + info.Level)))
		w.out.WriteString(":")
		w.out.WriteString(info.AnchorID)
		w.out.WriteString("] ")
		w.out.WriteString(info.Text)

	case events.EndHeading:
		w.out.WriteString(" [HEADING_END]\n")

	case events.StartList:
		w.inList = true
		w.out.WriteString("[LIST_START]")

	case events.EndList:
		w.inList = false
		w.out.WriteString("[LIST_END]\n")

	case events.StartTable:
		w.out.WriteString("[TABLE_START]\n")

	case events.EndTable:
		w.out.WriteString("[TABLE_END]\n")

	case events.StartTableRow:
		w.out.WriteString("[ROW_START]")

	case events.EndTableRow:
		w.out.WriteString("[ROW_END]\n")

	case events.StartTableCell:
		w.out.WriteString("[CELL_START]")

	case events.EndTableCell:
		w.out.WriteString("[CELL_END]")

	case events.StartFormatting:
		info := event.Arg.(events.FormatInfo)
		w.out.WriteString("[FORMAT_START:")
		w.out.WriteString(w.styleToString(info.Style))
		if info.URL != "" {
			w.out.WriteString(":")
			w.out.WriteString(info.URL)
		}
		w.out.WriteString("]")

	case events.EndFormatting:
		info := event.Arg.(events.FormatInfo)
		w.out.WriteString("[FORMAT_END:")
		w.out.WriteString(w.styleToString(info.Style))
		w.out.WriteString("]")

	case events.Text:
		text := event.Arg.(string)
		if w.inList {
			w.out.WriteString("[LIST_ITEM]")
		}
		// Show the raw text with visible newlines and spaces
		escaped := w.escapeWhitespace(text)
		w.out.WriteString(escaped)

	case events.Image:
		info := event.Arg.(events.ImageInfo)
		w.out.WriteString("[IMAGE:")
		w.out.WriteString(info.URL)
		w.out.WriteString(":")
		w.out.WriteString(info.Alt)
		w.out.WriteString("]")
	}
}

// styleToString converts a style bit flag to readable string
func (w *PlainTextWriter) styleToString(style events.Style) string {
	var parts []string
	if style.Has(events.Bold) {
		parts = append(parts, "BOLD")
	}
	if style.Has(events.Italic) {
		parts = append(parts, "ITALIC")
	}
	if style.Has(events.Underline) {
		parts = append(parts, "UNDERLINE")
	}
	if style.Has(events.Strike) {
		parts = append(parts, "STRIKE")
	}
	if style.Has(events.Highlight) {
		parts = append(parts, "HIGHLIGHT")
	}
	if style.Has(events.Sub) {
		parts = append(parts, "SUB")
	}
	if style.Has(events.Sup) {
		parts = append(parts, "SUP")
	}
	if style.Has(events.Link) {
		parts = append(parts, "LINK")
	}
	if len(parts) == 0 {
		return "NONE"
	}
	return strings.Join(parts, "+")
}

// escapeWhitespace makes whitespace characters visible
func (w *PlainTextWriter) escapeWhitespace(text string) string {
	// Replace newlines with visible markers
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	text = strings.ReplaceAll(text, "\t", "\\t")

	// Show multiple spaces
	if strings.Contains(text, "  ") {
		// Mark sequences of spaces
		for strings.Contains(text, "  ") {
			text = strings.ReplaceAll(text, "  ", " •")
		}
	}

	return text
}

// Result returns the final plain text string
func (w *PlainTextWriter) Result() string {
	return w.out.String()
}

// Reset clears the writer state for reuse
func (w *PlainTextWriter) Reset() {
	w.out.Reset()
	w.inList = false
}

// SetOutput sets the output destination (for StreamWriter interface)
func (w *PlainTextWriter) SetOutput(writer io.Writer) {
	// For string-based writers, we ignore this
	// The Result() method returns the final string
}
