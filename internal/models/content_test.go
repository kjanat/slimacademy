package models

import (
	"encoding/json"
	"testing"
)

func TestContentUnionType(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		expectedType string // "document" or "chapters"
		expectedLen  int    // for chapters array
		wantErr      bool
	}{
		{
			name: "content as document",
			jsonData: `{
				"documentId": "doc123",
				"revisionId": "rev456",
				"title": "Test Document",
				"body": {
					"content": []
				},
				"headers": {},
				"footers": {},
				"documentStyle": {
					"background": {},
					"pageSize": {
						"height": {"magnitude": 11, "unit": "IN"},
						"width": {"magnitude": 8.5, "unit": "IN"}
					},
					"marginTop": {"magnitude": 1, "unit": "IN"},
					"marginBottom": {"magnitude": 1, "unit": "IN"},
					"marginRight": {"magnitude": 1, "unit": "IN"},
					"marginLeft": {"magnitude": 1, "unit": "IN"},
					"marginHeader": {"magnitude": 0.5, "unit": "IN"},
					"marginFooter": {"magnitude": 0.5, "unit": "IN"},
					"pageNumberStart": 1,
					"useCustomHeaderFooterMargins": false,
					"defaultHeaderId": "",
					"defaultFooterId": "",
					"firstPageHeaderId": "",
					"firstPageFooterId": ""
				},
				"namedStyles": {
					"styles": []
				},
				"lists": {},
				"inlineObjects": {},
				"positionedObjects": {}
			}`,
			expectedType: "document",
			expectedLen:  0,
			wantErr:      false,
		},
		{
			name: "content as chapters array",
			jsonData: `[
				{
					"id": 1,
					"summaryId": 100,
					"title": "Chapter 1",
					"isFree": 1,
					"isSupplement": 0,
					"isLocked": 0,
					"isVisible": 1,
					"parentChapterId": null,
					"gDocsChapterId": "gdocs1",
					"sortIndex": 0,
					"subChapters": []
				},
				{
					"id": 2,
					"summaryId": 100,
					"title": "Chapter 2",
					"isFree": 0,
					"isSupplement": 0,
					"isLocked": 1,
					"isVisible": 1,
					"parentChapterId": null,
					"gDocsChapterId": "gdocs2",
					"sortIndex": 1,
					"subChapters": [
						{
							"id": 3,
							"summaryId": 100,
							"title": "Chapter 2.1",
							"isFree": 0,
							"isSupplement": 0,
							"isLocked": 1,
							"isVisible": 1,
							"parentChapterId": 2,
							"gDocsChapterId": "gdocs3",
							"sortIndex": 0,
							"subChapters": []
						}
					]
				}
			]`,
			expectedType: "chapters",
			expectedLen:  2,
			wantErr:      false,
		},
		{
			name:         "empty document",
			jsonData:     `{}`,
			expectedType: "document",
			expectedLen:  0,
			wantErr:      false,
		},
		{
			name:         "empty chapters array",
			jsonData:     `[]`,
			expectedType: "chapters",
			expectedLen:  0,
			wantErr:      false,
		},
		{
			name:         "null content",
			jsonData:     `null`,
			expectedType: "null",
			expectedLen:  0,
			wantErr:      false,
		},
		{
			name:         "invalid JSON",
			jsonData:     `{"invalid": json}`,
			expectedType: "",
			expectedLen:  0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := UnmarshalContent([]byte(tt.jsonData))

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalContent() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalContent() error = %v", err)
				return
			}

			switch tt.expectedType {
			case "document":
				if content.Document == nil {
					t.Errorf("Expected Document but got nil")
				}
				if content.Chapters != nil {
					t.Errorf("Expected Chapters to be nil but got %v", content.Chapters)
				}
			case "chapters":
				if content.Chapters == nil {
					t.Errorf("Expected Chapters but got nil")
				}
				if content.Document != nil {
					t.Errorf("Expected Document to be nil but got %v", content.Document)
				}
				if len(content.Chapters) != tt.expectedLen {
					t.Errorf("Expected %d chapters but got %d", tt.expectedLen, len(content.Chapters))
				}
			case "null":
				if content.Document != nil || content.Chapters != nil {
					t.Errorf("Expected both Document and Chapters to be nil")
				}
			}
		})
	}
}

func TestContentMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  *Content
		expected string
		wantErr  bool
	}{
		{
			name: "marshal document",
			content: &Content{
				Document: &Document{
					DocumentID: "doc123",
					Title:      "Test Doc",
					Body:       Body{Content: []StructuralElement{}},
					Headers:    map[string]HeaderFooter{},
					Footers:    map[string]HeaderFooter{},
					DocumentStyle: DocumentStyle{
						Background: Background{},
						PageSize: Size{
							Height: Dimension{Magnitude: 11, Unit: "IN"},
							Width:  Dimension{Magnitude: 8.5, Unit: "IN"},
						},
						MarginTop:    Dimension{Magnitude: 1, Unit: "IN"},
						MarginBottom: Dimension{Magnitude: 1, Unit: "IN"},
						MarginRight:  Dimension{Magnitude: 1, Unit: "IN"},
						MarginLeft:   Dimension{Magnitude: 1, Unit: "IN"},
						MarginHeader: Dimension{Magnitude: 0.5, Unit: "IN"},
						MarginFooter: Dimension{Magnitude: 0.5, Unit: "IN"},
					},
					NamedStyles:       NamedStyles{Styles: []NamedStyle{}},
					Lists:             map[string]List{},
					InlineObjects:     map[string]InlineObject{},
					PositionedObjects: map[string]PositionedObject{},
				},
				Chapters: nil,
			},
			wantErr: false,
		},
		{
			name: "marshal chapters",
			content: &Content{
				Document: nil,
				Chapters: []Chapter{
					{
						ID:             1,
						SummaryID:      100,
						Title:          "Chapter 1",
						IsFree:         BoolInt(true),
						GDocsChapterID: "gdocs1",
						SubChapters:    []Chapter{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "marshal empty content",
			content: &Content{
				Document: nil,
				Chapters: nil,
			},
			expected: "null",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("json.Marshal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			if tt.expected != "" && string(data) != tt.expected {
				t.Errorf("json.Marshal() = %s, expected %s", string(data), tt.expected)
			}

			// Test roundtrip
			var unmarshaled Content
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Roundtrip unmarshal error = %v", err)
				return
			}

			// Verify the type is preserved
			if (tt.content.Document != nil) != (unmarshaled.Document != nil) {
				t.Errorf("Document presence mismatch after roundtrip")
			}
			if (tt.content.Chapters != nil) != (unmarshaled.Chapters != nil) {
				t.Errorf("Chapters presence mismatch after roundtrip")
			}
		})
	}
}

func TestContentUnmarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectDoc   bool
		expectChaps bool
		wantErr     bool
	}{
		{
			name:        "document with minimal fields",
			jsonData:    `{"documentId": "", "title": ""}`,
			expectDoc:   true,
			expectChaps: false,
			wantErr:     false,
		},
		{
			name:        "single chapter in array",
			jsonData:    `[{"id": 1, "summaryId": 1, "title": "Ch1", "gDocsChapterId": ""}]`,
			expectDoc:   false,
			expectChaps: true,
			wantErr:     false,
		},
		{
			name:        "document with extra fields",
			jsonData:    `{"documentId": "doc", "title": "title", "extraField": "ignored", "body": {"content": []}}`,
			expectDoc:   true,
			expectChaps: false,
			wantErr:     false,
		},
		{
			name:        "chapters with extra fields",
			jsonData:    `[{"id": 1, "summaryId": 1, "title": "Ch1", "gDocsChapterId": "", "extraField": "ignored"}]`,
			expectDoc:   false,
			expectChaps: true,
			wantErr:     false,
		},
		{
			name:        "ambiguous object that could be either",
			jsonData:    `{"id": 1}`,
			expectDoc:   true, // Should default to document when ambiguous
			expectChaps: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var content Content
			err := json.Unmarshal([]byte(tt.jsonData), &content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("json.Unmarshal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}

			hasDoc := content.Document != nil
			hasChaps := content.Chapters != nil

			if hasDoc != tt.expectDoc {
				t.Errorf("Document presence = %v, expected %v", hasDoc, tt.expectDoc)
			}
			if hasChaps != tt.expectChaps {
				t.Errorf("Chapters presence = %v, expected %v", hasChaps, tt.expectChaps)
			}
		})
	}
}

