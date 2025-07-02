package models

import "encoding/json"

// Content represents either a Document or array of Chapters
type Content struct {
	Document *Document
	Chapters []Chapter
}

// UnmarshalJSON handles the union type for Content
func (c *Content) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array of chapters first
	var chapters []Chapter
	if err := json.Unmarshal(data, &chapters); err == nil {
		c.Chapters = chapters
		c.Document = nil
		return nil
	}

	// Try to unmarshal as document
	var doc Document
	if err := json.Unmarshal(data, &doc); err == nil {
		c.Document = &doc
		c.Chapters = nil
		return nil
	}

	return json.Unmarshal(data, &c)
}

// MarshalJSON handles the union type for Content
func (c Content) MarshalJSON() ([]byte, error) {
	if c.Chapters != nil {
		return json.Marshal(c.Chapters)
	}
	if c.Document != nil {
		return json.Marshal(c.Document)
	}
	return json.Marshal(nil)
}

// UnmarshalContent unmarshals JSON data into Content (handles union type)
func UnmarshalContent(data []byte) (*Content, error) {
	var content Content
	err := json.Unmarshal(data, &content)
	return &content, err
}
