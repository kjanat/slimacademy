package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kjanat/slimacademy/internal/models"
	testutils "github.com/kjanat/slimacademy/test/utils"
)

func TestBookParser_ParseBook(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name        string
		bookPath    string
		wantErr     bool
		expectTitle string
	}{
		{
			name:        "valid simple book",
			bookPath:    testutils.GetValidBookPath("simple_book"),
			wantErr:     false,
			expectTitle: "Test Book",
		},
		{
			name:     "non-existent directory",
			bookPath: "/path/that/does/not/exist",
			wantErr:  true,
		},
		{
			name:     "directory without book files",
			bookPath: testutils.GetInvalidDataPath("missing_files"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := parser.ParseBook(tt.bookPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("BookParser.ParseBook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if book == nil {
					t.Errorf("Expected book to be non-nil")
					return
				}

				if book.Title != tt.expectTitle {
					t.Errorf("Expected book title to be %q, got %q", tt.expectTitle, book.Title)
				}

				// Verify chapters were loaded
				if len(book.Chapters) == 0 {
					t.Errorf("Expected book to have chapters")
				}

				// Verify content was loaded
				if book.Content == nil {
					t.Errorf("Expected book content to be loaded")
				} else if book.Content.Document != nil && book.Content.Document.DocumentID == "" {
					t.Errorf("Expected book content document to be loaded")
				}
			}
		})
	}
}

func TestBookParser_ParseMetadata(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name     string
		jsonFile string
		wantErr  bool
	}{
		{
			name:     "valid metadata",
			jsonFile: filepath.Join(testutils.GetValidBookPath("simple_book"), "123.json"),
			wantErr:  false,
		},
		{
			name:     "malformed json",
			jsonFile: filepath.Join(testutils.GetInvalidDataPath("malformed_json"), "456.json"),
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			jsonFile: "/path/that/does/not/exist.json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := &models.Book{}
			err := parser.parseMetadata(tt.jsonFile, book)

			if (err != nil) != tt.wantErr {
				t.Errorf("BookParser.parseMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if book.ID == 0 {
					t.Errorf("Expected book ID to be set after parsing metadata")
				}
			}
		})
	}
}

func TestBookParser_ParseChapters(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name         string
		chaptersFile string
		wantErr      bool
		expectCount  int
	}{
		{
			name:         "valid chapters",
			chaptersFile: filepath.Join(testutils.GetValidBookPath("simple_book"), "chapters.json"),
			wantErr:      false,
			expectCount:  2, // Based on our test fixture
		},
		{
			name:         "non-existent chapters file",
			chaptersFile: "/path/that/does/not/exist.json",
			wantErr:      false, // Should not error, just skip
			expectCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := &models.Book{}
			err := parser.parseChapters(tt.chaptersFile, book)

			if (err != nil) != tt.wantErr {
				t.Errorf("BookParser.parseChapters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(book.Chapters) != tt.expectCount {
				t.Errorf("Expected %d chapters, got %d", tt.expectCount, len(book.Chapters))
			}

			if !tt.wantErr && tt.expectCount > 0 {
				// Verify chapter structure
				if book.Chapters[0].Title == "" {
					t.Errorf("Expected first chapter to have a title")
				}

				// Check for sub-chapters
				if len(book.Chapters) > 1 && len(book.Chapters[1].SubChapters) == 0 {
					t.Errorf("Expected second chapter to have sub-chapters based on test fixture")
				}
			}
		})
	}
}

func TestBookParser_ParseContent(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name        string
		contentFile string
		wantErr     bool
	}{
		{
			name:        "valid content",
			contentFile: filepath.Join(testutils.GetValidBookPath("simple_book"), "content.json"),
			wantErr:     false,
		},
		{
			name:        "non-existent content file",
			contentFile: "/path/that/does/not/exist.json",
			wantErr:     false, // Should not error, just skip
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book := &models.Book{}
			err := parser.parseContent(tt.contentFile, book)

			if (err != nil) != tt.wantErr {
				t.Errorf("BookParser.parseContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.contentFile != "/path/that/does/not/exist.json" {
				if book.Content == nil {
					t.Errorf("Expected content to be loaded")
				} else if book.Content.Document != nil {
					if book.Content.Document.DocumentID == "" {
						t.Errorf("Expected content document ID to be set")
					}

					if len(book.Content.Document.Body.Content) == 0 {
						t.Errorf("Expected content body to have elements")
					}
				}
			}
		})
	}
}