func TestContentTypeDiscrimination(t *testing.T) {
	// Test that the unmarshal logic correctly identifies the type

	t.Run("chapters array takes precedence", func(t *testing.T) {
		// Valid JSON that could parse as both but should be chapters
		jsonData := `[{"id": 1, "summaryId": 1, "title": "Test", "gDocsChapterId": ""}]`

		var content Content
		err := json.Unmarshal([]byte(jsonData), &content)
		if err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if content.Document != nil {
			t.Errorf("Expected Document to be nil for chapters array")
		}
		if content.Chapters == nil {
			t.Errorf("Expected Chapters to be non-nil for chapters array")
		}
		if len(content.Chapters) != 1 {
			t.Errorf("Expected 1 chapter, got %d", len(content.Chapters))
		}
	})

	t.Run("document when not valid chapters", func(t *testing.T) {
		// Valid document JSON that cannot be chapters
		jsonData := `{"documentId": "doc123", "title": "Test Doc", "revisionId": "rev123"}`

		var content Content
		err := json.Unmarshal([]byte(jsonData), &content)
		if err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if content.Document == nil {
			t.Errorf("Expected Document to be non-nil for document JSON")
		}
		if content.Chapters != nil {
			t.Errorf("Expected Chapters to be nil for document JSON")
		}
		if content.Document.DocumentID != "doc123" {
			t.Errorf("DocumentID = %v, expected doc123", content.Document.DocumentID)
		}
	})
}

func TestContentValidation(t *testing.T) {
	tests := []struct {
		name    string
		content *Content
		isValid bool
	}{
		{
			name: "valid document content",
			content: &Content{
				Document: &Document{
					DocumentID: "doc123",
					Title:      "Valid Doc",
				},
				Chapters: nil,
			},
			isValid: true,
		},
		{
			name: "valid chapters content",
			content: &Content{
				Document: nil,
				Chapters: []Chapter{
					{ID: 1, Title: "Chapter 1", GDocsChapterID: "gdocs1"},
				},
			},
			isValid: true,
		},
		{
			name: "invalid - both document and chapters",
			content: &Content{
				Document: &Document{DocumentID: "doc"},
				Chapters: []Chapter{{ID: 1}},
			},
			isValid: false,
		},
		{
			name: "valid - empty content (both nil)",
			content: &Content{
				Document: nil,
				Chapters: nil,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Custom validation logic
			isValid := (tt.content.Document == nil) != (tt.content.Chapters == nil) ||
				(tt.content.Document == nil && tt.content.Chapters == nil)

			if isValid != tt.isValid {
				t.Errorf("Content validation = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}

func TestContentWithNestedChapters(t *testing.T) {
	jsonData := `[
		{
			"id": 1,
			"summaryId": 100,
			"title": "Parent Chapter",
			"isFree": 1,
			"isSupplement": 0,
			"isLocked": 0,
			"isVisible": 1,
			"parentChapterId": null,
			"gDocsChapterId": "parent",
			"sortIndex": 0,
			"subChapters": [
				{
					"id": 2,
					"summaryId": 100,
					"title": "Sub Chapter 1",
					"isFree": 1,
					"isSupplement": 0,
					"isLocked": 0,
					"isVisible": 1,
					"parentChapterId": 1,
					"gDocsChapterId": "sub1",
					"sortIndex": 0,
					"subChapters": []
				},
				{
					"id": 3,
					"summaryId": 100,
					"title": "Sub Chapter 2",
					"isFree": 0,
					"isSupplement": 0,
					"isLocked": 1,
					"isVisible": 1,
					"parentChapterId": 1,
					"gDocsChapterId": "sub2",
					"sortIndex": 1,
					"subChapters": []
				}
			]
		}
	]`

	content, err := UnmarshalContent([]byte(jsonData))
	if err != nil {
		t.Fatalf("UnmarshalContent() error = %v", err)
	}

	if content.Chapters == nil {
		t.Fatal("Expected Chapters to be non-nil")
	}

	if len(content.Chapters) != 1 {
		t.Errorf("Expected 1 parent chapter, got %d", len(content.Chapters))
	}

	parent := content.Chapters[0]
	if parent.Title != "Parent Chapter" {
		t.Errorf("Parent title = %v, expected 'Parent Chapter'", parent.Title)
	}

	if len(parent.SubChapters) != 2 {
		t.Errorf("Expected 2 sub chapters, got %d", len(parent.SubChapters))
	}

	if parent.SubChapters[0].Title != "Sub Chapter 1" {
		t.Errorf("Sub chapter 1 title = %v, expected 'Sub Chapter 1'", parent.SubChapters[0].Title)
	}

	if !parent.SubChapters[1].IsLocked.Bool() {
		t.Errorf("Sub chapter 2 IsLocked = %v, expected true", parent.SubChapters[1].IsLocked)
	}
}
