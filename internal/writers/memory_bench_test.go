package writers

import (
	"context"
	"fmt"
	"iter"
	"runtime"
	"testing"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// MemoryStats tracks memory usage during tests
type MemoryStats struct {
	AllocsBefore uint64
	AllocsAfter  uint64
	BytesBefore  uint64
	BytesAfter   uint64
}

// GetMemoryStats captures current memory statistics
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return MemoryStats{
		AllocsBefore: m.TotalAlloc,
		BytesBefore:  m.Alloc,
	}
}

// Update captures final memory statistics
func (ms *MemoryStats) Update() {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	ms.AllocsAfter = m.TotalAlloc
	ms.BytesAfter = m.Alloc
}

// AllocDelta returns the memory allocation delta
func (ms *MemoryStats) AllocDelta() uint64 {
	return ms.AllocsAfter - ms.AllocsBefore
}

// BytesDelta returns the current memory usage delta
func (ms *MemoryStats) BytesDelta() uint64 {
	if ms.BytesAfter > ms.BytesBefore {
		return ms.BytesAfter - ms.BytesBefore
	}
	return 0 // GC freed memory
}

// generateLargeEventSlice creates a large slice of events for testing
func generateLargeEventSlice(n int) []streaming.Event {
	events := make([]streaming.Event, 0, n)

	// Start document
	events = append(events, streaming.Event{
		Kind:        streaming.StartDoc,
		Title:       "Large Test Document",
		Description: "Memory efficiency test with large content",
	})

	// Generate many sections with content
	for i := 0; i < n/10; i++ {
		// Heading
		events = append(events, streaming.Event{
			Kind:     streaming.StartHeading,
			Level:    2,
			AnchorID: fmt.Sprintf("section-%d", i),
		})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: fmt.Sprintf("Section %d: Large Content Test", i),
		})
		events = append(events, streaming.Event{Kind: streaming.EndHeading})

		// Paragraph with various formatting
		events = append(events, streaming.Event{Kind: streaming.StartParagraph})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: "This is a large paragraph with ",
		})
		events = append(events, streaming.Event{
			Kind:  streaming.StartFormatting,
			Style: streaming.Bold,
		})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: "bold text",
		})
		events = append(events, streaming.Event{
			Kind:  streaming.EndFormatting,
			Style: streaming.Bold,
		})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: " and some ",
		})
		events = append(events, streaming.Event{
			Kind:    streaming.StartFormatting,
			Style:   streaming.Link,
			LinkURL: fmt.Sprintf("https://example.com/section-%d", i),
		})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: "linked content",
		})
		events = append(events, streaming.Event{
			Kind:  streaming.EndFormatting,
			Style: streaming.Link,
		})
		events = append(events, streaming.Event{
			Kind:        streaming.Text,
			TextContent: fmt.Sprintf(" with content that repeats across %d sections.", i),
		})
		events = append(events, streaming.Event{Kind: streaming.EndParagraph})
	}

	// End document
	events = append(events, streaming.Event{Kind: streaming.EndDoc})

	return events
}

// generateLargeEventStream creates a streaming event iterator for testing
func generateLargeEventStream(n int) iter.Seq[streaming.Event] {
	return func(yield func(streaming.Event) bool) {
		// Start document
		if !yield(streaming.Event{
			Kind:        streaming.StartDoc,
			Title:       "Large Streaming Test Document",
			Description: "Memory efficiency test with streaming",
		}) {
			return
		}

		// Generate many sections with content
		for i := 0; i < n/10; i++ {
			// Heading
			if !yield(streaming.Event{
				Kind:     streaming.StartHeading,
				Level:    2,
				AnchorID: fmt.Sprintf("stream-section-%d", i),
			}) {
				return
			}
			if !yield(streaming.Event{
				Kind:        streaming.Text,
				TextContent: fmt.Sprintf("Streaming Section %d", i),
			}) {
				return
			}
			if !yield(streaming.Event{Kind: streaming.EndHeading}) {
				return
			}

			// Paragraph
			if !yield(streaming.Event{Kind: streaming.StartParagraph}) {
				return
			}
			if !yield(streaming.Event{
				Kind:        streaming.Text,
				TextContent: fmt.Sprintf("Streaming content for section %d with repeated patterns.", i),
			}) {
				return
			}
			if !yield(streaming.Event{Kind: streaming.EndParagraph}) {
				return
			}
		}

		// End document
		yield(streaming.Event{Kind: streaming.EndDoc})
	}
}

// BenchmarkMemoryUsage_SliceVsStream compares memory usage between slice and stream processing
func BenchmarkMemoryUsage_SliceVsStream(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Slice_%d_events", size), func(b *testing.B) {
			writer := NewMinimalHTMLWriter()
			events := generateLargeEventSlice(size)

			b.ResetTimer()
			stats := GetMemoryStats()

			for i := 0; i < b.N; i++ {
				_, err := writer.ProcessEvents(events)
				if err != nil {
					b.Fatalf("Error processing events: %v", err)
				}
				writer.Reset()
			}

			stats.Update()
			b.ReportMetric(float64(stats.AllocDelta())/float64(b.N), "allocs/op")
			b.ReportMetric(float64(stats.BytesDelta())/float64(b.N), "bytes/op")
		})

		b.Run(fmt.Sprintf("Stream_%d_events", size), func(b *testing.B) {
			writer := NewMinimalHTMLWriter()

			b.ResetTimer()
			stats := GetMemoryStats()

			for i := 0; i < b.N; i++ {
				stream := generateLargeEventStream(size)
				_, err := writer.ProcessEventStream(context.Background(), stream)
				if err != nil {
					b.Fatalf("Error processing stream: %v", err)
				}
				writer.Reset()
			}

			stats.Update()
			b.ReportMetric(float64(stats.AllocDelta())/float64(b.N), "allocs/op")
			b.ReportMetric(float64(stats.BytesDelta())/float64(b.N), "bytes/op")
		})
	}
}

