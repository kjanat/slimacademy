package models

// Table represents a table
type Table struct {
	Rows                  int64      `json:"rows"`
	Columns               int64      `json:"columns"`
	TableRows             []TableRow `json:"tableRows"`
	TableStyle            TableStyle `json:"tableStyle"`
	SuggestedDeletionIDs  any        `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any        `json:"suggestedInsertionIds"`
}

// TableRow represents a row in a table
type TableRow struct {
	StartIndex            int64         `json:"startIndex"`
	EndIndex              int64         `json:"endIndex"`
	TableCells            []TableCell   `json:"tableCells"`
	TableRowStyle         TableRowStyle `json:"tableRowStyle"`
	SuggestedDeletionIDs  any           `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any           `json:"suggestedInsertionIds"`
}

// TableCell represents a cell in a table
type TableCell struct {
	StartIndex            int64               `json:"startIndex"`
	EndIndex              int64               `json:"endIndex"`
	Content               []StructuralElement `json:"content"`
	TableCellStyle        TableCellStyle      `json:"tableCellStyle"`
	SuggestedDeletionIDs  any                 `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any                 `json:"suggestedInsertionIds"`
}

// TableCellStyle represents styling for a table cell
type TableCellStyle struct {
	RowSpan          int64      `json:"rowSpan"`
	ColumnSpan       int64      `json:"columnSpan"`
	BackgroundColor  *Color     `json:"backgroundColor,omitempty"`
	BorderLeft       *Border    `json:"borderLeft,omitempty"`
	BorderRight      *Border    `json:"borderRight,omitempty"`
	BorderTop        *Border    `json:"borderTop,omitempty"`
	BorderBottom     *Border    `json:"borderBottom,omitempty"`
	PaddingLeft      *Dimension `json:"paddingLeft,omitempty"`
	PaddingRight     *Dimension `json:"paddingRight,omitempty"`
	PaddingTop       *Dimension `json:"paddingTop,omitempty"`
	PaddingBottom    *Dimension `json:"paddingBottom,omitempty"`
	ContentAlignment *string    `json:"contentAlignment,omitempty"`
}

// TableRowStyle represents styling for a table row
type TableRowStyle struct {
	MinRowHeight    Dimension `json:"minRowHeight"`
	TableHeader     any       `json:"tableHeader"`
	PreventOverflow any       `json:"preventOverflow"`
}

// TableStyle represents styling for a table
type TableStyle struct {
	TableColumnProperties []TableColumnProperty `json:"tableColumnProperties"`
}

// TableColumnProperty represents properties for a table column
type TableColumnProperty struct {
	WidthType string     `json:"widthType"`
	Width     *Dimension `json:"width,omitempty"`
}
