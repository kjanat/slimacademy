package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/client"
)

// Test helper to create temporary directory
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "source-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// Test helper to create mock book data
func createMockBookData(id, title string) *client.BookData {
	return &client.BookData{
		ID:    id,
		Title: title,
		Summary: map[string]any{
			"id":          id,
			"title":       title,
			"description": fmt.Sprintf("Description for %s", title),
			"author":      "Test Author",
		},
		Chapters: []map[string]any{
			{
				"id":      "1",
				"title":   "Chapter 1",
				"content": "Content of chapter 1",
			},
			{
				"id":      "2",
				"title":   "Chapter 2",
				"content": "Content of chapter 2",
			},
		},
		Content: map[string]any{
			"metadata": map[string]any{
				"total_chapters": 2,
				"word_count":     1500,
			},
			"structure": []map[string]any{
				{"type": "chapter", "id": "1"},
				{"type": "chapter", "id": "2"},
			},
		},
		Notes: []map[string]any{
			{
				"id":      "note1",
				"content": "First note",
				"page":    1,
			},
			{
				"id":      "note2",
				"content": "Second note",
				"page":    15,
			},
		},
	}
}

// Test helper to verify JSON file contents
func verifyJSONFile(t *testing.T, filepath string, expected any) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	var actual any
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("Failed to unmarshal JSON from %s: %v", filepath, err)
	}

	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	if string(expectedJSON) != string(actualJSON) {
		t.Errorf("JSON content mismatch in %s.\nExpected: %s\nActual: %s",
			filepath, string(expectedJSON), string(actualJSON))
	}
}

func TestNewSourceManager(t *testing.T) {
	tests := []struct {
		name        string
		sourceDir   string
		expectedDir string
	}{
		{
			name:        "with custom directory",
			sourceDir:   "/custom/source",
			expectedDir: "/custom/source",
		},
		{
			name:        "with empty directory",
			sourceDir:   "",
			expectedDir: "source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSourceManager(tt.sourceDir)

			if sm == nil {
				t.Fatal("NewSourceManager returned nil")
			}

			if sm.sourceDir != tt.expectedDir {
				t.Errorf("Expected sourceDir %q, got %q", tt.expectedDir, sm.sourceDir)
			}
		})
	}
}

func TestSourceManager_SaveBookData(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("save valid book data", func(t *testing.T) {
		bookData := createMockBookData("3631", "Advanced Go Programming")

		err := sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("SaveBookData failed: %v", err)
		}

		// Check that directory was created
		expectedDir := filepath.Join(tempDir, "Advanced Go Programming")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to be created", expectedDir)
		}

		// Verify all expected files exist and have correct content
		files := map[string]any{
			"3631.json":       bookData.Summary,
			"chapters.json":   bookData.Chapters,
			"content.json":    bookData.Content,
			"list-notes.json": bookData.Notes,
		}

		for filename, expectedData := range files {
			filePath := filepath.Join(expectedDir, filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist", filePath)
				continue
			}
			verifyJSONFile(t, filePath, expectedData)
		}
	})

	t.Run("save book with problematic title", func(t *testing.T) {
		bookData := createMockBookData("1234", "Book/With\\Bad:Chars*?\"<>|")

		err := sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("SaveBookData failed: %v", err)
		}

		// Check that directory name was sanitized (backslash becomes dash too)
		expectedDir := filepath.Join(tempDir, "Book-With-Bad-Chars------")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			// List the actual directories to see what was created
			entries, _ := os.ReadDir(tempDir)
			var dirNames []string
			for _, entry := range entries {
				if entry.IsDir() {
					dirNames = append(dirNames, entry.Name())
				}
			}
			t.Errorf("Expected sanitized directory %s to be created. Found directories: %v", expectedDir, dirNames)
		}
	})

	t.Run("save book with empty title", func(t *testing.T) {
		bookData := createMockBookData("5678", "")

		err := sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("SaveBookData failed: %v", err)
		}

		// Check that fallback directory name was used
		expectedDir := filepath.Join(tempDir, "Book_5678")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Expected fallback directory %s to be created", expectedDir)
		}
	})

	t.Run("save book with whitespace-only title", func(t *testing.T) {
		bookData := createMockBookData("9999", "   \n\t   ")

		err := sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("SaveBookData failed: %v", err)
		}

		// Check that fallback directory name was used
		expectedDir := filepath.Join(tempDir, "Book_9999")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Expected fallback directory %s to be created", expectedDir)
		}
	})

	t.Run("overwrite existing book", func(t *testing.T) {
		bookData := createMockBookData("1111", "Test Book")

		// Save once
		err := sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("First SaveBookData failed: %v", err)
		}

		// Modify and save again
		bookData.Summary = map[string]any{
			"id":          bookData.ID,
			"title":       bookData.Title,
			"description": "Updated description",
		}

		err = sm.SaveBookData(bookData)
		if err != nil {
			t.Fatalf("Second SaveBookData failed: %v", err)
		}

		// Verify the updated content
		summaryPath := filepath.Join(tempDir, "Test Book", "1111.json")
		verifyJSONFile(t, summaryPath, bookData.Summary)
	})
}

