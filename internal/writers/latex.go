package writers

import (
	"fmt"
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/events"
)

// LaTeXWriter generates LaTeX from events
type LaTeXWriter struct {
	config      *config.LaTeXConfig
	out         *strings.Builder
	activeStyle events.Style
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
func (w *LaTeXWriter) Handle(event events.Event) {
	switch event.Kind {
	case events.StartDoc:
		title := event.Arg.(string)
		w.writeDocumentHeader(title)

	case events.EndDoc:
		w.out.WriteString("\\end{document}\n")

	case events.StartParagraph:
		// Paragraphs are separated by blank lines in LaTeX

	case events.EndParagraph:
		w.out.WriteString("\n\n")

	case events.StartHeading:
		info := event.Arg.(events.HeadingInfo)
		sectionCmd := w.getSectionCommand(info.Level)
		w.out.WriteString(fmt.Sprintf("%s%s}\n\\label{%s}\n\n",
			sectionCmd, w.escapeLaTeX(info.Text), info.AnchorID))

	case events.EndHeading:
		// Heading complete - nothing needed

	case events.StartList:
		w.inList = true
		w.listDepth++
		w.out.WriteString("\\begin{itemize}\n")

	case events.EndList:
		w.inList = false
		w.listDepth--
		w.out.WriteString("\\end{itemize}\n\n")

	case events.StartTable:
		w.inTable = true
		w.out.WriteString("\\begin{table}[h]\n\\centering\n\\begin{tabular}{")
		// Note: We'd need to know column count here, defaulting to 3
		w.out.WriteString("lll")
		w.out.WriteString("}\n\\hline\n")

	case events.EndTable:
		w.inTable = false
		w.out.WriteString("\\hline\n\\end{tabular}\n\\end{table}\n\n")

	case events.StartTableRow:
		// Row will be handled by content

	case events.EndTableRow:
		w.out.WriteString(" \\\\\n")

	case events.StartTableCell:
		// Cell content will be handled by other events

	case events.EndTableCell:
		w.out.WriteString(" & ")

	case events.StartFormatting:
		info := event.Arg.(events.FormatInfo)
		w.openLaTeXCommand(info.Style, info.URL)
		w.activeStyle |= info.Style
		if info.Style.Has(events.Link) {
			w.linkURL = info.URL
		}

	case events.EndFormatting:
		info := event.Arg.(events.FormatInfo)
		w.closeLaTeXCommand(info.Style)
		w.activeStyle &^= info.Style
		if info.Style.Has(events.Link) {
			w.linkURL = ""
		}

	case events.Text:
		text := event.Arg.(string)
		if w.inList {
			// Handle list items
			indent := strings.Repeat("  ", w.listDepth-1)
			w.out.WriteString(fmt.Sprintf("%s\\item %s\n", indent, w.escapeLaTeX(text)))
		} else {
			w.out.WriteString(w.escapeLaTeX(text))
		}

	case events.Image:
		info := event.Arg.(events.ImageInfo)
		w.out.WriteString(fmt.Sprintf("\\includegraphics[width=0.8\\textwidth]{%s}",
			w.escapeLaTeX(info.URL)))
	}
}

// writeDocumentHeader writes the LaTeX document header
func (w *LaTeXWriter) writeDocumentHeader(title string) {
	w.out.WriteString(w.config.GetDocumentPreamble())
	w.out.WriteString("\n")
	w.out.WriteString(fmt.Sprintf("\\title{%s}\n", w.escapeLaTeX(title)))
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
func (w *LaTeXWriter) openLaTeXCommand(style events.Style, linkURL string) {
	switch style {
	case events.Bold:
		w.out.WriteString(w.config.GetBoldCommand())
	case events.Italic:
		w.out.WriteString(w.config.GetItalicCommand())
	case events.Underline:
		w.out.WriteString(w.config.GetUnderlineCommand())
	case events.Strike:
		w.out.WriteString(w.config.GetStrikeCommand())
	case events.Highlight:
		w.out.WriteString(w.config.GetHighlightCommand())
	case events.Sub:
		w.out.WriteString(w.config.GetSubscriptCommand())
	case events.Sup:
		w.out.WriteString(w.config.GetSuperscriptCommand())
	case events.Link:
		w.out.WriteString(fmt.Sprintf("\\href{%s}{", w.escapeLaTeX(linkURL)))
	}
}

// closeLaTeXCommand closes a LaTeX command based on style
func (w *LaTeXWriter) closeLaTeXCommand(style events.Style) {
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
