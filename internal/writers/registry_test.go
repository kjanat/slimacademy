package writers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// MockWriter implements WriterV2 for testing
type MockWriter struct {
	events      []streaming.Event
	stats       WriterStats
	shouldError bool
	errorAfter  int
	mu          sync.Mutex
}

func NewMockWriter() *MockWriter {
	return &MockWriter{
		events: make([]streaming.Event, 0),
	}
}

func (w *MockWriter) Handle(event streaming.Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.shouldError && len(w.events) >= w.errorAfter {
		return fmt.Errorf("mock error after %d events", w.errorAfter)
	}

	w.events = append(w.events, event)
	w.stats.EventsProcessed++

	switch event.Kind {
	case streaming.Text:
		w.stats.TextChars += len(event.TextContent)
	case streaming.Image:
		w.stats.Images++
	case streaming.StartTable:
		w.stats.Tables++
	case streaming.StartHeading:
		w.stats.Headings++
	case streaming.StartList:
		w.stats.Lists++
	}

	return nil
}

func (w *MockWriter) Flush() ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.shouldError {
		return nil, fmt.Errorf("mock flush error")
	}

	var result strings.Builder
	result.WriteString("Mock Output:\n")
	for _, event := range w.events {
		result.WriteString(fmt.Sprintf("Event: %v\n", event.Kind))
	}
	return []byte(result.String()), nil
}

// ContentType returns the MIME type of the output
func (w *MockWriter) ContentType() string {
	return "text/mock"
}

// IsText returns true since this is a text-based mock writer
func (w *MockWriter) IsText() bool {
	return true
}

func (w *MockWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.events = w.events[:0]
	w.stats = WriterStats{}
}

func (w *MockWriter) Stats() WriterStats {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.stats
}

func (w *MockWriter) SetError(shouldError bool, afterEvents int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.shouldError = shouldError
	w.errorAfter = afterEvents
}

func (w *MockWriter) GetEvents() []streaming.Event {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Return a copy to prevent race conditions
	events := make([]streaming.Event, len(w.events))
	copy(events, w.events)
	return events
}

func TestNewWriterRegistry(t *testing.T) {
	registry := NewWriterRegistry()

	if registry == nil {
		t.Error("NewWriterRegistry should return non-nil registry")
	}
	if registry.writers == nil {
		t.Error("Registry should have initialized writers map")
	}
	if registry.metadata == nil {
		t.Error("Registry should have initialized metadata map")
	}
}

func TestWriterRegistry_RegisterAndGet(t *testing.T) {
	// Create a new registry to avoid conflicts with global registry
	testRegistry := NewWriterRegistry()

	// Save original registry and restore after test
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	// Register a mock writer
	factory := func(cfg *config.Config) WriterV2 { return NewMockWriter() }
	metadata := WriterMetadata{
		Name:        "Mock",
		Extension:   ".mock",
		Description: "Mock writer for testing",
		MimeType:    "text/mock",
	}

	Register("mock", factory, metadata)

	// Test Get
	retrievedFactory, exists := Get("mock")
	if !exists {
		t.Error("Registered writer should exist")
	}
	if retrievedFactory == nil {
		t.Error("Retrieved factory should not be nil")
	}

	// Test that factory creates proper writer
	cfg := &config.Config{}
	writer := retrievedFactory(cfg)
	if writer == nil {
		t.Error("Factory should create non-nil writer")
	}
	mockWriter := writer.(*MockWriter) //nolint:gocritic
	if mockWriter == nil {
		t.Error("MockWriter should not be nil")
	}

	// Test non-existent writer
	_, exists = Get("nonexistent")
	if exists {
		t.Error("Non-existent writer should not be found")
	}
}

func TestWriterRegistry_GetMetadata(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	metadata := WriterMetadata{
		Name:        "Test Writer",
		Extension:   ".test",
		Description: "A test writer",
		MimeType:    "text/test",
	}

	Register("test", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, metadata)

	// Test GetMetadata
	retrievedMetadata, exists := GetMetadata("test")
	if !exists {
		t.Error("Metadata for registered writer should exist")
	}
	if retrievedMetadata.Name != metadata.Name {
		t.Errorf("Expected name %q, got %q", metadata.Name, retrievedMetadata.Name)
	}
	if retrievedMetadata.Extension != metadata.Extension {
		t.Errorf("Expected extension %q, got %q", metadata.Extension, retrievedMetadata.Extension)
	}
	if retrievedMetadata.Description != metadata.Description {
		t.Errorf("Expected description %q, got %q", metadata.Description, retrievedMetadata.Description)
	}
	if retrievedMetadata.MimeType != metadata.MimeType {
		t.Errorf("Expected MIME type %q, got %q", metadata.MimeType, retrievedMetadata.MimeType)
	}

	// Test non-existent metadata
	_, exists = GetMetadata("nonexistent")
	if exists {
		t.Error("Metadata for non-existent writer should not be found")
	}
}

