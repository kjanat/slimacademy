package parser

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kjanat/slimacademy/internal/models"
)

type BookParser struct{}

func NewBookParser() *BookParser {
	return &BookParser{}
}

func (p *BookParser) ParseBook(bookDirPath string) (*models.Book, error) {
	book := &models.Book{}

	metadataPath := filepath.Join(bookDirPath, "*.json")
	matches, err := filepath.Glob(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find metadata files: %w", err)
	}

	var metadataFile string
	for _, match := range matches {
		name := filepath.Base(match)
		if name != "chapters.json" && name != "content.json" && name != "list-notes.json" {
			metadataFile = match
			break
		}
	}

	if metadataFile == "" {
		return nil, fmt.Errorf("no metadata file found in %s", bookDirPath)
	}

	if err := p.parseMetadata(metadataFile, book); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	chaptersPath := filepath.Join(bookDirPath, "chapters.json")
	if err := p.parseChapters(chaptersPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse chapters: %w", err)
	}

	contentPath := filepath.Join(bookDirPath, "content.json")
	if err := p.parseContent(contentPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return book, nil
}

func (p *BookParser) parseMetadata(filePath string, book *models.Book) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	if err := json.Unmarshal(data, book); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return nil
}

func (p *BookParser) parseChapters(filePath string, book *models.Book) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to check chapters file: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read chapters file: %w", err)
	}

	var chapters []models.Chapter
	if err := json.Unmarshal(data, &chapters); err != nil {
		return fmt.Errorf("failed to unmarshal chapters: %w", err)
	}

	book.Chapters = chapters
	return nil
}

func (p *BookParser) parseContent(filePath string, book *models.Book) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to check content file: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read content file: %w", err)
	}

	content := &models.Content{}
	if err := json.Unmarshal(data, content); err != nil {
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}

	book.Content = content

	// Build inline object map for image processing
	p.buildInlineObjectMap(book)

	return nil
}

// buildInlineObjectMap creates a map from inline object IDs to image URLs
func (p *BookParser) buildInlineObjectMap(book *models.Book) {
	book.InlineObjectMap = make(map[string]string)

	// First, try to get image URLs from the document's inline objects
	if book.Content != nil && book.Content.Document != nil {
		for objectID, inlineObj := range book.Content.Document.InlineObjects {
			if inlineObj.InlineObjectProperties.EmbeddedObject.ImageProperties != nil {
				imageURL := inlineObj.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentURI
				if imageURL != "" {
					book.InlineObjectMap[objectID] = imageURL
				}
			}
		}
	}

	// Also map from BookImage objects if available
	for _, img := range book.Images {
		if img.ObjectID != "" && img.ImageURL != "" {
			book.InlineObjectMap[img.ObjectID] = img.ImageURL
		}
	}
}

func (p *BookParser) FindAllBooks(rootDir string) ([]string, error) {
	var bookDirs []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		chaptersPath := filepath.Join(path, "chapters.json")
		contentPath := filepath.Join(path, "content.json")

		if _, err := os.Stat(chaptersPath); err == nil {
			if _, err := os.Stat(contentPath); err == nil {
				bookDirs = append(bookDirs, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return bookDirs, nil
}
