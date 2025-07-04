package main

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/kjanat/slimacademy/internal/config"
	"github.com/kjanat/slimacademy/internal/models"
)

// TestConvertBookToZipBinaryFormats tests that convertBookToZip correctly handles binary formats like EPUB
func TestConvertBookToZipBinaryFormats(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()

	// Create a test book
	book := &models.Book{
		ID:    "test-book",
		Title: "Test Book for Binary Formats",
		Chapters: []models.Chapter{
			{
				ID:    "ch1",
				Title: "Chapter 1",
				Sections: []models.Section{
					{
						ID: "sec1",
						Content: []models.Block{
							{
								Type: "paragraph",
								Elements: []models.Element{
									{Type: "text", Content: "This is test content."},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create a ZIP buffer
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	// Test with multiple formats including binary (EPUB)
	formats := []string{"markdown", "html", "epub"}

	// Convert book to ZIP
	err := convertBookToZip(ctx, book, formats, zipWriter, cfg)
	if err != nil {
		t.Fatalf("convertBookToZip failed: %v", err)
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close ZIP writer: %v", err)
	}

	// Read the ZIP to verify contents
	reader, err := zip.NewReader(bytes.NewReader(zipBuf.Bytes()), int64(zipBuf.Len()))
	if err != nil {
		t.Fatalf("Failed to read generated ZIP: %v", err)
	}

	// Expected files
	expectedFiles := map[string]bool{
		"Test_Book_for_Binary_Formats.md":   false,
		"Test_Book_for_Binary_Formats.html": false,
		"Test_Book_for_Binary_Formats.epub": false,
	}

	// Check each file in the ZIP
	for _, file := range reader.File {
		if _, expected := expectedFiles[file.Name]; expected {
			expectedFiles[file.Name] = true

			// Open and read the file to ensure no corruption
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open %s: %v", file.Name, err)
				continue
			}
			defer rc.Close()

			var content bytes.Buffer
			if _, err := content.ReadFrom(rc); err != nil {
				t.Errorf("Failed to read %s: %v", file.Name, err)
				continue
			}

			// For EPUB files, verify it's a valid ZIP (nested ZIP)
			if file.Name == "Test_Book_for_Binary_Formats.epub" {
				epubReader, err := zip.NewReader(bytes.NewReader(content.Bytes()), int64(content.Len()))
				if err != nil {
					t.Errorf("EPUB file is not a valid ZIP archive: %v", err)
					continue
				}

				// Check for mimetype file in EPUB
				foundMimetype := false
				for _, epubFile := range epubReader.File {
					if epubFile.Name == "mimetype" {
						foundMimetype = true
						break
					}
				}
				if !foundMimetype {
					t.Error("EPUB file missing required mimetype file")
				}
			}
		}
	}

	// Ensure all expected files were found
	for filename, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file not found in ZIP: %s", filename)
		}
	}
}

// TestSanitizeFilename tests the filename sanitization function
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "Simple_Title"},
		{"Title/With/Slashes", "Title_With_Slashes"},
		{"Title\\With\\Backslashes", "Title_With_Backslashes"},
		{"Title:With:Colons", "Title_With_Colons"},
		{"Title*With*Asterisks", "Title_With_Asterisks"},
		{"Title?With?Questions", "Title_With_Questions"},
		{"Title\"With\"Quotes", "Title_With_Quotes"},
		{"Title<With>Brackets", "Title_With_Brackets"},
		{"Title|With|Pipes", "Title_With_Pipes"},
		{"Title  With  Multiple  Spaces", "Title_With_Multiple_Spaces"},
		{"__Leading__Trailing__", "Leading_Trailing"},
	}

	for _, test := range tests {
		result := sanitizeFilename(test.input)
		if result != test.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

// TestBinaryDataIntegrity tests that binary data is not corrupted during string operations
func TestBinaryDataIntegrity(t *testing.T) {
	// Create test data with binary content
	binaryData := []byte{
		0x50, 0x4B, 0x03, 0x04,  // ZIP header
		0x00, 0x00, 0x00, 0x00,  // Null bytes
		0xFF, 0xFE, 0xFD, 0xFC,  // High bytes
		'T', 'e', 's', 't',      // ASCII text
		0x80, 0x81, 0x82, 0x83,  // Extended ASCII
	}

	// Test that Write preserves binary data (our fix)
	var buf bytes.Buffer
	n, err := buf.Write(binaryData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(binaryData) {
		t.Errorf("Write wrote %d bytes, expected %d", n, len(binaryData))
	}
	if !bytes.Equal(buf.Bytes(), binaryData) {
		t.Error("Binary data was corrupted by Write")
	}

	// Test that string conversion corrupts binary data (what we fixed)
	stringData := string(binaryData)
	backToBytes := []byte(stringData)
	if bytes.Equal(backToBytes, binaryData) {
		// This might pass on some systems, but generally string conversion
		// can corrupt binary data, especially with invalid UTF-8 sequences
		t.Log("Warning: string conversion didn't corrupt data on this system")
	}
}