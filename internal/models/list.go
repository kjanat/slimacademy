package models

// List represents a list definition
type List struct {
	ListProperties       ListProperties `json:"listProperties"`
	SuggestedDeletionIDs any            `json:"suggestedDeletionIds"`
	SuggestedInsertionID any            `json:"suggestedInsertionId"`
}

// ListProperties represents properties of a list
type ListProperties struct {
	NestingLevels []NestingLevel `json:"nestingLevels"`
}

// NestingLevel represents a nesting level in a list
type NestingLevel struct {
	BulletAlignment string    `json:"bulletAlignment"`
	GlyphType       *string   `json:"glyphType,omitempty"`
	GlyphFormat     *string   `json:"glyphFormat,omitempty"`
	GlyphSymbol     *string   `json:"glyphSymbol,omitempty"`
	StartNumber     int64     `json:"startNumber"`
	IndentFirstLine Dimension `json:"indentFirstLine"`
	IndentStart     Dimension `json:"indentStart"`
	TextStyle       TextStyle `json:"textStyle"`
}