func TestSourceManager_SaveMultipleBooks(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	books := []*client.BookData{
		createMockBookData("1001", "First Book"),
		createMockBookData("1002", "Second Book"),
		createMockBookData("1003", "Third Book"),
	}

	t.Run("save multiple valid books", func(t *testing.T) {
		err := sm.SaveMultipleBooks(books)
		if err != nil {
			t.Fatalf("SaveMultipleBooks failed: %v", err)
		}

		// Verify all books were saved
		for _, book := range books {
			bookDir := filepath.Join(tempDir, book.Title)
			if _, err := os.Stat(bookDir); os.IsNotExist(err) {
				t.Errorf("Expected directory for book %s to exist", book.Title)
			}

			summaryFile := filepath.Join(bookDir, fmt.Sprintf("%s.json", book.ID))
			if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
				t.Errorf("Expected summary file for book %s to exist", book.ID)
			}
		}
	})

	t.Run("save with one invalid book", func(t *testing.T) {
		// Create a mix of valid and invalid books
		mixedBooks := []*client.BookData{
			createMockBookData("2001", "Valid Book 1"),
			{ID: "2002", Title: "Invalid Book", Summary: make(chan int)}, // Invalid JSON
			createMockBookData("2003", "Valid Book 2"),
		}

		err := sm.SaveMultipleBooks(mixedBooks)
		if err == nil {
			t.Fatal("Expected SaveMultipleBooks to fail with invalid book data")
		}

		if !strings.Contains(err.Error(), "failed to save book Invalid Book") {
			t.Errorf("Expected error to mention failed book, got: %v", err)
		}
	})

	t.Run("save empty book list", func(t *testing.T) {
		err := sm.SaveMultipleBooks([]*client.BookData{})
		if err != nil {
			t.Errorf("SaveMultipleBooks should not fail with empty list: %v", err)
		}
	})
}

func TestSourceManager_ListExistingBooks(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("list from empty directory", func(t *testing.T) {
		books, err := sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("ListExistingBooks failed: %v", err)
		}

		if len(books) != 0 {
			t.Errorf("Expected 0 books in empty directory, got %d", len(books))
		}
	})

	t.Run("list from nonexistent directory", func(t *testing.T) {
		sm := NewSourceManager(filepath.Join(tempDir, "nonexistent"))
		books, err := sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("ListExistingBooks should not fail for nonexistent directory: %v", err)
		}

		if len(books) != 0 {
			t.Errorf("Expected 0 books for nonexistent directory, got %d", len(books))
		}
	})

	t.Run("list with books present", func(t *testing.T) {
		// Create some book directories manually
		bookDirs := []string{"Book One", "Book Two", "Book Three"}
		for _, bookDir := range bookDirs {
			if err := os.MkdirAll(filepath.Join(tempDir, bookDir), 0755); err != nil {
				t.Fatalf("Failed to create book directory: %v", err)
			}
		}

		// Create a hidden directory (should be ignored)
		if err := os.MkdirAll(filepath.Join(tempDir, ".hidden"), 0755); err != nil {
			t.Fatalf("Failed to create hidden directory: %v", err)
		}

		// Create a file (should be ignored)
		if err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		books, err := sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("ListExistingBooks failed: %v", err)
		}

		if len(books) != 3 {
			t.Errorf("Expected 3 books, got %d: %v", len(books), books)
		}

		for _, expectedBook := range bookDirs {
			found := false
			for _, actualBook := range books {
				if actualBook == expectedBook {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find book %q in list %v", expectedBook, books)
			}
		}

		// Verify hidden directory is not listed
		for _, book := range books {
			if book == ".hidden" {
				t.Error("Hidden directory should not be listed as a book")
			}
		}
	})
}

