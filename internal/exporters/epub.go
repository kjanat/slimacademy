package exporters

import (
	"fmt"
	"os"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/pkg/exporters"
)

type EPUBExporter struct{}

func NewEPUBExporter() exporters.Exporter {
	return &EPUBExporter{}
}

func (e *EPUBExporter) Export(book *models.Book, outputPath string) error {
	htmlExporter := &HTMLExporter{}
	htmlContent := htmlExporter.generateHTML(book)

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title>%s</title>
    <meta charset="UTF-8"/>
</head>
<body>
%s
</body>
</html>`, book.Title, htmlContent)

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func (e *EPUBExporter) GetExtension() string {
	return "epub.html"
}

func (e *EPUBExporter) GetName() string {
	return "EPUB"
}