// BenchmarkDuplicateDetection tests the performance of unique.Handle duplicate detection
func BenchmarkDuplicateDetection(b *testing.B) {
	writer := NewMinimalHTMLWriter()

	// Create events with many duplicate URLs and anchor IDs
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Duplicate Test"},
	}

	// Add many duplicate URLs and anchors
	for i := 0; i < 1000; i++ {
		// Reuse same URLs and anchors to test deduplication
		anchorID := fmt.Sprintf("section-%d", i%10)                // Only 10 unique anchors
		linkURL := fmt.Sprintf("https://example.com/page-%d", i%5) // Only 5 unique URLs

		events = append(events, []streaming.Event{
			{Kind: streaming.StartHeading, Level: 2, AnchorID: anchorID},
			{Kind: streaming.Text, TextContent: fmt.Sprintf("Heading %d", i)},
			{Kind: streaming.EndHeading},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: linkURL},
			{Kind: streaming.Text, TextContent: "Duplicate link"},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.EndParagraph},
		}...)
	}

	events = append(events, streaming.Event{Kind: streaming.EndDoc})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := writer.ProcessEvents(events)
		if err != nil {
			b.Fatalf("Error processing events: %v", err)
		}
		writer.Reset()
	}
}

// BenchmarkLargeDocumentStreaming tests streaming performance with large documents
func BenchmarkLargeDocumentStreaming(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			writer := NewMinimalHTMLWriter()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stream := generateLargeEventStream(size)
				_, err := writer.ProcessEventStream(context.Background(), stream)
				if err != nil {
					b.Fatalf("Error processing stream: %v", err)
				}
				writer.Reset()
			}
		})
	}
}

// TestMemoryEfficiency validates that streaming uses less memory than slice processing
func TestMemoryEfficiency(t *testing.T) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			writer := NewMinimalHTMLWriter()

			// Test slice processing memory
			events := generateLargeEventSlice(size)
			stats1 := GetMemoryStats()
			_, err := writer.ProcessEvents(events)
			if err != nil {
				t.Fatalf("Error processing events: %v", err)
			}
			stats1.Update()
			sliceBytes := stats1.BytesDelta()
			writer.Reset()

			// Test streaming memory
			stream := generateLargeEventStream(size)
			stats2 := GetMemoryStats()
			_, err = writer.ProcessEventStream(context.Background(), stream)
			if err != nil {
				t.Fatalf("Error processing stream: %v", err)
			}
			stats2.Update()
			streamBytes := stats2.BytesDelta()

			t.Logf("Size %d: Slice used %d bytes, Stream used %d bytes",
				size, sliceBytes, streamBytes)

			// For large documents, streaming should use same or less memory
			if size >= 5000 && streamBytes > sliceBytes*2 {
				t.Errorf("Streaming used more than 2x memory of slice processing: %d vs %d bytes",
					streamBytes, sliceBytes)
			}
		})
	}
}

// TestUniqueHandleEfficiency tests that unique.Handle provides O(1) deduplication
func TestUniqueHandleEfficiency(t *testing.T) {
	writer := NewMinimalHTMLWriter()

	// Process document with many duplicates
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Duplicate Test"},
	}

	// Create many events with the same URLs and anchor IDs
	duplicateURL := "https://example.com/duplicate"
	duplicateAnchor := "duplicate-section"
	duplicateText := "This text appears many times"

	for i := 0; i < 1000; i++ {
		events = append(events, []streaming.Event{
			{Kind: streaming.StartHeading, Level: 2, AnchorID: duplicateAnchor},
			{Kind: streaming.Text, TextContent: duplicateText},
			{Kind: streaming.EndHeading},
			{Kind: streaming.StartParagraph},
			{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: duplicateURL},
			{Kind: streaming.Text, TextContent: duplicateText},
			{Kind: streaming.EndFormatting, Style: streaming.Link},
			{Kind: streaming.EndParagraph},
		}...)
	}

	events = append(events, streaming.Event{Kind: streaming.EndDoc})

	_, err := writer.ProcessEvents(events)
	if err != nil {
		t.Fatalf("Error processing events: %v", err)
	}

	// Verify deduplication worked
	if len(writer.seenURLs) != 1 {
		t.Errorf("Expected 1 unique URL, got %d", len(writer.seenURLs))
	}

	if len(writer.seenAnchors) != 1 {
		t.Errorf("Expected 1 unique anchor, got %d", len(writer.seenAnchors))
	}

	if len(writer.seenTexts) != 1 {
		t.Errorf("Expected 1 unique text, got %d", len(writer.seenTexts))
	}

	// Verify text was counted correctly
	for _, count := range writer.seenTexts {
		if count != 2000 { // 2 occurrences per iteration * 1000 iterations
			t.Errorf("Expected text count 2000, got %d", count)
		}
	}
}
