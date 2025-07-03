// Package main implements the SlimAcademy CLI application for document transformation.
// It provides commands for converting books between formats, validation, listing sources,
// and exporting content with ZIP archive support and concurrent processing.
package main

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjanat/slimacademy/internal/client"
	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	"github.com/kjanat/slimacademy/internal/source"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/writers"
)

// main is the entry point for the slim command-line tool, dispatching to convert, check, or list commands based on user input.
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
	case "fetch":
		if err := handleFetch(ctx, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints usage instructions and example commands for the slim document transformation tool.
func printUsage() {
	fmt.Println(`slim - Document transformation tool

Usage:
  slim convert [options] <input>     Convert document(s) to various formats
  slim check <input>                 Check document for issues (sanitizer only)
  slim list [path]                   List available books
  slim fetch [options]               Fetch book data from Slim Academy API

Convert Options:
  --all                             Convert all books in directory to all formats as ZIP to stdout
  --format <format>                 Output format (markdown, html, latex, epub, plaintext)
  --formats <formats>               Multiple formats (comma-separated)
  --output <path>                   Output file/directory path
  --config <path>                   Configuration file path

Fetch Options:
  --login                           Login only (authenticate and save token)
  --all                             Fetch all books from library
  --id <id>                         Fetch specific book by ID
  --output <dir>                    Output directory (default: source)
  --clean                           Clean output directory before fetching

Examples:
  slim convert --all > all-books.zip                    # All books, all formats as ZIP
  slim convert --format markdown book1                  # Single book to markdown
  slim convert --formats "html,epub" book1 --output out # Multiple formats
  slim check book1                                       # Validate book1
  slim list source/                                      # List books in source/
  slim fetch --login                                     # Login and save authentication
  slim fetch --all                                       # Fetch all books to source/
  slim fetch --id 3631                                   # Fetch specific book
  slim fetch --all --output data/                        # Fetch all books to data/`)
}

// handleConvert processes the 'convert' command, converting books to specified formats based on command-line arguments.
// It supports converting a single book or all books in the source directory, delegating to the appropriate handler.
// Returns an error if argument parsing or conversion fails.
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

// handleCheck parses a book at the specified path, runs a sanitizer to detect issues, and prints a summary of any warnings found.
// Returns an error if the input path is missing or the book cannot be parsed.
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

// handleList lists all books found in the specified directory or the default "source" directory.
// It prints the paths of discovered books or a message if none are found.
// Returns an error if the directory cannot be scanned.
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

// parseConvertOptions parses command-line arguments for the convert command and returns a ConvertOptions struct.
// It supports flags for batch conversion, output formats, output path, and configuration file.
// Returns an error if required arguments are missing or invalid options are provided.
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

// convertAllToZip converts all books found in the "source" directory to all supported formats and writes them as separate files in a ZIP archive to standard output.
// Returns an error if no books are found or if the initial search fails; logs warnings for individual book failures but continues processing others.
func convertAllToZip(ctx context.Context, opts *ConvertOptions) (err error) {
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
		if closeErr := zipWriter.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				fmt.Fprintf(os.Stderr, "Error closing ZIP: %v\n", closeErr)
			}
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

// convertBookToZip converts a book to multiple formats and writes each format as a separate file entry in the provided ZIP archive.
// Each file is named using a sanitized version of the book's title and the appropriate file extension.
// Returns an error if conversion or writing to the ZIP archive fails.
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
	for _, result := range results {
		filename := fmt.Sprintf("%s%s", sanitizeFilename(book.Title), result.Extension)

		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		if _, err := writer.Write(result.Data); err != nil {
			return err
		}
	}

	return nil
}

