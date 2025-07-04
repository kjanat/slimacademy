// Package main implements the SlimAcademy CLI application with Cobra framework.
// It provides commands for converting books between formats, validation, listing sources,
// and exporting content with ZIP archive support and concurrent processing.
package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/writers"
)

// main is the entry point for the slim command-line tool using Cobra framework.
func main() {
	Execute()
}

// convertBookToZip converts a book to multiple formats and writes each format as a separate file entry in the provided ZIP archive.
// Each file is named using a sanitized version of the book's title and the appropriate file extension.
// Returns an error if conversion or writing to the ZIP archive fails.
func convertBookToZip(ctx context.Context, book *models.Book, formats []string, zipWriter *zip.Writer, appConfig *config.Config) error {
	// Create multi-writer with configuration
	multiWriter, err := writers.NewMultiWriter(ctx, formats, appConfig)
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

	// Get results and write to ZIP
	results, err := multiWriter.FlushAll()
	if err != nil {
		return fmt.Errorf("failed to get conversion results: %w", err)
	}

	// Write each format to ZIP
	for _, result := range results {
		// Generate filename
		baseTitle := sanitizeFilename(book.Title)
		filename := fmt.Sprintf("%s%s", baseTitle, result.Extension)

		// Create file in ZIP
		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry %s: %w", filename, err)
		}

		// Write content
		if _, err := fileWriter.Write(result.Data); err != nil {
			return fmt.Errorf("failed to write content to ZIP entry %s: %w", filename, err)
		}
	}

	return nil
}

// sanitizeFilename creates a safe filename from a book title
func sanitizeFilename(title string) string {
	// Replace problematic characters
	result := strings.ReplaceAll(title, " ", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	result = strings.ReplaceAll(result, ":", "_")
	result = strings.ReplaceAll(result, "*", "_")
	result = strings.ReplaceAll(result, "?", "_")
	result = strings.ReplaceAll(result, "\"", "_")
	result = strings.ReplaceAll(result, "<", "_")
	result = strings.ReplaceAll(result, ">", "_")
	result = strings.ReplaceAll(result, "|", "_")

	// Remove multiple consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// Trim underscores from ends
	return strings.Trim(result, "_")
}

// getFormatExtension returns the file extension for a given format
func getFormatExtension(format string) string {
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
		return format
	}
}
