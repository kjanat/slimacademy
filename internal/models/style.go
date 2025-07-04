package models

// TextStyle represents text styling
type TextStyle struct {
	Bold               *bool               `json:"bold,omitempty"`
	Italic             *bool               `json:"italic,omitempty"`
	Underline          *bool               `json:"underline,omitempty"`
	Strikethrough      *bool               `json:"strikethrough,omitempty"`
	SmallCaps          *bool               `json:"smallCaps,omitempty"`
	BackgroundColor    *Color              `json:"backgroundColor,omitempty"`
	ForegroundColor    *Color              `json:"foregroundColor,omitempty"`
	FontSize           *Dimension          `json:"fontSize,omitempty"`
	WeightedFontFamily *WeightedFontFamily `json:"weightedFontFamily,omitempty"`
	BaselineOffset     *string             `json:"baselineOffset,omitempty"`
	Link               *Link               `json:"link,omitempty"`
}

// ParagraphStyle represents paragraph styling
type ParagraphStyle struct {
	NamedStyleType      string     `json:"namedStyleType"`
	Alignment           *string    `json:"alignment,omitempty"`
	LineSpacing         *float64   `json:"lineSpacing,omitempty"`
	Direction           string     `json:"direction"`
	SpacingMode         *string    `json:"spacingMode,omitempty"`
	SpaceAbove          *Dimension `json:"spaceAbove,omitempty"`
	SpaceBelow          *Dimension `json:"spaceBelow,omitempty"`
	IndentFirstLine     *Dimension `json:"indentFirstLine,omitempty"`
	IndentStart         *Dimension `json:"indentStart,omitempty"`
	IndentEnd           *Dimension `json:"indentEnd,omitempty"`
	TabStops            []TabStop  `json:"tabStops,omitempty"`
	BorderTop           *Border    `json:"borderTop,omitempty"`
	BorderBottom        *Border    `json:"borderBottom,omitempty"`
	BorderLeft          *Border    `json:"borderLeft,omitempty"`
	BorderRight         *Border    `json:"borderRight,omitempty"`
	BorderBetween       *Border    `json:"borderBetween,omitempty"`
	Shading             *Shading   `json:"shading,omitempty"`
	HeadingID           *string    `json:"headingId,omitempty"`
	AvoidWidowAndOrphan *bool      `json:"avoidWidowAndOrphan,omitempty"`
	KeepLinesTogether   *bool      `json:"keepLinesTogether,omitempty"`
	KeepWithNext        *bool      `json:"keepWithNext,omitempty"`
	PageBreakBefore     *bool      `json:"pageBreakBefore,omitempty"`
}

// DocumentStyle represents document-level styling
type DocumentStyle struct {
	Background                   Background `json:"background"`
	PageSize                     Size       `json:"pageSize"`
	MarginTop                    Dimension  `json:"marginTop"`
	MarginBottom                 Dimension  `json:"marginBottom"`
	MarginRight                  Dimension  `json:"marginRight"`
	MarginLeft                   Dimension  `json:"marginLeft"`
	MarginHeader                 Dimension  `json:"marginHeader"`
	MarginFooter                 Dimension  `json:"marginFooter"`
	PageNumberStart              int64      `json:"pageNumberStart"`
	UseCustomHeaderFooterMargins bool       `json:"useCustomHeaderFooterMargins"`
	UseEvenPageHeaderFooter      any        `json:"useEvenPageHeaderFooter"`
	UseFirstPageHeaderFooter     any        `json:"useFirstPageHeaderFooter"`
	DefaultHeaderID              string     `json:"defaultHeaderId"`
	DefaultFooterID              string     `json:"defaultFooterId"`
	FirstPageHeaderID            string     `json:"firstPageHeaderId"`
	FirstPageFooterID            string     `json:"firstPageFooterId"`
	EvenPageHeaderID             any        `json:"evenPageHeaderId"`
	EvenPageFooterID             any        `json:"evenPageFooterId"`
}

// NamedStyles represents named styles in the document
type NamedStyles struct {
	Styles []NamedStyle `json:"styles"`
}

// NamedStyle represents a named style
type NamedStyle struct {
	NamedStyleType string         `json:"namedStyleType"`
	TextStyle      TextStyle      `json:"textStyle"`
	ParagraphStyle ParagraphStyle `json:"paragraphStyle"`
}

// RGBColor represents RGB color values
type RGBColor struct {
	Red   *float64 `json:"red,omitempty"`
	Green *float64 `json:"green,omitempty"`
	Blue  *float64 `json:"blue,omitempty"`
}

// WeightedFontFamily represents a font family with weight
type WeightedFontFamily struct {
	FontFamily string `json:"fontFamily"`
	Weight     int64  `json:"weight"`
}

// Link represents a hyperlink
type Link struct {
	URL        *string `json:"url,omitempty"`
	BookmarkID *string `json:"bookmarkId,omitempty"`
	HeadingID  *string `json:"headingId,omitempty"`
}

// Border represents a border style
type Border struct {
	Color     *Color    `json:"color,omitempty"`
	Width     Dimension `json:"width"`
	Padding   Dimension `json:"padding"`
	DashStyle string    `json:"dashStyle"`
}

// Shading represents background shading
type Shading struct {
	BackgroundColor *Color `json:"backgroundColor,omitempty"`
}

// TabStop represents a tab stop
type TabStop struct {
	Offset    Dimension `json:"offset"`
	Alignment string    `json:"alignment"`
}

// Background represents document background
type Background struct {
	Color *Color `json:"color,omitempty"`
}
