package models

import (
	"encoding/json"
	"fmt"
)

// Color represents a color value that can be either an object or an array
type Color struct {
	Color *RGBColor `json:"color,omitempty"`
}

// UnmarshalJSON handles both object and array representations of Color
func (c *Color) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as the expected object format
	type colorAlias Color
	var color colorAlias
	if err := json.Unmarshal(data, &color); err == nil {
		*c = Color(color)
		return nil
	}

	// If that fails, check if it's an array (legacy format)
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		// Handle array format - typically empty arrays in Google Docs
		if len(arr) == 0 {
			// Empty array means no color
			*c = Color{}
			return nil
		}
		// If non-empty array, we'd need to handle the specific format
		return fmt.Errorf("non-empty color array format not implemented")
	}

	// If neither worked, return error
	return fmt.Errorf("color must be either an object or an array")
}

// MarshalJSON marshals Color back to JSON
func (c Color) MarshalJSON() ([]byte, error) {
	if c.Color == nil {
		// Return null for empty color
		return []byte("null"), nil
	}
	// Use default marshaling
	type colorAlias Color
	return json.Marshal(colorAlias(c))
}
