package models

import (
	"encoding/json"
	"testing"
)

func TestBook_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name: "valid complete book",
			jsonData: `{
				"id": 123,
				"title": "Test Book",
				"description": "A test book",
				"availableDate": "2025-01-01",
				"examDate": "2025-01-15",
				"bachelorYearNumber": "Bachelor 1",
				"collegeStartYear": 2024,
				"shopUrl": "/shop/test",
				"isPurchased": 1,
				"lastOpenedAt": null,
				"readProgress": null,
				"pageCount": 10,
				"readPageCount": null,
				"readPercentage": null,
				"hasFreeChapters": 0,
				"supplements": [],
				"images": [],
				"formulasImages": [],
				"periods": ["Test Period"]
			}`,
			wantErr: false,
		},
		{
			name: "book with minimal fields",
			jsonData: `{
				"id": 456,
				"title": "Minimal Book",
				"description": "",
				"availableDate": "",
				"examDate": "",
				"bachelorYearNumber": "",
				"collegeStartYear": 0,
				"shopUrl": "",
				"isPurchased": 0,
				"pageCount": 0,
				"hasFreeChapters": 0,
				"supplements": [],
				"images": [],
				"formulasImages": [],
				"periods": []
			}`,
			wantErr: false,
		},
		{
			name: "book with null optional fields",
			jsonData: `{
				"id": 789,
				"title": "Null Fields Book",
				"description": "Test with nulls",
				"availableDate": "2025-01-01",
				"examDate": "2025-01-15",
				"bachelorYearNumber": "Bachelor 1",
				"collegeStartYear": 2024,
				"shopUrl": "/shop/test",
				"isPurchased": 1,
				"lastOpenedAt": null,
				"readProgress": null,
				"pageCount": 10,
				"readPageCount": null,
				"readPercentage": null,
				"hasFreeChapters": 0,
				"supplements": null,
				"images": null,
				"formulasImages": null,
				"periods": null
			}`,
			wantErr: false,
		},
		{
			name: "invalid json - missing required field",
			jsonData: `{
				"title": "Missing ID Book",
				"description": "Missing required ID field"
			}`,
			wantErr: false, // JSON unmarshaling doesn't enforce required fields
		},
		{
			name: "invalid json - wrong types",
			jsonData: `{
				"id": "not-a-number",
				"title": 123,
				"isPurchased": "not-a-number"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var book Book
			err := json.Unmarshal([]byte(tt.jsonData), &book)

			if (err != nil) != tt.wantErr {
				t.Errorf("Book.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Basic validation for successful unmarshaling
				if book.ID == 0 && tt.name != "invalid json - missing required field" {
					t.Errorf("Expected book ID to be set")
				}
			}
		})
	}
}

func TestImage_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"id": 1,
		"summaryId": 123,
		"createdAt": "2025-01-01 00:00:00",
		"objectId": "test-image-1",
		"mimeType": "image/png",
		"imageUrl": "/images/test-image-1.png"
	}`

	var image Image
	err := json.Unmarshal([]byte(jsonData), &image)
	if err != nil {
		t.Fatalf("Failed to unmarshal Image: %v", err)
	}

	if image.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", image.ID)
	}
	if image.SummaryID != 123 {
		t.Errorf("Expected SummaryID to be 123, got %d", image.SummaryID)
	}
	if image.MimeType != "image/png" {
		t.Errorf("Expected MimeType to be 'image/png', got %s", image.MimeType)
	}
}

func TestChapter_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"id": 1,
		"summaryId": 123,
		"title": "Test Chapter",
		"isFree": 0,
		"isSupplement": 0,
		"isLocked": 0,
		"isVisible": 1,
		"parentChapterId": null,
		"gDocsChapterId": "h.test",
		"sortIndex": 0,
		"subChapters": [
			{
				"id": 2,
				"summaryId": 123,
				"title": "Sub Chapter",
				"isFree": 0,
				"isSupplement": 0,
				"isLocked": 0,
				"isVisible": 1,
				"parentChapterId": 1,
				"gDocsChapterId": "h.sub",
				"sortIndex": 1,
				"subChapters": []
			}
		]
	}`

	var chapter Chapter
	err := json.Unmarshal([]byte(jsonData), &chapter)
	if err != nil {
		t.Fatalf("Failed to unmarshal Chapter: %v", err)
	}

	if chapter.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", chapter.ID)
	}
	if chapter.Title != "Test Chapter" {
		t.Errorf("Expected Title to be 'Test Chapter', got %s", chapter.Title)
	}
	if len(chapter.SubChapters) != 1 {
		t.Errorf("Expected 1 sub-chapter, got %d", len(chapter.SubChapters))
	}
	if chapter.SubChapters[0].ID != 2 {
		t.Errorf("Expected sub-chapter ID to be 2, got %d", chapter.SubChapters[0].ID)
	}
	if chapter.SubChapters[0].ParentChapterID == nil || *chapter.SubChapters[0].ParentChapterID != 1 {
		t.Errorf("Expected sub-chapter parent ID to be 1")
	}
}

func TestContent_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"documentId": "test-doc-123",
		"revisionId": "test-revision-123",
		"suggestionsViewMode": "SUGGESTIONS_INLINE",
		"title": "Test Content",
		"body": {
			"content": [
				{
					"endIndex": 1,
					"startIndex": null,
					"sectionBreak": {
						"sectionStyle": {
							"sectionType": "CONTINUOUS"
						}
					}
				},
				{
					"endIndex": 14,
					"startIndex": 1,
					"paragraph": {
						"elements": [
							{
								"endIndex": 14,
								"startIndex": 1,
								"textRun": {
									"content": "Test content\n",
									"textStyle": {
										"fontSize": {
											"magnitude": 12,
											"unit": "PT"
										}
									}
								}
							}
						],
						"paragraphStyle": {
							"namedStyleType": "NORMAL_TEXT"
						}
					}
				}
			]
		}
	}`

	var content Content
	err := json.Unmarshal([]byte(jsonData), &content)
	if err != nil {
		t.Fatalf("Failed to unmarshal Content: %v", err)
	}

	if content.DocumentID != "test-doc-123" {
		t.Errorf("Expected DocumentID to be 'test-doc-123', got %s", content.DocumentID)
	}
	if content.Title != "Test Content" {
		t.Errorf("Expected Title to be 'Test Content', got %s", content.Title)
	}
	if len(content.Body.Content) != 2 {
		t.Errorf("Expected 2 content elements, got %d", len(content.Body.Content))
	}

	// Test first element (section break)
	firstElement := content.Body.Content[0]
	if firstElement.SectionBreak == nil {
		t.Errorf("Expected first element to be a section break")
	}

	// Test second element (paragraph)
	secondElement := content.Body.Content[1]
	if secondElement.Paragraph == nil {
		t.Errorf("Expected second element to be a paragraph")
	}
	if len(secondElement.Paragraph.Elements) != 1 {
		t.Errorf("Expected paragraph to have 1 element, got %d", len(secondElement.Paragraph.Elements))
	}

	textRun := secondElement.Paragraph.Elements[0].TextRun
	if textRun == nil {
		t.Errorf("Expected paragraph element to be a text run")
		return
	}
	if textRun.Content != "Test content\n" {
		t.Errorf("Expected text content to be 'Test content\n', got %s", textRun.Content)
	}
}

