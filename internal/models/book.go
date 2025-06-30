package models

type Book struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	AvailableDate      string    `json:"availableDate"`
	ExamDate           string    `json:"examDate"`
	BachelorYearNumber string    `json:"bachelorYearNumber"`
	CollegeStartYear   int       `json:"collegeStartYear"`
	ShopURL            string    `json:"shopUrl"`
	IsPurchased        int       `json:"isPurchased"`
	LastOpenedAt       *string   `json:"lastOpenedAt"`
	ReadProgress       *int      `json:"readProgress"`
	PageCount          int       `json:"pageCount"`
	ReadPageCount      *int      `json:"readPageCount"`
	ReadPercentage     *float64  `json:"readPercentage"`
	HasFreeChapters    int       `json:"hasFreeChapters"`
	Supplements        []string  `json:"supplements"`
	Images             []Image   `json:"images"`
	FormulasImages     []string  `json:"formulasImages"`
	Periods            []string  `json:"periods"`
	Chapters           []Chapter `json:"chapters"`
	Content            Content   `json:"content"`
	
	// Added for inline object mapping
	InlineObjectMap    map[string]string `json:"-"`
}

type Image struct {
	ID        int    `json:"id"`
	SummaryID int    `json:"summaryId"`
	CreatedAt string `json:"createdAt"`
	ObjectID  string `json:"objectId"`
	MimeType  string `json:"mimeType"`
	ImageURL  string `json:"imageUrl"`
}

type Chapter struct {
	ID              int       `json:"id"`
	SummaryID       int       `json:"summaryId"`
	Title           string    `json:"title"`
	IsFree          int       `json:"isFree"`
	IsSupplement    int       `json:"isSupplement"`
	IsLocked        int       `json:"isLocked"`
	IsVisible       int       `json:"isVisible"`
	ParentChapterID *int      `json:"parentChapterId"`
	GDocsChapterID  string    `json:"gDocsChapterId"`
	SortIndex       int       `json:"sortIndex"`
	SubChapters     []Chapter `json:"subChapters"`
}

type Content struct {
	DocumentID          string                 `json:"documentId"`
	RevisionID          string                 `json:"revisionId"`
	SuggestionsViewMode string                 `json:"suggestionsViewMode"`
	Title               string                 `json:"title"`
	Body                Body                   `json:"body"`
	DocumentStyle       *DocumentStyle         `json:"documentStyle,omitempty"`
	Headers             map[string]interface{} `json:"headers,omitempty"`
	Footers             map[string]interface{} `json:"footers,omitempty"`
	InlineObjects       map[string]interface{} `json:"inlineObjects,omitempty"`
	Lists               map[string]interface{} `json:"lists,omitempty"`
	NamedStyles         map[string]interface{} `json:"namedStyles,omitempty"`
	PositionedObjects   map[string]interface{} `json:"positionedObjects,omitempty"`
}

type DocumentStyle struct {
	Background                   *Background `json:"background,omitempty"`
	DefaultFooterID              *string     `json:"defaultFooterId,omitempty"`
	DefaultHeaderID              *string     `json:"defaultHeaderId,omitempty"`
	EvenPageFooterID             *string     `json:"evenPageFooterId,omitempty"`
	EvenPageHeaderID             *string     `json:"evenPageHeaderId,omitempty"`
	FirstPageFooterID            *string     `json:"firstPageFooterId,omitempty"`
	FirstPageHeaderID            *string     `json:"firstPageHeaderId,omitempty"`
	MarginBottom                 *Dimension  `json:"marginBottom,omitempty"`
	MarginFooter                 *Dimension  `json:"marginFooter,omitempty"`
	MarginHeader                 *Dimension  `json:"marginHeader,omitempty"`
	MarginLeft                   *Dimension  `json:"marginLeft,omitempty"`
	MarginRight                  *Dimension  `json:"marginRight,omitempty"`
	MarginTop                    *Dimension  `json:"marginTop,omitempty"`
	PageNumberStart              *int        `json:"pageNumberStart,omitempty"`
	PageSize                     *Size       `json:"pageSize,omitempty"`
	UseCustomHeaderFooterMargins *bool       `json:"useCustomHeaderFooterMargins,omitempty"`
	UseEvenPageHeaderFooter      *bool       `json:"useEvenPageHeaderFooter,omitempty"`
	UseFirstPageHeaderFooter     *bool       `json:"useFirstPageHeaderFooter,omitempty"`
}

type Background struct {
	Color interface{} `json:"color,omitempty"`
}

type Color struct {
	RgbColor interface{} `json:"rgbColor,omitempty"`
}

type RgbColor struct {
	Blue  float64 `json:"blue"`
	Green float64 `json:"green"`
	Red   float64 `json:"red"`
}

