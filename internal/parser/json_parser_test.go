package parser

import (
	"os"
	"path/filepath"
	"testing"

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
				if book.Content.DocumentID == "" {
					t.Errorf("Expected book content to be loaded")
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
		name        string
		chaptersFile string
		wantErr     bool
		expectCount int
	}{
		{
			name:        "valid chapters",
			chaptersFile: filepath.Join(testutils.GetValidBookPath("simple_book"), "chapters.json"),
			wantErr:     false,
			expectCount: 2, // Based on our test fixture
		},
		{
			name:        "non-existent chapters file",
			chaptersFile: "/path/that/does/not/exist.json",
			wantErr:     false, // Should not error, just skip
			expectCount: 0,
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
				if book.Content.DocumentID == "" {
					t.Errorf("Expected content document ID to be set")
				}
				
				if len(book.Content.Body.Content) == 0 {
					t.Errorf("Expected content body to have elements")
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

func TestBookParser_MetadataFileDetection(t *testing.T) {
	// Create a temporary directory with various JSON files
	tempDir := testutils.CreateTempDir(t)
	
	// Create test files
	files := []string{
		"123.json",         // Should be detected as metadata
		"chapters.json",    // Should be ignored
		"content.json",     // Should be ignored
		"list-notes.json",  // Should be ignored
		"other.json",       // Should be detected as metadata
	}
	
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		var content string
		if file == "chapters.json" {
			content = `[]` // Chapters should be an array
		} else if file == "content.json" {
			content = `{"documentId": "", "body": {"content": []}}` // Minimal valid content
		} else {
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

