package models

import (
	"encoding/json"
	"testing"
)

func TestTextStyleSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected TextStyle
		wantErr  bool
	}{
		{
			name: "complete text style",
			jsonData: `{
				"bold": true,
				"italic": false,
				"underline": true,
				"strikethrough": false,
				"smallCaps": true,
				"backgroundColor": {
					"color": {
						"red": 1.0,
						"green": 0.0,
						"blue": 0.0
					}
				},
				"foregroundColor": {
					"color": {
						"red": 0.0,
						"green": 1.0,
						"blue": 0.0
					}
				},
				"fontSize": {
					"magnitude": 12,
					"unit": "PT"
				},
				"weightedFontFamily": {
					"fontFamily": "Arial",
					"weight": 400
				},
				"baselineOffset": "SUBSCRIPT",
				"link": {
					"url": "https://example.com"
				}
			}`,
			expected: TextStyle{
				Bold:          &[]bool{true}[0],
				Italic:        &[]bool{false}[0],
				Underline:     &[]bool{true}[0],
				Strikethrough: &[]bool{false}[0],
				SmallCaps:     &[]bool{true}[0],
				BackgroundColor: &Color{
					Color: &RGBColor{
						Red:   &[]float64{1.0}[0],
						Green: &[]float64{0.0}[0],
						Blue:  &[]float64{0.0}[0],
					},
				},
				ForegroundColor: &Color{
					Color: &RGBColor{
						Red:   &[]float64{0.0}[0],
						Green: &[]float64{1.0}[0],
						Blue:  &[]float64{0.0}[0],
					},
				},
				FontSize: &Dimension{
					Magnitude: 12,
					Unit:      "PT",
				},
				WeightedFontFamily: &WeightedFontFamily{
					FontFamily: "Arial",
					Weight:     400,
				},
				BaselineOffset: &[]string{"SUBSCRIPT"}[0],
				Link: &Link{
					URL: &[]string{"https://example.com"}[0],
				},
			},
			wantErr: false,
		},
		{
			name:     "empty text style",
			jsonData: `{}`,
			expected: TextStyle{},
			wantErr:  false,
		},
		{
			name: "text style with null values",
			jsonData: `{
				"bold": null,
				"italic": null,
				"fontSize": null,
				"link": null
			}`,
			expected: TextStyle{
				Bold:     nil,
				Italic:   nil,
				FontSize: nil,
				Link:     nil,
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"bold": "invalid"}`,
			expected: TextStyle{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var style TextStyle
			err := json.Unmarshal([]byte(tt.jsonData), &style)

			if tt.wantErr {
				if err == nil {
					t.Errorf("TextStyle unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("TextStyle unmarshal error = %v", err)
				return
			}

			// Test boolean pointers
			if !compareBoolPointers(style.Bold, tt.expected.Bold) {
				t.Errorf("Bold mismatch")
			}
			if !compareBoolPointers(style.Italic, tt.expected.Italic) {
				t.Errorf("Italic mismatch")
			}
			if !compareBoolPointers(style.Underline, tt.expected.Underline) {
				t.Errorf("Underline mismatch")
			}

			// Test font size
			if !compareDimensionPointers(style.FontSize, tt.expected.FontSize) {
				t.Errorf("FontSize mismatch")
			}

			// Test font family
			if !compareWeightedFontFamilyPointers(style.WeightedFontFamily, tt.expected.WeightedFontFamily) {
				t.Errorf("WeightedFontFamily mismatch")
			}

			// Test link
			if !compareLinkPointers(style.Link, tt.expected.Link) {
				t.Errorf("Link mismatch")
			}
		})
	}
}

func TestParagraphStyleSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected ParagraphStyle
		wantErr  bool
	}{
		{
			name: "complete paragraph style",
			jsonData: `{
				"namedStyleType": "HEADING_1",
				"alignment": "CENTER",
				"lineSpacing": 1.5,
				"direction": "LEFT_TO_RIGHT",
				"spacingMode": "COLLAPSED",
				"spaceAbove": {"magnitude": 12, "unit": "PT"},
				"spaceBelow": {"magnitude": 6, "unit": "PT"},
				"indentFirstLine": {"magnitude": 0.5, "unit": "IN"},
				"indentStart": {"magnitude": 0, "unit": "PT"},
				"indentEnd": {"magnitude": 0, "unit": "PT"},
				"tabStops": [
					{
						"offset": {"magnitude": 1, "unit": "IN"},
						"alignment": "START"
					}
				],
				"borderTop": {
					"color": {"color": {"red": 0, "green": 0, "blue": 0}},
					"width": {"magnitude": 1, "unit": "PT"},
					"padding": {"magnitude": 0, "unit": "PT"},
					"dashStyle": "SOLID"
				},
				"shading": {
					"backgroundColor": {
						"color": {"red": 0.9, "green": 0.9, "blue": 0.9}
					}
				},
				"headingId": "heading_123",
				"avoidWidowAndOrphan": true,
				"keepLinesTogether": false,
				"keepWithNext": true,
				"pageBreakBefore": false
			}`,
			expected: ParagraphStyle{
				NamedStyleType:  "HEADING_1",
				Alignment:       &[]string{"CENTER"}[0],
				LineSpacing:     &[]float64{1.5}[0],
				Direction:       "LEFT_TO_RIGHT",
				SpacingMode:     &[]string{"COLLAPSED"}[0],
				SpaceAbove:      &Dimension{Magnitude: 12, Unit: "PT"},
				SpaceBelow:      &Dimension{Magnitude: 6, Unit: "PT"},
				IndentFirstLine: &Dimension{Magnitude: 0.5, Unit: "IN"},
				IndentStart:     &Dimension{Magnitude: 0, Unit: "PT"},
				IndentEnd:       &Dimension{Magnitude: 0, Unit: "PT"},
				TabStops: []TabStop{
					{
						Offset:    Dimension{Magnitude: 1, Unit: "IN"},
						Alignment: "START",
					},
				},
				HeadingID:           &[]string{"heading_123"}[0],
				AvoidWidowAndOrphan: &[]bool{true}[0],
				KeepLinesTogether:   &[]bool{false}[0],
				KeepWithNext:        &[]bool{true}[0],
				PageBreakBefore:     &[]bool{false}[0],
			},
			wantErr: false,
		},
		{
			name: "minimal paragraph style",
			jsonData: `{
				"namedStyleType": "NORMAL_TEXT",
				"direction": "LEFT_TO_RIGHT"
			}`,
			expected: ParagraphStyle{
				NamedStyleType: "NORMAL_TEXT",
				Direction:      "LEFT_TO_RIGHT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var style ParagraphStyle
			err := json.Unmarshal([]byte(tt.jsonData), &style)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParagraphStyle unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParagraphStyle unmarshal error = %v", err)
				return
			}

			if style.NamedStyleType != tt.expected.NamedStyleType {
				t.Errorf("NamedStyleType = %v, expected %v",
					style.NamedStyleType, tt.expected.NamedStyleType)
			}
			if style.Direction != tt.expected.Direction {
				t.Errorf("Direction = %v, expected %v",
					style.Direction, tt.expected.Direction)
			}

			// Test pointer fields
			if !compareStringPointers(style.Alignment, tt.expected.Alignment) {
				t.Errorf("Alignment mismatch")
			}
			if !compareFloat64Pointers(style.LineSpacing, tt.expected.LineSpacing) {
				t.Errorf("LineSpacing mismatch")
			}
			if !compareStringPointers(style.HeadingID, tt.expected.HeadingID) {
				t.Errorf("HeadingID mismatch")
			}

			// Test tab stops
			if len(style.TabStops) != len(tt.expected.TabStops) {
				t.Errorf("TabStops length = %v, expected %v",
					len(style.TabStops), len(tt.expected.TabStops))
			}
		})
	}
}

func TestColorSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Color
		wantErr  bool
	}{
		{
			name: "RGB color",
			jsonData: `{
				"color": {
					"red": 0.5,
					"green": 0.7,
					"blue": 0.9
				}
			}`,
			expected: Color{
				Color: &RGBColor{
					Red:   &[]float64{0.5}[0],
					Green: &[]float64{0.7}[0],
					Blue:  &[]float64{0.9}[0],
				},
			},
			wantErr: false,
		},
		{
			name: "color with missing components",
			jsonData: `{
				"color": {
					"red": 1.0
				}
			}`,
			expected: Color{
				Color: &RGBColor{
					Red:   &[]float64{1.0}[0],
					Green: nil,
					Blue:  nil,
				},
			},
			wantErr: false,
		},
		{
			name:     "empty color",
			jsonData: `{}`,
			expected: Color{Color: nil},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var color Color
			err := json.Unmarshal([]byte(tt.jsonData), &color)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Color unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Color unmarshal error = %v", err)
				return
			}

			if (color.Color == nil) != (tt.expected.Color == nil) {
				t.Errorf("Color presence mismatch")
			}

			if color.Color != nil && tt.expected.Color != nil {
				if !compareFloat64Pointers(color.Color.Red, tt.expected.Color.Red) {
					t.Errorf("Red component mismatch")
				}
				if !compareFloat64Pointers(color.Color.Green, tt.expected.Color.Green) {
					t.Errorf("Green component mismatch")
				}
				if !compareFloat64Pointers(color.Color.Blue, tt.expected.Color.Blue) {
					t.Errorf("Blue component mismatch")
				}
			}
		})
	}
}

func TestLinkSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Link
		wantErr  bool
	}{
		{
			name: "URL link",
			jsonData: `{
				"url": "https://example.com"
			}`,
			expected: Link{
				URL:        &[]string{"https://example.com"}[0],
				BookmarkID: nil,
				HeadingID:  nil,
			},
			wantErr: false,
		},
		{
			name: "bookmark link",
			jsonData: `{
				"bookmarkId": "bookmark_123"
			}`,
			expected: Link{
				URL:        nil,
				BookmarkID: &[]string{"bookmark_123"}[0],
				HeadingID:  nil,
			},
			wantErr: false,
		},
		{
			name: "heading link",
			jsonData: `{
				"headingId": "heading_456"
			}`,
			expected: Link{
				URL:        nil,
				BookmarkID: nil,
				HeadingID:  &[]string{"heading_456"}[0],
			},
			wantErr: false,
		},
		{
			name: "multiple link types (should be valid)",
			jsonData: `{
				"url": "https://example.com",
				"bookmarkId": "bookmark_123",
				"headingId": "heading_456"
			}`,
			expected: Link{
				URL:        &[]string{"https://example.com"}[0],
				BookmarkID: &[]string{"bookmark_123"}[0],
				HeadingID:  &[]string{"heading_456"}[0],
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var link Link
			err := json.Unmarshal([]byte(tt.jsonData), &link)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Link unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Link unmarshal error = %v", err)
				return
			}

			if !compareStringPointers(link.URL, tt.expected.URL) {
				t.Errorf("URL mismatch")
			}
			if !compareStringPointers(link.BookmarkID, tt.expected.BookmarkID) {
				t.Errorf("BookmarkID mismatch")
			}
			if !compareStringPointers(link.HeadingID, tt.expected.HeadingID) {
				t.Errorf("HeadingID mismatch")
			}
		})
	}
}

func TestBorderSerialization(t *testing.T) {
	jsonData := `{
		"color": {
			"color": {
				"red": 0.0,
				"green": 0.0,
				"blue": 0.0
			}
		},
		"width": {
			"magnitude": 2,
			"unit": "PT"
		},
		"padding": {
			"magnitude": 1,
			"unit": "PT"
		},
		"dashStyle": "SOLID"
	}`

	var border Border
	err := json.Unmarshal([]byte(jsonData), &border)
	if err != nil {
		t.Fatalf("Border unmarshal error = %v", err)
	}

	if border.Width.Magnitude != 2 {
		t.Errorf("Width magnitude = %v, expected 2", border.Width.Magnitude)
	}
	if border.Width.Unit != "PT" {
		t.Errorf("Width unit = %v, expected PT", border.Width.Unit)
	}
	if border.DashStyle != "SOLID" {
		t.Errorf("DashStyle = %v, expected SOLID", border.DashStyle)
	}

	// Test color
	if border.Color == nil || border.Color.Color == nil {
		t.Errorf("Expected color to be present")
	}
}

func TestNamedStyleSerialization(t *testing.T) {
	jsonData := `{
		"namedStyleType": "HEADING_1",
		"textStyle": {
			"bold": true,
			"fontSize": {
				"magnitude": 18,
				"unit": "PT"
			}
		},
		"paragraphStyle": {
			"namedStyleType": "HEADING_1",
			"direction": "LEFT_TO_RIGHT",
			"alignment": "START"
		}
	}`

	var namedStyle NamedStyle
	err := json.Unmarshal([]byte(jsonData), &namedStyle)
	if err != nil {
		t.Fatalf("NamedStyle unmarshal error = %v", err)
	}

	if namedStyle.NamedStyleType != "HEADING_1" {
		t.Errorf("NamedStyleType = %v, expected HEADING_1", namedStyle.NamedStyleType)
	}

	// Test text style
	if namedStyle.TextStyle.Bold == nil || !*namedStyle.TextStyle.Bold {
		t.Errorf("Expected TextStyle.Bold to be true")
	}
	if namedStyle.TextStyle.FontSize == nil || namedStyle.TextStyle.FontSize.Magnitude != 18 {
		t.Errorf("Expected FontSize to be 18PT")
	}

	// Test paragraph style
	if namedStyle.ParagraphStyle.NamedStyleType != "HEADING_1" {
		t.Errorf("ParagraphStyle.NamedStyleType = %v, expected HEADING_1",
			namedStyle.ParagraphStyle.NamedStyleType)
	}
}

func TestStyleMarshalRoundtrip(t *testing.T) {
	// Test complex style with all features
	original := TextStyle{
		Bold:      &[]bool{true}[0],
		Italic:    &[]bool{false}[0],
		Underline: &[]bool{true}[0],
		FontSize: &Dimension{
			Magnitude: 14,
			Unit:      "PT",
		},
		ForegroundColor: &Color{
			Color: &RGBColor{
				Red:   &[]float64{0.2}[0],
				Green: &[]float64{0.4}[0],
				Blue:  &[]float64{0.8}[0],
			},
		},
		WeightedFontFamily: &WeightedFontFamily{
			FontFamily: "Times New Roman",
			Weight:     700,
		},
		Link: &Link{
			URL: &[]string{"https://test.com"}[0],
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var roundtrip TextStyle
	err = json.Unmarshal(data, &roundtrip)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify roundtrip
	if !compareBoolPointers(original.Bold, roundtrip.Bold) {
		t.Errorf("Bold roundtrip failed")
	}
	if !compareBoolPointers(original.Italic, roundtrip.Italic) {
		t.Errorf("Italic roundtrip failed")
	}
	if !compareDimensionPointers(original.FontSize, roundtrip.FontSize) {
		t.Errorf("FontSize roundtrip failed")
	}
	if !compareWeightedFontFamilyPointers(original.WeightedFontFamily, roundtrip.WeightedFontFamily) {
		t.Errorf("WeightedFontFamily roundtrip failed")
	}
	if !compareLinkPointers(original.Link, roundtrip.Link) {
		t.Errorf("Link roundtrip failed")
	}
}

// Helper functions for comparing pointer values
func compareBoolPointers(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareFloat64Pointers(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareStringPointers(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareDimensionPointers(a, b *Dimension) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Magnitude == b.Magnitude && a.Unit == b.Unit
}

func compareWeightedFontFamilyPointers(a, b *WeightedFontFamily) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.FontFamily == b.FontFamily && a.Weight == b.Weight
}

func compareLinkPointers(a, b *Link) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return compareStringPointers(a.URL, b.URL) &&
		compareStringPointers(a.BookmarkID, b.BookmarkID) &&
		compareStringPointers(a.HeadingID, b.HeadingID)
}
