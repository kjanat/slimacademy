package writers

import (
	"fmt"
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/events"
)

// HTMLWriter generates clean HTML from events
type HTMLWriter struct {
	config          *config.HTMLConfig
	out             *strings.Builder
	activeStyle     events.Style
	linkURL         string
	inList          bool
	inTable         bool
	tableIsFirstRow bool
}

// NewHTMLWriter creates a new HTML writer
func NewHTMLWriter() *HTMLWriter {
	return NewHTMLWriterWithConfig(nil)
}

// NewHTMLWriterWithConfig creates a new HTML writer with custom config
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
func (w *HTMLWriter) Handle(event events.Event) {
	switch event.Kind {
	case events.StartDoc:
		title := event.Arg.(string)
		w.out.WriteString("<!DOCTYPE html>\n")
		w.out.WriteString("<html lang=\"en\">\n")
		w.out.WriteString("<head>\n")
		w.out.WriteString("    <meta charset=\"UTF-8\">\n")
		w.out.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
		w.out.WriteString(fmt.Sprintf("    <title>%s</title>\n", w.escapeHTML(title)))
		w.out.WriteString("    <style>\n")
		w.out.WriteString(w.getCSS())
		w.out.WriteString("    </style>\n")
		w.out.WriteString("</head>\n")
		w.out.WriteString("<body>\n")
		w.out.WriteString(fmt.Sprintf("    <h1>%s</h1>\n", w.escapeHTML(title)))

	case events.EndDoc:
		w.out.WriteString("</body>\n")
		w.out.WriteString("</html>\n")

	case events.StartParagraph:
		w.out.WriteString("    <p>")

	case events.EndParagraph:
		w.out.WriteString("</p>\n")

	case events.StartHeading:
		info := event.Arg.(events.HeadingInfo)
		fmt.Fprintf(w.out, "    <h%d id=\"%s\">%s</h%d>\n",
			info.Level, info.AnchorID, w.escapeHTML(info.Text), info.Level)

	case events.EndHeading:
		// Heading complete - nothing needed (handled in StartHeading)

	case events.StartList:
		w.inList = true
		w.out.WriteString("    <ul>\n")

	case events.EndList:
		w.inList = false
		w.out.WriteString("    </ul>\n")

	case events.StartTable:
		w.inTable = true
		w.tableIsFirstRow = true
		w.out.WriteString("    <table style=\"border-collapse: collapse; width: 100%; margin: 20px 0;\">\n")

	case events.EndTable:
		w.inTable = false
		w.out.WriteString("    </table>\n")

	case events.StartTableRow:
		w.out.WriteString("        <tr>\n")

	case events.EndTableRow:
		w.out.WriteString("        </tr>\n")
		w.tableIsFirstRow = false

	case events.StartTableCell:
		tag := "td"
		style := "border: 1px solid #ddd; padding: 8px;"
		if w.tableIsFirstRow {
			tag = "th"
			style += " background-color: #f2f2f2; font-weight: bold;"
		}
		w.out.WriteString(fmt.Sprintf("            <%s style=\"%s\">", tag, style))

	case events.EndTableCell:
		tag := "td"
		if w.tableIsFirstRow {
			tag = "th"
		}
		w.out.WriteString(fmt.Sprintf("</%s>\n", tag))

	case events.StartFormatting:
		info := event.Arg.(events.FormatInfo)
		w.openHTMLTag(info.Style, info.URL)
		w.activeStyle |= info.Style
		if info.Style.Has(events.Link) {
			w.linkURL = info.URL
		}

	case events.EndFormatting:
		info := event.Arg.(events.FormatInfo)
		w.closeHTMLTag(info.Style)
		w.activeStyle &^= info.Style
		if info.Style.Has(events.Link) {
			w.linkURL = ""
		}

	case events.Text:
		text := event.Arg.(string)
		if w.inTable {
			// Convert newlines to HTML breaks in tables
			text = strings.ReplaceAll(text, "\n", "<br>")
		}
		if w.inList {
			// Handle list items
			w.out.WriteString("        <li>")
		}
		w.out.WriteString(w.escapeHTML(text))
		if w.inList {
			w.out.WriteString("</li>\n")
		}

	case events.Image:
		info := event.Arg.(events.ImageInfo)
		fmt.Fprintf(w.out, "<img src=\"%s\" alt=\"%s\" style=\"max-width: 100%%; height: auto;\" />",
			w.escapeHTML(info.URL), w.escapeHTML(info.Alt))
	}
}

// openHTMLTag opens an HTML tag based on style
func (w *HTMLWriter) openHTMLTag(style events.Style, linkURL string) {
	switch style {
	case events.Bold:
		open, _ := w.config.GetBoldTags()
		w.out.WriteString(open)
	case events.Italic:
		open, _ := w.config.GetItalicTags()
		w.out.WriteString(open)
	case events.Underline:
		open, _ := w.config.GetUnderlineTags()
		w.out.WriteString(open)
	case events.Strike:
		open, _ := w.config.GetStrikeTags()
		w.out.WriteString(open)
	case events.Highlight:
		open, _ := w.config.GetHighlightTags()
		w.out.WriteString(open)
	case events.Sub:
		open, _ := w.config.GetSubscriptTags()
		w.out.WriteString(open)
	case events.Sup:
		open, _ := w.config.GetSuperscriptTags()
		w.out.WriteString(open)
	case events.Link:
		w.out.WriteString(fmt.Sprintf("<a href=\"%s\">", w.escapeHTML(linkURL)))
	}
}

// closeHTMLTag closes an HTML tag based on style
func (w *HTMLWriter) closeHTMLTag(style events.Style) {
	switch style {
	case events.Bold:
		_, close := w.config.GetBoldTags()
		w.out.WriteString(close)
	case events.Italic:
		_, close := w.config.GetItalicTags()
		w.out.WriteString(close)
	case events.Underline:
		_, close := w.config.GetUnderlineTags()
		w.out.WriteString(close)
	case events.Strike:
		_, close := w.config.GetStrikeTags()
		w.out.WriteString(close)
	case events.Highlight:
		_, close := w.config.GetHighlightTags()
		w.out.WriteString(close)
	case events.Sub:
		_, close := w.config.GetSubscriptTags()
		w.out.WriteString(close)
	case events.Sup:
		_, close := w.config.GetSuperscriptTags()
		w.out.WriteString(close)
	case events.Link:
		w.out.WriteString("</a>")
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
	w.inTable = false
	w.tableIsFirstRow = false
}

// SetOutput sets the output destination (for StreamWriter interface)
func (w *HTMLWriter) SetOutput(writer io.Writer) {
	// For string-based writers, we ignore this
	// The Result() method returns the final string
}
