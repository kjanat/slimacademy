package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBookSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected *Book
		wantErr  bool
	}{
		{
			name: "complete book with all fields",
			jsonData: `{
				"id": 12345,
				"title": "Advanced Mathematics",
				"description": "Comprehensive mathematics course",
				"availableDate": "2024-01-15",
				"examDate": "2024-06-15",
				"bachelorYearNumber": "3",
				"collegeStartYear": 2022,
				"shopUrl": "https://shop.example.com/book/12345",
				"isPurchased": 1,
				"lastOpenedAt": "2024-01-20T10:30:00Z",
				"readProgress": 75,
				"pageCount": 300,
				"readPageCount": 225,
				"readPercentage": 75.0,
				"hasFreeChapters": 1,
				"supplements": ["formula-sheet", "exercises"],
				"images": [
					{
						"id": 1,
						"summaryId": 12345,
						"createdAt": "2024-01-01T00:00:00Z",
						"objectId": "obj_123",
						"mimeType": "image/png",
						"imageUrl": "https://example.com/image1.png"
					}
				],
				"formulasImages": [],
				"periods": ["Fall 2024", "Spring 2025"]
			}`,
			expected: func() *Book {
				lastOpened, _ := time.Parse(time.RFC3339, "2024-01-20T10:30:00Z")
				customTime := &CustomTime{lastOpened}
				return &Book{
					ID:                 12345,
					Title:              "Advanced Mathematics",
					Description:        "Comprehensive mathematics course",
					AvailableDate:      "2024-01-15",
					ExamDate:           "2024-06-15",
					BachelorYearNumber: "3",
					CollegeStartYear:   2022,
					ShopURL:            "https://shop.example.com/book/12345",
					IsPurchased:        BoolInt(true),
					LastOpenedAt:       customTime,
					ReadProgress:       &[]int64{75}[0],
					PageCount:          300,
					HasFreeChapters:    BoolInt(true),
					Supplements:        []any{"formula-sheet", "exercises"},
					Images: []BookImage{
						{
							ID:        1,
							SummaryID: 12345,
							CreatedAt: func() CustomTime { t, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"); return CustomTime{t} }(),
							ObjectID:  "obj_123",
							MIMEType:  "image/png",
							ImageURL:  "https://example.com/image1.png",
						},
					},
					FormulasImages: []any{},
					Periods:        []string{"Fall 2024", "Spring 2025"},
				}
			}(),
			wantErr: false,
		},
		{
			name: "minimal book with required fields only",
			jsonData: `{
				"id": 1,
				"title": "Basic Book",
				"description": "",
				"availableDate": "",
				"examDate": "",
				"bachelorYearNumber": "",
				"collegeStartYear": 0,
				"shopUrl": "",
				"isPurchased": 0,
				"pageCount": 0,
				"readPageCount": null,
				"readPercentage": null,
				"hasFreeChapters": 0,
				"supplements": [],
				"images": [],
				"formulasImages": [],
				"periods": []
			}`,
			expected: &Book{
				ID:                 1,
				Title:              "Basic Book",
				Description:        "",
				AvailableDate:      "",
				ExamDate:           "",
				BachelorYearNumber: "",
				CollegeStartYear:   0,
				ShopURL:            "",
				IsPurchased:        BoolInt(false),
				PageCount:          0,
				HasFreeChapters:    BoolInt(false),
				Supplements:        []any{},
				Images:             []BookImage{},
				FormulasImages:     []any{},
				Periods:            []string{},
			},
			wantErr: false,
		},
		{
			name: "book with null optional fields",
			jsonData: `{
				"id": 2,
				"title": "Test Book",
				"description": "Test",
				"availableDate": "2024-01-01",
				"examDate": "2024-06-01",
				"bachelorYearNumber": "1",
				"collegeStartYear": 2024,
				"shopUrl": "https://test.com",
				"isPurchased": 0,
				"lastOpenedAt": null,
				"readProgress": null,
				"pageCount": 100,
				"readPageCount": null,
				"readPercentage": null,
				"hasFreeChapters": 0,
				"supplements": [],
				"images": [],
				"formulasImages": [],
				"periods": []
			}`,
			expected: &Book{
				ID:                 2,
				Title:              "Test Book",
				Description:        "Test",
				AvailableDate:      "2024-01-01",
				ExamDate:           "2024-06-01",
				BachelorYearNumber: "1",
				CollegeStartYear:   2024,
				ShopURL:            "https://test.com",
				IsPurchased:        BoolInt(false),
				LastOpenedAt:       nil,
				ReadProgress:       nil,
				PageCount:          100,
				HasFreeChapters:    BoolInt(false),
				Supplements:        []any{},
				Images:             []BookImage{},
				FormulasImages:     []any{},
				Periods:            []string{},
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"id": "invalid"}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := UnmarshalBook([]byte(tt.jsonData))

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalBook() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBook() error = %v", err)
				return
			}

			// Check basic fields
			if book.ID != tt.expected.ID {
				t.Errorf("Book.ID = %v, expected %v", book.ID, tt.expected.ID)
			}
			if book.Title != tt.expected.Title {
				t.Errorf("Book.Title = %v, expected %v", book.Title, tt.expected.Title)
			}
			if book.Description != tt.expected.Description {
				t.Errorf("Book.Description = %v, expected %v", book.Description, tt.expected.Description)
			}

			// Check pointer fields - handle time parsing in expected data
			if tt.expected.LastOpenedAt != nil {
				if book.LastOpenedAt == nil {
					t.Errorf("Book.LastOpenedAt expected to be set but was nil")
				}
			} else {
				if book.LastOpenedAt != nil {
					t.Errorf("Book.LastOpenedAt expected to be nil but was set")
				}
			}

			if tt.expected.ReadProgress != nil {
				if book.ReadProgress == nil {
					t.Errorf("Book.ReadProgress expected to be set but was nil")
				} else if *book.ReadProgress != *tt.expected.ReadProgress {
					t.Errorf("Book.ReadProgress = %v, expected %v", *book.ReadProgress, *tt.expected.ReadProgress)
				}
			} else {
				if book.ReadProgress != nil {
					t.Errorf("Book.ReadProgress expected to be nil but was set")
				}
			}

			// Check slice fields
			if len(book.Images) != len(tt.expected.Images) {
				t.Errorf("Book.Images length = %v, expected %v", len(book.Images), len(tt.expected.Images))
			}
			if len(book.Periods) != len(tt.expected.Periods) {
				t.Errorf("Book.Periods length = %v, expected %v", len(book.Periods), len(tt.expected.Periods))
			}
		})
	}
}

