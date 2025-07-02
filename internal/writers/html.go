package writers

import (
	"fmt"
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// init registers the HTML writer with the writer registry, providing its factory function and metadata.
func init() {
	Register("html", func() WriterV2 {
		return &HTMLWriterV2{
			HTMLWriter: NewHTMLWriter(),
		}
	}, WriterMetadata{
		Name:        "HTML",
		Extension:   ".html",
		Description: "Clean HTML format",
		MimeType:    "text/html",
	})
}

// HTMLWriter generates clean HTML from events
type HTMLWriter struct {
	config              *config.HTMLConfig
	out                 *strings.Builder
	activeStyle         streaming.StyleFlags
	linkURL             string
	inList              bool
	inListItem          bool
	inTable             bool
	tableIsFirstRow     bool
	currentHeadingLevel int
}

// NewHTMLWriter returns a new HTMLWriter instance with default HTML configuration.
func NewHTMLWriter() *HTMLWriter {
	return NewHTMLWriterWithConfig(nil)
}

// NewHTMLWriterWithConfig returns a new HTMLWriter using the provided configuration, or the default configuration if nil.
func NewHTMLWriterWithConfig(cfg *config.HTMLConfig) *HTMLWriter {
	if cfg == nil {
		cfg = config.DefaultHTMLConfig()
	}
	return &HTMLWriter{
		config: cfg,
		out:    &strings.Builder{},
	}
}

// Handle processes a single event
func (w *HTMLWriter) Handle(event streaming.Event) {
	switch event.Kind {
	case streaming.StartDoc:
		title := event.Title
		w.out.WriteString("<!DOCTYPE html>\n")
		w.out.WriteString("<html lang=\"en\">\n")
		w.out.WriteString("<head>\n")
		w.out.WriteString("    <meta charset=\"UTF-8\">\n")
		w.out.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
		fmt.Fprintf(w.out, "    <title>%s</title>\n", w.escapeHTML(title))
		w.out.WriteString("    <style>\n")
		w.out.WriteString(w.getCSS())
		w.out.WriteString("    </style>\n")
		w.out.WriteString("</head>\n")
		w.out.WriteString("<body>\n")
		fmt.Fprintf(w.out, "    <h1>%s</h1>\n", w.escapeHTML(title))

	case streaming.EndDoc:
		w.out.WriteString("</body>\n")
		w.out.WriteString("</html>\n")

	case streaming.StartParagraph:
		if w.inListItem {
			// Close previous list item before starting a new paragraph
			w.out.WriteString("</li>\n")
			w.inListItem = false
		}
		w.out.WriteString("    <p>")

	case streaming.EndParagraph:
		w.out.WriteString("</p>\n")

	case streaming.StartHeading:
		if w.inListItem {
			// Close previous list item before starting a heading
			w.out.WriteString("</li>\n")
			w.inListItem = false
		}
		w.currentHeadingLevel = event.Level
		fmt.Fprintf(w.out, "    <h%d id=\"%s\">", event.Level, event.AnchorID)

	case streaming.EndHeading:
		fmt.Fprintf(w.out, "</h%d>\n", w.currentHeadingLevel)

	case streaming.StartList:
		w.inList = true
		w.out.WriteString("    <ul>\n")

	case streaming.EndList:
		if w.inListItem {
			// Close the last list item
			w.out.WriteString("</li>\n")
			w.inListItem = false
		}
		w.inList = false
		w.out.WriteString("    </ul>\n")

	case streaming.StartTable:
		w.inTable = true
		w.tableIsFirstRow = true
		w.out.WriteString("    <table style=\"border-collapse: collapse; width: 100%; margin: 20px 0;\">\n")

	case streaming.EndTable:
		w.inTable = false
		w.out.WriteString("    </table>\n")

	case streaming.StartTableRow:
		w.out.WriteString("        <tr>\n")

	case streaming.EndTableRow:
		w.out.WriteString("        </tr>\n")
		w.tableIsFirstRow = false

	case streaming.StartTableCell:
		tag := "td"
		style := "border: 1px solid #ddd; padding: 8px;"
		if w.tableIsFirstRow {
			tag = "th"
			style += " background-color: #f2f2f2; font-weight: bold;"
		}
		fmt.Fprintf(w.out, "            <%s style=\"%s\">", tag, style)

	case streaming.EndTableCell:
		tag := "td"
		if w.tableIsFirstRow {
			tag = "th"
		}
		fmt.Fprintf(w.out, "</%s>\n", tag)

	case streaming.StartFormatting:
		w.openHTMLTag(event.Style, event.LinkURL)
		w.activeStyle |= event.Style
		if event.Style&streaming.Link != 0 {
			w.linkURL = event.LinkURL
		}

	case streaming.EndFormatting:
		w.closeHTMLTag(event.Style)
		w.activeStyle &^= event.Style
		if event.Style&streaming.Link != 0 {
			w.linkURL = ""
		}

	case streaming.Text:
		text := event.TextContent
		if w.inTable {
			// Convert newlines to HTML breaks in tables
			text = strings.ReplaceAll(text, "\n", "<br>")
		}
		if w.inList && !w.inListItem {
			// Start a new list item
			w.out.WriteString("        <li>")
			w.inListItem = true
		}
		w.out.WriteString(w.escapeHTML(text))

	case streaming.Image:
		fmt.Fprintf(w.out, "<img src=\"%s\" alt=\"%s\" style=\"max-width: 100%%; height: auto;\" />",
			w.escapeHTML(event.ImageURL), w.escapeHTML(event.ImageAlt))
	}
}

// openHTMLTag opens an HTML tag based on style
func (w *HTMLWriter) openHTMLTag(style streaming.StyleFlags, linkURL string) {
	if style&streaming.Bold != 0 {
		open, _ := w.config.GetBoldTags()
		w.out.WriteString(open)
	}
	if style&streaming.Italic != 0 {
		open, _ := w.config.GetItalicTags()
		w.out.WriteString(open)
	}
	if style&streaming.Underline != 0 {
		open, _ := w.config.GetUnderlineTags()
		w.out.WriteString(open)
	}
	if style&streaming.Strike != 0 {
		open, _ := w.config.GetStrikeTags()
		w.out.WriteString(open)
	}
	if style&streaming.Highlight != 0 {
		open, _ := w.config.GetHighlightTags()
		w.out.WriteString(open)
	}
	if style&streaming.Sub != 0 {
		open, _ := w.config.GetSubscriptTags()
		w.out.WriteString(open)
	}
	if style&streaming.Sup != 0 {
		open, _ := w.config.GetSuperscriptTags()
		w.out.WriteString(open)
	}
	if style&streaming.Link != 0 {
		fmt.Fprintf(w.out, "<a href=\"%s\">", w.escapeHTML(linkURL))
	}
}

// closeHTMLTag closes an HTML tag based on style
func (w *HTMLWriter) closeHTMLTag(style streaming.StyleFlags) {
	// Close in reverse order
	if style&streaming.Link != 0 {
		w.out.WriteString("</a>")
	}
	if style&streaming.Sup != 0 {
		_, close := w.config.GetSuperscriptTags()
		w.out.WriteString(close)
	}
	if style&streaming.Sub != 0 {
		_, close := w.config.GetSubscriptTags()
		w.out.WriteString(close)
	}
	if style&streaming.Highlight != 0 {
		_, close := w.config.GetHighlightTags()
		w.out.WriteString(close)
	}
	if style&streaming.Strike != 0 {
		_, close := w.config.GetStrikeTags()
		w.out.WriteString(close)
	}
	if style&streaming.Underline != 0 {
		_, close := w.config.GetUnderlineTags()
		w.out.WriteString(close)
	}
	if style&streaming.Italic != 0 {
		_, close := w.config.GetItalicTags()
		w.out.WriteString(close)
	}
	if style&streaming.Bold != 0 {
		_, close := w.config.GetBoldTags()
		w.out.WriteString(close)
	}
}

// escapeHTML escapes HTML special characters
func (w *HTMLWriter) escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// getCSS returns the CSS styles for the HTML document
func (w *HTMLWriter) getCSS() string {
	return `
        body {
            font-family: 'Georgia', serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }

        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 2em;
            margin-bottom: 1em;
        }

        h1 {
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }

        p {
            margin: 1em 0;
            text-align: justify;
        }

        a {
            color: #3498db;
        }

        ul {
            margin: 1em 0;
            padding-left: 2em;
        }

        table {
            margin: 1em 0;
        }

        mark {
            background-color: #fff3cd;
            padding: 0.2em;
        }
    `
}

// Result returns the final HTML string
func (w *HTMLWriter) Result() string {
	return w.out.String()
}

// Reset clears the writer state for reuse
func (w *HTMLWriter) Reset() {
	w.out.Reset()
	w.activeStyle = 0
	w.linkURL = ""
	w.inList = false
	w.inListItem = false
	w.inTable = false
	w.tableIsFirstRow = false
}

// SetOutput sets the output destination (for StreamWriter interface)
func (w *HTMLWriter) SetOutput(writer io.Writer) {
	// For string-based writers, we ignore this
	// The Result() method returns the final string
}

// HTMLWriterV2 implements the WriterV2 interface
type HTMLWriterV2 struct {
	*HTMLWriter
	stats WriterStats
}

// Handle processes a single event with error handling
func (w *HTMLWriterV2) Handle(event streaming.Event) error {
	w.HTMLWriter.Handle(event)
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
func (w *HTMLWriterV2) Flush() (string, error) {
	return w.Result(), nil
}

// Reset clears the writer state for reuse
func (w *HTMLWriterV2) Reset() {
	w.HTMLWriter.Reset()
	w.stats = WriterStats{}
}

// Stats returns processing statistics
func (w *HTMLWriterV2) Stats() WriterStats {
	return w.stats
}
