package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	testutils "github.com/kjanat/slimacademy/test/utils"
)

// Integration tests for Parser + Sanitizer workflow
func TestParserSanitizer_Integration(t *testing.T) {
	parser := NewBookParser()
	s := sanitizer.NewSanitizer()

	tests := []struct {
		name               string
		bookData           map[string]string
		expectedWarnings   int
		expectCleanContent bool
	}{
		{
			name: "clean book requires no sanitization",
			bookData: map[string]string{
				"123.json": `{
					"id": 123,
					"title": "Clean Book",
					"description": "A perfectly clean book"
				}`,
				"chapters.json": `[
					{"id": 1, "title": "Chapter 1"},
					{"id": 2, "title": "Chapter 2"}
				]`,
				"content.json": `{
					"documentId": "doc-123",
					"body": {
						"content": [
							{
								"paragraph": {
									"elements": [
										{
											"textRun": {
												"content": "Clean paragraph content"
											}
										}
									]
								}
							}
						]
					}
				}`,
			},
			expectedWarnings:   0,
			expectCleanContent: true,
		},
		{
			name: "dirty book requires sanitization",
			bookData: map[string]string{
				"456.json": `{
					"id": 456,
					"title": "Dirty Book",
					"description": "A book with problematic content"
				}`,
				"chapters.json": `[
					{"id": 1, "title": "Good Chapter"},
					{"id": 2, "title": "   "}
				]`,
				"content.json": `{
					"documentId": "doc-456",
					"body": {
						"content": [
							{
								"paragraph": {
									"elements": [
										{
											"textRun": {
												"content": "Text\u0000with\u0001control\u0002characters"
											}
										}
									],
									"paragraphStyle": {
										"namedStyleType": "HEADING_1"
									}
								}
							}
						]
					}
				}`,
			},
			expectedWarnings:   2, // Empty chapter title + dirty text
			expectCleanContent: false,
		},
		{
			name: "book with empty headings",
			bookData: map[string]string{
				"789.json": `{
					"id": 789,
					"title": "Book with Empty Headings"
				}`,
				"chapters.json": "[]",
				"content.json": `{
					"documentId": "doc-789",
					"body": {
						"content": [
							{
								"paragraph": {
									"elements": [
										{
											"textRun": {
												"content": "   "
											}
										}
									],
									"paragraphStyle": {
										"namedStyleType": "HEADING_2"
									}
								}
							}
						]
					}
				}`,
			},
			expectedWarnings:   2, // Text sanitization + empty heading
			expectCleanContent: false,
		},
		{
			name: "book with malformed links",
			bookData: map[string]string{
				"999.json": `{
					"id": 999,
					"title": "Book with Bad Links"
				}`,
				"chapters.json": "[]",
				"content.json": `{
					"documentId": "doc-999",
					"body": {
						"content": [
							{
								"paragraph": {
									"elements": [
										{
											"textRun": {
												"content": "Link text",
												"textStyle": {
													"link": {
														"url": "  https://example.com<script>alert('xss')</script>  "
													}
												}
											}
										}
									]
								}
							}
						]
					}
				}`,
			},
			expectedWarnings:   2, // Whitespace + HTML tags in URL
			expectCleanContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with test files
			tempDir := testutils.CreateTempDir(t)
			for filename, content := range tt.bookData {
				path := filepath.Join(tempDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create %s: %v", filename, err)
				}
			}

			// Parse the book
			book, err := parser.ParseBook(tempDir)
			if err != nil {
				t.Fatalf("Failed to parse book: %v", err)
			}

			// Sanitize the book
			result := s.Sanitize(book)
			if result == nil {
				t.Fatal("Sanitizer returned nil result")
			}

			// Verify sanitizer created a deep copy
			if result.Book == book {
				t.Error("Sanitizer should create a deep copy, not return original")
			}

			// Check warning count
			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
				for i, w := range result.Warnings {
					t.Logf("Warning %d: %s at %s (original: %q, fixed: %q)",
						i+1, w.Issue, w.Location, w.Original, w.Fixed)
				}
			}

			// Verify book integrity after sanitization
			if result.Book.ID != book.ID {
				t.Error("Book ID should be preserved during sanitization")
			}
			if result.Book.Title != book.Title {
				t.Error("Book title should be preserved during sanitization")
			}

			// Verify deep copy independence
			originalTitle := book.Title
			result.Book.Title = "Modified Title"
			if book.Title != originalTitle {
				t.Error("Modifying sanitized book should not affect original")
			}
		})
	}
}

