package models

// Paragraph represents a paragraph element
type Paragraph struct {
	Elements            []ParagraphElement `json:"elements"`
	ParagraphStyle      ParagraphStyle     `json:"paragraphStyle"`
	Bullet              *Bullet            `json:"bullet,omitempty"`
	PositionedObjectIDs any                `json:"positionedObjectIds"`
}

// ParagraphElement represents an element within a paragraph
type ParagraphElement struct {
	StartIndex          int64                `json:"startIndex"`
	EndIndex            int64                `json:"endIndex"`
	TextRun             *TextRun             `json:"textRun,omitempty"`
	InlineObjectElement *InlineObjectElement `json:"inlineObjectElement,omitempty"`
	PageBreak           *PageBreak           `json:"pageBreak,omitempty"`
	HorizontalRule      *HorizontalRule      `json:"horizontalRule,omitempty"`
	AutoText            *AutoText            `json:"autoText,omitempty"`
}

// TextRun represents a run of text with consistent styling
type TextRun struct {
	Content               string    `json:"content"`
	TextStyle             TextStyle `json:"textStyle"`
	SuggestedDeletionIDs  any       `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any       `json:"suggestedInsertionIds"`
}

// Bullet represents a list bullet
type Bullet struct {
	ListID       string    `json:"listId"`
	NestingLevel *int64    `json:"nestingLevel,omitempty"`
	TextStyle    TextStyle `json:"textStyle"`
}

// InlineObjectElement represents an inline object element
type InlineObjectElement struct {
	InlineObjectID        string    `json:"inlineObjectId"`
	TextStyle             TextStyle `json:"textStyle"`
	SuggestedDeletionIDs  any       `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any       `json:"suggestedInsertionIds"`
}

// PageBreak represents a page break
type PageBreak struct {
	TextStyle             TextStyle `json:"textStyle"`
	SuggestedDeletionIDs  any       `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any       `json:"suggestedInsertionIds"`
}

// HorizontalRule represents a horizontal rule
type HorizontalRule struct {
	TextStyle             TextStyle `json:"textStyle"`
	SuggestedDeletionIDs  any       `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any       `json:"suggestedInsertionIds"`
}

// AutoText represents automatic text (like page numbers)
type AutoText struct {
	Type                  string    `json:"type"`
	TextStyle             TextStyle `json:"textStyle"`
	SuggestedDeletionIDs  any       `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any       `json:"suggestedInsertionIds"`
}
