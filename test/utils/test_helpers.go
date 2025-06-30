package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
)

// GetTestDataPath returns the absolute path to test data directory
func GetTestDataPath() string {
	// Get the current working directory and navigate to test fixtures
	wd, _ := os.Getwd()
	// Navigate up to project root if we're in a subdirectory
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			// Reached filesystem root without finding go.mod
			break
		}
		wd = parent
	}
	return filepath.Join(wd, "test", "fixtures")
}

// GetValidBookPath returns path to a valid test book
func GetValidBookPath(bookName string) string {
	return filepath.Join(GetTestDataPath(), "valid_books", bookName)
}

// GetInvalidDataPath returns path to invalid test data
func GetInvalidDataPath(dataType string) string {
	return filepath.Join(GetTestDataPath(), "invalid_data", dataType)
}

// LoadTestBook loads a complete test book from fixtures
func LoadTestBook(t *testing.T, bookName string) *models.Book {
	bookPath := GetValidBookPath(bookName)
	
	// Find metadata file
	files, err := os.ReadDir(bookPath)
	if err != nil {
		t.Fatalf("Failed to read book directory: %v", err)
	}
	
	var metadataFile string
	for _, file := range files {
		name := file.Name()
		if filepath.Ext(name) == ".json" && 
		   name != "chapters.json" && 
		   name != "content.json" && 
		   name != "list-notes.json" {
			metadataFile = filepath.Join(bookPath, name)
			break
		}
	}
	
	if metadataFile == "" {
		t.Fatalf("No metadata file found in %s", bookPath)
	}
	
	book := &models.Book{}
	
	// Load metadata
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}
	if err := json.Unmarshal(data, book); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}
	
	// Load chapters
	chaptersPath := filepath.Join(bookPath, "chapters.json")
	if data, err := os.ReadFile(chaptersPath); err == nil {
		if err := json.Unmarshal(data, &book.Chapters); err != nil {
			t.Fatalf("Failed to unmarshal chapters: %v", err)
		}
	}
	
	// Load content
	contentPath := filepath.Join(bookPath, "content.json")
	if data, err := os.ReadFile(contentPath); err == nil {
		if err := json.Unmarshal(data, &book.Content); err != nil {
			t.Fatalf("Failed to unmarshal content: %v", err)
		}
	}
	
	return book
}

// CreateTempDir creates a temporary directory for tests
func CreateTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "slimacademy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	
	return dir
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", filePath)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, filePath string) {
	if _, err := os.Stat(filePath); err == nil {
		t.Errorf("Expected file to not exist: %s", filePath)
	}
}

// ReadFileString reads a file and returns its content as string
func ReadFileString(t *testing.T, filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	return string(data)
}

// AssertStringContains checks if a string contains a substring
func AssertStringContains(t *testing.T, haystack, needle string) {
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected string to contain %q, but it didn't.\nActual: %s", needle, haystack)
	}
}

// AssertStringNotContains checks if a string does not contain a substring
func AssertStringNotContains(t *testing.T, haystack, needle string) {
	if strings.Contains(haystack, needle) {
		t.Errorf("Expected string to not contain %q, but it did.\nActual: %s", needle, haystack)
	}
}