// Test parser and sanitizer with real SlimAcademy data structure
func TestParserSanitizer_RealDataIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real data integration test in short mode")
	}

	parser := NewBookParser()
	s := sanitizer.NewSanitizer()

	// Test with source directory if available
	sourceDir := filepath.Join("..", "..", "source")
	bookDirs, err := parser.FindAllBooks(sourceDir)
	if err != nil {
		t.Skip("Source directory not available for integration testing")
	}

	if len(bookDirs) == 0 {
		t.Skip("No books found for integration testing")
	}

	// Test with first book only to keep test time reasonable
	bookDir := bookDirs[0]
	t.Run("real_book_integration", func(t *testing.T) {
		// Parse book
		book, err := parser.ParseBook(bookDir)
		if err != nil {
			t.Fatalf("Failed to parse real book: %v", err)
		}

		// Sanitize book
		result := s.Sanitize(book)
		if result == nil {
			t.Fatal("Sanitizer returned nil result")
		}

		// Log warnings if any (real data might have issues)
		if len(result.Warnings) > 0 {
			t.Logf("Found %d warnings in real book data:", len(result.Warnings))
			for i, w := range result.Warnings {
				if i < 10 { // Limit output to first 10 warnings
					t.Logf("  %s at %s", w.Issue, w.Location)
				}
			}
			if len(result.Warnings) > 10 {
				t.Logf("  ... and %d more warnings", len(result.Warnings)-10)
			}
		}

		// Verify sanitized book is valid
		if result.Book.ID == 0 {
			t.Error("Sanitized book should have valid ID")
		}
		if result.Book.Title == "" {
			t.Error("Sanitized book should have title")
		}

		// Verify content structure is preserved
		if book.Content != nil && result.Book.Content == nil {
			t.Error("Content should be preserved during sanitization")
		}
		if book.Chapters != nil && result.Book.Chapters == nil {
			t.Error("Chapters should be preserved during sanitization")
		}

		t.Logf("Successfully processed real book: %s", result.Book.Title)
	})
}

