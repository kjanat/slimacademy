package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	internalExporters "github.com/kjanat/slimacademy/internal/exporters"
	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/transformer"
)

func main() {
	var (
		inputDir  = flag.String("input", "./source", "Input directory containing book folders")
		outputDir = flag.String("output", "./output", "Output directory for exported files")
		format    = flag.String("format", "plaintext", "Single export format (markdown, html, latex, epub, plaintext) - deprecated, use --formats")
		formats   = flag.String("formats", "plaintext", "Comma-separated list of export formats (markdown, html, latex, epub, plaintext)")
		book      = flag.String("book", "", "Specific book directory to transform (optional)")
		listBooks = flag.Bool("list", false, "List all available books")
	)
	flag.Parse()

	bookParser := parser.NewBookParser()

	if *listBooks {
		books, err := bookParser.FindAllBooks(*inputDir)
		if err != nil {
			log.Fatalf("Failed to find books: %v", err)
		}

		fmt.Println("Available books:")
		for _, bookDir := range books {
			fmt.Printf("  %s\n", filepath.Base(bookDir))
		}
		return
	}

	var bookDirs []string
	var err error

	if *book != "" {
		bookPath := filepath.Join(*inputDir, *book)
		if _, err := os.Stat(bookPath); os.IsNotExist(err) {
			log.Fatalf("Book directory does not exist: %s", bookPath)
		}
		bookDirs = []string{bookPath}
	} else {
		bookDirs, err = bookParser.FindAllBooks(*inputDir)
		if err != nil {
			log.Fatalf("Failed to find books: %v", err)
		}
	}

	if len(bookDirs) == 0 {
		log.Fatalf("No books found in %s", *inputDir)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	transformer := transformer.NewTransformer()

	// Parse requested formats
	var requestedFormats []internalExporters.ExportFormat
	formatsList := *formats
	if formatsList == "plaintext" && *format != "plaintext" {
		// If formats not specified but format is, use the legacy format flag
		formatsList = *format
	}

	for _, f := range strings.Split(formatsList, ",") {
		f = strings.TrimSpace(f)
		switch f {
		case "markdown", "md":
			requestedFormats = append(requestedFormats, internalExporters.FormatMarkdown)
		case "html":
			requestedFormats = append(requestedFormats, internalExporters.FormatHTML)
		case "latex", "tex":
			requestedFormats = append(requestedFormats, internalExporters.FormatLaTeX)
		case "epub":
			requestedFormats = append(requestedFormats, internalExporters.FormatEPUB)
		case "plaintext", "txt":
			requestedFormats = append(requestedFormats, internalExporters.FormatPlainText)
		default:
			log.Fatalf("Unsupported format: %s", f)
		}
	}

	if len(requestedFormats) == 0 {
		log.Fatalf("No valid formats specified")
	}

	multiExporter := internalExporters.NewMultiExporter(nil)

	for _, bookDir := range bookDirs {
		fmt.Printf("Processing book: %s\n", filepath.Base(bookDir))

		book, err := bookParser.ParseBook(bookDir)
		if err != nil {
			log.Printf("Failed to parse book %s: %v", bookDir, err)
			continue
		}

		transformedBook, err := transformer.Transform(book)
		if err != nil {
			log.Printf("Failed to transform book %s: %v", bookDir, err)
			continue
		}

		// Create base output path (without extension)
		baseOutputPath := filepath.Join(*outputDir, sanitizeFilename(book.Title))

		if err := multiExporter.ExportFormats(transformedBook, baseOutputPath, requestedFormats); err != nil {
			log.Printf("Failed to export book %s: %v", bookDir, err)
			continue
		}

		// Report which files were created
		for _, format := range requestedFormats {
			ext := getExtensionForFormat(format)
			outputPath := fmt.Sprintf("%s.%s", baseOutputPath, ext)
			fmt.Printf("Exported %s to: %s\n", format, outputPath)
		}
	}

	fmt.Printf("Processing complete. Output written to: %s\n", *outputDir)
}

func sanitizeFilename(filename string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

func getExtensionForFormat(format internalExporters.ExportFormat) string {
	switch format {
	case internalExporters.FormatMarkdown:
		return "md"
	case internalExporters.FormatHTML:
		return "html"
	case internalExporters.FormatLaTeX:
		return "tex"
	case internalExporters.FormatEPUB:
		return "epub"
	case internalExporters.FormatPlainText:
		return "txt"
	default:
		return "txt"
	}
}
