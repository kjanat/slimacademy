package writers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// Integration tests for the complete conversion pipeline

func TestIntegration_CompleteAcademicDocument(t *testing.T) {
	// Create a comprehensive academic document
	book := createAcademicTestBook()

	// Test conversion to HTML with academic features
	ctx := context.Background()
	result := convertBookToHTML(ctx, t, book)

	// Verify academic header structure
	verifyAcademicHeader(t, result, book)

	// Verify table of contents
	verifyTableOfContents(t, result, book.Chapters)

	// Verify content structure
	verifyContentStructure(t, result)

	// Verify HTML is well-formed
	verifyWellFormedHTML(t, result)
}

func TestIntegration_MultiFormatConsistency(t *testing.T) {
	book := createTestBook()
	ctx := context.Background()

	// Test formats that are currently implemented
	formats := []string{"html"}

	// Convert to all formats
	results := make(map[string]string)
	for _, format := range formats {
		switch format {
		case "html":
			results[format] = convertBookToHTML(ctx, t, book)
		default:
			t.Logf("Skipping format %s - not implemented yet", format)
		}
	}

	// Verify all formats produced output
	for format, result := range results {
		if result == "" {
			t.Errorf("Format %s produced empty output", format)
		}
		if len(result) < 100 {
			t.Errorf("Format %s produced suspiciously short output: %d chars", format, len(result))
		}
	}

	// TODO: Add cross-format consistency checks when more formats are implemented
}

func TestIntegration_StreamingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a large document
	book := createLargeTestBook(1000)
	ctx := context.Background()

	start := time.Now()
	result := convertBookToHTML(ctx, t, book)
	duration := time.Since(start)

	// Verify result is reasonable
	if result == "" {
		t.Error("Large document conversion produced empty result")
	}

	// Performance expectations (adjust as needed)
	maxDuration := 5 * time.Second
	if duration > maxDuration {
		t.Errorf("Large document conversion took too long: %v (max: %v)", duration, maxDuration)
	}

	t.Logf("Converted large document (%d paragraphs) in %v", 1000, duration)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		book        *models.Book
		expectError bool
	}{
		{
			name:        "nil book",
			book:        nil,
			expectError: false, // Should handle gracefully
		},
		{
			name:        "empty book",
			book:        &models.Book{},
			expectError: false,
		},
		{
			name: "book with nil content",
			book: &models.Book{
				ID:      1,
				Title:   "Test Book",
				Content: nil,
			},
			expectError: false,
		},
		{
			name: "book with empty content",
			book: &models.Book{
				ID:      1,
				Title:   "Test Book",
				Content: &models.Content{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertBookToHTML(ctx, t, tt.book)

			if tt.expectError {
				if result != "" {
					t.Error("Expected error case to produce empty result")
				}
			} else {
				// Even error cases should produce some basic HTML structure
				if !strings.Contains(result, "<!DOCTYPE html>") {
					t.Error("Result should contain valid HTML structure")
				}
			}
		})
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	book := createLargeTestBook(100)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context during processing
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Should handle cancellation gracefully
	result := convertBookToHTML(ctx, t, book)

	// Result might be incomplete but should not cause panic
	// Basic structure should still be present if any processing occurred
	if result != "" && !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("Partial result should still have valid HTML start")
	}
}

func TestIntegration_ConcurrentConversion(t *testing.T) {
	book := createTestBook()
	numGoroutines := 10

	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Start concurrent conversions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ctx := context.Background()
			result := convertBookToHTML(ctx, t, book)
			results <- result
		}(i)
	}

	// Collect results
	var allResults []string
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			allResults = append(allResults, result)
		case err := <-errors:
			t.Errorf("Concurrent conversion failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Error("Concurrent conversion timed out")
			return
		}
	}

	// Verify all results are identical (deterministic output)
	if len(allResults) != numGoroutines {
		t.Errorf("Expected %d results, got %d", numGoroutines, len(allResults))
	}

	firstResult := allResults[0]
	for i, result := range allResults {
		if result != firstResult {
			t.Errorf("Result %d differs from first result", i)
			// Only show first few differences to avoid overwhelming output
			if i < 3 {
				t.Logf("First result length: %d", len(firstResult))
				t.Logf("Result %d length: %d", i, len(result))
			}
		}
	}
}

