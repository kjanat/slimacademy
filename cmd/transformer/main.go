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
	"github.com/kjanat/slimacademy/pkg/exporters"
)

func main() {
	var (
		inputDir   = flag.String("input", ".", "Input directory containing book folders")
		outputDir  = flag.String("output", "./output", "Output directory for exported files")
		format     = flag.String("format", "markdown", "Export format (markdown, html, epub)")
		book       = flag.String("book", "", "Specific book directory to transform (optional)")
		listBooks  = flag.Bool("list", false, "List all available books")
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

	var exporter exporters.Exporter
	switch *format {
	case "markdown":
		exporter = internalExporters.NewMarkdownExporter()
	case "html":
		exporter = internalExporters.NewHTMLExporter()
	case "epub":
		exporter = internalExporters.NewEPUBExporter()
	default:
		log.Fatalf("Unsupported format: %s", *format)
	}

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

		outputPath := filepath.Join(*outputDir, fmt.Sprintf("%s.%s", 
			sanitizeFilename(book.Title), exporter.GetExtension()))

		if err := exporter.Export(transformedBook, outputPath); err != nil {
			log.Printf("Failed to export book %s: %v", bookDir, err)
			continue
		}

		fmt.Printf("Exported to: %s\n", outputPath)
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