package models

import (
	"encoding/json"
	"testing"
)

func TestDocumentSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected *Document
		wantErr  bool
	}{
		{
			name: "complete document with all fields",
			jsonData: `{
				"documentId": "doc123",
				"revisionId": "rev456",
				"suggestionsViewMode": "PREVIEW_SUGGESTIONS_ACCEPTED",
				"title": "Test Document",
				"body": {
					"content": [
						{
							"startIndex": 0,
							"endIndex": 10,
							"paragraph": {
								"elements": [
									{
										"startIndex": 0,
										"endIndex": 10,
										"textRun": {
											"content": "Hello World",
											"textStyle": {}
										}
									}
								],
								"paragraphStyle": {
									"namedStyleType": "NORMAL_TEXT",
									"direction": "LEFT_TO_RIGHT"
								}
							}
						}
					]
				},
				"headers": {
					"header1": {
						"headerId": "header1",
						"content": []
					}
				},
				"footers": {
					"footer1": {
						"footerId": "footer1",
						"content": []
					}
				},
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
					"defaultHeaderId": "header1",
					"defaultFooterId": "footer1",
					"firstPageHeaderId": "header1",
					"firstPageFooterId": "footer1"
				},
				"namedStyles": {
					"styles": [
						{
							"namedStyleType": "NORMAL_TEXT",
							"textStyle": {},
							"paragraphStyle": {
								"namedStyleType": "NORMAL_TEXT",
								"direction": "LEFT_TO_RIGHT"
							}
						}
					]
				},
				"lists": {},
				"inlineObjects": {},
				"positionedObjects": {}
			}`,
			expected: &Document{
				DocumentID:          "doc123",
				RevisionID:          "rev456",
				SuggestionsViewMode: "PREVIEW_SUGGESTIONS_ACCEPTED",
				Title:               "Test Document",
				Body: Body{
					Content: []StructuralElement{
						{
							StartIndex: 0,
							EndIndex:   10,
							Paragraph: &Paragraph{
								Elements: []ParagraphElement{
									{
										StartIndex: 0,
										EndIndex:   10,
										TextRun: &TextRun{
											Content:   "Hello World",
											TextStyle: TextStyle{},
										},
									},
								},
								ParagraphStyle: ParagraphStyle{
									NamedStyleType: "NORMAL_TEXT",
									Direction:      "LEFT_TO_RIGHT",
								},
							},
						},
					},
				},
				Headers: map[string]HeaderFooter{
					"header1": {
						HeaderID: "header1",
						Content:  []StructuralElement{},
					},
				},
				Footers: map[string]HeaderFooter{
					"footer1": {
						FooterID: "footer1",
						Content:  []StructuralElement{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal document",
			jsonData: `{
				"documentId": "minimal",
				"title": "Minimal Doc",
				"body": {"content": []},
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
				"namedStyles": {"styles": []},
				"lists": {},
				"inlineObjects": {},
				"positionedObjects": {}
			}`,
			expected: &Document{
				DocumentID: "minimal",
				Title:      "Minimal Doc",
				Body:       Body{Content: []StructuralElement{}},
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"documentId": invalid}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := UnmarshalDocument([]byte(tt.jsonData))

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalDocument() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalDocument() error = %v", err)
				return
			}

			if doc.DocumentID != tt.expected.DocumentID {
				t.Errorf("DocumentID = %v, expected %v", doc.DocumentID, tt.expected.DocumentID)
			}
			if doc.Title != tt.expected.Title {
				t.Errorf("Title = %v, expected %v", doc.Title, tt.expected.Title)
			}

			// Check body content length
			if len(doc.Body.Content) != len(tt.expected.Body.Content) {
				t.Errorf("Body content length = %v, expected %v",
					len(doc.Body.Content), len(tt.expected.Body.Content))
			}
		})
	}
}

func TestStructuralElementSerialization(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectPara  bool
		expectTable bool
		expectTOC   bool
		expectSB    bool
		wantErr     bool
	}{
		{
			name: "paragraph element",
			jsonData: `{
				"startIndex": 0,
				"endIndex": 15,
				"paragraph": {
					"elements": [{
						"startIndex": 0,
						"endIndex": 15,
						"textRun": {
							"content": "Test paragraph",
							"textStyle": {}
						}
					}],
					"paragraphStyle": {
						"namedStyleType": "NORMAL_TEXT",
						"direction": "LEFT_TO_RIGHT"
					}
				}
			}`,
			expectPara:  true,
			expectTable: false,
			expectTOC:   false,
			expectSB:    false,
			wantErr:     false,
		},
		{
			name: "table element",
			jsonData: `{
				"startIndex": 0,
				"endIndex": 50,
				"table": {
					"rows": 2,
					"columns": 2,
					"tableRows": [
						{
							"startIndex": 0,
							"endIndex": 25,
							"tableCells": [
								{
									"startIndex": 0,
									"endIndex": 12,
									"content": [],
									"tableCellStyle": {
										"rowSpan": 1,
										"columnSpan": 1
									}
								}
							],
							"tableRowStyle": {
								"minRowHeight": {"magnitude": 0, "unit": "PT"}
							}
						}
					],
					"tableStyle": {
						"tableColumnProperties": []
					}
				}
			}`,
			expectPara:  false,
			expectTable: true,
			expectTOC:   false,
			expectSB:    false,
			wantErr:     false,
		},
		{
			name: "table of contents element",
			jsonData: `{
				"startIndex": 0,
				"endIndex": 20,
				"tableOfContents": {
					"content": []
				}
			}`,
			expectPara:  false,
			expectTable: false,
			expectTOC:   true,
			expectSB:    false,
			wantErr:     false,
		},
		{
			name: "section break element",
			jsonData: `{
				"startIndex": 100,
				"endIndex": 101,
				"sectionBreak": {
					"sectionStyle": {
						"columnSeparatorStyle": "NONE",
						"contentDirection": "LEFT_TO_RIGHT",
						"sectionType": "CONTINUOUS"
					}
				}
			}`,
			expectPara:  false,
			expectTable: false,
			expectTOC:   false,
			expectSB:    true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var element StructuralElement
			err := json.Unmarshal([]byte(tt.jsonData), &element)

			if tt.wantErr {
				if err == nil {
					t.Errorf("StructuralElement unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("StructuralElement unmarshal error = %v", err)
				return
			}

			hasPara := element.Paragraph != nil
			hasTable := element.Table != nil
			hasTOC := element.TableOfContents != nil
			hasSB := element.SectionBreak != nil

			if hasPara != tt.expectPara {
				t.Errorf("Paragraph presence = %v, expected %v", hasPara, tt.expectPara)
			}
			if hasTable != tt.expectTable {
				t.Errorf("Table presence = %v, expected %v", hasTable, tt.expectTable)
			}
			if hasTOC != tt.expectTOC {
				t.Errorf("TableOfContents presence = %v, expected %v", hasTOC, tt.expectTOC)
			}
			if hasSB != tt.expectSB {
				t.Errorf("SectionBreak presence = %v, expected %v", hasSB, tt.expectSB)
			}
		})
	}
}

func TestDocumentStyleSerialization(t *testing.T) {
	jsonData := `{
		"background": {
			"color": {
				"color": {
					"red": 1.0,
					"green": 1.0,
					"blue": 1.0
				}
			}
		},
		"pageSize": {
			"height": {"magnitude": 11, "unit": "IN"},
			"width": {"magnitude": 8.5, "unit": "IN"}
		},
		"marginTop": {"magnitude": 1.0, "unit": "IN"},
		"marginBottom": {"magnitude": 1.0, "unit": "IN"},
		"marginRight": {"magnitude": 1.0, "unit": "IN"},
		"marginLeft": {"magnitude": 1.0, "unit": "IN"},
		"marginHeader": {"magnitude": 0.5, "unit": "IN"},
		"marginFooter": {"magnitude": 0.5, "unit": "IN"},
		"pageNumberStart": 1,
		"useCustomHeaderFooterMargins": true,
		"defaultHeaderId": "header1",
		"defaultFooterId": "footer1",
		"firstPageHeaderId": "firstHeader",
		"firstPageFooterId": "firstFooter"
	}`

	var docStyle DocumentStyle
	err := json.Unmarshal([]byte(jsonData), &docStyle)
	if err != nil {
		t.Fatalf("DocumentStyle unmarshal error = %v", err)
	}

	// Test page size
	if docStyle.PageSize.Height.Magnitude != 11 {
		t.Errorf("PageSize height = %v, expected 11", docStyle.PageSize.Height.Magnitude)
	}
	if docStyle.PageSize.Width.Unit != "IN" {
		t.Errorf("PageSize width unit = %v, expected IN", docStyle.PageSize.Width.Unit)
	}

	// Test margins
	if docStyle.MarginTop.Magnitude != 1.0 {
		t.Errorf("MarginTop = %v, expected 1.0", docStyle.MarginTop.Magnitude)
	}

	// Test boolean field
	if !docStyle.UseCustomHeaderFooterMargins {
		t.Errorf("UseCustomHeaderFooterMargins = %v, expected true", docStyle.UseCustomHeaderFooterMargins)
	}

	// Test header/footer IDs
	if docStyle.DefaultHeaderID != "header1" {
		t.Errorf("DefaultHeaderID = %v, expected header1", docStyle.DefaultHeaderID)
	}
}

func TestHeaderFooterSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		isHeader bool
		wantErr  bool
	}{
		{
			name: "header with content",
			jsonData: `{
				"headerId": "header123",
				"content": [
					{
						"startIndex": 0,
						"endIndex": 10,
						"paragraph": {
							"elements": [{
								"startIndex": 0,
								"endIndex": 10,
								"textRun": {
									"content": "Header text",
									"textStyle": {}
								}
							}],
							"paragraphStyle": {
								"namedStyleType": "NORMAL_TEXT",
								"direction": "LEFT_TO_RIGHT"
							}
						}
					}
				]
			}`,
			isHeader: true,
			wantErr:  false,
		},
		{
			name: "footer with content",
			jsonData: `{
				"footerId": "footer123",
				"content": [
					{
						"startIndex": 0,
						"endIndex": 15,
						"paragraph": {
							"elements": [{
								"startIndex": 0,
								"endIndex": 15,
								"textRun": {
									"content": "Footer text",
									"textStyle": {}
								}
							}],
							"paragraphStyle": {
								"namedStyleType": "NORMAL_TEXT",
								"direction": "LEFT_TO_RIGHT"
							}
						}
					}
				]
			}`,
			isHeader: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hf HeaderFooter
			err := json.Unmarshal([]byte(tt.jsonData), &hf)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HeaderFooter unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("HeaderFooter unmarshal error = %v", err)
				return
			}

			if tt.isHeader {
				if hf.HeaderID == "" {
					t.Errorf("Expected HeaderID to be set")
				}
				if hf.FooterID != "" {
					t.Errorf("Expected FooterID to be empty, got %v", hf.FooterID)
				}
			} else {
				if hf.FooterID == "" {
					t.Errorf("Expected FooterID to be set")
				}
				if hf.HeaderID != "" {
					t.Errorf("Expected HeaderID to be empty, got %v", hf.HeaderID)
				}
			}

			if len(hf.Content) == 0 {
				t.Errorf("Expected content to be present")
			}
		})
	}
}