func TestIntegration_ComplexFormatting(t *testing.T) {
	// Create book with complex nested formatting
	book := createComplexFormattingBook()
	ctx := context.Background()

	result := convertBookToHTML(ctx, t, book)

	// Verify complex formatting is handled correctly
	if !strings.Contains(result, "<strong>") {
		t.Error("Result should contain bold formatting")
	}
	if !strings.Contains(result, "<em>") {
		t.Error("Result should contain italic formatting")
	}
	if !strings.Contains(result, "<a href=") {
		t.Error("Result should contain links")
	}

	// Verify nested formatting
	if !strings.Contains(result, "<strong><em>") || !strings.Contains(result, "</em></strong>") {
		t.Error("Result should contain properly nested formatting")
	}

	// Verify HTML is still well-formed with complex formatting
	verifyWellFormedHTML(t, result)
}

func TestIntegration_LargeTable(t *testing.T) {
	book := createLargeTableBook()
	ctx := context.Background()

	result := convertBookToHTML(ctx, t, book)

	// Verify table structure
	if !strings.Contains(result, "<table") {
		t.Error("Result should contain table")
	}
	if !strings.Contains(result, "<th") {
		t.Error("Result should contain table headers")
	}
	if !strings.Contains(result, "<td") {
		t.Error("Result should contain table cells")
	}

	// Count rows and cells to verify structure
	thCount := strings.Count(result, "<th")
	tdCount := strings.Count(result, "<td")
	trCount := strings.Count(result, "<tr")

	if thCount == 0 {
		t.Error("Should have table headers")
	}
	if tdCount == 0 {
		t.Error("Should have table data cells")
	}
	if trCount == 0 {
		t.Error("Should have table rows")
	}

	// Verify table is properly closed
	if strings.Count(result, "<table") != strings.Count(result, "</table>") {
		t.Error("Table tags should be balanced")
	}
}

// Helper functions for integration tests

func convertBookToHTML(ctx context.Context, t *testing.T, book *models.Book) string {
	// Handle nil book case
	if book == nil {
		book = &models.Book{
			ID:    0,
			Title: "Empty Book",
		}
	}

	// Sanitize book
	sanitizer := sanitizer.NewSanitizer()
	sanitizeResult := sanitizer.Sanitize(book)

	// Create HTML writer
	writer := &HTMLWriterV2{
		HTMLWriter: NewHTMLWriter(),
	}

	// Create streamer
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	// Stream events to writer
	for event := range streamer.Stream(ctx, sanitizeResult.Book) {
		err := writer.Handle(event)
		if err != nil {
			t.Errorf("Writer failed to handle event: %v", err)
			break
		}
	}

	// Get result
	result, err := writer.Flush()
	if err != nil {
		t.Errorf("Writer failed to flush: %v", err)
		return ""
	}

	return result
}