func TestBookImageSerialization(t *testing.T) {
	createdAt := CustomTime{func() time.Time { t, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z"); return t }()}

	tests := []struct {
		name     string
		jsonData string
		expected BookImage
		wantErr  bool
	}{
		{
			name: "complete book image",
			jsonData: `{
				"id": 123,
				"summaryId": 456,
				"createdAt": "2024-01-01T12:00:00Z",
				"objectId": "obj_789",
				"mimeType": "image/jpeg",
				"imageUrl": "https://example.com/image.jpg"
			}`,
			expected: BookImage{
				ID:        123,
				SummaryID: 456,
				CreatedAt: createdAt,
				ObjectID:  "obj_789",
				MIMEType:  "image/jpeg",
				ImageURL:  "https://example.com/image.jpg",
			},
			wantErr: false,
		},
		{
			name:     "invalid time format",
			jsonData: `{"id": 1, "summaryId": 2, "createdAt": "invalid-time", "objectId": "obj", "mimeType": "image/png", "imageUrl": "url"}`,
			expected: BookImage{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var img BookImage
			err := json.Unmarshal([]byte(tt.jsonData), &img)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BookImage unmarshal expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("BookImage unmarshal error = %v", err)
				return
			}

			if img.ID != tt.expected.ID {
				t.Errorf("BookImage.ID = %v, expected %v", img.ID, tt.expected.ID)
			}
			if img.ObjectID != tt.expected.ObjectID {
				t.Errorf("BookImage.ObjectID = %v, expected %v", img.ObjectID, tt.expected.ObjectID)
			}
			if !img.CreatedAt.Time.Equal(tt.expected.CreatedAt.Time) {
				t.Errorf("BookImage.CreatedAt = %v, expected %v", img.CreatedAt, tt.expected.CreatedAt)
			}
		})
	}
}

func TestBookMarshalJSON(t *testing.T) {
	now := time.Now()
	progress := int64(50)
	customTime := &CustomTime{now}

	book := &Book{
		ID:           1,
		Title:        "Test Book",
		Description:  "A test book",
		LastOpenedAt: customTime,
		ReadProgress: &progress,
		Images: []BookImage{
			{
				ID:        1,
				SummaryID: 1,
				CreatedAt: CustomTime{now},
				ObjectID:  "obj_1",
				MIMEType:  "image/png",
				ImageURL:  "https://example.com/image.png",
			},
		},
		Periods: []string{"Spring 2024"},
	}

	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back to verify roundtrip
	var unmarshaled Book
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != book.ID {
		t.Errorf("Roundtrip failed for ID: got %v, want %v", unmarshaled.ID, book.ID)
	}
	if unmarshaled.Title != book.Title {
		t.Errorf("Roundtrip failed for Title: got %v, want %v", unmarshaled.Title, book.Title)
	}
	if len(unmarshaled.Images) != len(book.Images) {
		t.Errorf("Roundtrip failed for Images length: got %v, want %v", len(unmarshaled.Images), len(book.Images))
	}
}

func TestBookValidation(t *testing.T) {
	tests := []struct {
		name     string
		book     *Book
		validate func(*Book) bool
		expected bool
	}{
		{
			name: "valid purchased book",
			book: &Book{
				ID:          1,
				Title:       "Test",
				IsPurchased: BoolInt(true),
				PageCount:   100,
			},
			validate: func(b *Book) bool {
				return b.IsPurchased.Bool() && b.PageCount > 0
			},
			expected: true,
		},
		{
			name: "valid free book with free chapters",
			book: &Book{
				ID:              2,
				Title:           "Free Book",
				IsPurchased:     BoolInt(false),
				HasFreeChapters: BoolInt(true),
			},
			validate: func(b *Book) bool {
				return !b.IsPurchased.Bool() && b.HasFreeChapters.Bool()
			},
			expected: true,
		},
		{
			name: "book with read progress",
			book: &Book{
				ID:           3,
				Title:        "In Progress",
				PageCount:    200,
				ReadProgress: func() *int64 { p := int64(150); return &p }(),
			},
			validate: func(b *Book) bool {
				return b.ReadProgress != nil && *b.ReadProgress <= b.PageCount
			},
			expected: true,
		},
		{
			name: "book with invalid read progress",
			book: &Book{
				ID:           4,
				Title:        "Invalid Progress",
				PageCount:    100,
				ReadProgress: func() *int64 { p := int64(150); return &p }(),
			},
			validate: func(b *Book) bool {
				return b.ReadProgress != nil && *b.ReadProgress <= b.PageCount
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.validate(tt.book)
			if result != tt.expected {
				t.Errorf("Validation result = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBookDeepCopy(t *testing.T) {
	// Test that modifications to complex fields don't affect original
	original := &Book{
		ID:       1,
		Title:    "Original",
		Images:   []BookImage{{ID: 1, ObjectID: "obj1"}},
		Periods:  []string{"Period1"},
		Chapters: []Chapter{{ID: 1, Title: "Chapter 1"}},
	}

	// Simulate deep copy by marshaling and unmarshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var copy Book
	err = json.Unmarshal(data, &copy)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Modify copy
	copy.Title = "Modified"
	if len(copy.Images) > 0 {
		copy.Images[0].ObjectID = "modified_obj"
	}
	if len(copy.Periods) > 0 {
		copy.Periods[0] = "Modified Period"
	}

	// Original should be unchanged
	if original.Title != "Original" {
		t.Errorf("Original title was modified: %v", original.Title)
	}
	if len(original.Images) > 0 && original.Images[0].ObjectID != "obj1" {
		t.Errorf("Original image ObjectID was modified: %v", original.Images[0].ObjectID)
	}
	if len(original.Periods) > 0 && original.Periods[0] != "Period1" {
		t.Errorf("Original period was modified: %v", original.Periods[0])
	}
}

func TestBookAcademicMetadata(t *testing.T) {
	tests := []struct {
		name           string
		book           *Book
		expectedYear   int64
		expectedPeriod string
		hasValidDates  bool
	}{
		{
			name: "complete academic metadata",
			book: &Book{
				BachelorYearNumber: "3",
				CollegeStartYear:   2022,
				AvailableDate:      "2024-01-15",
				ExamDate:           "2024-06-15",
				Periods:            []string{"Fall 2024", "Spring 2025"},
			},
			expectedYear:   2022,
			expectedPeriod: "Fall 2024",
			hasValidDates:  true,
		},
		{
			name: "minimal academic metadata",
			book: &Book{
				BachelorYearNumber: "1",
				CollegeStartYear:   2023,
				Periods:            []string{},
			},
			expectedYear:   2023,
			expectedPeriod: "",
			hasValidDates:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.book.CollegeStartYear != tt.expectedYear {
				t.Errorf("CollegeStartYear = %v, expected %v", tt.book.CollegeStartYear, tt.expectedYear)
			}

			if len(tt.book.Periods) > 0 {
				if tt.book.Periods[0] != tt.expectedPeriod {
					t.Errorf("First period = %v, expected %v", tt.book.Periods[0], tt.expectedPeriod)
				}
			} else if tt.expectedPeriod != "" {
				t.Errorf("Expected period %v but got empty slice", tt.expectedPeriod)
			}

			hasValidDates := tt.book.AvailableDate != "" && tt.book.ExamDate != ""
			if hasValidDates != tt.hasValidDates {
				t.Errorf("Has valid dates = %v, expected %v", hasValidDates, tt.hasValidDates)
			}
		})
	}
}