type Size struct {
	Height *Dimension `json:"height,omitempty"`
	Width  *Dimension `json:"width,omitempty"`
}

type Body struct {
	Content []ContentElement `json:"content"`
}

type ContentElement struct {
	EndIndex        int              `json:"endIndex"`
	StartIndex      *int             `json:"startIndex"`
	SectionBreak    *SectionBreak    `json:"sectionBreak,omitempty"`
	Paragraph       *Paragraph       `json:"paragraph,omitempty"`
	Table           *Table           `json:"table,omitempty"`
	TableOfContents *TableOfContents `json:"tableOfContents,omitempty"`
}

type SectionBreak struct {
	SuggestedDeletionIds  []string     `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string     `json:"suggestedInsertionIds"`
	SectionStyle          SectionStyle `json:"sectionStyle"`
}

type SectionStyle struct {
	ColumnSeparatorStyle     string  `json:"columnSeparatorStyle"`
	ContentDirection         string  `json:"contentDirection"`
	DefaultFooterID          *string `json:"defaultFooterId"`
	DefaultHeaderID          *string `json:"defaultHeaderId"`
	EvenPageFooterID         *string `json:"evenPageFooterId"`
	EvenPageHeaderID         *string `json:"evenPageHeaderId"`
	FirstPageFooterID        *string `json:"firstPageFooterId"`
	FirstPageHeaderID        *string `json:"firstPageHeaderId"`
	PageNumberStart          *int    `json:"pageNumberStart"`
	SectionType              string  `json:"sectionType"`
	UseFirstPageHeaderFooter *bool   `json:"useFirstPageHeaderFooter"`
}

type Table struct {
	Columns               int         `json:"columns"`
	Rows                  int         `json:"rows"`
	SuggestedDeletionIds  []string    `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string    `json:"suggestedInsertionIds"`
	TableRows             []TableRow  `json:"tableRows"`
	TableStyle            *TableStyle `json:"tableStyle,omitempty"`
}

type TableRow struct {
	EndIndex              int            `json:"endIndex"`
	StartIndex            int            `json:"startIndex"`
	SuggestedDeletionIds  []string       `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string       `json:"suggestedInsertionIds"`
	TableCells            []TableCell    `json:"tableCells"`
	TableRowStyle         *TableRowStyle `json:"tableRowStyle,omitempty"`
}

type TableCell struct {
	Content               []ContentElement `json:"content"`
	EndIndex              int              `json:"endIndex"`
	StartIndex            int              `json:"startIndex"`
	SuggestedDeletionIds  []string         `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string         `json:"suggestedInsertionIds"`
	TableCellStyle        *TableCellStyle  `json:"tableCellStyle,omitempty"`
}

type TableStyle struct {
	TableColumnProperties []TableColumnProperties `json:"tableColumnProperties,omitempty"`
}

type TableColumnProperties struct {
	Width     *Dimension `json:"width,omitempty"`
	WidthType string     `json:"widthType,omitempty"`
}

type TableRowStyle struct {
	MinRowHeight *Dimension `json:"minRowHeight,omitempty"`
}

type TableCellStyle struct {
	BackgroundColor  interface{}      `json:"backgroundColor,omitempty"`
	BorderBottom     *TableCellBorder `json:"borderBottom,omitempty"`
	BorderLeft       *TableCellBorder `json:"borderLeft,omitempty"`
	BorderRight      *TableCellBorder `json:"borderRight,omitempty"`
	BorderTop        *TableCellBorder `json:"borderTop,omitempty"`
	ColumnSpan       *int             `json:"columnSpan,omitempty"`
	ContentAlignment string           `json:"contentAlignment,omitempty"`
	PaddingBottom    *Dimension       `json:"paddingBottom,omitempty"`
	PaddingLeft      *Dimension       `json:"paddingLeft,omitempty"`
	PaddingRight     *Dimension       `json:"paddingRight,omitempty"`
	PaddingTop       *Dimension       `json:"paddingTop,omitempty"`
	RowSpan          *int             `json:"rowSpan,omitempty"`
}

type TableCellBorder struct {
	Color     interface{} `json:"color,omitempty"`
	DashStyle string      `json:"dashStyle,omitempty"`
	Width     *Dimension  `json:"width,omitempty"`
}

