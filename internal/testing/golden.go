package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/kjanat/slimacademy/internal/models"
	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	"github.com/kjanat/slimacademy/internal/streaming"
	"github.com/kjanat/slimacademy/internal/writers"
)

// GoldenTest represents a single golden test case
type GoldenTest struct {
	Name        string
	BookPath    string
	Format      string
	GoldenPath  string
	Description string
}

// GoldenTestSuite manages golden tests for the document transformation system
type GoldenTestSuite struct {
	testDir    string
	goldenDir  string
	fixtureDir string
}

// NewGoldenTestSuite creates a new golden test suite
func NewGoldenTestSuite(testDir string) *GoldenTestSuite {
	return &GoldenTestSuite{
		testDir:    testDir,
		goldenDir:  filepath.Join(testDir, "golden"),
		fixtureDir: filepath.Join(testDir, "fixtures"),
	}
}

// RunGoldenTests executes all golden tests for the specified formats
func (suite *GoldenTestSuite) RunGoldenTests(t *testing.T, formats []string) {
	tests, err := suite.discoverTests(formats)
	if err != nil {
		t.Fatalf("Failed to discover golden tests: %v", err)
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			suite.runSingleTest(t, test)
		})
	}
}

// runSingleTest executes a single golden test
func (suite *GoldenTestSuite) runSingleTest(t *testing.T, test GoldenTest) {
	// Parse the book
	bookParser := parser.NewBookParser()
	book, err := bookParser.ParseBook(test.BookPath)
	if err != nil {
		t.Fatalf("Failed to parse book %s: %v", test.BookPath, err)
	}

	// Generate output using the new architecture
	ctx := context.Background()
	output, err := suite.generateOutput(ctx, book, test.Format)
	if err != nil {
		t.Fatalf("Failed to generate output: %v", err)
	}

	// Compare with golden file
	if suite.shouldUpdateGolden() {
		suite.updateGoldenFile(t, test.GoldenPath, output)
		t.Logf("Updated golden file: %s", test.GoldenPath)
		return
	}

	expected, err := os.ReadFile(test.GoldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", test.GoldenPath, err)
	}

	if string(expected) != output {
		t.Errorf("Output doesn't match golden file %s", test.GoldenPath)
		suite.showDiff(t, string(expected), output)
	}
}

// generateOutput creates the transformed output using the new streaming architecture
func (suite *GoldenTestSuite) generateOutput(ctx context.Context, book *models.Book, format string) (string, error) {
	// Sanitize the book first
	sanitizer := sanitizer.NewSanitizer()
	result := sanitizer.Sanitize(book)

	// Create writer for the format
	multiWriter, err := writers.NewMultiWriter(ctx, []string{format})
	if err != nil {
		return "", err
	}
	defer multiWriter.Close()

	// Create streamer
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	// Process events
	if err := multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for event := range streamer.Stream(ctx, result.Book) {
			if !yield(event) {
				break
			}
		}
	}); err != nil {
		return "", err
	}

	// Get results
	results, err := multiWriter.FlushAll()
	if err != nil {
		return "", err
	}

	output, exists := results[format]
	if !exists {
		return "", fmt.Errorf("no output generated for format %s", format)
	}

	return output, nil
}

// discoverTests finds all golden test cases in the test directory
func (suite *GoldenTestSuite) discoverTests(formats []string) ([]GoldenTest, error) {
	var tests []GoldenTest

	// Walk through fixture directories
	err := filepath.WalkDir(suite.fixtureDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Check if this is a book directory (contains content.json)
		contentPath := filepath.Join(path, "content.json")
		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			return nil
		}

		// Generate tests for each format
		bookName := filepath.Base(path)
		for _, format := range formats {
			goldenFile := fmt.Sprintf("%s.%s", bookName, getExtension(format))
			goldenPath := filepath.Join(suite.goldenDir, goldenFile)

			tests = append(tests, GoldenTest{
				Name:        fmt.Sprintf("%s_%s", bookName, format),
				BookPath:    path,
				Format:      format,
				GoldenPath:  goldenPath,
				Description: fmt.Sprintf("Golden test for %s in %s format", bookName, format),
			})
		}

		return nil
	})

	return tests, err
}

// shouldUpdateGolden checks if golden files should be updated
func (suite *GoldenTestSuite) shouldUpdateGolden() bool {
	return os.Getenv("UPDATE_GOLDEN") == "1"
}

// updateGoldenFile writes new golden file content
func (suite *GoldenTestSuite) updateGoldenFile(t *testing.T, goldenPath, content string) {
	// Ensure directory exists
	dir := filepath.Dir(goldenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create golden directory: %v", err)
	}

	// Write golden file
	if err := os.WriteFile(goldenPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write golden file: %v", err)
	}
}

