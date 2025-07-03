package models

import (
	"encoding/json"
	"testing"
)

func TestBoolInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BoolInt
		wantErr  bool
	}{
		{
			name:     "integer 1 to true",
			input:    "1",
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "integer 0 to false",
			input:    "0",
			expected: BoolInt(false),
			wantErr:  false,
		},
		{
			name:     "string 1 to true",
			input:    `"1"`,
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "string 0 to false",
			input:    `"0"`,
			expected: BoolInt(false),
			wantErr:  false,
		},
		{
			name:     "string true to true",
			input:    `"true"`,
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "string false to false",
			input:    `"false"`,
			expected: BoolInt(false),
			wantErr:  false,
		},
		{
			name:     "boolean true to true",
			input:    "true",
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "boolean false to false",
			input:    "false",
			expected: BoolInt(false),
			wantErr:  false,
		},
		{
			name:     "float 1.0 to true",
			input:    "1.0",
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "float 0.0 to false",
			input:    "0.0",
			expected: BoolInt(false),
			wantErr:  false,
		},
		{
			name:     "negative number to true",
			input:    "-1",
			expected: BoolInt(true),
			wantErr:  false,
		},
		{
			name:     "large number to true",
			input:    "42",
			expected: BoolInt(true),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b BoolInt
			err := json.Unmarshal([]byte(tt.input), &b)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if b != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestBoolInt_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    BoolInt
		expected string
	}{
		{
			name:     "true to 1",
			input:    BoolInt(true),
			expected: "1",
		},
		{
			name:     "false to 0",
			input:    BoolInt(false),
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if string(result) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}

func TestBoolInt_Bool(t *testing.T) {
	tests := []struct {
		name     string
		input    BoolInt
		expected bool
	}{
		{
			name:     "BoolInt(true) to true",
			input:    BoolInt(true),
			expected: true,
		},
		{
			name:     "BoolInt(false) to false",
			input:    BoolInt(false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Bool()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBoolInt_String(t *testing.T) {
	tests := []struct {
		name     string
		input    BoolInt
		expected string
	}{
		{
			name:     "BoolInt(true) to 1",
			input:    BoolInt(true),
			expected: "1",
		},
		{
			name:     "BoolInt(false) to 0",
			input:    BoolInt(false),
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestChapter_BoolIntIntegration(t *testing.T) {
	// Test that Chapter can properly handle BoolInt fields from JSON
	jsonData := `{
		"id": 1,
		"summaryId": 123,
		"title": "Test Chapter",
		"isFree": 1,
		"isSupplement": 0,
		"isLocked": 1,
		"isVisible": 0,
		"gDocsChapterId": "test-chapter",
		"sortIndex": 1,
		"subChapters": []
	}`

	var chapter Chapter
	err := json.Unmarshal([]byte(jsonData), &chapter)
	if err != nil {
		t.Fatalf("Failed to unmarshal chapter: %v", err)
	}

	// Verify boolean conversions
	if !chapter.IsFree.Bool() {
		t.Error("Expected IsFree to be true")
	}
	if chapter.IsSupplement.Bool() {
		t.Error("Expected IsSupplement to be false")
	}
	if !chapter.IsLocked.Bool() {
		t.Error("Expected IsLocked to be true")
	}
	if chapter.IsVisible.Bool() {
		t.Error("Expected IsVisible to be false")
	}

	// Test marshaling back to JSON
	data, err := json.Marshal(chapter)
	if err != nil {
		t.Fatalf("Failed to marshal chapter: %v", err)
	}

	// Verify that boolean fields are serialized as 0/1
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["isFree"].(float64) != 1 {
		t.Errorf("Expected isFree to be 1, got %v", result["isFree"])
	}
	if result["isSupplement"].(float64) != 0 {
		t.Errorf("Expected isSupplement to be 0, got %v", result["isSupplement"])
	}
	if result["isLocked"].(float64) != 1 {
		t.Errorf("Expected isLocked to be 1, got %v", result["isLocked"])
	}
	if result["isVisible"].(float64) != 0 {
		t.Errorf("Expected isVisible to be 0, got %v", result["isVisible"])
	}
}

func TestBook_BoolIntIntegration(t *testing.T) {
	// Test that Book can properly handle BoolInt fields from JSON
	jsonData := `{
		"id": 1,
		"title": "Test Book",
		"description": "A test book",
		"isPurchased": 1,
		"pageCount": 100,
		"hasFreeChapters": 0,
		"supplements": [],
		"images": [],
		"formulasImages": [],
		"periods": []
	}`

	var book Book
	err := json.Unmarshal([]byte(jsonData), &book)
	if err != nil {
		t.Fatalf("Failed to unmarshal book: %v", err)
	}

	// Verify boolean conversions
	if !book.IsPurchased.Bool() {
		t.Error("Expected IsPurchased to be true")
	}
	if book.HasFreeChapters.Bool() {
		t.Error("Expected HasFreeChapters to be false")
	}

	// Test marshaling back to JSON
	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal book: %v", err)
	}

	// Verify that boolean fields are serialized as 0/1
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["isPurchased"].(float64) != 1 {
		t.Errorf("Expected isPurchased to be 1, got %v", result["isPurchased"])
	}
	if result["hasFreeChapters"].(float64) != 0 {
		t.Errorf("Expected hasFreeChapters to be 0, got %v", result["hasFreeChapters"])
	}
}
