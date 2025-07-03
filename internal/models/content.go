package models

import (
	"encoding/json"
	"fmt"
)

// Content represents either a Document or array of Chapters
type Content struct {
	Document *Document
	Chapters []Chapter
}

// UnmarshalJSON handles the union type for Content
func (c *Content) UnmarshalJSON(data []byte) error {
	// First check what type of data we have
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if it's an array (chapters)
	if len(raw) > 0 && raw[0] == '[' {
		var chapters []Chapter
		if err := json.Unmarshal(raw, &chapters); err != nil {
			return fmt.Errorf("failed to unmarshal chapters array: %w", err)
		}
		c.Chapters = chapters
		c.Document = nil
		return nil
	}

	// Otherwise, it should be a document object
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		// If direct unmarshal fails, check if it's an object at all
		var test map[string]interface{}
		if testErr := json.Unmarshal(raw, &test); testErr != nil {
			return fmt.Errorf("content is neither an array nor an object: %w", testErr)
		}
		// It's an object but failed to unmarshal as Document
		return fmt.Errorf("failed to unmarshal as Document: %w", err)
	}

	c.Document = &doc
	c.Chapters = nil
	return nil
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
