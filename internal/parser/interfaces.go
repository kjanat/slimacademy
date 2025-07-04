package parser

import (
	"io"

	"github.com/kjanat/slimacademy/internal/models"
)

// Parser defines the interface for parsing book data
type Parser interface {
	// ParseBook parses a complete book from a directory
	ParseBook(bookDirPath string) (*models.Book, error)

	// ParseBookStreaming parses a book using streaming for large files
	ParseBookStreaming(bookDirPath string) (*models.Book, error)
}

// FileParser defines the interface for parsing individual files
type FileParser interface {
	// ParseMetadata parses book metadata from a reader
	ParseMetadata(r io.Reader, book *models.Book) error

	// ParseChapters parses chapters from a reader
	ParseChapters(r io.Reader) ([]models.Chapter, error)

	// ParseContent parses content from a reader (supports streaming)
	ParseContent(r io.Reader) (*models.Content, error)

	// ParseContentStreaming parses large content files using streaming
	ParseContentStreaming(r io.Reader, handler ContentHandler) error
}

// ContentHandler processes content elements as they are parsed
type ContentHandler interface {
	// HandleParagraph is called for each paragraph element
	HandleParagraph(para models.Paragraph) error

	// HandleStructuralElement is called for other structural elements
	HandleStructuralElement(elem models.StructuralElement) error

	// Finalize is called after all content has been processed
	Finalize() (*models.Content, error)
}

// ContentElementHandler processes individual content elements (simplified interface)
type ContentElementHandler interface {
	HandleElement(elem interface{}) error
}

// BookFinder defines the interface for discovering books
type BookFinder interface {
	// FindAllBooks discovers all book directories in a root directory
	FindAllBooks(rootDir string) ([]string, error)

	// IsBookDirectory checks if a directory contains book files
	IsBookDirectory(dirPath string) (bool, error)
}

// ImageMapper defines the interface for mapping inline objects to images
type ImageMapper interface {
	// BuildInlineObjectMap creates a mapping from object IDs to image URLs
	BuildInlineObjectMap(book *models.Book) map[string]string
}

// ParserConfig holds configuration for the parser
type ParserConfig struct {
	// StreamingThreshold is the file size in bytes above which streaming is used
	StreamingThreshold int64

	// MaxMemoryBuffer is the maximum memory buffer size for streaming
	MaxMemoryBuffer int

	// ValidateContent enables content validation during parsing
	ValidateContent bool

	// StrictMode fails on any parsing errors (vs best-effort)
	StrictMode bool
}

// DefaultParserConfig returns a default parser configuration
func DefaultParserConfig() *ParserConfig {
	return &ParserConfig{
		StreamingThreshold: 10 * 1024 * 1024, // 10MB
		MaxMemoryBuffer:    1024 * 1024,      // 1MB buffer
		ValidateContent:    true,
		StrictMode:         false,
	}
}
