package events

import "iter"

// Kind represents the type of event in the document stream
type Kind int

const (
	StartDoc Kind = iota
	EndDoc
	StartParagraph
	EndParagraph
	StartHeading
	EndHeading
	StartList
	EndList
	StartTable
	EndTable
	StartTableRow
	EndTableRow
	StartTableCell
	EndTableCell
	StartFormatting
	EndFormatting
	Text
	Image
)

// Event represents a single structural or formatting event in the document
type Event struct {
	Kind Kind
	Arg  any // carries Style, level, text, URLs, etc.
}

// Style represents active formatting as bit flags for efficient diffing
type Style uint16

const (
	Bold Style = 1 << iota
	Italic
	Underline
	Strike
	Highlight
	Sub
	Sup
	Link
)

// precedenceOrder defines the canonical nesting order for formatting markers
// Link is outermost, subscript/superscript are innermost
var precedenceOrder = []Style{Link, Bold, Italic, Underline, Strike, Highlight, Sub, Sup}

// Has checks if a specific style flag is set
func (s Style) Has(flag Style) bool {
	return s&flag != 0
}

// diff returns the styles that need to be closed and opened to transition from current to next
func (s Style) diff(next Style) (close, open []Style) {
	changed := s ^ next
	closing := s & changed    // what we had but don't want
	opening := next & changed // what we want but don't have

	// Special case: if Link status changes, we need to close/reopen everything
	// because Link is outermost and affects nesting order
	linkChanged := (s.Has(Link)) != (next.Has(Link))

	if linkChanged {
		// Close all current styles in reverse precedence order
		for i := len(precedenceOrder) - 1; i >= 0; i-- {
			style := precedenceOrder[i]
			if s.Has(style) {
				close = append(close, style)
			}
		}

		// Open all next styles in forward precedence order
		for _, style := range precedenceOrder {
			if next.Has(style) {
				open = append(open, style)
			}
		}
	} else {
		// Normal case: only handle changed styles
		// Return closing styles in reverse precedence order (innermost first)
		for i := len(precedenceOrder) - 1; i >= 0; i-- {
			style := precedenceOrder[i]
			if closing&style != 0 {
				close = append(close, style)
			}
		}

		// Return opening styles in forward precedence order (outermost first)
		for _, style := range precedenceOrder {
			if opening&style != 0 {
				open = append(open, style)
			}
		}
	}

	return close, open
}

// HeadingInfo carries heading-specific information
type HeadingInfo struct {
	Level    int
	Text     string
	AnchorID string
}

// ListInfo carries list-specific information
type ListInfo struct {
	Level   int
	Ordered bool
}

// TableInfo carries table-specific information
type TableInfo struct {
	Columns int
	Rows    int
}

// ImageInfo carries image-specific information
type ImageInfo struct {
	URL string
	Alt string
}

// FormatInfo carries formatting information for events
type FormatInfo struct {
	Style Style
	URL   string
}

// StreamFunc is the function signature for event streaming
type StreamFunc func(yield func(Event) bool)

// Seq is an alias for the Go 1.23 iterator type
type Seq[T any] iter.Seq[T]