func TestSourceManager_CleanSourceDirectory(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("clean nonexistent directory", func(t *testing.T) {
		sm := NewSourceManager(filepath.Join(tempDir, "nonexistent"))
		err := sm.CleanSourceDirectory()
		if err != nil {
			t.Errorf("CleanSourceDirectory should not fail for nonexistent directory: %v", err)
		}
	})

	t.Run("clean empty directory", func(t *testing.T) {
		// Create the directory first
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			t.Fatalf("Failed to create source directory: %v", err)
		}

		err := sm.CleanSourceDirectory()
		if err != nil {
			t.Errorf("CleanSourceDirectory failed on empty directory: %v", err)
		}
	})

	t.Run("clean directory with contents", func(t *testing.T) {
		// Create some test content
		testDirs := []string{"Book1", "Book2", "Book3"}
		for _, dir := range testDirs {
			bookDir := filepath.Join(tempDir, dir)
			if err := os.MkdirAll(bookDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Add some files in each directory
			testFile := filepath.Join(bookDir, "test.json")
			if err := os.WriteFile(testFile, []byte(`{"test": true}`), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		// Add a test file in the root
		rootFile := filepath.Join(tempDir, "root.txt")
		if err := os.WriteFile(rootFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create root file: %v", err)
		}

		// Clean the directory
		err := sm.CleanSourceDirectory()
		if err != nil {
			t.Fatalf("CleanSourceDirectory failed: %v", err)
		}

		// Verify everything was removed
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read directory after cleaning: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("Expected directory to be empty after cleaning, found %d entries: %v", len(entries), entries)
		}
	})

	t.Run("clean with permission error", func(t *testing.T) {
		// Create a directory with restricted permissions
		restrictedDir := filepath.Join(tempDir, "restricted")
		if err := os.MkdirAll(restrictedDir, 0755); err != nil {
			t.Fatalf("Failed to create restricted directory: %v", err)
		}

		// Create a file inside it
		testFile := filepath.Join(restrictedDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make the directory read-only
		if err := os.Chmod(restrictedDir, 0444); err != nil {
			t.Fatalf("Failed to change directory permissions: %v", err)
		}

		// Cleanup: restore permissions so temp dir can be cleaned up
		defer func() {
			os.Chmod(restrictedDir, 0755)
		}()

		err := sm.CleanSourceDirectory()
		// This should fail due to permission issues
		if err == nil {
			t.Error("Expected CleanSourceDirectory to fail with permission error")
		}
	})
}

func TestSourceManager_BookExists(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("book does not exist", func(t *testing.T) {
		exists := sm.BookExists("Nonexistent Book")
		if exists {
			t.Error("BookExists should return false for nonexistent book")
		}
	})

	t.Run("book exists", func(t *testing.T) {
		bookTitle := "Existing Book"
		bookDir := filepath.Join(tempDir, bookTitle)
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			t.Fatalf("Failed to create book directory: %v", err)
		}

		exists := sm.BookExists(bookTitle)
		if !exists {
			t.Error("BookExists should return true for existing book")
		}
	})

	t.Run("book with sanitized title exists", func(t *testing.T) {
		bookTitle := "Book/With\\Special:Chars"
		sanitizedTitle := "Book-With-Special-Chars"
		bookDir := filepath.Join(tempDir, sanitizedTitle)
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			t.Fatalf("Failed to create book directory: %v", err)
		}

		exists := sm.BookExists(bookTitle)
		if !exists {
			t.Error("BookExists should return true for book with sanitized title")
		}
	})
}

