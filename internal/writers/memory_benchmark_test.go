package writers

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// BenchmarkMemoryEfficiency benchmarks O(1) memory usage with iter.Seq
func BenchmarkMemoryEfficiency(b *testing.B) {
	// Create test books of varying sizes
	sizes := []int{100, 1000, 10000, 50000} // Number of paragraphs

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d_paragraphs", size), func(b *testing.B) {
			book := createLargeTestBookForBench(size)

			b.ResetTimer()
			b.ReportAllocs()

			// Measure memory before processing
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			for i := 0; i < b.N; i++ {
				// Use streaming with iter.Seq for O(1) memory
				streamer := streaming.NewStreamer(streaming.StreamOptions{
					SanitizeText: true,
					SkipEmpty:    true,
				})

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				eventStream := streamer.Stream(ctx, book)

				writer := NewMinimalHTMLWriter()
				html, err := writer.ProcessEventStream(ctx, eventStream)
				if err != nil {
					b.Fatalf("Failed to process events: %v", err)
				}

				// Verify we got valid output
				if len(html) < 100 {
					b.Fatalf("Output too small: %d bytes", len(html))
				}

				cancel()
			}

			// Measure memory after processing
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Calculate memory growth
			memGrowth := memAfter.HeapAlloc - memBefore.HeapAlloc
			b.Logf("Document size: %d paragraphs", size)
			b.Logf("Memory growth: %d bytes (%.2f MB)", memGrowth, float64(memGrowth)/1024/1024)
			b.Logf("Heap objects growth: %d", memAfter.HeapObjects-memBefore.HeapObjects)
		})
	}
}

// BenchmarkTraditionalVsStreaming compares memory usage
func BenchmarkTraditionalVsStreaming(b *testing.B) {
	book := createLargeTestBookForBench(10000) // 10k paragraphs

	b.Run("Traditional_Slice_Processing", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Traditional approach: load all events into slice
			streamer := streaming.NewStreamer(streaming.StreamOptions{
				SanitizeText: true,
				SkipEmpty:    true,
			})

			ctx := context.Background()
			eventStream := streamer.Stream(ctx, book)

			// Load all events into memory (traditional approach)
			var events []streaming.Event
			for event := range eventStream {
				events = append(events, event)
			}

			// Process all events
			writer := NewMinimalHTMLWriter()
			html, err := writer.ProcessEvents(events)
			if err != nil {
				b.Fatalf("Failed to process events: %v", err)
			}

			if len(html) < 100 {
				b.Fatalf("Output too small: %d bytes", len(html))
			}
		}
	})

	b.Run("Streaming_Iterator_Processing", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Streaming approach: O(1) memory with iter.Seq
			streamer := streaming.NewStreamer(streaming.StreamOptions{
				SanitizeText: true,
				SkipEmpty:    true,
			})

			ctx := context.Background()
			eventStream := streamer.Stream(ctx, book)

			writer := NewMinimalHTMLWriter()
			html, err := writer.ProcessEventStream(ctx, eventStream)
			if err != nil {
				b.Fatalf("Failed to process events: %v", err)
			}

			if len(html) < 100 {
				b.Fatalf("Output too small: %d bytes", len(html))
			}
		}
	})
}

// BenchmarkUniqueHandlePerformance tests unique.Handle[string] performance
func BenchmarkUniqueHandlePerformance(b *testing.B) {
	// Test duplicate detection performance with unique.Handle
	headings := []string{
		"Introduction", "Overview", "Introduction", "Conclusion",
		"Overview", "Methods", "Results", "Introduction", "Discussion",
		"Methods", "Introduction", "Overview", // Many duplicates
	}

	b.Run("UniqueHandle_Interning", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			streamer := streaming.NewStreamer(streaming.StreamOptions{})

			// Process headings - this will use unique.Handle internally
			ctx := context.Background()
			book := &models.Book{
				Title: "Test Document",
				Content: &models.Content{
					Document: &models.Document{
						Body: models.Body{
							Content: createHeadingElements(headings),
						},
					},
				},
			}

			eventStream := streamer.Stream(ctx, book)

			// Count events to ensure processing happens
			count := 0
			for _ = range eventStream {
				count++
			}

			if count == 0 {
				b.Fatal("No events generated")
			}
		}
	})
}

// Helper functions

func createLargeTestBookForBench(paragraphCount int) *models.Book {
	elements := make([]models.StructuralElement, paragraphCount)

	for i := 0; i < paragraphCount; i++ {
		content := fmt.Sprintf("This is paragraph number %d. It contains some text to simulate a real document with multiple sentences and formatting.", i+1)

		elements[i] = models.StructuralElement{
			Paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: content,
						},
					},
				},
			},
		}
	}

	return &models.Book{
		Title:       fmt.Sprintf("Large Test Document (%d paragraphs)", paragraphCount),
		Description: "A large document for testing memory efficiency",
		Content: &models.Content{
			Document: &models.Document{
				Body: models.Body{
					Content: elements,
				},
			},
		},
	}
}

func createHeadingElements(headings []string) []models.StructuralElement {
	elements := make([]models.StructuralElement, len(headings))

	for i, heading := range headings {
		elements[i] = models.StructuralElement{
			Paragraph: &models.Paragraph{
				ParagraphStyle: models.ParagraphStyle{
					NamedStyleType: "HEADING_1",
				},
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: heading,
						},
					},
				},
			},
		}
	}

	return elements
}
