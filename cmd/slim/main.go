package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/writers"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "convert":
		if err := handleConvert(ctx, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := handleCheck(ctx, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := handleList(ctx, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`slim - Document transformation tool

Usage:
  slim convert [options] <input>     Convert document(s) to various formats
  slim check <input>                 Check document for issues (sanitizer only)
  slim list [path]                   List available books

Convert Options:
  --all                             Convert all books in directory to all formats as ZIP to stdout
  --format <format>                 Output format (markdown, html, latex, epub, plaintext)
  --formats <formats>               Multiple formats (comma-separated)
  --output <path>                   Output file/directory path
  --config <path>                   Configuration file path

Examples:
  slim convert --all > all-books.zip                    # All books, all formats as ZIP
  slim convert --format markdown book1                  # Single book to markdown
  slim convert --formats "html,epub" book1 --output out # Multiple formats
  slim check book1                                       # Validate book1
  slim list source/                                      # List books in source/`)
}

func handleConvert(ctx context.Context, args []string) error {
	opts, err := parseConvertOptions(args)
	if err != nil {
		return err
	}

	if opts.All {
		return convertAllToZip(ctx, opts)
	}

	return convertSingle(ctx, opts)
}

func handleCheck(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("check command requires input path")
	}

	bookPath := args[0]

	// Parse the book
	bookParser := parser.NewBookParser()
	book, err := bookParser.ParseBook(bookPath)
	if err != nil {
		return fmt.Errorf("failed to parse book: %w", err)
	}

	// Run sanitizer
	sanitizer := sanitizer.NewSanitizer()
	result := sanitizer.Sanitize(book)

	// Report results
	if len(result.Warnings) == 0 {
		fmt.Println("✓ No issues found")
		return nil
	}

	fmt.Printf("Found %d issues:\n\n", len(result.Warnings))
	for i, warning := range result.Warnings {
		fmt.Printf("%d. %s: %s\n", i+1, warning.Location, warning.Issue)
		if warning.Original != warning.Fixed {
			fmt.Printf("   Original: %q\n", warning.Original)
			fmt.Printf("   Fixed:    %q\n", warning.Fixed)
		}
		fmt.Println()
	}

	return nil
}

func handleList(ctx context.Context, args []string) error {
	rootDir := "source"
	if len(args) > 0 {
		rootDir = args[0]
	}

	parser := parser.NewBookParser()
	books, err := parser.FindAllBooks(rootDir)
	if err != nil {
		return fmt.Errorf("failed to find books: %w", err)
	}

	if len(books) == 0 {
		fmt.Println("No books found")
		return nil
	}

	fmt.Printf("Found %d books:\n", len(books))
	for _, book := range books {
		fmt.Printf("  %s\n", book)
	}

	return nil
}

type ConvertOptions struct {
	All        bool
	Formats    []string
	InputPath  string
	OutputPath string
	ConfigPath string
}

func parseConvertOptions(args []string) (*ConvertOptions, error) {
	opts := &ConvertOptions{
		Formats: []string{"markdown"}, // Default format
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			opts.All = true
		case "--format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--format requires a value")
			}
			opts.Formats = []string{args[i+1]}
			i++
		case "--formats":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--formats requires a value")
			}
			opts.Formats = strings.Split(args[i+1], ",")
			i++
		case "--output":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			opts.OutputPath = args[i+1]
			i++
		case "--config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--config requires a value")
			}
			opts.ConfigPath = args[i+1]
			i++
		default:
			if !strings.HasPrefix(args[i], "--") {
				opts.InputPath = args[i]
			} else {
				return nil, fmt.Errorf("unknown option: %s", args[i])
			}
		}
	}

	if !opts.All && opts.InputPath == "" {
		return nil, fmt.Errorf("input path is required (or use --all)")
	}

	return opts, nil
}

func convertAllToZip(ctx context.Context, opts *ConvertOptions) error {
	// Find all books
	parser := parser.NewBookParser()
	books, err := parser.FindAllBooks("source")
	if err != nil {
		return fmt.Errorf("failed to find books: %w", err)
	}

	if len(books) == 0 {
		return fmt.Errorf("no books found")
	}

	// Create ZIP writer to stdout
	zipWriter := zip.NewWriter(os.Stdout)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing ZIP: %v\n", err)
		}
	}()

	// All supported formats for --all
	allFormats := []string{"markdown", "html", "latex", "epub", "plaintext"}

	// Process each book
	for _, bookPath := range books {
		book, err := parser.ParseBook(bookPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", bookPath, err)
			continue
		}

		// Convert to all formats
		if err := convertBookToZip(ctx, book, allFormats, zipWriter); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to convert %s: %v\n", book.Title, err)
			continue
		}
	}

	return nil
}

func convertBookToZip(ctx context.Context, book *models.Book, formats []string, zipWriter *zip.Writer) error {
	// Create multi-writer
	multiWriter, err := writers.NewMultiWriter(ctx, formats)
	if err != nil {
		return err
	}
	defer multiWriter.Close()

	// Create streamer
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	// Process events
	if err := multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for event := range streamer.Stream(ctx, book) {
			if !yield(event) {
				break
			}
		}
	}); err != nil {
		return err
	}

	// Get results
	results, err := multiWriter.FlushAll()
	if err != nil {
		return err
	}

	// Add files to ZIP
	for format, content := range results {
		filename := fmt.Sprintf("%s.%s", sanitizeFilename(book.Title), getExtension(format))

		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(writer, content); err != nil {
			return err
		}
	}

	return nil
}

func convertSingle(ctx context.Context, opts *ConvertOptions) error {
	// Parse book
	parser := parser.NewBookParser()
	book, err := parser.ParseBook(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to parse book: %w", err)
	}

	// Load configuration if specified
	validator := config.NewValidator()
	if opts.ConfigPath != "" {
		// Load and validate config based on format
		// Implementation would depend on the specific format
		_ = validator // Use validator here
	}

	// Create multi-writer
	multiWriter, err := writers.NewMultiWriter(ctx, opts.Formats)
	if err != nil {
		return err
	}
	defer multiWriter.Close()

	// Create streamer
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	// Process events
	if err := multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for event := range streamer.Stream(ctx, book) {
			if !yield(event) {
				break
			}
		}
	}); err != nil {
		return err
	}

	// Get results and write files
	results, err := multiWriter.FlushAll()
	if err != nil {
		return err
	}

	// Write output files
	for format, content := range results {
		outputPath := opts.OutputPath
		if outputPath == "" {
			outputPath = fmt.Sprintf("%s.%s", sanitizeFilename(book.Title), getExtension(format))
		} else if len(opts.Formats) > 1 {
			// Multiple formats: create files with format suffix
			ext := filepath.Ext(outputPath)
			base := strings.TrimSuffix(outputPath, ext)
			outputPath = fmt.Sprintf("%s.%s", base, getExtension(format))
		}

		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}

		fmt.Printf("Written: %s\n", outputPath)
	}

	return nil
}

func sanitizeFilename(name string) string {
	// Replace problematic characters for filenames
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "*", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, "<", "-")
	name = strings.ReplaceAll(name, ">", "-")
	name = strings.ReplaceAll(name, "|", "-")
	return strings.TrimSpace(name)
}

func getExtension(format string) string {
	switch format {
	case "markdown":
		return "md"
	case "html":
		return "html"
	case "latex":
		return "tex"
	case "epub":
		return "epub"
	case "plaintext":
		return "txt"
	default:
		return "txt"
	}
}
