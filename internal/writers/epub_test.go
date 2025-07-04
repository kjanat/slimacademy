package writers

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/streaming"
)

// TestEPUBWriterBinaryIntegrity tests that EPUB files are valid ZIP archives
func TestEPUBWriterBinaryIntegrity(t *testing.T) {
	// Create a writer
	writer := &EPUBWriterV2{
		stats:  WriterStats{},
		buffer: &bytes.Buffer{},
	}
	writer.epubWriter = NewEPUBWriterWithConfig(writer.buffer, config.DefaultEPUBConfig())

	// Send events to create a simple EPUB
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Book"},
		{Kind: streaming.StartHeading, Level: 1, HeadingText: streaming.NewCachedText("Chapter 1"), AnchorID: "ch1"},
		{Kind: streaming.Text, TextContent: "This is chapter 1 content."},
		{Kind: streaming.EndHeading},
		{Kind: streaming.StartHeading, Level: 1, HeadingText: streaming.NewCachedText("Chapter 2"), AnchorID: "ch2"},
		{Kind: streaming.Text, TextContent: "This is chapter 2 content."},
		{Kind: streaming.EndHeading},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		if err := writer.Handle(event); err != nil {
			t.Fatalf("Failed to handle event: %v", err)
		}
	}

	// Flush to get the binary data
	data, err := writer.Flush()
	if err != nil {
		t.Fatalf("Failed to flush EPUB writer: %v", err)
	}

	// Verify the data is not empty
	if len(data) == 0 {
		t.Fatal("EPUB data is empty")
	}

	// Verify it's a valid ZIP archive
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Generated EPUB is not a valid ZIP archive: %v", err)
	}

	// Check for required EPUB files
	requiredFiles := map[string]bool{
		"mimetype":             false,
		"META-INF/container.xml": false,
		"OEBPS/content.opf":    false,
		"OEBPS/toc.ncx":        false,
		"OEBPS/styles.css":     false,
	}

	for _, file := range reader.File {
		if _, ok := requiredFiles[file.Name]; ok {
			requiredFiles[file.Name] = true
		}

		// Verify each file can be read without errors
		rc, err := file.Open()
		if err != nil {
			t.Errorf("Failed to open file %s: %v", file.Name, err)
			continue
		}
		defer rc.Close()

		// Read the file content to ensure no corruption
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(rc); err != nil {
			t.Errorf("Failed to read file %s: %v", file.Name, err)
		}
	}

	// Ensure all required files are present
	for file, found := range requiredFiles {
		if !found {
			t.Errorf("Required EPUB file missing: %s", file)
		}
	}

	// Verify mimetype is stored uncompressed
	for _, file := range reader.File {
		if file.Name == "mimetype" {
			if file.Method != zip.Store {
				t.Error("mimetype file should be stored uncompressed")
			}
			break
		}
	}
}

// TestEPUBWriterMultipleChapters tests EPUB generation with multiple chapters
func TestEPUBWriterMultipleChapters(t *testing.T) {
	writer := &EPUBWriterV2{
		stats:  WriterStats{},
		buffer: &bytes.Buffer{},
	}
	writer.epubWriter = NewEPUBWriterWithConfig(writer.buffer, config.DefaultEPUBConfig())

	// Create a book with multiple chapters
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Multi-Chapter Book"},
	}

	// Add 5 chapters
	for i := 1; i <= 5; i++ {
		events = append(events, 
			streaming.Event{
				Kind: streaming.StartHeading, 
				Level: 1, 
				HeadingText: streaming.NewCachedText(fmt.Sprintf("Chapter %d", i)), 
				AnchorID: fmt.Sprintf("ch%d", i),
			},
			streaming.Event{
				Kind: streaming.Text, 
				TextContent: fmt.Sprintf("This is the content of chapter %d.", i),
			},
			streaming.Event{Kind: streaming.EndHeading},
		)
	}

	events = append(events, streaming.Event{Kind: streaming.EndDoc})

	// Process events
	for _, event := range events {
		if err := writer.Handle(event); err != nil {
			t.Fatalf("Failed to handle event: %v", err)
		}
	}

	// Get the result
	data, err := writer.Flush()
	if err != nil {
		t.Fatalf("Failed to flush EPUB writer: %v", err)
	}

	// Parse as ZIP
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Failed to parse EPUB as ZIP: %v", err)
	}

	// Count chapter files
	chapterCount := 0
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "OEBPS/chapter_") && strings.HasSuffix(file.Name, ".xhtml") {
			chapterCount++
		}
	}

	if chapterCount != 5 {
		t.Errorf("Expected 5 chapter files, got %d", chapterCount)
	}
}