type TableOfContents struct {
	Content               []ContentElement `json:"content"`
	SuggestedDeletionIds  []string         `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string         `json:"suggestedInsertionIds"`
}

type Paragraph struct {
	PositionedObjectIds []string       `json:"positionedObjectIds"`
	Elements            []Element      `json:"elements"`
	ParagraphStyle      ParagraphStyle `json:"paragraphStyle"`
	Bullet              *Bullet        `json:"bullet,omitempty"`
}

type Bullet struct {
	ListID       string     `json:"listId"`
	NestingLevel *int       `json:"nestingLevel,omitempty"`
	TextStyle    *TextStyle `json:"textStyle,omitempty"`
}

type Element struct {
	EndIndex            int                  `json:"endIndex"`
	StartIndex          int                  `json:"startIndex"`
	TextRun             *TextRun             `json:"textRun,omitempty"`
	InlineObjectElement *InlineObjectElement `json:"inlineObjectElement,omitempty"`
	PageBreak           *PageBreak           `json:"pageBreak,omitempty"`
	ColumnBreak         *ColumnBreak         `json:"columnBreak,omitempty"`
	FootnoteReference   *FootnoteReference   `json:"footnoteReference,omitempty"`
	HorizontalRule      *HorizontalRule      `json:"horizontalRule,omitempty"`
	Equation            *Equation            `json:"equation,omitempty"`
}

type TextRun struct {
	Content               string    `json:"content"`
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type InlineObjectElement struct {
	InlineObjectID        string    `json:"inlineObjectId"`
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type PageBreak struct {
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type ColumnBreak struct {
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type FootnoteReference struct {
	FootnoteID            string    `json:"footnoteId"`
	FootnoteNumber        string    `json:"footnoteNumber"`
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type HorizontalRule struct {
	SuggestedDeletionIds  []string  `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string  `json:"suggestedInsertionIds"`
	TextStyle             TextStyle `json:"textStyle"`
}

type Equation struct {
	SuggestedDeletionIds  []string `json:"suggestedDeletionIds"`
	SuggestedInsertionIds []string `json:"suggestedInsertionIds"`
}

type TextStyle struct {
	BaselineOffset     *string             `json:"baselineOffset"`
	Bold               *bool               `json:"bold"`
	Italic             *bool               `json:"italic"`
	SmallCaps          *bool               `json:"smallCaps"`
	Strikethrough      *bool               `json:"strikethrough"`
	Underline          *bool               `json:"underline"`
	FontSize           *FontSize           `json:"fontSize"`
	ForegroundColor    interface{}         `json:"foregroundColor"`
	BackgroundColor    interface{}         `json:"backgroundColor"`
	WeightedFontFamily *WeightedFontFamily `json:"weightedFontFamily"`
	Link               *Link               `json:"link"`
}

type FontSize struct {
	Magnitude float64 `json:"magnitude"`
	Unit      string  `json:"unit"`
}

type WeightedFontFamily struct {
	FontFamily string `json:"fontFamily"`
	Weight     int    `json:"weight"`
}

type Link struct {
	BookmarkID *string `json:"bookmarkId"`
	HeadingID  *string `json:"headingId"`
	URL        string  `json:"url"`
}

type ParagraphStyle struct {
	Alignment              string           `json:"alignment"`
	AvoidWidowAndOrphan    *bool            `json:"avoidWidowAndOrphan"`
	BorderBetween          *ParagraphBorder `json:"borderBetween,omitempty"`
	BorderBottom           *ParagraphBorder `json:"borderBottom,omitempty"`
	BorderLeft             *ParagraphBorder `json:"borderLeft,omitempty"`
	BorderRight            *ParagraphBorder `json:"borderRight,omitempty"`
	BorderTop              *ParagraphBorder `json:"borderTop,omitempty"`
	Direction              string           `json:"direction"`
	HeadingID              *string          `json:"headingId"`
	IndentEnd              *Dimension       `json:"indentEnd,omitempty"`
	IndentFirstLine        *Dimension       `json:"indentFirstLine,omitempty"`
	IndentStart            *Dimension       `json:"indentStart,omitempty"`
	KeepLinesTogether      *bool            `json:"keepLinesTogether"`
	KeepWithNext           *bool            `json:"keepWithNext"`
	LineSpacing            float64          `json:"lineSpacing"`
	NamedStyleType         string           `json:"namedStyleType"`
	PageBreakBefore        bool             `json:"pageBreakBefore"`
	ShadingBackgroundColor interface{}      `json:"shadingBackgroundColor,omitempty"`
	SpaceAbove             *Dimension       `json:"spaceAbove"`
	SpaceBelow             *Dimension       `json:"spaceBelow"`
	SpacingMode            string           `json:"spacingMode"`
	TabStops               []TabStop        `json:"tabStops,omitempty"`
}

type ParagraphBorder struct {
	Color     interface{} `json:"color,omitempty"`
	DashStyle string      `json:"dashStyle,omitempty"`
	Padding   *Dimension  `json:"padding,omitempty"`
	Width     *Dimension  `json:"width,omitempty"`
}

type TabStop struct {
	Alignment string     `json:"alignment"`
	Offset    *Dimension `json:"offset"`
}

type Dimension struct {
	Magnitude *float64 `json:"magnitude"`
	Unit      string   `json:"unit"`
}
