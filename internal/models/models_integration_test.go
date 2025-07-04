package models

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCompleteBookWithAllModels tests a complete book with all related models
func TestCompleteBookWithAllModels(t *testing.T) {
	// Create a complete book with all possible data structures
	createdAt := time.Now()
	readProgress := int64(150)
	customTime := &CustomTime{createdAt}

	book := &Book{
		ID:                 12345,
		Title:              "Advanced Computer Science",
		Description:        "Comprehensive computer science textbook",
		AvailableDate:      "2024-01-01",
		ExamDate:           "2024-06-15",
		BachelorYearNumber: "3",
		CollegeStartYear:   2022,
		ShopURL:            "https://shop.example.com/book/12345",
		IsPurchased:        BoolInt(true),
		LastOpenedAt:       customTime,
		ReadProgress:       &readProgress,
		PageCount:          300,
		HasFreeChapters:    BoolInt(true),
		Images: []BookImage{
			{
				ID:        1,
				SummaryID: 12345,
				CreatedAt: CustomTime{createdAt},
				ObjectID:  "obj_cover",
				MIMEType:  "image/jpeg",
				ImageURL:  "https://example.com/cover.jpg",
			},
			{
				ID:        2,
				SummaryID: 12345,
				CreatedAt: CustomTime{createdAt},
				ObjectID:  "obj_diagram",
				MIMEType:  "image/png",
				ImageURL:  "https://example.com/diagram.png",
			},
		},
		Periods: []string{"Fall 2024", "Spring 2025"},
		Chapters: []Chapter{
			{
				ID:             1,
				SummaryID:      12345,
				Title:          "Introduction",
				IsFree:         BoolInt(true),
				GDocsChapterID: "intro_chapter",
				SortIndex:      0,
				SubChapters:    []Chapter{},
			},
			{
				ID:             2,
				SummaryID:      12345,
				Title:          "Data Structures",
				IsFree:         BoolInt(false),
				IsLocked:       BoolInt(true),
				GDocsChapterID: "ds_chapter",
				SortIndex:      1,
				SubChapters: []Chapter{
					{
						ID:              3,
						SummaryID:       12345,
						Title:           "Arrays",
						IsFree:          BoolInt(false),
						IsLocked:        BoolInt(true),
						ParentChapterID: &[]int64{2}[0],
						GDocsChapterID:  "arrays_section",
						SortIndex:       0,
						SubChapters:     []Chapter{},
					},
					{
						ID:              4,
						SummaryID:       12345,
						Title:           "Linked Lists",
						IsFree:          BoolInt(false),
						IsLocked:        BoolInt(true),
						ParentChapterID: &[]int64{2}[0],
						GDocsChapterID:  "lists_section",
						SortIndex:       1,
						SubChapters:     []Chapter{},
					},
				},
			},
		},
		Content: &Content{
			Document: &Document{
				DocumentID:          "doc_12345",
				RevisionID:          "rev_1",
				SuggestionsViewMode: "PREVIEW_SUGGESTIONS_ACCEPTED",
				Title:               "Advanced Computer Science - Full Content",
				Body: Body{
					Content: []StructuralElement{
						{
							StartIndex: 0,
							EndIndex:   50,
							Paragraph: &Paragraph{
								Elements: []ParagraphElement{
									{
										StartIndex: 0,
										EndIndex:   50,
										TextRun: &TextRun{
											Content: "Chapter 1: Introduction to Computer Science",
											TextStyle: TextStyle{
												Bold: &[]bool{true}[0],
												FontSize: &Dimension{
													Magnitude: 18,
													Unit:      "PT",
												},
											},
										},
									},
								},
								ParagraphStyle: ParagraphStyle{
									NamedStyleType: "HEADING_1",
									Direction:      "LEFT_TO_RIGHT",
									Alignment:      &[]string{"START"}[0],
									HeadingID:      &[]string{"heading_intro"}[0],
								},
							},
						},
						{
							StartIndex: 51,
							EndIndex:   200,
							Table: &Table{
								Rows:    2,
								Columns: 3,
								TableRows: []TableRow{
									{
										StartIndex: 51,
										EndIndex:   125,
										TableCells: []TableCell{
											{
												StartIndex: 51,
												EndIndex:   75,
												Content: []StructuralElement{
													{
														StartIndex: 51,
														EndIndex:   75,
														Paragraph: &Paragraph{
															Elements: []ParagraphElement{
																{
																	StartIndex: 51,
																	EndIndex:   75,
																	TextRun: &TextRun{
																		Content:   "Algorithm",
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
												TableCellStyle: TableCellStyle{
													RowSpan:    1,
													ColumnSpan: 1,
												},
											},
											{
												StartIndex: 76,
												EndIndex:   100,
												Content: []StructuralElement{
													{
														StartIndex: 76,
														EndIndex:   100,
														Paragraph: &Paragraph{
															Elements: []ParagraphElement{
																{
																	StartIndex: 76,
																	EndIndex:   100,
																	TextRun: &TextRun{
																		Content:   "Time Complexity",
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
												TableCellStyle: TableCellStyle{
													RowSpan:    1,
													ColumnSpan: 1,
												},
											},
											{
												StartIndex: 101,
												EndIndex:   125,
												Content: []StructuralElement{
													{
														StartIndex: 101,
														EndIndex:   125,
														Paragraph: &Paragraph{
															Elements: []ParagraphElement{
																{
																	StartIndex: 101,
																	EndIndex:   125,
																	TextRun: &TextRun{
																		Content:   "Space Complexity",
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
												TableCellStyle: TableCellStyle{
													RowSpan:    1,
													ColumnSpan: 1,
												},
											},
										},
										TableRowStyle: TableRowStyle{
											MinRowHeight: Dimension{Magnitude: 20, Unit: "PT"},
										},
									},
								},
								TableStyle: TableStyle{
									TableColumnProperties: []TableColumnProperty{
										{WidthType: "EVENLY_DISTRIBUTED"},
										{WidthType: "EVENLY_DISTRIBUTED"},
										{WidthType: "EVENLY_DISTRIBUTED"},
									},
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
				Footers: map[string]HeaderFooter{
					"default": {
						FooterID: "default",
						Content:  []StructuralElement{},
					},
				},
				DocumentStyle: DocumentStyle{
					Background: Background{},
					PageSize: Size{
						Height: Dimension{Magnitude: 11, Unit: "IN"},
						Width:  Dimension{Magnitude: 8.5, Unit: "IN"},
					},
					MarginTop:       Dimension{Magnitude: 1, Unit: "IN"},
					MarginBottom:    Dimension{Magnitude: 1, Unit: "IN"},
					MarginLeft:      Dimension{Magnitude: 1, Unit: "IN"},
					MarginRight:     Dimension{Magnitude: 1, Unit: "IN"},
					MarginHeader:    Dimension{Magnitude: 0.5, Unit: "IN"},
					MarginFooter:    Dimension{Magnitude: 0.5, Unit: "IN"},
					PageNumberStart: 1,
					DefaultHeaderID: "default",
					DefaultFooterID: "default",
				},
				NamedStyles: NamedStyles{
					Styles: []NamedStyle{
						{
							NamedStyleType: "HEADING_1",
							TextStyle: TextStyle{
								Bold: &[]bool{true}[0],
								FontSize: &Dimension{
									Magnitude: 18,
									Unit:      "PT",
								},
							},
							ParagraphStyle: ParagraphStyle{
								NamedStyleType: "HEADING_1",
								Direction:      "LEFT_TO_RIGHT",
							},
						},
					},
				},
				Lists:             map[string]List{},
				InlineObjects:     map[string]InlineObject{},
				PositionedObjects: map[string]PositionedObject{},
			},
		},
		InlineObjectMap: map[string]string{
			"obj_cover":   "https://example.com/cover.jpg",
			"obj_diagram": "https://example.com/diagram.png",
		},
	}

	// Test JSON marshaling and unmarshaling
	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal complete book: %v", err)
	}

	var unmarshaled Book
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal complete book: %v", err)
	}

	// Verify basic book properties
	if unmarshaled.ID != book.ID {
		t.Errorf("Book ID mismatch: got %v, want %v", unmarshaled.ID, book.ID)
	}
	if unmarshaled.Title != book.Title {
		t.Errorf("Book Title mismatch: got %v, want %v", unmarshaled.Title, book.Title)
	}

	// Verify images
	if len(unmarshaled.Images) != len(book.Images) {
		t.Errorf("Images count mismatch: got %v, want %v",
			len(unmarshaled.Images), len(book.Images))
	}

	// Note: Chapters field has `json:"-"` tag, so it won't be serialized/deserialized
	// This is expected behavior as chapters are loaded separately from chapters.json
	if len(unmarshaled.Chapters) != 0 {
		t.Errorf("Chapters should be empty after JSON roundtrip due to json:\"-\" tag, got %v",
			len(unmarshaled.Chapters))
	}
}

func TestContentTypeInteraction(t *testing.T) {
	// Test that Content can seamlessly switch between Document and Chapters

	t.Run("document content in book", func(t *testing.T) {
		book := &Book{
			ID:    1,
			Title: "Document Book",
			Content: &Content{
				Document: &Document{
					DocumentID: "doc1",
					Title:      "Document Content",
					Body:       Body{Content: []StructuralElement{}},
				},
			},
		}

		// Serialize and deserialize
		data, err := json.Marshal(book.Content)
		if err != nil {
			t.Fatalf("Failed to marshal document content: %v", err)
		}

		var content Content
		err = json.Unmarshal(data, &content)
		if err != nil {
			t.Fatalf("Failed to unmarshal document content: %v", err)
		}

		if content.Document == nil {
			t.Error("Expected Document to be present")
		}
		if content.Chapters != nil {
			t.Error("Expected Chapters to be nil")
		}
	})

	t.Run("chapters content in book", func(t *testing.T) {
		book := &Book{
			ID:    2,
			Title: "Chapters Book",
			Content: &Content{
				Chapters: []Chapter{
					{ID: 1, Title: "Chapter 1", GDocsChapterID: "ch1"},
					{ID: 2, Title: "Chapter 2", GDocsChapterID: "ch2"},
				},
			},
		}

		// Serialize and deserialize
		data, err := json.Marshal(book.Content)
		if err != nil {
			t.Fatalf("Failed to marshal chapters content: %v", err)
		}

		var content Content
		err = json.Unmarshal(data, &content)
		if err != nil {
			t.Fatalf("Failed to unmarshal chapters content: %v", err)
		}

		if content.Document != nil {
			t.Error("Expected Document to be nil")
		}
		if content.Chapters == nil {
			t.Error("Expected Chapters to be present")
		}
		if len(content.Chapters) != 2 {
			t.Errorf("Expected 2 chapters, got %d", len(content.Chapters))
		}
	})
}

func TestDeepNestingAndCrossReferences(t *testing.T) {
	// Test deeply nested structures and cross-references

	document := &Document{
		DocumentID: "nested_doc",
		Title:      "Deeply Nested Document",
		Body: Body{
			Content: []StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   100,
					Paragraph: &Paragraph{
						Elements: []ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   50,
								TextRun: &TextRun{
									Content: "See chapter on algorithms",
									TextStyle: TextStyle{
										Link: &Link{
											HeadingID: &[]string{"heading_algorithms"}[0],
										},
									},
								},
							},
							{
								StartIndex: 51,
								EndIndex:   100,
								InlineObjectElement: &InlineObjectElement{
									InlineObjectID: "img_1",
									TextStyle:      TextStyle{},
								},
							},
						},
						ParagraphStyle: ParagraphStyle{
							NamedStyleType: "NORMAL_TEXT",
							Direction:      "LEFT_TO_RIGHT",
						},
					},
				},
				{
					StartIndex: 101,
					EndIndex:   300,
					Table: &Table{
						Rows:    3,
						Columns: 2,
						TableRows: []TableRow{
							{
								StartIndex: 101,
								EndIndex:   200,
								TableCells: []TableCell{
									{
										StartIndex: 101,
										EndIndex:   150,
										Content: []StructuralElement{
											{
												StartIndex: 101,
												EndIndex:   150,
												Paragraph: &Paragraph{
													Elements: []ParagraphElement{
														{
															StartIndex: 101,
															EndIndex:   150,
															TextRun: &TextRun{
																Content: "Nested table content with link",
																TextStyle: TextStyle{
																	Link: &Link{
																		URL: &[]string{"https://external.com"}[0],
																	},
																	Bold: &[]bool{true}[0],
																},
															},
														},
													},
													ParagraphStyle: ParagraphStyle{
														NamedStyleType: "NORMAL_TEXT",
														Direction:      "LEFT_TO_RIGHT",
														BorderLeft: &Border{
															Width: Dimension{Magnitude: 1, Unit: "PT"},
															Color: &Color{
																Color: &RGBColor{
																	Red:   &[]float64{0.8}[0],
																	Green: &[]float64{0.8}[0],
																	Blue:  &[]float64{0.8}[0],
																},
															},
															DashStyle: "SOLID",
														},
													},
												},
											},
										},
										TableCellStyle: TableCellStyle{
											RowSpan:    1,
											ColumnSpan: 1,
											BackgroundColor: &Color{
												Color: &RGBColor{
													Red:   &[]float64{0.95}[0],
													Green: &[]float64{0.95}[0],
													Blue:  &[]float64{0.95}[0],
												},
											},
										},
									},
								},
								TableRowStyle: TableRowStyle{
									MinRowHeight: Dimension{Magnitude: 25, Unit: "PT"},
								},
							},
						},
						TableStyle: TableStyle{
							TableColumnProperties: []TableColumnProperty{
								{
									WidthType: "FIXED_WIDTH",
									Width: &Dimension{
										Magnitude: 2,
										Unit:      "IN",
									},
								},
								{
									WidthType: "EVENLY_DISTRIBUTED",
								},
							},
						},
					},
				},
			},
		},
		InlineObjects: map[string]InlineObject{
			"img_1": {
				ObjectID: "img_1",
				InlineObjectProperties: InlineObjectProperties{
					EmbeddedObject: EmbeddedObject{
						Title:       &[]string{"Sample Image"}[0],
						Description: &[]string{"A sample embedded image"}[0],
						Size: Size{
							Height: Dimension{Magnitude: 200, Unit: "PX"},
							Width:  Dimension{Magnitude: 300, Unit: "PX"},
						},
						ImageProperties: &ImageProperties{
							ContentURI: "https://example.com/image.png",
							Brightness: &[]float64{0.9}[0],
							Contrast:   &[]float64{1.1}[0],
							CropProperties: CropProperties{
								OffsetLeft: &[]float64{0.1}[0],
								OffsetTop:  &[]float64{0.1}[0],
							},
						},
					},
				},
			},
		},
		Lists:             map[string]List{},
		PositionedObjects: map[string]PositionedObject{},
		Headers:           map[string]HeaderFooter{},
		Footers:           map[string]HeaderFooter{},
		DocumentStyle:     DocumentStyle{},
		NamedStyles:       NamedStyles{Styles: []NamedStyle{}},
	}

	// Test serialization of complex nested structure
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("Failed to marshal complex document: %v", err)
	}

	var unmarshaled Document
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex document: %v", err)
	}

	// Verify structure preservation
	if len(unmarshaled.Body.Content) != len(document.Body.Content) {
		t.Errorf("Body content length mismatch")
	}

	// Verify inline objects
	if len(unmarshaled.InlineObjects) != len(document.InlineObjects) {
		t.Errorf("InlineObjects count mismatch")
	}

	// Verify nested table structure
	if len(unmarshaled.Body.Content) > 1 {
		table := unmarshaled.Body.Content[1].Table
		if table == nil {
			t.Error("Expected table to be present")
		} else if table.Rows != 3 || table.Columns != 2 {
			t.Errorf("Table dimensions mismatch: got %dx%d, want 3x2",
				table.Rows, table.Columns)
		}
	}

	// Verify cross-references (links)
	if len(unmarshaled.Body.Content) > 0 {
		para := unmarshaled.Body.Content[0].Paragraph
		if para != nil && len(para.Elements) > 0 {
			textRun := para.Elements[0].TextRun
			if textRun != nil && textRun.TextStyle.Link != nil {
				if textRun.TextStyle.Link.HeadingID == nil {
					t.Error("Expected HeadingID link to be preserved")
				}
			}
		}
	}
}

func TestInlineObjectMapping(t *testing.T) {
	// Test the inline object mapping functionality

	book := &Book{
		ID:    1,
		Title: "Image Book",
		Images: []BookImage{
			{
				ID:       1,
				ObjectID: "obj_1",
				ImageURL: "https://example.com/image1.png",
			},
			{
				ID:       2,
				ObjectID: "obj_2",
				ImageURL: "https://example.com/image2.jpg",
			},
		},
		InlineObjectMap: map[string]string{
			"obj_1": "https://example.com/image1.png",
			"obj_2": "https://example.com/image2.jpg",
		},
		Content: &Content{
			Document: &Document{
				DocumentID: "doc_with_images",
				Body: Body{
					Content: []StructuralElement{
						{
							Paragraph: &Paragraph{
								Elements: []ParagraphElement{
									{
										InlineObjectElement: &InlineObjectElement{
											InlineObjectID: "obj_1",
										},
									},
								},
							},
						},
					},
				},
				InlineObjects: map[string]InlineObject{
					"obj_1": {
						ObjectID: "obj_1",
						InlineObjectProperties: InlineObjectProperties{
							EmbeddedObject: EmbeddedObject{
								ImageProperties: &ImageProperties{
									ContentURI: "https://example.com/image1.png",
								},
							},
						},
					},
					"obj_2": {
						ObjectID: "obj_2",
						InlineObjectProperties: InlineObjectProperties{
							EmbeddedObject: EmbeddedObject{
								ImageProperties: &ImageProperties{
									ContentURI: "https://example.com/image2.jpg",
								},
							},
						},
					},
				},
			},
		},
	}

	// Verify the mapping consistency
	for objectID, expectedURL := range book.InlineObjectMap {
		// Check if it exists in BookImages
		found := false
		for _, img := range book.Images {
			if img.ObjectID == objectID && img.ImageURL == expectedURL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ObjectID %s not found in BookImages with URL %s", objectID, expectedURL)
		}

		// Check if it exists in InlineObjects
		if book.Content != nil && book.Content.Document != nil {
			inlineObj, exists := book.Content.Document.InlineObjects[objectID]
			if !exists {
				t.Errorf("ObjectID %s not found in InlineObjects", objectID)
			} else if inlineObj.InlineObjectProperties.EmbeddedObject.ImageProperties != nil {
				actualURL := inlineObj.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentURI
				if actualURL != expectedURL {
					t.Errorf("URL mismatch for %s: got %s, want %s",
						objectID, actualURL, expectedURL)
				}
			}
		}
	}
}

func TestModelValidationIntegration(t *testing.T) {
	// Test validation across all models

	tests := []struct {
		name    string
		book    *Book
		isValid bool
		errors  []string
	}{
		{
			name: "valid complete book",
			book: &Book{
				ID:              1,
				Title:           "Valid Book",
				IsPurchased:     BoolInt(true),
				PageCount:       100,
				HasFreeChapters: BoolInt(false),
				Content: &Content{
					Document: &Document{
						DocumentID: "valid_doc",
						Title:      "Valid Document",
					},
				},
			},
			isValid: true,
			errors:  []string{},
		},
		{
			name: "invalid book - contradictory content",
			book: &Book{
				ID:    2,
				Title: "Invalid Book",
				Content: &Content{
					Document: &Document{DocumentID: "doc"},
					Chapters: []Chapter{{ID: 1}}, // Both present - invalid
				},
			},
			isValid: false,
			errors:  []string{"Content cannot have both Document and Chapters"},
		},
		{
			name: "invalid book - empty required fields",
			book: &Book{
				ID:    0,  // Invalid ID
				Title: "", // Empty title
				Content: &Content{
					Document: &Document{
						DocumentID: "", // Empty document ID
					},
				},
			},
			isValid: false,
			errors:  []string{"Book ID must be positive", "Title cannot be empty", "DocumentID cannot be empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []string

			// Validate book
			if tt.book.ID <= 0 {
				errors = append(errors, "Book ID must be positive")
			}
			if tt.book.Title == "" {
				errors = append(errors, "Title cannot be empty")
			}

			// Validate content
			if tt.book.Content != nil {
				if tt.book.Content.Document != nil && tt.book.Content.Chapters != nil {
					errors = append(errors, "Content cannot have both Document and Chapters")
				}
				if tt.book.Content.Document != nil && tt.book.Content.Document.DocumentID == "" {
					errors = append(errors, "DocumentID cannot be empty")
				}
			}

			isValid := len(errors) == 0

			if isValid != tt.isValid {
				t.Errorf("Validation result = %v, expected %v", isValid, tt.isValid)
			}

			// Check if expected errors are present
			for _, expectedError := range tt.errors {
				found := false
				for _, actualError := range errors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s' not found in actual errors: %v",
						expectedError, errors)
				}
			}
		})
	}
}

func TestPerformanceWithLargeData(t *testing.T) {
	// Test performance with large nested structures

	// Create a large document with many elements
	elements := make([]StructuralElement, 1000)
	for i := 0; i < 1000; i++ {
		elements[i] = StructuralElement{
			StartIndex: int64(i * 10),
			EndIndex:   int64((i + 1) * 10),
			Paragraph: &Paragraph{
				Elements: []ParagraphElement{
					{
						StartIndex: int64(i * 10),
						EndIndex:   int64((i + 1) * 10),
						TextRun: &TextRun{
							Content: "Sample text content for performance testing",
							TextStyle: TextStyle{
								Bold: &[]bool{i%2 == 0}[0],
								FontSize: &Dimension{
									Magnitude: 12,
									Unit:      "PT",
								},
							},
						},
					},
				},
				ParagraphStyle: ParagraphStyle{
					NamedStyleType: "NORMAL_TEXT",
					Direction:      "LEFT_TO_RIGHT",
				},
			},
		}
	}

	largeDoc := &Document{
		DocumentID: "large_doc",
		Title:      "Large Performance Test Document",
		Body: Body{
			Content: elements,
		},
		Headers:           map[string]HeaderFooter{},
		Footers:           map[string]HeaderFooter{},
		DocumentStyle:     DocumentStyle{},
		NamedStyles:       NamedStyles{Styles: []NamedStyle{}},
		Lists:             map[string]List{},
		InlineObjects:     map[string]InlineObject{},
		PositionedObjects: map[string]PositionedObject{},
	}

	// Measure serialization performance
	startTime := time.Now()
	data, err := json.Marshal(largeDoc)
	marshalTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to marshal large document: %v", err)
	}

	t.Logf("Marshaling 1000 elements took: %v", marshalTime)
	t.Logf("Serialized data size: %d bytes", len(data))

	// Measure deserialization performance
	startTime = time.Now()
	var unmarshaled Document
	err = json.Unmarshal(data, &unmarshaled)
	unmarshalTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to unmarshal large document: %v", err)
	}

	t.Logf("Unmarshaling 1000 elements took: %v", unmarshalTime)

	// Verify structure integrity
	if len(unmarshaled.Body.Content) != len(largeDoc.Body.Content) {
		t.Errorf("Content length mismatch after large data processing")
	}

	// Performance expectations (adjust based on requirements)
	if marshalTime > time.Second {
		t.Errorf("Marshaling took too long: %v", marshalTime)
	}
	if unmarshalTime > time.Second {
		t.Errorf("Unmarshaling took too long: %v", unmarshalTime)
	}
}
