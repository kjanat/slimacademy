// Package writers provides a modern streaming-based document format conversion system.
// It implements the WriterV2 interface pattern with auto-registration, concurrent processing,
// and support for multiple output formats including markdown, HTML, EPUB, LaTeX, and plain text.
package writers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// WriterV2 represents the improved writer interface with error handling
type WriterV2 interface {
	// Handle processes a single event and returns an error if processing fails
	Handle(event streaming.Event) error
	// Flush finalizes any pending operations and returns the result
	Flush() ([]byte, error)
	// Reset clears the writer state for reuse
	Reset()
	// Stats returns processing statistics
	Stats() WriterStats
	// ContentType returns the MIME type of the output
	ContentType() string
	// IsText returns true if the output is text-based
	IsText() bool
}

// WriterStats contains processing statistics for observability
type WriterStats struct {
	EventsProcessed  int
	TextChars        int
	Images           int
	Tables           int
	Headings         int
	Lists            int
	Errors           int
	ProcessingTimeMs int64 // Processing time in milliseconds
	MemoryUsageBytes int64 // Peak memory usage in bytes
	StartTime        int64 // Unix timestamp when processing started
	EndTime          int64 // Unix timestamp when processing ended
}

// WriterFactory creates new writer instances with configuration
type WriterFactory func(cfg *config.Config) WriterV2

// WriterRegistry manages format writers with automatic registration
type WriterRegistry struct {
	mu       sync.RWMutex
	writers  map[string]WriterFactory
	metadata map[string]WriterMetadata
}

// WriterMetadata contains information about a writer format
type WriterMetadata struct {
	Name        string
	Extension   string
	Description string
	MimeType    string
	IsBinary    bool // Indicates if the output is binary data
}

var (
	// Global registry instance
	registry = NewWriterRegistry()
)

// NewWriterRegistry returns a new, empty WriterRegistry for managing writer factories and their metadata.
func NewWriterRegistry() *WriterRegistry {
	return &WriterRegistry{
		writers:  make(map[string]WriterFactory),
		metadata: make(map[string]WriterMetadata),
	}
}

// Register adds a writer factory and its metadata to the global registry under the specified format key.
func Register(format string, factory WriterFactory, metadata WriterMetadata) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.writers[format] = factory
	registry.metadata[format] = metadata
}

// Get returns the writer factory associated with the given format, along with a boolean indicating whether the format is registered.
func Get(format string) (WriterFactory, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	factory, exists := registry.writers[format]
	return factory, exists
}

// GetMetadata returns the metadata associated with the specified writer format.
// It returns the metadata and a boolean indicating whether the format exists in the registry.
func GetMetadata(format string) (WriterMetadata, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	metadata, exists := registry.metadata[format]
	return metadata, exists
}

// ListFormats returns a slice of all registered writer format names.
func ListFormats() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	formats := make([]string, 0, len(registry.writers))
	for format := range registry.writers {
		formats = append(formats, format)
	}
	return formats
}

// MultiWriter processes events through multiple writers concurrently with error handling
type MultiWriter struct {
	writers map[string]WriterV2
	ctx     context.Context
	cancel  context.CancelFunc
	errCh   chan error
}

// NewMultiWriter creates a MultiWriter that manages multiple WriterV2 instances for the specified formats, using the provided context for lifecycle and cancellation control.
// Returns an error if any requested format is not registered.
func NewMultiWriter(ctx context.Context, formats []string, cfg *config.Config) (*MultiWriter, error) {
	writers := make(map[string]WriterV2)

	// Create writer instances for each format
	for _, format := range formats {
		factory, exists := Get(format)
		if !exists {
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
		writers[format] = factory(cfg)
	}

	writerCtx, cancel := context.WithCancel(ctx)

	return &MultiWriter{
		writers: writers,
		ctx:     writerCtx,
		cancel:  cancel,
		errCh:   make(chan error, len(writers)),
	}, nil
}

// ProcessEvents drives all writers through the event stream with error propagation
func (mw *MultiWriter) ProcessEvents(eventStream func(yield func(streaming.Event) bool)) error {
	defer mw.cancel()

	// Get logger from context if available
	logger := slog.Default()
	if ctxLogger, ok := mw.ctx.Value("logger").(*slog.Logger); ok {
		logger = ctxLogger
	}

	startTime := time.Now()
	eventCount := 0

	// Process events through all writers
	eventStream(func(event streaming.Event) bool {
		eventCount++

		// Check if context was cancelled
		select {
		case <-mw.ctx.Done():
			logger.Debug("Event processing cancelled", "events_processed", eventCount)
			return false // Stop iteration
		default:
		}

		// Process event through all writers
		for format, writer := range mw.writers {
			if err := writer.Handle(event); err != nil {
				logger.Error("Writer failed processing event",
					"format", format,
					"event_kind", event.Kind.String(),
					"error", err,
					"events_processed", eventCount)

				// Send error and cancel all operations
				select {
				case mw.errCh <- fmt.Errorf("writer %s failed: %w", format, err):
				default:
				}
				mw.cancel()
				return false // Stop iteration
			}
		}

		// Log progress for large documents
		if eventCount%1000 == 0 {
			logger.Debug("Processing progress",
				"events_processed", eventCount,
				"elapsed_ms", time.Since(startTime).Milliseconds())
		}

		return true // Continue iteration
	})

	processingTime := time.Since(startTime)
	logger.Info("Event processing completed",
		"total_events", eventCount,
		"processing_time_ms", processingTime.Milliseconds(),
		"formats", len(mw.writers))

	// Check for errors
	select {
	case err := <-mw.errCh:
		return err
	default:
	}

	return nil
}

// OutputResult contains the result of a writer operation
type OutputResult struct {
	Format      string
	Data        []byte
	ContentType string
	IsText      bool
	Extension   string
}

// FlushAll finalizes all writers and returns their results
func (mw *MultiWriter) FlushAll() ([]OutputResult, error) {
	results := make([]OutputResult, 0, len(mw.writers))

	for format, writer := range mw.writers {
		data, err := writer.Flush()
		if err != nil {
			return nil, fmt.Errorf("flush failed for %s: %w", format, err)
		}

		metadata, exists := GetMetadata(format)
		if !exists {
			return nil, fmt.Errorf("metadata not found for format: %s", format)
		}

		results = append(results, OutputResult{
			Format:      format,
			Data:        data,
			ContentType: writer.ContentType(),
			IsText:      writer.IsText(),
			Extension:   metadata.Extension,
		})
	}

	return results, nil
}

// GetStats returns combined statistics from all writers
func (mw *MultiWriter) GetStats() map[string]WriterStats {
	stats := make(map[string]WriterStats)

	for format, writer := range mw.writers {
		stats[format] = writer.Stats()
	}

	return stats
}

// Close cancels all operations and cleans up resources
func (mw *MultiWriter) Close() {
	mw.cancel()
}
