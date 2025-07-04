package main

import (
	"archive/zip"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/writers"
	"github.com/spf13/cobra"
)

var (
	// Convert command flags
	convertAll    bool
	outputFormats []string
	outputPath    string
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [input]",
	Short: "Convert document(s) to various formats",
	Long: `Convert SlimAcademy books to various output formats.

Supports converting single books or batch processing all books in a directory.
Output formats include markdown, HTML, LaTeX, EPUB, and plain text.

Examples:
  slim convert book1                                    # Convert to markdown (default)
  slim convert --format html book1                     # Convert to HTML
  slim convert --formats html,epub book1               # Convert to multiple formats
  slim convert --all > all-books.zip                   # All books as ZIP archive
  slim convert book1 --output /tmp/output.md           # Specify output path
  slim convert --config config.yaml book1              # Use custom configuration`,

	Args: func(cmd *cobra.Command, args []string) error {
		if convertAll {
			return nil // --all doesn't require input argument
		}
		if len(args) < 1 {
			return fmt.Errorf("input path required when not using --all")
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if convertAll {
			return runConvertAll(ctx)
		}

		return runConvertSingle(ctx, args[0])
	},
}

func runConvertAll(ctx context.Context) error {
	logger := slog.Default().With("command", "convert-all")
	logger.Info("Starting batch conversion", "formats", outputFormats)

	// Load configuration
	loader := config.NewLoader()
	appConfig, err := loader.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find all books
	parser := parser.NewBookParser()
	books, err := parser.FindAllBooks("source")
	if err != nil {
		return fmt.Errorf("failed to find books: %w", err)
	}

	if len(books) == 0 {
		return fmt.Errorf("no books found")
	}

	logger.Info("Found books for conversion", "count", len(books))

	// Create ZIP writer to stdout
	zipWriter := zip.NewWriter(os.Stdout)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			logger.Error("Error closing ZIP writer", "error", closeErr)
		}
	}()

	// Process each book
	for _, bookPath := range books {
		book, err := parser.ParseBook(bookPath)
		if err != nil {
			logger.Warn("Failed to parse book", "path", bookPath, "error", err)
			continue
		}

		if err := convertBookToZip(ctx, book, outputFormats, zipWriter, appConfig); err != nil {
			logger.Warn("Failed to convert book", "title", book.Title, "error", err)
			continue
		}

		logger.Debug("Book converted successfully", "title", book.Title)
	}

	return nil
}

func runConvertSingle(ctx context.Context, inputPath string) error {
	logger := slog.Default().With("command", "convert-single", "input", inputPath)
	logger.Info("Starting single book conversion", "formats", outputFormats, "output", outputPath)

	// Parse book
	parser := parser.NewBookParser()
	book, err := parser.ParseBook(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse book: %w", err)
	}

	logger.Debug("Book parsed successfully", "title", book.Title, "chapters", len(book.Chapters))

	// Load configuration
	loader := config.NewLoader()
	appConfig, err := loader.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create multi-writer with configuration
	multiWriter, err := writers.NewMultiWriter(ctx, outputFormats, appConfig)
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
		return fmt.Errorf("failed to process events: %w", err)
	}

	// Write output files
	results, err := multiWriter.FlushAll()
	if err != nil {
		return fmt.Errorf("failed to get conversion results: %w", err)
	}

	for _, result := range results {
		filename := outputPath
		if filename == "" {
			// Generate filename based on book title and format
			filename = fmt.Sprintf("%s%s", sanitizeFilename(book.Title), result.Extension)
		}

		if err := os.WriteFile(filename, result.Data, 0644); err != nil {
			return fmt.Errorf("failed to write %s file: %w", result.Format, err)
		}

		logger.Info("Output file written", "format", result.Format, "filename", filename, "size_bytes", len(result.Data))
	}

	return nil
}

func getExtension(format string) string {
	return getFormatExtension(format)
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Convert-specific flags
	convertCmd.Flags().BoolVar(&convertAll, "all", false, "Convert all books in directory to all formats as ZIP to stdout")
	convertCmd.Flags().StringSliceVarP(&outputFormats, "formats", "f", []string{"markdown"}, "Output formats (markdown,html,latex,epub,plaintext)")
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file/directory path")

	// Deprecated --format flag for backwards compatibility
	convertCmd.Flags().String("format", "", "Single output format (deprecated, use --formats)")
	convertCmd.Flags().MarkDeprecated("format", "use --formats instead")
}