func createAcademicTestBook() *models.Book {
	readProgress := int64(75)
	return &models.Book{
		ID:                 1,
		Title:              "Advanced Computer Science",
		Description:        "A comprehensive guide to advanced computer science topics",
		AvailableDate:      "2024-01-15",
		ExamDate:           "2024-06-30",
		BachelorYearNumber: "Year 3",
		CollegeStartYear:   2021,
		ReadProgress:       &readProgress,
		ReadPercentage:     75.5,
		PageCount:          300,
		HasFreeChapters:    models.BoolInt(true),
		Periods:            []string{"Q1", "Q2", "Q3"},
		Images: []models.BookImage{
			{ID: 1, ImageURL: "/images/algorithms.png"},
			{ID: 2, ImageURL: "/images/datastructures.jpg"},
		},
		Chapters: []models.Chapter{
			{
				ID:       1,
				Title:    "Introduction",
				IsFree:   models.BoolInt(true),
				IsLocked: models.BoolInt(false),
				SubChapters: []models.Chapter{
					{ID: 2, Title: "Course Overview", IsFree: models.BoolInt(true)},
					{ID: 3, Title: "Prerequisites", IsFree: models.BoolInt(true)},
				},
			},
			{
				ID:       4,
				Title:    "Data Structures",
				IsFree:   models.BoolInt(false),
				IsLocked: models.BoolInt(true),
				SubChapters: []models.Chapter{
					{ID: 5, Title: "Arrays and Lists", IsLocked: models.BoolInt(true)},
					{ID: 6, Title: "Trees and Graphs", IsLocked: models.BoolInt(true)},
				},
			},
			{
				ID:       7,
				Title:    "Algorithms",
				IsFree:   models.BoolInt(false),
				IsLocked: models.BoolInt(true),
			},
		},
		Content: &models.Content{
			Document: &models.Document{
				DocumentID: "academic-cs-book",
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "Introduction to Computer Science",
										},
									},
								},
								ParagraphStyle: models.ParagraphStyle{
									NamedStyleType: "HEADING_1",
								},
							},
						},
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "This comprehensive guide covers advanced topics in computer science, including data structures, algorithms, and computational complexity.",
										},
									},
								},
								ParagraphStyle: models.ParagraphStyle{
									NamedStyleType: "NORMAL_TEXT",
								},
							},
						},
					},
				},
			},
		},
	}
}