func TestWriterRegistry_ListFormats(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	// Register multiple writers
	formats := []string{"format1", "format2", "format3"}
	for _, format := range formats {
		Register(format, func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{
			Name: fmt.Sprintf("Writer %s", format),
		})
	}

	listedFormats := ListFormats()

	if len(listedFormats) != len(formats) {
		t.Errorf("Expected %d formats, got %d", len(formats), len(listedFormats))
	}

	// Check that all registered formats are in the list
	formatMap := make(map[string]bool)
	for _, format := range listedFormats {
		formatMap[format] = true
	}

	for _, format := range formats {
		if !formatMap[format] {
			t.Errorf("Format %q should be in listed formats", format)
		}
	}
}

func TestNewMultiWriter(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	// Register test writers
	Register("writer1", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 1"})
	Register("writer2", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 2"})

	ctx := context.Background()
	cfg := &config.Config{}
	formats := []string{"writer1", "writer2"}

	multiWriter, err := NewMultiWriter(ctx, formats, cfg)
	if err != nil {
		t.Fatalf("NewMultiWriter should not return error: %v", err)
	}
	if multiWriter == nil {
		t.Error("NewMultiWriter should return non-nil MultiWriter")
	}
	if len(multiWriter.writers) != 2 {
		t.Errorf("Expected 2 writers, got %d", len(multiWriter.writers))
	}

	// Test with unsupported format
	_, err = NewMultiWriter(ctx, []string{"unsupported"}, cfg)
	if err == nil {
		t.Error("NewMultiWriter should return error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

func TestMultiWriter_ProcessEvents(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	// Register test writers
	Register("writer1", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 1"})
	Register("writer2", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 2"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"writer1", "writer2"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.Text, TextContent: "Hello, World!"},
		{Kind: streaming.EndDoc},
	}

	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	})

	if err != nil {
		t.Errorf("ProcessEvents should not return error: %v", err)
	}

	// Verify all writers received events
	for format, writer := range multiWriter.writers {
		mockWriter := writer.(*MockWriter) //nolint:gocritic
		receivedEvents := mockWriter.GetEvents()
		if len(receivedEvents) != len(events) {
			t.Errorf("Writer %s should have received %d events, got %d", format, len(events), len(receivedEvents))
		}

		for i, expectedEvent := range events {
			if receivedEvents[i].Kind != expectedEvent.Kind {
				t.Errorf("Writer %s event %d: expected %v, got %v", format, i, expectedEvent.Kind, receivedEvents[i].Kind)
			}
		}
	}
}

func TestMultiWriter_ProcessEvents_WithError(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	// Register test writers, one that will error
	Register("good", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Good Writer"})
	Register("bad", func(cfg *config.Config) WriterV2 {
		mock := NewMockWriter()
		mock.SetError(true, 1) // Error after 1 event
		return mock
	}, WriterMetadata{Name: "Bad Writer"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"good", "bad"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.Text, TextContent: "This should cause error"},
		{Kind: streaming.EndDoc},
	}

	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	})

	if err == nil {
		t.Error("ProcessEvents should return error when writer fails")
	}
	if !strings.Contains(err.Error(), "writer bad failed") {
		t.Errorf("Error should mention failed writer, got: %v", err)
	}
}

func TestMultiWriter_ProcessEvents_ContextCancellation(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("test", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Test Writer"})

	ctx, cancel := context.WithCancel(context.Background())
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"test"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	// Cancel context immediately
	cancel()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.Text, TextContent: "This should be cancelled"},
		{Kind: streaming.EndDoc},
	}

	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	})

	// Should not return error due to cancellation, just stop processing
	if err != nil {
		t.Errorf("ProcessEvents should handle cancellation gracefully: %v", err)
	}
}

func TestMultiWriter_FlushAll(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("format1", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Format 1"})
	Register("format2", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Format 2"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"format1", "format2"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	// Process some events first
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.EndDoc},
	}

	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	})
	if err != nil {
		t.Fatalf("ProcessEvents failed: %v", err)
	}

	// Flush all writers
	results, err := multiWriter.FlushAll()
	if err != nil {
		t.Errorf("FlushAll should not return error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if len(result.Data) == 0 {
			t.Errorf("Result for format %s should not be empty", result.Format)
		}
		if !strings.Contains(string(result.Data), "Mock Output") {
			t.Errorf("Result for format %s should contain mock output", result.Format)
		}
	}
}

