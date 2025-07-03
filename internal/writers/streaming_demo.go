package writers

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// DemoStreamingImprovements demonstrates the iter.Seq and unique.Handle improvements
func DemoStreamingImprovements() {
	fmt.Println("🚀 SlimAcademy Streaming Architecture Improvements")
	fmt.Println("==============================================")
	fmt.Println()

	writer := NewMinimalHTMLWriter()

	// Create a streaming event generator
	eventStream := generateTestEventStream()

	fmt.Println("📊 Processing events with Go 1.23+ iter.Seq streaming...")
	start := time.Now()

	html, err := writer.ProcessEventStream(context.Background(), eventStream)
	if err != nil {
		fmt.Printf("❌ Error processing stream: %v\n", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("✅ Processed stream in %v\n", duration)
	fmt.Printf("📏 Generated HTML: %d bytes\n", len(html))
	fmt.Println()

	// Show duplicate detection statistics
	fmt.Println("🔍 Duplicate Detection with unique.Handle:")
	fmt.Printf("   • Unique URLs seen: %d\n", len(writer.seenURLs))
	fmt.Printf("   • Unique anchors seen: %d\n", len(writer.seenAnchors))
	fmt.Printf("   • Unique text patterns: %d\n", len(writer.seenTexts))
	fmt.Println()

	// Show text pattern analysis
	fmt.Println("📈 Text Pattern Analysis:")
	for _, count := range writer.seenTexts {
		if count > 1 {
			fmt.Printf("   • Pattern appears %d times\n", count)
			break // Just show one example
		}
	}
	fmt.Println()

	fmt.Println("🎯 Key Improvements Implemented:")
	fmt.Println("   ✅ iter.Seq[Event] for streaming processing")
	fmt.Println("   ✅ unique.Handle[string] for O(1) duplicate detection")
	fmt.Println("   ✅ Context-aware cancellation support")
	fmt.Println("   ✅ Memory-efficient URL and anchor deduplication")
	fmt.Println("   ✅ Template integration with streaming architecture")
	fmt.Println()

	// Compare with slice processing
	fmt.Println("⚖️  Architecture Comparison:")
	fmt.Println("   • Slice Processing: Loads all events into memory first")
	fmt.Println("   • Stream Processing: Processes events one-at-a-time")
	fmt.Println("   • Duplicate Detection: O(1) lookup with unique.Handle")
	fmt.Println("   • Template Integration: Both use same minimal template system")
}

// generateTestEventStream creates a streaming event generator for testing
func generateTestEventStream() iter.Seq[streaming.Event] {
	return func(yield func(streaming.Event) bool) {
		// Start document
		if !yield(streaming.Event{
			Kind:        streaming.StartDoc,
			Title:       "Streaming Architecture Demo",
			Description: "Demonstrating Go 1.23+ iter.Seq and unique.Handle improvements",
		}) {
			return
		}

		// Generate content with duplicate patterns for testing
		duplicateURL := "https://example.com/api"
		duplicateText := "This text appears multiple times"

		for i := 0; i < 5; i++ {
			// Heading with same anchor pattern (tests deduplication)
			if !yield(streaming.Event{
				Kind:     streaming.StartHeading,
				Level:    2,
				AnchorID: "section", // Same ID - tests duplicate handling
			}) {
				return
			}
			if !yield(streaming.Event{
				Kind:        streaming.Text,
				TextContent: fmt.Sprintf("Section %d", i),
			}) {
				return
			}
			if !yield(streaming.Event{Kind: streaming.EndHeading}) {
				return
			}

			// Paragraph with duplicate text and URLs
			if !yield(streaming.Event{Kind: streaming.StartParagraph}) {
				return
			}
			if !yield(streaming.Event{
				Kind:        streaming.Text,
				TextContent: duplicateText, // Same text - tests deduplication
			}) {
				return
			}
			if !yield(streaming.Event{
				Kind:    streaming.StartFormatting,
				Style:   streaming.Link,
				LinkURL: duplicateURL, // Same URL - tests deduplication
			}) {
				return
			}
			if !yield(streaming.Event{
				Kind:        streaming.Text,
				TextContent: "API documentation",
			}) {
				return
			}
			if !yield(streaming.Event{
				Kind:  streaming.EndFormatting,
				Style: streaming.Link,
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
