package writers

import (
	"fmt"
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

func init() {
	Register("latex", func() WriterV2 {
		return &LaTeXWriterV2{
			LaTeXWriter: NewLaTeXWriter(nil),
		}
	}, WriterMetadata{
		Name:        "LaTeX",
		Extension:   ".tex",
		Description: "LaTeX document format",
		MimeType:    "text/x-tex",
	})
}

// LaTeXWriter generates LaTeX from events
type LaTeXWriter struct {
	config      *config.LaTeXConfig
	out         *strings.Builder
	activeStyle streaming.StyleFlags
	linkURL     string
	inList      bool
	inTable     bool
	listDepth   int
}

// NewLaTeXWriter creates a new LaTeX writer
func NewLaTeXWriter(cfg *config.LaTeXConfig) *LaTeXWriter {
	if cfg == nil {
		cfg = config.DefaultLaTeXConfig()
	}
	return &LaTeXWriter{
		config: cfg,
		out:    &strings.Builder{},
	}
}

// Handle processes a single event
func (w *LaTeXWriter) Handle(event streaming.Event) {
	switch event.Kind {
	case streaming.StartDoc:
		title := event.Title
		w.writeDocumentHeader(title)

	case streaming.EndDoc:
		w.out.WriteString("\\end{document}\n")

	case streaming.StartParagraph:
		// Paragraphs are separated by blank lines in LaTeX

	case streaming.EndParagraph:
		w.out.WriteString("\n\n")

	case streaming.StartHeading:
		sectionCmd := w.getSectionCommand(event.Level)
		fmt.Fprintf(w.out, "%s%s}\n\\label{%s}\n\n",
			sectionCmd, w.escapeLaTeX(event.HeadingText.Value()), event.AnchorID)

	case streaming.EndHeading:
		// Heading complete - nothing needed

	case streaming.StartList:
		w.inList = true
		w.listDepth++
		w.out.WriteString("\\begin{itemize}\n")

	case streaming.EndList:
		w.inList = false
		w.listDepth--
		w.out.WriteString("\\end{itemize}\n\n")

	case streaming.StartTable:
		w.inTable = true
		w.out.WriteString("\\begin{table}[h]\n\\centering\n\\begin{tabular}{")
		// Note: We'd need to know column count here, defaulting to 3
		w.out.WriteString("lll")
		w.out.WriteString("}\n\\hline\n")

	case streaming.EndTable:
		w.inTable = false
		w.out.WriteString("\\hline\n\\end{tabular}\n\\end{table}\n\n")

	case streaming.StartTableRow:
		// Row will be handled by content

	case streaming.EndTableRow:
		w.out.WriteString(" \\\\\n")

	case streaming.StartTableCell:
		// Cell content will be handled by other events

	case streaming.EndTableCell:
		w.out.WriteString(" & ")

	case streaming.StartFormatting:
		w.openLaTeXCommand(event.Style, event.LinkURL)
		w.activeStyle |= event.Style
		if event.Style&streaming.Link != 0 {
			w.linkURL = event.LinkURL
		}

	case streaming.EndFormatting:
		w.closeLaTeXCommand(event.Style)
		w.activeStyle &^= event.Style
		if event.Style&streaming.Link != 0 {
			w.linkURL = ""
		}

	case streaming.Text:
		text := event.TextContent
		if w.inList {
			// Handle list items
			indent := strings.Repeat("  ", w.listDepth-1)
			fmt.Fprintf(w.out, "%s\\item %s\n", indent, w.escapeLaTeX(text))
		} else {
			w.out.WriteString(w.escapeLaTeX(text))
		}

	case streaming.Image:
		fmt.Fprintf(w.out, "\\includegraphics[width=0.8\\textwidth]{%s}",
			w.escapeLaTeX(event.ImageURL))
	}
}

// writeDocumentHeader writes the LaTeX document header
func (w *LaTeXWriter) writeDocumentHeader(title string) {
	w.out.WriteString(w.config.GetDocumentPreamble())
	w.out.WriteString("\n")
	fmt.Fprintf(w.out, "\\title{%s}\n", w.escapeLaTeX(title))
	w.out.WriteString("\\author{}\n")
	w.out.WriteString("\\date{}\n\n")
	w.out.WriteString("\\begin{document}\n")
	w.out.WriteString("\\maketitle\n\n")
}

// getSectionCommand returns the appropriate LaTeX section command for the level
func (w *LaTeXWriter) getSectionCommand(level int) string {
	return w.config.GetHeadingCommand(level)
}

// openLaTeXCommand opens a LaTeX command based on style
func (w *LaTeXWriter) openLaTeXCommand(style streaming.StyleFlags, linkURL string) {
	if style&streaming.Bold != 0 {
		w.out.WriteString(w.config.GetBoldCommand())
	}
	if style&streaming.Italic != 0 {
		w.out.WriteString(w.config.GetItalicCommand())
	}
	if style&streaming.Underline != 0 {
		w.out.WriteString(w.config.GetUnderlineCommand())
	}
	if style&streaming.Strike != 0 {
		w.out.WriteString(w.config.GetStrikeCommand())
	}
	if style&streaming.Highlight != 0 {
		w.out.WriteString(w.config.GetHighlightCommand())
	}
	if style&streaming.Sub != 0 {
		w.out.WriteString(w.config.GetSubscriptCommand())
	}
	if style&streaming.Sup != 0 {
		w.out.WriteString(w.config.GetSuperscriptCommand())
	}
	if style&streaming.Link != 0 {
		fmt.Fprintf(w.out, "\\href{%s}{", w.escapeLaTeX(linkURL))
	}
}

// closeLaTeXCommand closes a LaTeX command based on style
func (w *LaTeXWriter) closeLaTeXCommand(style streaming.StyleFlags) {
	// All LaTeX commands end with closing brace
	w.out.WriteString("}")
}

// escapeLaTeX escapes special LaTeX characters
func (w *LaTeXWriter) escapeLaTeX(text string) string {
	// Escape common LaTeX special characters
	text = strings.ReplaceAll(text, "\\", "\\textbackslash{}")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	text = strings.ReplaceAll(text, "$", "\\$")
	text = strings.ReplaceAll(text, "&", "\\&")
	text = strings.ReplaceAll(text, "%", "\\%")
	text = strings.ReplaceAll(text, "#", "\\#")
	text = strings.ReplaceAll(text, "^", "\\textasciicircum{}")
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "~", "\\textasciitilde{}")
	return text
}

// Result returns the final LaTeX string
func (w *LaTeXWriter) Result() string {
	return w.out.String()
}

// Reset clears the writer state for reuse
func (w *LaTeXWriter) Reset() {
	w.out.Reset()
	w.activeStyle = 0
	w.linkURL = ""
	w.inList = false
	w.inTable = false
	w.listDepth = 0
}

// SetOutput sets the output destination (for StreamWriter interface)
func (w *LaTeXWriter) SetOutput(writer io.Writer) {
	// For string-based writers, we ignore this
	// The Result() method returns the final string
}

// LaTeXWriterV2 implements the WriterV2 interface
type LaTeXWriterV2 struct {
	*LaTeXWriter
	stats WriterStats
}

// Handle processes a single event with error handling
func (w *LaTeXWriterV2) Handle(event streaming.Event) error {
	w.LaTeXWriter.Handle(event)
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
func (w *LaTeXWriterV2) Flush() (string, error) {
	return w.Result(), nil
}

// Reset clears the writer state for reuse
func (w *LaTeXWriterV2) Reset() {
	w.LaTeXWriter.Reset()
	w.stats = WriterStats{}
}

// Stats returns processing statistics
func (w *LaTeXWriterV2) Stats() WriterStats {
	return w.stats
}