func TestBookParser_FindAllBooks(t *testing.T) {
	parser := NewBookParser()

	// Test with our test fixtures directory
	testDataPath := testutils.GetTestDataPath()

	books, err := parser.FindAllBooks(filepath.Join(testDataPath, "valid_books"))
	if err != nil {
		t.Fatalf("FindAllBooks() failed: %v", err)
	}

	if len(books) == 0 {
		t.Errorf("Expected to find at least one book in test fixtures")
	}

	// Verify that the found paths are valid book directories
	for _, bookPath := range books {
		// Check that it contains the required files
		chaptersPath := filepath.Join(bookPath, "chapters.json")
		contentPath := filepath.Join(bookPath, "content.json")

		if _, err := os.Stat(chaptersPath); os.IsNotExist(err) {
			t.Errorf("Book directory %s missing chapters.json", bookPath)
		}

		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			t.Errorf("Book directory %s missing content.json", bookPath)
		}
	}

	// Test with non-existent directory
	_, err = parser.FindAllBooks("/path/that/does/not/exist")
	if err == nil {
		t.Errorf("Expected error when searching non-existent directory")
	}

	// Test with empty directory
	tempDir := testutils.CreateTempDir(t)
	emptyBooks, err := parser.FindAllBooks(tempDir)
	if err != nil {
		t.Errorf("Unexpected error with empty directory: %v", err)
	}
	if len(emptyBooks) != 0 {
		t.Errorf("Expected no books in empty directory, got %d", len(emptyBooks))
	}
}

// Test Content Union Type Handling
func TestBookParser_ContentUnionHandling(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name        string
		contentJSON string
		expectDoc   bool
		expectChaps bool
		expectError bool
	}{
		{
			name: "document content",
			contentJSON: `{
				"documentId": "doc-123",
				"title": "Test Document",
				"body": {"content": []}
			}`,
			expectDoc:   true,
			expectChaps: false,
		},
		{
			name: "chapters content",
			contentJSON: `[
				{"id": 1, "title": "Chapter 1"},
				{"id": 2, "title": "Chapter 2"}
			]`,
			expectDoc:   false,
			expectChaps: true,
		},
		{
			name:        "invalid JSON",
			contentJSON: `{invalid json}`,
			expectError: true,
		},
		{
			name:        "empty object",
			contentJSON: `{}`,
			expectDoc:   true, // Empty document is valid
			expectChaps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := testutils.CreateTempDir(t)

			// Create test files
			metadataContent := `{"id": 123, "title": "Test Book"}`
			contentFile := filepath.Join(tempDir, "content.json")
			metadataFile := filepath.Join(tempDir, "123.json")
			chaptersFile := filepath.Join(tempDir, "chapters.json")

			if err := os.WriteFile(metadataFile, []byte(metadataContent), 0644); err != nil {
				t.Fatalf("Failed to create metadata file: %v", err)
			}
			if err := os.WriteFile(chaptersFile, []byte("[]"), 0644); err != nil {
				t.Fatalf("Failed to create chapters file: %v", err)
			}
			if err := os.WriteFile(contentFile, []byte(tt.contentJSON), 0644); err != nil {
				t.Fatalf("Failed to create content file: %v", err)
			}

			book, err := parser.ParseBook(tempDir)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid content")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if book.Content == nil {
				t.Error("Expected content to be loaded")
				return
			}

			if tt.expectDoc && book.Content.Document == nil {
				t.Error("Expected document content")
			}
			if tt.expectChaps && book.Content.Chapters == nil {
				t.Error("Expected chapters content")
			}
			if !tt.expectDoc && book.Content.Document != nil {
				t.Error("Did not expect document content")
			}
			if !tt.expectChaps && book.Content.Chapters != nil {
				t.Error("Did not expect chapters content")
			}
		})
	}
}

