// Package source provides source file management and discovery for SlimAcademy books.
// It handles scanning directories, validating book structures, and managing
// book metadata and content file organization.
package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjanat/slimacademy/internal/client"
)

// SourceManager handles saving book data to the source directory
type SourceManager struct {
	sourceDir string
}

// NewSourceManager creates a new source manager
func NewSourceManager(sourceDir string) *SourceManager {
	if sourceDir == "" {
		sourceDir = "source"
	}
	return &SourceManager{sourceDir: sourceDir}
}

// SaveBookData saves a book's data to the source directory in the expected format
func (sm *SourceManager) SaveBookData(bookData *client.BookData) error {
	// Create book directory name from title
	bookDir := sm.sanitizeDirectoryName(bookData.Title)
	if bookDir == "" {
		bookDir = fmt.Sprintf("Book_%s", bookData.ID)
	}

	bookPath := filepath.Join(sm.sourceDir, bookDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(bookPath, 0755); err != nil {
		return fmt.Errorf("failed to create book directory %s: %w", bookPath, err)
	}

	// Save summary data as {id}.json (matching existing format)
	summaryFile := filepath.Join(bookPath, fmt.Sprintf("%s.json", bookData.ID))
	if err := sm.saveJSONFile(summaryFile, bookData.Summary); err != nil {
		return fmt.Errorf("failed to save summary file: %w", err)
	}

	// Save chapters.json
	chaptersFile := filepath.Join(bookPath, "chapters.json")
	if err := sm.saveJSONFile(chaptersFile, bookData.Chapters); err != nil {
		return fmt.Errorf("failed to save chapters file: %w", err)
	}

	// Save content.json
	contentFile := filepath.Join(bookPath, "content.json")
	if err := sm.saveJSONFile(contentFile, bookData.Content); err != nil {
		return fmt.Errorf("failed to save content file: %w", err)
	}

	// Save list-notes.json (matching bash script naming)
	notesFile := filepath.Join(bookPath, "list-notes.json")
	if err := sm.saveJSONFile(notesFile, bookData.Notes); err != nil {
		return fmt.Errorf("failed to save notes file: %w", err)
	}

	return nil
}

// SaveMultipleBooks saves multiple books to the source directory
func (sm *SourceManager) SaveMultipleBooks(books []*client.BookData) error {
	for _, book := range books {
		if err := sm.SaveBookData(book); err != nil {
			return fmt.Errorf("failed to save book %s (%s): %w", book.Title, book.ID, err)
		}
	}
	return nil
}

// saveJSONFile saves data as a pretty-printed JSON file
func (sm *SourceManager) saveJSONFile(filepath string, data any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON data: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	return nil
}

// sanitizeDirectoryName cleans up a title to be used as a directory name
func (sm *SourceManager) sanitizeDirectoryName(title string) string {
	// Replace problematic characters with safe alternatives
	title = strings.ReplaceAll(title, "/", "-")
	title = strings.ReplaceAll(title, "\\", "-")
	title = strings.ReplaceAll(title, ":", "-")
	title = strings.ReplaceAll(title, "*", "-")
	title = strings.ReplaceAll(title, "?", "-")
	title = strings.ReplaceAll(title, "\"", "-")
	title = strings.ReplaceAll(title, "<", "-")
	title = strings.ReplaceAll(title, ">", "-")
	title = strings.ReplaceAll(title, "|", "-")
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")

	// Collapse multiple spaces and trim
	title = strings.TrimSpace(title)
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}

	return title
}

// ListExistingBooks returns a list of books already in the source directory
func (sm *SourceManager) ListExistingBooks() ([]string, error) {
	entries, err := os.ReadDir(sm.sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}

	var books []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			books = append(books, entry.Name())
		}
	}

	return books, nil
}

// CleanSourceDirectory removes all books from the source directory
func (sm *SourceManager) CleanSourceDirectory() error {
	if _, err := os.Stat(sm.sourceDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	entries, err := os.ReadDir(sm.sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(sm.sourceDir, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entryPath, err)
		}
	}

	return nil
}

// BookExists checks if a book with the given title already exists
func (sm *SourceManager) BookExists(title string) bool {
	bookDir := sm.sanitizeDirectoryName(title)
	bookPath := filepath.Join(sm.sourceDir, bookDir)

	if _, err := os.Stat(bookPath); err == nil {
		return true
	}
	return false
}

// GetBookInfo returns basic information about books in the source directory
func (sm *SourceManager) GetBookInfo() ([]map[string]any, error) {
	books, err := sm.ListExistingBooks()
	if err != nil {
		return nil, err
	}

	var bookInfo []map[string]any
	for _, bookName := range books {
		bookPath := filepath.Join(sm.sourceDir, bookName)

		// Try to find the main JSON file (should be numbered like 3631.json)
		entries, err := os.ReadDir(bookPath)
		if err != nil {
			continue
		}

		var mainJSONFile string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") &&
				entry.Name() != "chapters.json" &&
				entry.Name() != "content.json" &&
				entry.Name() != "list-notes.json" {
				mainJSONFile = entry.Name()
				break
			}
		}

		info := map[string]any{
			"name":    bookName,
			"path":    bookPath,
			"id_file": mainJSONFile,
		}

		// Try to read basic info from the main JSON file
		if mainJSONFile != "" {
			mainJSONPath := filepath.Join(bookPath, mainJSONFile)
			if data, err := os.ReadFile(mainJSONPath); err == nil {
				var jsonData map[string]any
				if err := json.Unmarshal(data, &jsonData); err == nil {
					if title, ok := jsonData["title"]; ok {
						info["title"] = title
					}
					if id, ok := jsonData["id"]; ok {
						info["id"] = id
					}
				}
			}
		}

		bookInfo = append(bookInfo, info)
	}

	return bookInfo, nil
}