func TestDocumentMarshalRoundtrip(t *testing.T) {
	// Create a document with various elements
	original := &Document{
		DocumentID: "test-doc",
		Title:      "Test Document",
		RevisionID: "rev-123",
		Body: Body{
			Content: []StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   20,
					Paragraph: &Paragraph{
						Elements: []ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   20,
								TextRun: &TextRun{
									Content: "Hello, World!",
									TextStyle: TextStyle{
										Bold: func() *bool { b := true; return &b }(),
									},
								},
							},
						},
						ParagraphStyle: ParagraphStyle{
							NamedStyleType: "HEADING_1",
							Direction:      "LEFT_TO_RIGHT",
						},
					},
				},
			},
		},
		Headers: map[string]HeaderFooter{
			"default": {
				HeaderID: "default",
				Content:  []StructuralElement{},
			},
		},
		Footers: map[string]HeaderFooter{},
		DocumentStyle: DocumentStyle{
			PageSize: Size{
				Height: Dimension{Magnitude: 11, Unit: "IN"},
				Width:  Dimension{Magnitude: 8.5, Unit: "IN"},
			},
			MarginTop:    Dimension{Magnitude: 1, Unit: "IN"},
			MarginBottom: Dimension{Magnitude: 1, Unit: "IN"},
			MarginLeft:   Dimension{Magnitude: 1, Unit: "IN"},
			MarginRight:  Dimension{Magnitude: 1, Unit: "IN"},
			MarginHeader: Dimension{Magnitude: 0.5, Unit: "IN"},
			MarginFooter: Dimension{Magnitude: 0.5, Unit: "IN"},
		},
		NamedStyles:       NamedStyles{Styles: []NamedStyle{}},
		Lists:             map[string]List{},
		InlineObjects:     map[string]InlineObject{},
		PositionedObjects: map[string]PositionedObject{},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var roundtrip Document
	err = json.Unmarshal(data, &roundtrip)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify key fields
	if roundtrip.DocumentID != original.DocumentID {
		t.Errorf("DocumentID roundtrip failed: got %v, want %v",
			roundtrip.DocumentID, original.DocumentID)
	}
	if roundtrip.Title != original.Title {
		t.Errorf("Title roundtrip failed: got %v, want %v",
			roundtrip.Title, original.Title)
	}
	if len(roundtrip.Body.Content) != len(original.Body.Content) {
		t.Errorf("Body content length roundtrip failed: got %v, want %v",
			len(roundtrip.Body.Content), len(original.Body.Content))
	}

	// Verify complex nested data
	if len(roundtrip.Body.Content) > 0 && len(original.Body.Content) > 0 {
		origPara := original.Body.Content[0].Paragraph
		rtPara := roundtrip.Body.Content[0].Paragraph

		if origPara != nil && rtPara != nil {
			if origPara.ParagraphStyle.NamedStyleType != rtPara.ParagraphStyle.NamedStyleType {
				t.Errorf("Paragraph style roundtrip failed")
			}
		}
	}
}

func TestDocumentValidation(t *testing.T) {
	tests := []struct {
		name    string
		doc     *Document
		isValid bool
	}{
		{
			name: "valid document",
			doc: &Document{
				DocumentID: "valid-doc",
				Title:      "Valid Document",
				Body:       Body{Content: []StructuralElement{}},
			},
			isValid: true,
		},
		{
			name: "document without ID",
			doc: &Document{
				DocumentID: "",
				Title:      "No ID Document",
				Body:       Body{Content: []StructuralElement{}},
			},
			isValid: false,
		},
		{
			name: "document with nil body content",
			doc: &Document{
				DocumentID: "nil-body-doc",
				Title:      "Nil Body Document",
				Body:       Body{Content: nil},
			},
			isValid: true, // nil content is acceptable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation logic
			isValid := tt.doc.DocumentID != ""

			if isValid != tt.isValid {
				t.Errorf("Document validation = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}