// Test Inline Object Mapping Generation
func TestBookParser_InlineObjectMapping(t *testing.T) {
	parser := NewBookParser()

	// Create test book with images and inline objects
	tempDir := testutils.CreateTempDir(t)

	metadataContent := `{
		"id": 123,
		"title": "Test Book",
		"images": [
			{
				"id": 1,
				"objectId": "kix.test1",
				"imageUrl": "/uploads/test1.png"
			},
			{
				"id": 2,
				"objectId": "kix.test2",
				"imageUrl": "/uploads/test2.png"
			}
		]
	}`

	contentContent := `{
		"documentId": "doc-123",
		"inlineObjects": {
			"kix.inline1": {
				"inlineObjectProperties": {
					"embeddedObject": {
						"imageProperties": {
							"contentUri": "/uploads/inline1.png"
						}
					}
				}
			},
			"kix.inline2": {
				"inlineObjectProperties": {
					"embeddedObject": {
						"imageProperties": {
							"contentUri": "/uploads/inline2.png"
						}
					}
				}
			}
		},
		"body": {"content": []}
	}`

	// Create test files
	files := map[string]string{
		"123.json":      metadataContent,
		"chapters.json": "[]",
		"content.json":  contentContent,
	}

	for filename, content := range files {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	book, err := parser.ParseBook(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse book: %v", err)
	}

	// Verify inline object map was built
	if book.InlineObjectMap == nil {
		t.Error("Expected InlineObjectMap to be initialized")
		return
	}

	expectedMappings := map[string]string{
		"kix.test1":   "https://api.slimacademy.nl/uploads/test1.png",   // From images array (normalized)
		"kix.test2":   "https://api.slimacademy.nl/uploads/test2.png",   // From images array (normalized)
		"kix.inline1": "https://api.slimacademy.nl/uploads/inline1.png", // From document inline objects (normalized)
		"kix.inline2": "https://api.slimacademy.nl/uploads/inline2.png", // From document inline objects (normalized)
	}

	for objectID, expectedURL := range expectedMappings {
		actualURL, exists := book.InlineObjectMap[objectID]
		if !exists {
			t.Errorf("Expected object ID %s to be in InlineObjectMap", objectID)
		} else if actualURL != expectedURL {
			t.Errorf("Expected URL %s for object %s, got %s", expectedURL, objectID, actualURL)
		}
	}
}

// Test Academic Metadata Parsing
func TestBookParser_AcademicMetadata(t *testing.T) {
	parser := NewBookParser()

	now := time.Now()
	metadataContent := fmt.Sprintf(`{
		"id": 123,
		"title": "Advanced Medical Studies",
		"description": "Comprehensive medical course material",
		"availableDate": "2025-01-15",
		"examDate": "2025-02-15",
		"bachelorYearNumber": "Bachelor 2",
		"collegeStartYear": 2024,
		"shopUrl": "/shop/medical-studies",
		"isPurchased": 1,
		"lastOpenedAt": "%s",
		"readProgress": 75,
		"pageCount": 250,
		"readPageCount": 150,
		"readPercentage": 60.5,
		"hasFreeChapters": 1,
		"supplements": ["video", "quiz"],
		"formulasImages": ["formula1.png", "formula2.png"],
		"periods": ["Q1 2025", "Q2 2025"],
		"images": [
			{
				"id": 1,
				"summaryId": 123,
				"createdAt": "%s",
				"objectId": "img-1",
				"mimeType": "image/png",
				"imageUrl": "/uploads/img1.png"
			}
		]
	}`, now.Format(time.RFC3339), now.Format(time.RFC3339))

	tempDir := testutils.CreateTempDir(t)
	files := map[string]string{
		"123.json":      metadataContent,
		"chapters.json": "[]",
		"content.json":  `{"documentId": "doc-123", "body": {"content": []}}`,
	}

	for filename, content := range files {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	book, err := parser.ParseBook(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse book: %v", err)
	}

	// Verify all metadata fields
	if book.ID != 123 {
		t.Errorf("Expected ID 123, got %d", book.ID)
	}
	if book.Title != "Advanced Medical Studies" {
		t.Errorf("Expected title 'Advanced Medical Studies', got %s", book.Title)
	}
	if book.BachelorYearNumber != "Bachelor 2" {
		t.Errorf("Expected bachelor year 'Bachelor 2', got %s", book.BachelorYearNumber)
	}
	if book.CollegeStartYear != 2024 {
		t.Errorf("Expected college start year 2024, got %d", book.CollegeStartYear)
	}
	if !book.IsPurchased.Bool() {
		t.Errorf("Expected isPurchased true, got %v", book.IsPurchased)
	}
	if book.ReadProgress == nil || *book.ReadProgress != 75 {
		t.Errorf("Expected read progress 75, got %v", book.ReadProgress)
	}
	if book.PageCount != 250 {
		t.Errorf("Expected page count 250, got %d", book.PageCount)
	}
	if !book.HasFreeChapters.Bool() {
		t.Errorf("Expected has free chapters true, got %v", book.HasFreeChapters)
	}

	// Verify arrays
	if len(book.Supplements) != 2 {
		t.Errorf("Expected 2 supplements, got %d", len(book.Supplements))
	}
	if len(book.FormulasImages) != 2 {
		t.Errorf("Expected 2 formula images, got %d", len(book.FormulasImages))
	}
	if len(book.Periods) != 2 {
		t.Errorf("Expected 2 periods, got %d", len(book.Periods))
	}

	// Verify BookImage structure
	if len(book.Images) != 1 {
		t.Errorf("Expected 1 image, got %d", len(book.Images))
	} else {
		img := book.Images[0]
		if img.ObjectID != "img-1" {
			t.Errorf("Expected object ID 'img-1', got %s", img.ObjectID)
		}
		if img.MIMEType != "image/png" {
			t.Errorf("Expected MIME type 'image/png', got %s", img.MIMEType)
		}
		if img.ImageURL != "https://api.slimacademy.nl/uploads/img1.png" {
			t.Errorf("Expected image URL 'https://api.slimacademy.nl/uploads/img1.png', got %s", img.ImageURL)
		}
	}
}

// Test Error Handling and Recovery
func TestBookParser_ErrorHandlingEnhanced(t *testing.T) {
	parser := NewBookParser()

	tests := []struct {
		name        string
		setupFiles  map[string]string
		expectError bool
		errorMatch  string
	}{
		{
			name: "malformed metadata JSON",
			setupFiles: map[string]string{
				"123.json":      `{"id": 123, invalid json}`,
				"chapters.json": "[]",
				"content.json":  `{"documentId": "doc", "body": {"content": []}}`,
			},
			expectError: true,
			errorMatch:  "failed to parse metadata",
		},
		{
			name: "malformed chapters JSON",
			setupFiles: map[string]string{
				"123.json":      `{"id": 123, "title": "Test"}`,
				"chapters.json": `[{"id": 1, invalid json}]`,
				"content.json":  `{"documentId": "doc", "body": {"content": []}}`,
			},
			expectError: true,
			errorMatch:  "failed to parse chapters",
		},
		{
			name: "malformed content JSON",
			setupFiles: map[string]string{
				"123.json":      `{"id": 123, "title": "Test"}`,
				"chapters.json": "[]",
				"content.json":  `{"documentId": invalid json}`,
			},
			expectError: true,
			errorMatch:  "failed to parse content",
		},
		{
			name:        "missing all files",
			setupFiles:  map[string]string{},
			expectError: true,
			errorMatch:  "no metadata file found",
		},
		{
			name: "only excluded files present",
			setupFiles: map[string]string{
				"chapters.json":   "[]",
				"content.json":    `{"documentId": "doc", "body": {"content": []}}`,
				"list-notes.json": "[]",
			},
			expectError: true,
			errorMatch:  "no metadata file found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := testutils.CreateTempDir(t)

			// Create test files
			for filename, content := range tt.setupFiles {
				path := filepath.Join(tempDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create %s: %v", filename, err)
				}
			}

			_, err := parser.ParseBook(tempDir)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMatch, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test Real SlimAcademy Data Format
func TestBookParser_RealDataFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real data test in short mode")
	}

	parser := NewBookParser()

	// Test with real SlimAcademy source directory if available
	sourceDir := filepath.Join("..", "..", "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Skip("Source directory not available for real data testing")
	}

	bookDirs, err := parser.FindAllBooks(sourceDir)
	if err != nil {
		t.Fatalf("Failed to find books in source directory: %v", err)
	}

	if len(bookDirs) == 0 {
		t.Skip("No books found in source directory")
	}

	// Test parsing the first few books (limit to avoid long test times)
	maxBooks := 3
	if len(bookDirs) < maxBooks {
		maxBooks = len(bookDirs)
	}

	for i := 0; i < maxBooks; i++ {
		bookDir := bookDirs[i]
		t.Run(fmt.Sprintf("book_%d", i), func(t *testing.T) {
			book, err := parser.ParseBook(bookDir)
			if err != nil {
				t.Errorf("Failed to parse real book at %s: %v", bookDir, err)
				return
			}

			// Verify essential fields are present
			if book.ID == 0 {
				t.Error("Book ID should be set")
			}
			if book.Title == "" {
				t.Error("Book title should be set")
			}

			// Verify inline object map was built if there are images
			if len(book.Images) > 0 && book.InlineObjectMap == nil {
				t.Error("InlineObjectMap should be initialized when images are present")
			}

			// Verify content structure
			if book.Content != nil {
				if book.Content.Document != nil && book.Content.Chapters != nil {
					t.Error("Content should have either Document or Chapters, not both")
				}
				if book.Content.Document == nil && book.Content.Chapters == nil {
					t.Error("Content should have either Document or Chapters")
				}
			}

			t.Logf("Successfully parsed book: %s (ID: %d)", book.Title, book.ID)
		})
	}
}

// Test Performance with Large Documents
func TestBookParser_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	parser := NewBookParser()

	// Create a large test document
	largeContent := createLargeDocumentContent(10000) // 10k paragraphs

	tempDir := testutils.CreateTempDir(t)
	files := map[string]string{
		"123.json":      `{"id": 123, "title": "Large Book"}`,
		"chapters.json": "[]",
		"content.json":  largeContent,
	}

	for filename, content := range files {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// Time the parsing
	start := time.Now()
	book, err := parser.ParseBook(tempDir)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to parse large book: %v", err)
	}

	if book.Content == nil || book.Content.Document == nil {
		t.Error("Expected document content to be parsed")
	}

	t.Logf("Parsed large document in %v", duration)

	// Performance threshold (adjust based on requirements)
	if duration > 5*time.Second {
		t.Errorf("Parsing took too long: %v (expected < 5s)", duration)
	}
}

// Helper function to create large document content
func createLargeDocumentContent(paragraphCount int) string {
	var content struct {
		DocumentID string `json:"documentId"`
		Body       struct {
			Content []map[string]interface{} `json:"content"`
		} `json:"body"`
	}

	content.DocumentID = "large-doc"
	content.Body.Content = make([]map[string]interface{}, paragraphCount)

	for i := 0; i < paragraphCount; i++ {
		content.Body.Content[i] = map[string]interface{}{
			"paragraph": map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"textRun": map[string]interface{}{
							"content": fmt.Sprintf("This is paragraph number %d with some content.", i+1),
						},
					},
				},
			},
		}
	}

	data, _ := json.Marshal(content)
	return string(data)
}