// convertSingle parses a single book and converts it to one or more specified formats, writing the output files to disk.
// If multiple formats are requested, each output file is named with the appropriate extension.
// Returns an error if parsing, conversion, or file writing fails.
func convertSingle(ctx context.Context, opts *ConvertOptions) error {
	// Parse book
	parser := parser.NewBookParser()
	book, err := parser.ParseBook(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to parse book: %w", err)
	}

	// Load configuration if specified
	loader := config.NewLoader()
	appConfig, err := loader.LoadConfig(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// TODO: Pass appConfig to writers when creating them
	// Current writers use defaults, but loaded config is validated and ready
	_ = appConfig // Config is loaded and validated but not yet used by writers

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
	for _, result := range results {
		outputPath := opts.OutputPath
		if outputPath == "" {
			outputPath = fmt.Sprintf("%s%s", sanitizeFilename(book.Title), result.Extension)
		} else {
			info, err := os.Stat(outputPath)
			if err == nil && info.IsDir() {
				// Output path is a directory, create file inside it
				filename := fmt.Sprintf("%s%s", sanitizeFilename(book.Title), result.Extension)
				outputPath = filepath.Join(outputPath, filename)
			} else if len(opts.Formats) > 1 {
				// Multiple formats with a file prefix
				ext := filepath.Ext(outputPath)
				base := strings.TrimSuffix(outputPath, ext)
				outputPath = fmt.Sprintf("%s%s", base, result.Extension)
			}
			// If single format, outputPath is used as is
		}

		if err := os.WriteFile(outputPath, result.Data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}

		fmt.Printf("Written: %s\n", outputPath)
	}

	return nil
}

// sanitizeFilename replaces characters that are invalid or problematic in filenames with hyphens and trims surrounding whitespace.
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

// getExtension returns the file extension corresponding to the given output format name.
// Supported formats include "markdown", "html", "latex", "epub", and "plaintext".
// Returns "txt" for unknown or unsupported formats.
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

type FetchOptions struct {
	LoginOnly bool
	All       bool
	BookID    string
	OutputDir string
	Clean     bool
}

// handleFetch processes the 'fetch' command, handling authentication and book data fetching from Slim Academy API.
// It supports login-only mode, fetching all books, or fetching specific books by ID.
// Returns an error if authentication fails or if fetching operations fail.
func handleFetch(ctx context.Context, args []string) error {
	opts, err := parseFetchOptions(args)
	if err != nil {
		return err
	}

	// Create API client
	apiClient := client.NewSlimClient(opts.OutputDir)

	// Handle login-only mode
	if opts.LoginOnly {
		fmt.Println("Authenticating with Slim Academy...")
		if err := apiClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Println("✓ Successfully authenticated and saved token")
		return nil
	}

	// Create source manager
	sourceManager := source.NewSourceManager(opts.OutputDir)

	// Clean directory if requested
	if opts.Clean {
		fmt.Printf("Cleaning output directory %s...\n", opts.OutputDir)
		if err := sourceManager.CleanSourceDirectory(); err != nil {
			return fmt.Errorf("failed to clean directory: %w", err)
		}
		fmt.Println("✓ Directory cleaned")
	}

	// Ensure authentication
	if err := apiClient.EnsureAuthenticated(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	switch {
	case opts.All:
		// Fetch all books from library
		fmt.Println("Fetching library...")
		books, err := apiClient.FetchLibraryBooks(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch library: %w", err)
		}

		fmt.Printf("Found %d books, saving to %s...\n", len(books), opts.OutputDir)
		if err := sourceManager.SaveMultipleBooks(books); err != nil {
			return fmt.Errorf("failed to save books: %w", err)
		}

		fmt.Printf("✓ Successfully saved %d books\n", len(books))

	case opts.BookID != "":
		// Fetch specific book
		fmt.Printf("Fetching book %s...\n", opts.BookID)
		bookData, err := apiClient.FetchAllBookData(ctx, opts.BookID)
		if err != nil {
			return fmt.Errorf("failed to fetch book %s: %w", opts.BookID, err)
		}

		fmt.Printf("Saving book '%s' to %s...\n", bookData.Title, opts.OutputDir)
		if err := sourceManager.SaveBookData(bookData); err != nil {
			return fmt.Errorf("failed to save book: %w", err)
		}

		fmt.Printf("✓ Successfully saved book '%s'\n", bookData.Title)

	default:
		return fmt.Errorf("specify --login, --all, or --id <id>")
	}

	return nil
}

// parseFetchOptions parses command-line arguments for the fetch command and returns a FetchOptions struct.
// It supports flags for login-only mode, fetching all books, specific book ID, output directory, and cleaning.
// Returns an error if invalid options are provided or required values are missing.
func parseFetchOptions(args []string) (*FetchOptions, error) {
	opts := &FetchOptions{
		OutputDir: "source", // Default output directory
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--login":
			opts.LoginOnly = true
		case "--all":
			opts.All = true
		case "--id":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--id requires a value")
			}
			opts.BookID = args[i+1]
			i++
		case "--output":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			opts.OutputDir = args[i+1]
			i++
		case "--clean":
			opts.Clean = true
		default:
			return nil, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	// Validate options
	if !opts.LoginOnly && !opts.All && opts.BookID == "" {
		return nil, fmt.Errorf("specify --login, --all, or --id <id>")
	}

	if opts.LoginOnly && (opts.All || opts.BookID != "") {
		return nil, fmt.Errorf("--login cannot be combined with other fetch options")
	}

	if opts.All && opts.BookID != "" {
		return nil, fmt.Errorf("--all and --id cannot be used together")
	}

	return opts, nil
}