// showDiff displays a simple diff between expected and actual output
func (suite *GoldenTestSuite) showDiff(t *testing.T, expected, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	t.Logf("Diff (first 20 lines):")
	for i := 0; i < maxLines && i < 20; i++ {
		var expectedLine, actualLine string
		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			t.Logf("Line %d:", i+1)
			t.Logf("  Expected: %q", expectedLine)
			t.Logf("  Actual:   %q", actualLine)
		}
	}
}

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

// PropertyTest represents a property-based test case
type PropertyTest struct {
	Name        string
	Property    func(*models.Book) bool
	Description string
}

// PropertyTestSuite manages property-based tests
type PropertyTestSuite struct {
	tests []PropertyTest
}

// NewPropertyTestSuite creates a new property test suite
func NewPropertyTestSuite() *PropertyTestSuite {
	suite := &PropertyTestSuite{}
	suite.registerBuiltinTests()
	return suite
}

// registerBuiltinTests adds built-in property tests
func (suite *PropertyTestSuite) registerBuiltinTests() {
	suite.tests = []PropertyTest{
		{
			Name: "balanced_markers",
			Property: func(book *models.Book) bool {
				return suite.checkBalancedMarkers(book)
			},
			Description: "Event stream should have balanced open/close markers",
		},
		{
			Name: "no_empty_headings",
			Property: func(book *models.Book) bool {
				return suite.checkNoEmptyHeadings(book)
			},
			Description: "No empty headings should be present after sanitization",
		},
		{
			Name: "valid_utf8",
			Property: func(book *models.Book) bool {
				return suite.checkValidUTF8(book)
			},
			Description: "All text content should be valid UTF-8",
		},
	}
}

// RunPropertyTests executes all property tests
func (suite *PropertyTestSuite) RunPropertyTests(t *testing.T, books []*models.Book) {
	for _, test := range suite.tests {
		t.Run(test.Name, func(t *testing.T) {
			for i, book := range books {
				if !test.Property(book) {
					t.Errorf("Property %s failed for book %d (%s)", test.Name, i, book.Title)
				}
			}
		})
	}
}

// Property test implementations

func (suite *PropertyTestSuite) checkBalancedMarkers(book *models.Book) bool {
	ctx := context.Background()
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	stack := make([]streaming.EventKind, 0)

	for event := range streamer.Stream(ctx, book) {
		switch event.Kind {
		case streaming.StartDoc, streaming.StartParagraph, streaming.StartHeading,
			streaming.StartList, streaming.StartTable, streaming.StartTableRow,
			streaming.StartTableCell, streaming.StartFormatting:
			stack = append(stack, event.Kind)
		case streaming.EndDoc, streaming.EndParagraph, streaming.EndHeading,
			streaming.EndList, streaming.EndTable, streaming.EndTableRow,
			streaming.EndTableCell, streaming.EndFormatting:
			if len(stack) == 0 {
				return false // Unmatched close
			}
			// Check if the close matches the most recent open
			expected := suite.getMatchingStart(event.Kind)
			if stack[len(stack)-1] != expected {
				return false // Mismatched pair
			}
			stack = stack[:len(stack)-1] // Pop
		}
	}

	return len(stack) == 0 // All markers should be balanced
}

func (suite *PropertyTestSuite) checkNoEmptyHeadings(book *models.Book) bool {
	// After sanitization, there should be no empty headings
	sanitizer := sanitizer.NewSanitizer()
	result := sanitizer.Sanitize(book)

	ctx := context.Background()
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	for event := range streamer.Stream(ctx, result.Book) {
		if event.Kind == streaming.StartHeading {
			if strings.TrimSpace(event.HeadingText.Value()) == "" {
				return false
			}
		}
	}

	return true
}

func (suite *PropertyTestSuite) checkValidUTF8(book *models.Book) bool {
	ctx := context.Background()
	streamer := streaming.NewStreamer(streaming.DefaultStreamOptions())

	for event := range streamer.Stream(ctx, book) {
		if event.Kind == streaming.Text {
			if !utf8.ValidString(event.TextContent) {
				return false
			}
		}
	}

	return true
}

func (suite *PropertyTestSuite) getMatchingStart(endKind streaming.EventKind) streaming.EventKind {
	switch endKind {
	case streaming.EndDoc:
		return streaming.StartDoc
	case streaming.EndParagraph:
		return streaming.StartParagraph
	case streaming.EndHeading:
		return streaming.StartHeading
	case streaming.EndList:
		return streaming.StartList
	case streaming.EndTable:
		return streaming.StartTable
	case streaming.EndTableRow:
		return streaming.StartTableRow
	case streaming.EndTableCell:
		return streaming.StartTableCell
	case streaming.EndFormatting:
		return streaming.StartFormatting
	default:
		return streaming.StartDoc // Fallback
	}
}