func TestBookParser_MetadataFileDetection(t *testing.T) {
	// Create a temporary directory with various JSON files
	tempDir := testutils.CreateTempDir(t)

	// Create test files
	files := []string{
		"123.json",        // Should be detected as metadata
		"chapters.json",   // Should be ignored
		"content.json",    // Should be ignored
		"list-notes.json", // Should be ignored
		"other.json",      // Should be detected as metadata
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		var content string
		switch file {
		case "chapters.json":
			content = `[]` // Chapters should be an array
		case "content.json":
			content = `{"documentId": "", "body": {"content": []}}` // Minimal valid content
		default:
			content = `{"id": 123, "title": "Test"}` // Metadata format
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	parser := NewBookParser()

	// The parser should find a metadata file (either 123.json or other.json)
	book, err := parser.ParseBook(tempDir)
	if err != nil {
		t.Fatalf("Expected successful parsing, got error: %v", err)
	}

	if book.ID != 123 {
		t.Errorf("Expected book ID to be 123, got %d", book.ID)
	}
	if book.Title != "Test" {
		t.Errorf("Expected book title to be 'Test', got %s", book.Title)
	}
}

func TestBookParser_ErrorHandling(t *testing.T) {
	parser := NewBookParser()

	// Test with directory that has no metadata files
	tempDir := testutils.CreateTempDir(t)

	// Create only non-metadata files
	files := map[string]string{
		"chapters.json":   "[]",
		"content.json":    `{"documentId": "", "body": {"content": []}}`,
		"list-notes.json": "[]",
	}
	for file, content := range files {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	_, err := parser.ParseBook(tempDir)
	if err == nil {
		t.Errorf("Expected error when no metadata file found")
	}
}
