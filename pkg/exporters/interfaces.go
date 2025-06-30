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