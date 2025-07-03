// Package parser provides JSON parsing functionality for SlimAcademy document content.
// It includes streaming JSON decoders with configurable memory thresholds and
// support for large document processing with efficient memory management.
package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kjanat/slimacademy/internal/models"
)

// BookParser is the legacy parser interface maintained for backward compatibility
type BookParser struct {
	config *ParserConfig
}

// NewBookParser creates a new BookParser with default configuration
func NewBookParser() *BookParser {
	return &BookParser{
		config: DefaultParserConfig(),
	}
}

// ParseBook parses a complete book from a directory
func (p *BookParser) ParseBook(bookDirPath string) (*models.Book, error) {
	book := &models.Book{}
	fmt.Printf("Parsing book in: %s\n", bookDirPath)

	// Check if we should use streaming based on content file size
	contentPath := filepath.Join(bookDirPath, "content.json")
	if info, err := os.Stat(contentPath); err == nil && info.Size() > p.config.StreamingThreshold {
		return p.parseBookStreaming(bookDirPath)
	}

	// Find and parse metadata file
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

	// Parse chapters
	chaptersPath := filepath.Join(bookDirPath, "chapters.json")
	if err := p.parseChapters(chaptersPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse chapters: %w", err)
	}

	// Parse content
	if err := p.parseContent(contentPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return book, nil
}

// parseBookStreaming parses a book using streaming for large content files
func (p *BookParser) parseBookStreaming(bookDirPath string) (*models.Book, error) {
	book := &models.Book{}
	fmt.Printf("Parsing book in streaming mode: %s\n", bookDirPath)

	// Metadata and chapters are usually small, parse them normally
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

	// Parse chapters
	chaptersPath := filepath.Join(bookDirPath, "chapters.json")
	if err := p.parseChapters(chaptersPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse chapters: %w", err)
	}

	// Parse content with streaming
	contentPath := filepath.Join(bookDirPath, "content.json")
	if err := p.parseContentStreaming(contentPath, book); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return book, nil
}

func (p *BookParser) parseMetadata(filePath string, book *models.Book) error {
	fmt.Printf("Parsing metadata: %s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(book); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return nil
}

func (p *BookParser) parseChapters(filePath string, book *models.Book) error {
	fmt.Printf("Parsing chapters: %s\n", filePath)

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Chapters file is optional
		}
		return fmt.Errorf("failed to check chapters file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open chapters file: %w", err)
	}
	defer file.Close()

	var chapters []models.Chapter
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&chapters); err != nil {
		return fmt.Errorf("failed to unmarshal chapters: %w", err)
	}

	book.Chapters = chapters
	return nil
}

func (p *BookParser) parseContent(filePath string, book *models.Book) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Content file is optional
		}
		return fmt.Errorf("failed to check content file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open content file: %w", err)
	}
	defer file.Close()

	content := &models.Content{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(content); err != nil {
		return fmt.Errorf("failed to unmarshal content: %w", err)
	}

	book.Content = content

	// Build inline object map for image processing
	p.buildInlineObjectMap(book)

	return nil
}

func (p *BookParser) parseContentStreaming(filePath string, book *models.Book) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Content file is optional
		}
		return fmt.Errorf("failed to check content file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open content file: %w", err)
	}
	defer file.Close()

	// For now, use regular parsing with json.Decoder which is more memory efficient
	// than reading the entire file into memory
	content := &models.Content{}
	decoder := json.NewDecoder(file)

	// Don't use DisallowUnknownFields for now as it might be too strict

	if err := decoder.Decode(content); err != nil {
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

// FindAllBooks discovers all book directories in the given root directory
func (p *BookParser) FindAllBooks(rootDir string) ([]string, error) {
	var bookDirs []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Check if this directory contains book files
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

// StreamingContentParser processes large content files element by element
type StreamingContentParser struct {
	decoder *json.Decoder
	handler ContentElementHandler
}

// NewStreamingContentParser creates a new streaming parser
func NewStreamingContentParser(r io.Reader, handler ContentElementHandler) *StreamingContentParser {
	return &StreamingContentParser{
		decoder: json.NewDecoder(r),
		handler: handler,
	}
}

// Parse processes the content stream
func (p *StreamingContentParser) Parse() error {
	// This is a simplified implementation
	// A full implementation would parse the JSON structure element by element
	var data interface{}
	if err := p.decoder.Decode(&data); err != nil {
		return err
	}

	return p.handler.HandleElement(data)
}
