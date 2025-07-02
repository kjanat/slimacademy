package models

import "encoding/json"

// Document represents a Google Docs-style document
type Document struct {
	DocumentID          string                      `json:"documentId"`
	RevisionID          string                      `json:"revisionId"`
	SuggestionsViewMode string                      `json:"suggestionsViewMode"`
	Title               string                      `json:"title"`
	Body                Body                        `json:"body"`
	Headers             map[string]HeaderFooter     `json:"headers"`
	Footers             map[string]HeaderFooter     `json:"footers"`
	DocumentStyle       DocumentStyle               `json:"documentStyle"`
	NamedStyles         NamedStyles                 `json:"namedStyles"`
	Lists               map[string]List             `json:"lists"`
	InlineObjects       map[string]InlineObject     `json:"inlineObjects"`
	PositionedObjects   map[string]PositionedObject `json:"positionedObjects"`
}

// Body represents the document body
type Body struct {
	Content []StructuralElement `json:"content"`
}

// StructuralElement represents a top-level element in the document
type StructuralElement struct {
	StartIndex      int64            `json:"startIndex"`
	EndIndex        int64            `json:"endIndex"`
	Paragraph       *Paragraph       `json:"paragraph,omitempty"`
	Table           *Table           `json:"table,omitempty"`
	SectionBreak    *SectionBreak    `json:"sectionBreak,omitempty"`
	TableOfContents *TableOfContents `json:"tableOfContents,omitempty"`
}

// HeaderFooter represents a header or footer
type HeaderFooter struct {
	HeaderID string              `json:"headerId,omitempty"`
	FooterID string              `json:"footerId,omitempty"`
	Content  []StructuralElement `json:"content"`
}

// TableOfContents represents a table of contents
type TableOfContents struct {
	Content               []StructuralElement `json:"content"`
	SuggestedDeletionIDs  any                 `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any                 `json:"suggestedInsertionIds"`
}

// SectionBreak represents a section break
type SectionBreak struct {
	SectionStyle          SectionStyle `json:"sectionStyle"`
	SuggestedDeletionIDs  any          `json:"suggestedDeletionIds"`
	SuggestedInsertionIDs any          `json:"suggestedInsertionIds"`
}

// SectionStyle represents styling for a section
type SectionStyle struct {
	ColumnSeparatorStyle     string `json:"columnSeparatorStyle"`
	ContentDirection         string `json:"contentDirection"`
	SectionType              string `json:"sectionType"`
	DefaultHeaderID          any    `json:"defaultHeaderId"`
	DefaultFooterID          any    `json:"defaultFooterId"`
	FirstPageHeaderID        any    `json:"firstPageHeaderId"`
	FirstPageFooterID        any    `json:"firstPageFooterId"`
	EvenPageHeaderID         any    `json:"evenPageHeaderId"`
	EvenPageFooterID         any    `json:"evenPageFooterId"`
	PageNumberStart          any    `json:"pageNumberStart"`
	UseFirstPageHeaderFooter any    `json:"useFirstPageHeaderFooter"`
}

// UnmarshalDocument unmarshals JSON data into a Document
func UnmarshalDocument(data []byte) (*Document, error) {
	var doc Document
	err := json.Unmarshal(data, &doc)
	return &doc, err
}