func TestSourceManager_GetBookInfo(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("empty directory", func(t *testing.T) {
		info, err := sm.GetBookInfo()
		if err != nil {
			t.Fatalf("GetBookInfo failed: %v", err)
		}

		if len(info) != 0 {
			t.Errorf("Expected 0 book info entries, got %d", len(info))
		}
	})

	t.Run("directory with books", func(t *testing.T) {
		// Save some test books
		books := []*client.BookData{
			createMockBookData("1001", "First Book"),
			createMockBookData("1002", "Second Book"),
		}

		err := sm.SaveMultipleBooks(books)
		if err != nil {
			t.Fatalf("Failed to save test books: %v", err)
		}

		info, err := sm.GetBookInfo()
		if err != nil {
			t.Fatalf("GetBookInfo failed: %v", err)
		}

		if len(info) != 2 {
			t.Errorf("Expected 2 book info entries, got %d", len(info))
		}

		// Check that all expected fields are present
		for _, bookInfo := range info {
			requiredFields := []string{"name", "path", "id_file"}
			for _, field := range requiredFields {
				if _, exists := bookInfo[field]; !exists {
					t.Errorf("Expected book info to contain field %q", field)
				}
			}

			// Check optional fields that should be present for our test data
			if title, exists := bookInfo["title"]; exists {
				if title == "" {
					t.Error("Expected non-empty title in book info")
				}
			}

			if id, exists := bookInfo["id"]; exists {
				if id == "" {
					t.Error("Expected non-empty id in book info")
				}
			}
		}
	})

	t.Run("directory with incomplete books", func(t *testing.T) {
		// Create a book directory without proper JSON files
		incompleteDir := filepath.Join(tempDir, "Incomplete Book")
		if err := os.MkdirAll(incompleteDir, 0755); err != nil {
			t.Fatalf("Failed to create incomplete book directory: %v", err)
		}

		// Create only some files
		if err := os.WriteFile(filepath.Join(incompleteDir, "chapters.json"), []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create chapters file: %v", err)
		}

		info, err := sm.GetBookInfo()
		if err != nil {
			t.Fatalf("GetBookInfo failed: %v", err)
		}

		// Should still include the directory, but without title/id info
		found := false
		for _, bookInfo := range info {
			if bookInfo["name"] == "Incomplete Book" {
				found = true
				if bookInfo["id_file"] != "" {
					t.Error("Expected empty id_file for incomplete book")
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find incomplete book in info list")
		}
	})
}

func TestSourceManager_SanitizeDirectoryName(t *testing.T) {
	sm := NewSourceManager("test")

	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "Simple Title"},
		{"Title/With/Slashes", "Title-With-Slashes"},
		{"Title\\With\\Backslashes", "Title-With-Backslashes"},
		{"Title:With:Colons", "Title-With-Colons"},
		{"Title*With*Asterisks", "Title-With-Asterisks"},
		{"Title?With?Questions", "Title-With-Questions"},
		{"Title\"With\"Quotes", "Title-With-Quotes"},
		{"Title<With>Brackets", "Title-With-Brackets"},
		{"Title|With|Pipes", "Title-With-Pipes"},
		{"Title\nWith\nNewlines", "Title With Newlines"},
		{"Title\rWith\rCarriageReturns", "Title With CarriageReturns"},
		{"  Title  With  Multiple  Spaces  ", "Title With Multiple Spaces"},
		{"Title\n\r\t   With\n\r\t   Mixed\n\r\t   Whitespace", "Title \t With \t Mixed \t Whitespace"},
		{"", ""},
		{"   ", ""},
		{"All/\\:*?\"<>|\n\r_Bad", "All--------- _Bad"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sanitize_%q", tt.input), func(t *testing.T) {
			result := sm.sanitizeDirectoryName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSourceManager_SaveJSONFile(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	t.Run("save valid JSON data", func(t *testing.T) {
		testData := map[string]any{
			"title":       "Test Book",
			"description": "This is a test",
			"chapters":    []string{"ch1", "ch2"},
			"metadata": map[string]any{
				"count":  42,
				"active": true,
			},
		}

		testFile := filepath.Join(tempDir, "test.json")
		err := sm.saveJSONFile(testFile, testData)
		if err != nil {
			t.Fatalf("saveJSONFile failed: %v", err)
		}

		// Verify file exists and has correct content
		verifyJSONFile(t, testFile, testData)

		// Verify it's pretty-printed (has indentation)
		data, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "  ") {
			t.Error("Expected JSON to be pretty-printed with indentation")
		}
	})

	t.Run("save invalid JSON data", func(t *testing.T) {
		// Channel types cannot be marshaled to JSON
		invalidData := map[string]any{
			"valid":   "data",
			"invalid": make(chan int),
		}

		testFile := filepath.Join(tempDir, "invalid.json")
		err := sm.saveJSONFile(testFile, invalidData)
		if err == nil {
			t.Error("Expected saveJSONFile to fail with invalid JSON data")
		}

		if !strings.Contains(err.Error(), "failed to marshal JSON data") {
			t.Errorf("Expected JSON marshal error, got: %v", err)
		}

		// Verify file was not created
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("Expected file not to be created on JSON marshal error")
		}
	})

	t.Run("save to read-only directory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.MkdirAll(readOnlyDir, 0444); err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		// Cleanup: restore permissions so temp dir can be cleaned up
		defer func() {
			os.Chmod(readOnlyDir, 0755)
		}()

		testData := map[string]any{"test": "data"}
		testFile := filepath.Join(readOnlyDir, "test.json")

		err := sm.saveJSONFile(testFile, testData)
		if err == nil {
			t.Error("Expected saveJSONFile to fail with read-only directory")
		}

		if !strings.Contains(err.Error(), "failed to write file") {
			t.Errorf("Expected file write error, got: %v", err)
		}
	})
}

