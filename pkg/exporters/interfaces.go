// Package exporters provides legacy exporter interfaces for document format conversion.
// This package is deprecated in favor of the new streaming-based writers architecture
// in internal/writers. New code should use the WriterV2 interface pattern instead.
package exporters

import "github.com/kjanat/slimacademy/internal/models"

type Exporter interface {
	Export(book *models.Book, outputPath string) error
	GetExtension() string
	GetName() string
}

type ExportFormat string

const (
	FormatMarkdown ExportFormat = "markdown"
	FormatHTML     ExportFormat = "html"
	FormatEPUB     ExportFormat = "epub"
)