func createTestBook() *models.Book {
	return &models.Book{
		ID:    1,
		Title: "Test Book",
		Content: &models.Content{
			Document: &models.Document{
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "Test paragraph content",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createLargeTestBook(paragraphCount int) *models.Book {
	elements := make([]models.StructuralElement, paragraphCount)
	for i := range elements {
		elements[i] = models.StructuralElement{
			Paragraph: &models.Paragraph{
				Elements: []models.ParagraphElement{
					{
						TextRun: &models.TextRun{
							Content: fmt.Sprintf("This is paragraph %d with some content to make it realistic for testing purposes.", i+1),
						},
					},
				},
			},
		}
	}

	return &models.Book{
		ID:    1,
		Title: fmt.Sprintf("Large Test Book (%d paragraphs)", paragraphCount),
		Content: &models.Content{
			Document: &models.Document{
				Body: models.Body{
					Content: elements,
				},
			},
		},
	}
}

func createComplexFormattingBook() *models.Book {
	bold := true
	italic := true
	url := "https://example.com"

	return &models.Book{
		ID:    1,
		Title: "Complex Formatting Test",
		Content: &models.Content{
			Document: &models.Document{
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Paragraph: &models.Paragraph{
								Elements: []models.ParagraphElement{
									{
										TextRun: &models.TextRun{
											Content: "This text has ",
										},
									},
									{
										TextRun: &models.TextRun{
											Content: "bold",
											TextStyle: models.TextStyle{
												Bold: &bold,
											},
										},
									},
									{
										TextRun: &models.TextRun{
											Content: " and ",
										},
									},
									{
										TextRun: &models.TextRun{
											Content: "italic",
											TextStyle: models.TextStyle{
												Italic: &italic,
											},
										},
									},
									{
										TextRun: &models.TextRun{
											Content: " and ",
										},
									},
									{
										TextRun: &models.TextRun{
											Content: "bold italic link",
											TextStyle: models.TextStyle{
												Bold:   &bold,
												Italic: &italic,
												Link: &models.Link{
													URL: &url,
												},
											},
										},
									},
									{
										TextRun: &models.TextRun{
											Content: " formatting.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createLargeTableBook() *models.Book {
	// Create a 5x10 table
	rows := make([]models.TableRow, 10)
	for i := range rows {
		cells := make([]models.TableCell, 5)
		for j := range cells {
			cells[j] = models.TableCell{
				Content: []models.StructuralElement{
					{
						Paragraph: &models.Paragraph{
							Elements: []models.ParagraphElement{
								{
									TextRun: &models.TextRun{
										Content: fmt.Sprintf("Cell %d,%d", i+1, j+1),
									},
								},
							},
						},
					},
				},
			}
		}
		rows[i] = models.TableRow{
			TableCells: cells,
		}
	}

	return &models.Book{
		ID:    1,
		Title: "Large Table Test",
		Content: &models.Content{
			Document: &models.Document{
				Body: models.Body{
					Content: []models.StructuralElement{
						{
							Table: &models.Table{
								Rows:      10,
								Columns:   5,
								TableRows: rows,
							},
						},
					},
				},
			},
		},
	}
}

func verifyAcademicHeader(t *testing.T, result string, book *models.Book) {
	if !strings.Contains(result, `class="academic-header"`) {
		t.Error("Result should contain academic header")
	}
	if !strings.Contains(result, book.Title) {
		t.Error("Result should contain book title")
	}
	if !strings.Contains(result, book.Description) {
		t.Error("Result should contain book description")
	}
	if !strings.Contains(result, book.BachelorYearNumber) {
		t.Error("Result should contain bachelor year number")
	}
}

func verifyTableOfContents(t *testing.T, result string, chapters []models.Chapter) {
	if len(chapters) == 0 {
		return
	}

	if !strings.Contains(result, `class="table-of-contents"`) {
		t.Error("Result should contain table of contents")
	}
	if !strings.Contains(result, "Table of Contents") {
		t.Error("Result should contain TOC title")
	}

	// Check for chapter titles in TOC
	for _, chapter := range chapters {
		if !strings.Contains(result, chapter.Title) {
			t.Errorf("TOC should contain chapter title: %s", chapter.Title)
		}
	}
}

func verifyContentStructure(t *testing.T, result string) {
	if !strings.Contains(result, `class="content"`) {
		t.Error("Result should contain content section")
	}
	if !strings.Contains(result, "<main") {
		t.Error("Result should contain main element")
	}
}

func verifyWellFormedHTML(t *testing.T, result string) {
	// Basic HTML structure checks
	if !strings.HasPrefix(result, "<!DOCTYPE html>") {
		t.Error("HTML should start with DOCTYPE declaration")
	}
	if !strings.Contains(result, "<html") {
		t.Error("HTML should contain html element")
	}
	if !strings.Contains(result, "<head>") {
		t.Error("HTML should contain head element")
	}
	if !strings.Contains(result, "<body>") {
		t.Error("HTML should contain body element")
	}
	if !strings.HasSuffix(strings.TrimSpace(result), "</html>") {
		t.Error("HTML should end with closing html tag")
	}

	// Check that basic tags are balanced
	checkBalancedTags := []string{"html", "body", "title", "h1", "h2", "p", "ul", "li", "table", "tr", "td", "th"}
	for _, tag := range checkBalancedTags {
		openTag := fmt.Sprintf("<%s", tag)
		closeTag := fmt.Sprintf("</%s>", tag)
		openCount := strings.Count(result, openTag)
		closeCount := strings.Count(result, closeTag)

		// For self-closing or content-dependent tags, only check if there are any
		if openCount > 0 && openCount != closeCount {
			t.Errorf("Tag %s is not balanced: %d open, %d close", tag, openCount, closeCount)
		}
	}

	// Special check for head tag (appears in meta tags)
	headOpenCount := strings.Count(result, "<head>")
	headCloseCount := strings.Count(result, "</head>")
	if headOpenCount != headCloseCount {
		t.Errorf("Head tag is not balanced: %d open, %d close", headOpenCount, headCloseCount)
	}
}

// Benchmark integration tests
func BenchmarkIntegration_CompleteConversion(b *testing.B) {
	book := createTestBook()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertBookToHTML(ctx, &testing.T{}, book)
	}
}

func BenchmarkIntegration_AcademicDocument(b *testing.B) {
	book := createAcademicTestBook()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertBookToHTML(ctx, &testing.T{}, book)
	}
}