// Test edge cases in parser-sanitizer workflow
func TestParserSanitizer_EdgeCases(t *testing.T) {
	parser := NewBookParser()
	s := sanitizer.NewSanitizer()

	tests := []struct {
		name     string
		bookData map[string]string
		testFunc func(t *testing.T, book *models.Book, result *sanitizer.Result)
	}{
		{
			name: "very large book",
			bookData: map[string]string{
				"123.json":      `{"id": 123, "title": "Large Book"}`,
				"chapters.json": generateLargeChaptersJSON(1000),
				"content.json":  generateLargeContentJSON(5000),
			},
			testFunc: func(t *testing.T, book *models.Book, result *sanitizer.Result) {
				if len(result.Book.Chapters) != 1000 {
					t.Errorf("Expected 1000 chapters, got %d", len(result.Book.Chapters))
				}
				// Performance check
				if result.Book.Content == nil {
					t.Error("Large content should be preserved")
				}
			},
		},
		{
			name: "book with unicode content",
			bookData: map[string]string{
				"456.json": `{
					"id": 456,
					"title": "Unicode Test Book 🎓",
					"description": "Tëst böök wïth ünïcödë çhäräctërs"
				}`,
				"chapters.json": `[
					{"id": 1, "title": "Français"},
					{"id": 2, "title": "Español"},
					{"id": 3, "title": "Português"},
					{"id": 4, "title": "中文"},
					{"id": 5, "title": "العربية"}
				]`,
				"content.json": `{
					"documentId": "unicode-doc",
					"body": {
						"content": [
							{
								"paragraph": {
									"elements": [
										{
											"textRun": {
												"content": "Café naïve résumé 中文测试 العربية"
											}
										}
									]
								}
							}
						]
					}
				}`,
			},
			testFunc: func(t *testing.T, book *models.Book, result *sanitizer.Result) {
				// Unicode should be preserved
				if !contains(result.Book.Title, "🎓") {
					t.Error("Unicode emoji should be preserved in title")
				}
				if len(result.Book.Chapters) != 5 {
					t.Error("All unicode chapter titles should be preserved")
				}
			},
		},
		{
			name: "book with nested subchapters",
			bookData: map[string]string{
				"789.json": `{"id": 789, "title": "Nested Book"}`,
				"chapters.json": `[
					{
						"id": 1,
						"title": "Main Chapter",
						"subChapters": [
							{
								"id": 2,
								"title": "Sub Chapter 1",
								"subChapters": [
									{"id": 3, "title": "Sub Sub Chapter"}
								]
							}
						]
					}
				]`,
				"content.json": `{"documentId": "nested-doc", "body": {"content": []}}`,
			},
			testFunc: func(t *testing.T, book *models.Book, result *sanitizer.Result) {
				if len(result.Book.Chapters) != 1 {
					t.Error("Should have one main chapter")
				}
				if len(result.Book.Chapters[0].SubChapters) != 1 {
					t.Error("Should have one sub chapter")
				}
				if len(result.Book.Chapters[0].SubChapters[0].SubChapters) != 1 {
					t.Error("Should have one sub-sub chapter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with test files
			tempDir := testutils.CreateTempDir(t)
			for filename, content := range tt.bookData {
				path := filepath.Join(tempDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create %s: %v", filename, err)
				}
			}

			// Parse and sanitize
			book, err := parser.ParseBook(tempDir)
			if err != nil {
				t.Fatalf("Failed to parse book: %v", err)
			}

			result := s.Sanitize(book)
			if result == nil {
				t.Fatal("Sanitizer returned nil result")
			}

			// Run custom test function
			tt.testFunc(t, book, result)
		})
	}
}

// Test concurrent parser-sanitizer operations
func TestParserSanitizer_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	parser := NewBookParser()

	// Create test books
	numBooks := 10
	bookDirs := make([]string, numBooks)

	for i := 0; i < numBooks; i++ {
		tempDir := testutils.CreateTempDir(t)
		bookDirs[i] = tempDir

		bookData := map[string]string{
			"book.json": fmt.Sprintf(`{
				"id": %d,
				"title": "Concurrent Book %d"
			}`, i+1, i+1),
			"chapters.json": `[{"id": 1, "title": "Chapter 1"}]`,
			"content.json":  `{"documentId": "doc", "body": {"content": []}}`,
		}

		for filename, content := range bookData {
			path := filepath.Join(tempDir, filename)
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create %s: %v", filename, err)
			}
		}
	}

	// Test concurrent parsing and sanitization
	results := make(chan struct {
		bookID int
		err    error
	}, numBooks)

	for i, bookDir := range bookDirs {
		go func(id int, dir string) {
			defer func() {
				if r := recover(); r != nil {
					results <- struct {
						bookID int
						err    error
					}{id, fmt.Errorf("panic: %v", r)}
				}
			}()

			// Parse book
			book, err := parser.ParseBook(dir)
			if err != nil {
				results <- struct {
					bookID int
					err    error
				}{id, err}
				return
			}

			// Sanitize book
			s := sanitizer.NewSanitizer()
			result := s.Sanitize(book)
			if result == nil {
				results <- struct {
					bookID int
					err    error
				}{id, fmt.Errorf("sanitizer returned nil")}
				return
			}

			results <- struct {
				bookID int
				err    error
			}{id, nil}
		}(i, bookDir)
	}

	// Collect results
	for i := 0; i < numBooks; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("Book %d failed: %v", result.bookID, result.err)
		}
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(s, substr)))
}

func generateLargeChaptersJSON(count int) string {
	var parts []string
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf(`{"id": %d, "title": "Chapter %d"}`, i+1, i+1))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func generateLargeContentJSON(paragraphCount int) string {
	var paragraphs []string
	for i := 0; i < paragraphCount; i++ {
		paragraphs = append(paragraphs, fmt.Sprintf(`{
			"paragraph": {
				"elements": [
					{
						"textRun": {
							"content": "Paragraph %d content."
						}
					}
				]
			}
		}`, i+1))
	}

	return fmt.Sprintf(`{
		"documentId": "large-doc",
		"body": {
			"content": [%s]
		}
	}`, strings.Join(paragraphs, ","))
}