func TestTextStyle_ColorHandling(t *testing.T) {
	// Test that our interface{} fields can handle different color formats
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name: "color as object",
			jsonData: `{
				"fontSize": {"magnitude": 12, "unit": "PT"},
				"foregroundColor": {"rgbColor": {"red": 1.0, "green": 0.5, "blue": 0.0}}
			}`,
			wantErr: false,
		},
		{
			name: "color as array",
			jsonData: `{
				"fontSize": {"magnitude": 12, "unit": "PT"},
				"backgroundColor": [1.0, 0.5, 0.0]
			}`,
			wantErr: false,
		},
		{
			name: "null colors",
			jsonData: `{
				"fontSize": {"magnitude": 12, "unit": "PT"},
				"foregroundColor": null,
				"backgroundColor": null
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var style TextStyle
			err := json.Unmarshal([]byte(tt.jsonData), &style)

			if (err != nil) != tt.wantErr {
				t.Errorf("TextStyle.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComplexContentStructure(t *testing.T) {
	// Test with more complex content including tables and inline objects
	jsonData := `{
		"documentId": "complex-doc",
		"revisionId": "complex-revision",
		"suggestionsViewMode": "SUGGESTIONS_INLINE",
		"title": "Complex Content",
		"body": {
			"content": [
				{
					"endIndex": 50,
					"startIndex": 1,
					"table": {
						"columns": 2,
						"rows": 1,
						"tableRows": [
							{
								"endIndex": 25,
								"startIndex": 10,
								"tableCells": [
									{
										"endIndex": 15,
										"startIndex": 10,
										"content": []
									},
									{
										"endIndex": 25,
										"startIndex": 15,
										"content": []
									}
								]
							}
						]
					}
				},
				{
					"endIndex": 60,
					"startIndex": 50,
					"paragraph": {
						"elements": [
							{
								"endIndex": 55,
								"startIndex": 50,
								"inlineObjectElement": {
									"inlineObjectId": "test-object",
									"textStyle": {}
								}
							},
							{
								"endIndex": 60,
								"startIndex": 55,
								"textRun": {
									"content": "Text\n",
									"textStyle": {}
								}
							}
						],
						"paragraphStyle": {
							"namedStyleType": "NORMAL_TEXT"
						}
					}
				}
			]
		}
	}`

	var content Content
	err := json.Unmarshal([]byte(jsonData), &content)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex content: %v", err)
	}

	if len(content.Body.Content) != 2 {
		t.Errorf("Expected 2 content elements, got %d", len(content.Body.Content))
	}

	// Test table element
	tableElement := content.Body.Content[0]
	if tableElement.Table == nil {
		t.Errorf("Expected first element to be a table")
	} else {
		if tableElement.Table.Columns != 2 {
			t.Errorf("Expected table to have 2 columns, got %d", tableElement.Table.Columns)
		}
		if tableElement.Table.Rows != 1 {
			t.Errorf("Expected table to have 1 row, got %d", tableElement.Table.Rows)
		}
	}

	// Test paragraph with inline object
	paragraphElement := content.Body.Content[1]
	if paragraphElement.Paragraph == nil {
		t.Errorf("Expected second element to be a paragraph")
	} else {
		if len(paragraphElement.Paragraph.Elements) != 2 {
			t.Errorf("Expected paragraph to have 2 elements, got %d", len(paragraphElement.Paragraph.Elements))
		}

		if paragraphElement.Paragraph.Elements[0].InlineObjectElement == nil {
			t.Errorf("Expected first paragraph element to be an inline object")
		}

		if paragraphElement.Paragraph.Elements[1].TextRun == nil {
			t.Errorf("Expected second paragraph element to be a text run")
		}
	}
}
