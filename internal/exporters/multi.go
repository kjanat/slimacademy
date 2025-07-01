package exporters

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/events"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/writers"
	"github.com/kjanat/slimacademy/pkg/exporters"
)

// ExportFormat represents the supported output formats
type ExportFormat string

const (
	FormatMarkdown  ExportFormat = "markdown"
	FormatHTML      ExportFormat = "html"
	FormatLaTeX     ExportFormat = "latex"
	FormatEPUB      ExportFormat = "epub"
	FormatPlainText ExportFormat = "plaintext"
)

// MultiExporter orchestrates multiple format writers simultaneously
type MultiExporter struct {
	markdownConfig *config.MarkdownConfig
	htmlConfig     *config.HTMLConfig
	latexConfig    *config.LaTeXConfig
	epubConfig     *config.EPUBConfig
}

// NewMultiExporter creates a new multi-format exporter
func NewMultiExporter(cfg *config.MarkdownConfig) *MultiExporter {
	if cfg == nil {
		cfg = config.DefaultMarkdownConfig()
	}
	return &MultiExporter{
		markdownConfig: cfg,
		htmlConfig:     config.DefaultHTMLConfig(),
		latexConfig:    config.DefaultLaTeXConfig(),
		epubConfig:     config.DefaultEPUBConfig(),
	}
}

// ExportFormats exports a book to multiple formats simultaneously
func (e *MultiExporter) ExportFormats(book *models.Book, outputPath string, formats []ExportFormat) error {
	// Create writers for requested formats
	activeWriters := make(map[ExportFormat]writers.Writer)

	for _, format := range formats {
		switch format {
		case FormatMarkdown:
			activeWriters[format] = writers.NewMarkdownWriter(e.markdownConfig)
		case FormatHTML:
			activeWriters[format] = writers.NewHTMLWriterWithConfig(e.htmlConfig)
		case FormatLaTeX:
			activeWriters[format] = writers.NewLaTeXWriter(e.latexConfig)
		case FormatPlainText:
			activeWriters[format] = writers.NewPlainTextWriter()
		case FormatEPUB:
			// EPUB needs special handling for ZIP output
			continue
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	}

	// Handle EPUB separately if requested
	var epubWriter *writers.EPUBWriter
	var epubBuffer *bytes.Buffer
	if contains(formats, FormatEPUB) {
		epubBuffer = &bytes.Buffer{}
		epubWriter = writers.NewEPUBWriterWithConfig(epubBuffer, e.epubConfig)
	}

	// Single pass through the document, driving all writers
	for event := range events.Stream(book) {
		// Drive all active writers
		for _, writer := range activeWriters {
			writer.Handle(event)
		}

		// Drive EPUB writer if active
		if epubWriter != nil {
			epubWriter.Handle(event)
		}
	}

	// Write output files
	baseDir := filepath.Dir(outputPath)
	baseName := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))

	for format, writer := range activeWriters {
		filename := fmt.Sprintf("%s.%s", baseName, getExtension(format))
		fullPath := filepath.Join(baseDir, filename)

		content := writer.Result()
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", format, err)
		}
	}

	// Write EPUB file if requested
	if epubWriter != nil {
		filename := fmt.Sprintf("%s.epub", baseName)
		fullPath := filepath.Join(baseDir, filename)

		if err := os.WriteFile(fullPath, epubBuffer.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write EPUB: %w", err)
		}
	}

	return nil
}

// Export implements the legacy Exporter interface for single format
func (e *MultiExporter) Export(book *models.Book, outputPath string) error {
	// Default to Markdown for legacy compatibility
	return e.ExportFormats(book, outputPath, []ExportFormat{FormatMarkdown})
}

// GetExtension returns the file extension for the default format
func (e *MultiExporter) GetExtension() string {
	return "md"
}

// GetName returns the exporter name
func (e *MultiExporter) GetName() string {
	return "Multi-Format"
}

// getExtension returns the file extension for a format
func getExtension(format ExportFormat) string {
	switch format {
	case FormatMarkdown:
		return "md"
	case FormatHTML:
		return "html"
	case FormatLaTeX:
		return "tex"
	case FormatEPUB:
		return "epub"
	case FormatPlainText:
		return "txt"
	default:
		return "txt"
	}
}

// contains checks if a slice contains a value
func contains(slice []ExportFormat, item ExportFormat) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Legacy wrappers for individual format exporters

// MarkdownExporter provides a legacy-compatible Markdown exporter
type MarkdownExporter struct {
	multi *MultiExporter
}

// NewMarkdownExporter creates a legacy Markdown exporter
func NewMarkdownExporter() exporters.Exporter {
	return &MarkdownExporter{
		multi: NewMultiExporter(nil),
	}
}

// NewMarkdownExporterWithConfig creates a legacy Markdown exporter with config
func NewMarkdownExporterWithConfig(cfg *config.MarkdownConfig) exporters.Exporter {
	return &MarkdownExporter{
		multi: NewMultiExporter(cfg),
	}
}

func (e *MarkdownExporter) Export(book *models.Book, outputPath string) error {
	return e.multi.ExportFormats(book, outputPath, []ExportFormat{FormatMarkdown})
}

func (e *MarkdownExporter) GetExtension() string {
	return "md"
}

func (e *MarkdownExporter) GetName() string {
	return "Markdown"
}

// HTMLExporter provides a legacy-compatible HTML exporter
type HTMLExporter struct {
	multi *MultiExporter
}

// NewHTMLExporter creates a legacy HTML exporter
func NewHTMLExporter() exporters.Exporter {
	return &HTMLExporter{
		multi: NewMultiExporter(nil),
	}
}

func (e *HTMLExporter) Export(book *models.Book, outputPath string) error {
	return e.multi.ExportFormats(book, outputPath, []ExportFormat{FormatHTML})
}

func (e *HTMLExporter) GetExtension() string {
	return "html"
}

func (e *HTMLExporter) GetName() string {
	return "HTML"
}

// EPUBExporter provides a legacy-compatible EPUB exporter
type EPUBExporter struct {
	multi *MultiExporter
}

// NewEPUBExporter creates a legacy EPUB exporter
func NewEPUBExporter() exporters.Exporter {
	return &EPUBExporter{
		multi: NewMultiExporter(nil),
	}
}

func (e *EPUBExporter) Export(book *models.Book, outputPath string) error {
	return e.multi.ExportFormats(book, outputPath, []ExportFormat{FormatEPUB})
}

func (e *EPUBExporter) GetExtension() string {
	return "epub"
}

func (e *EPUBExporter) GetName() string {
	return "EPUB"
}

// PlainTextExporter provides a legacy-compatible PlainText exporter
type PlainTextExporter struct {
	multi *MultiExporter
}

// NewPlainTextExporter creates a legacy PlainText exporter
func NewPlainTextExporter() exporters.Exporter {
	return &PlainTextExporter{
		multi: NewMultiExporter(nil),
	}
}

func (e *PlainTextExporter) Export(book *models.Book, outputPath string) error {
	return e.multi.ExportFormats(book, outputPath, []ExportFormat{FormatPlainText})
}

func (e *PlainTextExporter) GetExtension() string {
	return "txt"
}

func (e *PlainTextExporter) GetName() string {
	return "PlainText"
}
