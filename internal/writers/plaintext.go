package writers

import (
	"io"
	"strconv"
	"strings"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// init registers the "plaintext" writer, providing a plain text streaming writer with debugging markers and associated metadata.
func init() {
	Register("plaintext", func() WriterV2 {
		return &PlainTextWriterV2{
			PlainTextWriter: NewPlainTextWriter(),
		}
	}, WriterMetadata{
		Name:        "Plain Text",
		Extension:   ".txt",
		Description: "Plain text format for debugging",
		MimeType:    "text/plain",
	})
}

// PlainTextWriter generates raw plain text from events for debugging
type PlainTextWriter struct {
	out    *strings.Builder
	inList bool
}

// NewPlainTextWriter returns a new PlainTextWriter that accumulates plain text output for streaming document events.
func NewPlainTextWriter() *PlainTextWriter {
	return &PlainTextWriter{
		out: &strings.Builder{},
	}
}

// Handle processes a single event
func (w *PlainTextWriter) Handle(event streaming.Event) {
	switch event.Kind {
	case streaming.StartDoc:
		title := event.Title
		w.out.WriteString("=== DOCUMENT START ===\n")
		w.out.WriteString("TITLE: ")
		w.out.WriteString(title)
		w.out.WriteString("\n\n")

	case streaming.EndDoc:
		w.out.WriteString("\n=== DOCUMENT END ===\n")

	case streaming.StartParagraph:
		w.out.WriteString("[PARA_START]")

	case streaming.EndParagraph:
		w.out.WriteString("[PARA_END]\n")

	case streaming.StartHeading:
		w.out.WriteString("[HEADING_START:")
		w.out.WriteString(strconv.Itoa(event.Level))
		w.out.WriteString(":")
		w.out.WriteString(event.AnchorID)
		w.out.WriteString("] ")
		w.out.WriteString(event.HeadingText.Value())

	case streaming.EndHeading:
		w.out.WriteString(" [HEADING_END]\n")

	case streaming.StartList:
		w.inList = true
		w.out.WriteString("[LIST_START]")

	case streaming.EndList:
		w.inList = false
		w.out.WriteString("[LIST_END]\n")

	case streaming.StartTable:
		w.out.WriteString("[TABLE_START]\n")

	case streaming.EndTable:
		w.out.WriteString("[TABLE_END]\n")

	case streaming.StartTableRow:
		w.out.WriteString("[ROW_START]")

	case streaming.EndTableRow:
		w.out.WriteString("[ROW_END]\n")

	case streaming.StartTableCell:
		w.out.WriteString("[CELL_START]")

	case streaming.EndTableCell:
		w.out.WriteString("[CELL_END]")

	case streaming.StartFormatting:
		w.out.WriteString("[FORMAT_START:")
		w.out.WriteString(w.styleToString(event.Style))
		if event.LinkURL != "" {
			w.out.WriteString(":")
			w.out.WriteString(event.LinkURL)
		}
		w.out.WriteString("]")

	case streaming.EndFormatting:
		w.out.WriteString("[FORMAT_END:")
		w.out.WriteString(w.styleToString(event.Style))
		w.out.WriteString("]")

	case streaming.Text:
		text := event.TextContent
		if w.inList {
			w.out.WriteString("[LIST_ITEM]")
		}
		// Show the raw text with visible newlines and spaces
		escaped := w.escapeWhitespace(text)
		w.out.WriteString(escaped)

	case streaming.Image:
		w.out.WriteString("[IMAGE:")
		w.out.WriteString(event.ImageURL)
		w.out.WriteString(":")
		w.out.WriteString(event.ImageAlt)
		w.out.WriteString("]")
	}
}

// styleToString converts a style bit flag to readable string
func (w *PlainTextWriter) styleToString(style streaming.StyleFlags) string {
	var parts []string
	if style&streaming.Bold != 0 {
		parts = append(parts, "BOLD")
	}
	if style&streaming.Italic != 0 {
		parts = append(parts, "ITALIC")
	}
	if style&streaming.Underline != 0 {
		parts = append(parts, "UNDERLINE")
	}
	if style&streaming.Strike != 0 {
		parts = append(parts, "STRIKE")
	}
	if style&streaming.Highlight != 0 {
		parts = append(parts, "HIGHLIGHT")
	}
	if style&streaming.Sub != 0 {
		parts = append(parts, "SUB")
	}
	if style&streaming.Sup != 0 {
		parts = append(parts, "SUP")
	}
	if style&streaming.Link != 0 {
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

// PlainTextWriterV2 implements the WriterV2 interface
type PlainTextWriterV2 struct {
	*PlainTextWriter
	stats WriterStats
}

// Handle processes a single event with error handling
func (w *PlainTextWriterV2) Handle(event streaming.Event) error {
	w.PlainTextWriter.Handle(event)
	w.stats.EventsProcessed++

	switch event.Kind {
	case streaming.Text:
		w.stats.TextChars += len(event.TextContent)
	case streaming.Image:
		w.stats.Images++
	case streaming.StartTable:
		w.stats.Tables++
	case streaming.StartHeading:
		w.stats.Headings++
	case streaming.StartList:
		w.stats.Lists++
	}

	return nil
}

// Flush finalizes any pending operations and returns the result
func (w *PlainTextWriterV2) Flush() (string, error) {
	return w.Result(), nil
}

// Reset clears the writer state for reuse
func (w *PlainTextWriterV2) Reset() {
	w.PlainTextWriter.Reset()
	w.stats = WriterStats{}
}

// Stats returns processing statistics
func (w *PlainTextWriterV2) Stats() WriterStats {
	return w.stats
}