func TestMultiWriter_FlushAll_WithError(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("good", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Good Writer"})
	Register("bad", func(cfg *config.Config) WriterV2 {
		mock := NewMockWriter()
		mock.SetError(true, 0) // Error on flush
		return mock
	}, WriterMetadata{Name: "Bad Writer"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"good", "bad"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	_, err = multiWriter.FlushAll()
	if err == nil {
		t.Error("FlushAll should return error when writer fails to flush")
	}
	if !strings.Contains(err.Error(), "flush failed for bad") {
		t.Errorf("Error should mention failed format, got: %v", err)
	}
}

func TestMultiWriter_GetStats(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("writer1", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 1"})
	Register("writer2", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Writer 2"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"writer1", "writer2"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	// Process events to generate stats
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.Text, TextContent: "Hello"},
		{Kind: streaming.Image, ImageURL: "/image.jpg"},
		{Kind: streaming.EndDoc},
	}

	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	})
	if err != nil {
		t.Fatalf("ProcessEvents failed: %v", err)
	}

	stats := multiWriter.GetStats()

	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 writers, got %d", len(stats))
	}

	for format, stat := range stats {
		if stat.EventsProcessed != len(events) {
			t.Errorf("Writer %s should have processed %d events, got %d", format, len(events), stat.EventsProcessed)
		}
		if stat.TextChars != 5 { // "Hello" = 5 chars
			t.Errorf("Writer %s should have 5 text chars, got %d", format, stat.TextChars)
		}
		if stat.Images != 1 {
			t.Errorf("Writer %s should have 1 image, got %d", format, stat.Images)
		}
	}
}

func TestMultiWriter_Close(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("test", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Test Writer"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"test"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}

	// Close should not panic or error
	multiWriter.Close()

	// Context should be cancelled after close
	select {
	case <-multiWriter.ctx.Done():
		// Expected - context should be cancelled
	default:
		t.Error("Context should be cancelled after Close()")
	}
}

// Test concurrent access to registry
func TestWriterRegistry_ConcurrentAccess(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	const numGoroutines = 100
	const numFormats = 10

	var wg sync.WaitGroup

	// Concurrent registration
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			format := fmt.Sprintf("format%d", id%numFormats)
			Register(format, func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{
				Name: fmt.Sprintf("Writer %d", id),
			})
		}(i)
	}

	// Concurrent reading
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			format := fmt.Sprintf("format%d", id%numFormats)
			_, _ = Get(format)
			_, _ = GetMetadata(format)
			_ = ListFormats()
		}(i)
	}

	wg.Wait()

	// Verify final state
	formats := ListFormats()
	if len(formats) != numFormats {
		t.Errorf("Expected %d unique formats, got %d", numFormats, len(formats))
	}
}

// Test MultiWriter with timeout
func TestMultiWriter_ProcessEvents_Timeout(t *testing.T) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("test", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Test Writer"})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"test"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	// Process events with artificial delay
	err = multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		events := []streaming.Event{
			{Kind: streaming.StartDoc, Title: "Test Document"},
			{Kind: streaming.Text, TextContent: "Text 1"},
			{Kind: streaming.Text, TextContent: "Text 2"},
			{Kind: streaming.EndDoc},
		}

		for _, event := range events {
			// Add delay to trigger timeout
			time.Sleep(20 * time.Millisecond)
			if !yield(event) {
				break
			}
		}
	})

	// Should handle timeout gracefully
	if err != nil {
		t.Errorf("ProcessEvents should handle timeout gracefully: %v", err)
	}
}

// Benchmark tests
func BenchmarkWriterRegistry_Get(b *testing.B) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("benchmark", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Benchmark Writer"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Get("benchmark")
	}
}

func BenchmarkMultiWriter_ProcessEvents(b *testing.B) {
	testRegistry := NewWriterRegistry()
	originalRegistry := registry
	registry = testRegistry
	defer func() { registry = originalRegistry }()

	Register("bench1", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Bench Writer 1"})
	Register("bench2", func(cfg *config.Config) WriterV2 { return NewMockWriter() }, WriterMetadata{Name: "Bench Writer 2"})

	ctx := context.Background()
	cfg := &config.Config{}
	multiWriter, err := NewMultiWriter(ctx, []string{"bench1", "bench2"}, cfg)
	if err != nil {
		b.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Benchmark Document"},
		{Kind: streaming.StartParagraph},
		{Kind: streaming.Text, TextContent: "Benchmark text content for performance testing."},
		{Kind: streaming.EndParagraph},
		{Kind: streaming.EndDoc},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset writers for each iteration
		for _, writer := range multiWriter.writers {
			writer.Reset()
		}

		err := multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
			for _, event := range events {
				if !yield(event) {
					break
				}
			}
		})
		if err != nil {
			b.Fatalf("ProcessEvents failed: %v", err)
		}
	}
}