// Integration tests
func TestSourceManager_Integration(t *testing.T) {
	tempDir := createTempDir(t)
	sm := NewSourceManager(tempDir)

	// Test complete workflow
	t.Run("complete workflow", func(t *testing.T) {
		// 1. Start with empty directory
		books, err := sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("Initial ListExistingBooks failed: %v", err)
		}
		if len(books) != 0 {
			t.Errorf("Expected empty directory, found %d books", len(books))
		}

		// 2. Save multiple books
		testBooks := []*client.BookData{
			createMockBookData("1001", "Go Programming"),
			createMockBookData("1002", "Rust Programming"),
			createMockBookData("1003", "Python Programming"),
		}

		err = sm.SaveMultipleBooks(testBooks)
		if err != nil {
			t.Fatalf("SaveMultipleBooks failed: %v", err)
		}

		// 3. Verify books exist
		for _, book := range testBooks {
			if !sm.BookExists(book.Title) {
				t.Errorf("Expected book %q to exist", book.Title)
			}
		}

		// 4. List books again
		books, err = sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("ListExistingBooks after save failed: %v", err)
		}
		if len(books) != 3 {
			t.Errorf("Expected 3 books after save, found %d", len(books))
		}

		// 5. Get book info
		bookInfo, err := sm.GetBookInfo()
		if err != nil {
			t.Fatalf("GetBookInfo failed: %v", err)
		}
		if len(bookInfo) != 3 {
			t.Errorf("Expected 3 book info entries, got %d", len(bookInfo))
		}

		// 6. Clean directory
		err = sm.CleanSourceDirectory()
		if err != nil {
			t.Fatalf("CleanSourceDirectory failed: %v", err)
		}

		// 7. Verify directory is empty
		books, err = sm.ListExistingBooks()
		if err != nil {
			t.Fatalf("Final ListExistingBooks failed: %v", err)
		}
		if len(books) != 0 {
			t.Errorf("Expected empty directory after clean, found %d books", len(books))
		}

		// 8. Verify books no longer exist
		for _, book := range testBooks {
			if sm.BookExists(book.Title) {
				t.Errorf("Expected book %q to not exist after clean", book.Title)
			}
		}
	})
}

// Benchmark tests
func BenchmarkSourceManager_SaveBookData(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "benchmark-*")
	defer os.RemoveAll(tempDir)

	sm := NewSourceManager(tempDir)
	bookData := createMockBookData("bench1", "Benchmark Book")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different book ID each time to avoid conflicts
		bookData.ID = fmt.Sprintf("bench%d", i)
		bookData.Title = fmt.Sprintf("Benchmark Book %d", i)

		err := sm.SaveBookData(bookData)
		if err != nil {
			b.Fatalf("SaveBookData failed: %v", err)
		}
	}
}

func BenchmarkSourceManager_ListExistingBooks(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "benchmark-*")
	defer os.RemoveAll(tempDir)

	sm := NewSourceManager(tempDir)

	// Create some books first
	for i := 0; i < 100; i++ {
		bookData := createMockBookData(fmt.Sprintf("book%d", i), fmt.Sprintf("Book %d", i))
		sm.SaveBookData(bookData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sm.ListExistingBooks()
		if err != nil {
			b.Fatalf("ListExistingBooks failed: %v", err)
		}
	}
}

func BenchmarkSourceManager_SanitizeDirectoryName(b *testing.B) {
	sm := NewSourceManager("test")
	testTitle := "Complex/Book\\Title:With*Many?Special\"Characters<>|And\nNewlines"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.sanitizeDirectoryName(testTitle)
	}
}