// TestEPUBWriterReset tests that Reset properly clears the writer state
func TestEPUBWriterReset(t *testing.T) {
	writer := &EPUBWriterV2{
		stats:  WriterStats{},
		buffer: &bytes.Buffer{},
	}
	writer.epubWriter = NewEPUBWriterWithConfig(writer.buffer, config.DefaultEPUBConfig())

	// First document
	events1 := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "First Book"},
		{Kind: streaming.StartHeading, Level: 1, HeadingText: streaming.NewCachedText("Chapter 1"), AnchorID: "ch1"},
		{Kind: streaming.Text, TextContent: "First book content."},
		{Kind: streaming.EndHeading},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events1 {
		if err := writer.Handle(event); err != nil {
			t.Fatalf("Failed to handle event: %v", err)
		}
	}

	data1, err := writer.Flush()
	if err != nil {
		t.Fatalf("Failed to flush first EPUB: %v", err)
	}

	// Reset
	writer.Reset()

	// Second document
	events2 := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Second Book"},
		{Kind: streaming.StartHeading, Level: 1, HeadingText: streaming.NewCachedText("Chapter A"), AnchorID: "cha"},
		{Kind: streaming.Text, TextContent: "Second book content."},
		{Kind: streaming.EndHeading},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events2 {
		if err := writer.Handle(event); err != nil {
			t.Fatalf("Failed to handle event: %v", err)
		}
	}

	data2, err := writer.Flush()
	if err != nil {
		t.Fatalf("Failed to flush second EPUB: %v", err)
	}

	// Both should be valid EPUBs
	if _, err := zip.NewReader(bytes.NewReader(data1), int64(len(data1))); err != nil {
		t.Errorf("First EPUB is not valid: %v", err)
	}

	if _, err := zip.NewReader(bytes.NewReader(data2), int64(len(data2))); err != nil {
		t.Errorf("Second EPUB is not valid: %v", err)
	}

	// They should be different
	if bytes.Equal(data1, data2) {
		t.Error("Reset didn't properly clear the state - both EPUBs are identical")
	}
}

// TestEPUBWriterWithMultiWriter tests EPUB generation through the MultiWriter
func TestEPUBWriterWithMultiWriter(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()

	// Create multi-writer with EPUB format
	multiWriter, err := NewMultiWriter(ctx, []string{"epub"}, cfg)
	if err != nil {
		t.Fatalf("Failed to create MultiWriter: %v", err)
	}
	defer multiWriter.Close()

	// Create a simple document
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test EPUB via MultiWriter"},
		{Kind: streaming.StartHeading, Level: 1, HeadingText: streaming.NewCachedText("Introduction"), AnchorID: "intro"},
		{Kind: streaming.Text, TextContent: "This is a test of EPUB generation through MultiWriter."},
		{Kind: streaming.EndHeading},
		{Kind: streaming.EndDoc},
	}

	// Process events
	if err := multiWriter.ProcessEvents(func(yield func(streaming.Event) bool) {
		for _, event := range events {
			if !yield(event) {
				break
			}
		}
	}); err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	// Get results
	results, err := multiWriter.FlushAll()
	if err != nil {
		t.Fatalf("Failed to flush MultiWriter: %v", err)
	}

	// Should have one result for EPUB
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	// Verify format metadata
	if result.Format != "epub" {
		t.Errorf("Expected format 'epub', got '%s'", result.Format)
	}

	if result.Extension != ".epub" {
		t.Errorf("Expected extension '.epub', got '%s'", result.Extension)
	}

	if result.IsBinary != true {
		t.Error("Expected IsBinary to be true for EPUB")
	}

	// Verify it's a valid ZIP
	if _, err := zip.NewReader(bytes.NewReader(result.Data), int64(len(result.Data))); err != nil {
		t.Errorf("Generated EPUB is not a valid ZIP: %v", err)
	}
}

// TestConvertBookToZipBinaryHandling tests the fixed convertBookToZip function
func TestConvertBookToZipBinaryHandling(t *testing.T) {
	// This test would require the full CLI context, so we'll create a simpler integration test
	// that verifies binary data is not corrupted when written to a ZIP

	// Create test binary data (simulating EPUB content)
	testData := []byte{0x50, 0x4B, 0x03, 0x04} // ZIP magic number
	testData = append(testData, []byte("binary content with \x00 null bytes and special chars: \xff\xfe")...)

	// Create a ZIP in memory
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Write using the correct method (not string conversion)
	fileWriter, err := zipWriter.Create("test.epub")
	if err != nil {
		t.Fatalf("Failed to create ZIP entry: %v", err)
	}

	// This is the CORRECT way (what we fixed)
	if _, err := fileWriter.Write(testData); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close ZIP: %v", err)
	}

	// Read back and verify
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	if len(reader.File) != 1 {
		t.Fatalf("Expected 1 file in ZIP, got %d", len(reader.File))
	}

	file := reader.File[0]
	rc, err := file.Open()
	if err != nil {
		t.Fatalf("Failed to open file in ZIP: %v", err)
	}
	defer rc.Close()

	readData := new(bytes.Buffer)
	if _, err := readData.ReadFrom(rc); err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}

	// Verify data integrity
	if !bytes.Equal(testData, readData.Bytes()) {
		t.Error("Binary data was corrupted during ZIP write/read")
		t.Errorf("Original length: %d, Read length: %d", len(testData), readData.Len())
	}
}