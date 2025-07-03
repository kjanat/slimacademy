package writers

import (
	"testing"

	"github.com/kjanat/slimacademy/internal/streaming"
)

// TestMarkdownWriter_ListFormattingRegression tests that formatting markers
// are only applied to list item content, not to the list markers themselves
func TestMarkdownWriter_ListFormattingRegression(t *testing.T) {
	writer := NewMarkdownWriter(nil)

	// Simulate a list with bold formatting
	// Each list item is a separate paragraph with formatting
	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Test Document"},
		{Kind: streaming.StartList, ListOrdered: false},

		// First list item: bold text
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "Bold list item"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.EndParagraph},

		// Second list item: italic text
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "Italic list item"},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.EndParagraph},

		{Kind: streaming.EndList},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// The list markers should NOT have formatting applied
	expectedLines := []string{
		"# Test Document",
		"",
		"- **Bold list item**",
		"",
		"",
		"- _Italic list item_",
		"",
		"",
		"",
	}

	lines := splitLines(result)

	// Check each expected line
	for i, expected := range expectedLines {
		if i >= len(lines) {
			t.Errorf("Missing line %d, expected: %q", i, expected)
			continue
		}
		if lines[i] != expected {
			t.Errorf("Line %d mismatch:\nExpected: %q\nGot:      %q", i, expected, lines[i])
		}
	}

	// Verify we don't have formatting on the list markers
	if containsFormattedMarker(result) {
		t.Error("REGRESSION: List markers should not have formatting applied")
		t.Logf("Full result:\n%s", result)
	}
}

// TestMarkdownWriter_OrderedListFormattingRegression tests ordered lists with formatting
func TestMarkdownWriter_OrderedListFormattingRegression(t *testing.T) {
	writer := NewMarkdownWriter(nil)

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Ordered List Test"},
		{Kind: streaming.StartList, ListOrdered: true},

		// First list item: bold text
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Bold},
		{Kind: streaming.Text, TextContent: "First bold item"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold},
		{Kind: streaming.EndParagraph},

		// Second list item: italic text
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Italic},
		{Kind: streaming.Text, TextContent: "Second italic item"},
		{Kind: streaming.EndFormatting, Style: streaming.Italic},
		{Kind: streaming.EndParagraph},

		{Kind: streaming.EndList},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	expectedLines := []string{
		"# Ordered List Test",
		"",
		"1. **First bold item**",
		"",
		"",
		"2. _Second italic item_",
		"",
		"",
		"",
	}

	lines := splitLines(result)

	for i, expected := range expectedLines {
		if i >= len(lines) {
			t.Errorf("Missing line %d, expected: %q", i, expected)
			continue
		}
		if lines[i] != expected {
			t.Errorf("Line %d mismatch:\nExpected: %q\nGot:      %q", i, expected, lines[i])
		}
	}

	// Verify numbered markers don't have formatting
	if containsFormattedOrderedMarker(result) {
		t.Error("REGRESSION: Ordered list markers should not have formatting applied")
		t.Logf("Full result:\n%s", result)
	}
}

// TestMarkdownWriter_ComplexListFormatting tests complex formatting scenarios
func TestMarkdownWriter_ComplexListFormatting(t *testing.T) {
	writer := NewMarkdownWriter(nil)

	events := []streaming.Event{
		{Kind: streaming.StartDoc, Title: "Complex Test"},
		{Kind: streaming.StartList, ListOrdered: false},

		// First list item: bold and italic
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Bold | streaming.Italic},
		{Kind: streaming.Text, TextContent: "Bold and italic"},
		{Kind: streaming.EndFormatting, Style: streaming.Bold | streaming.Italic},
		{Kind: streaming.EndParagraph},

		// Second list item: link
		{Kind: streaming.StartParagraph},
		{Kind: streaming.StartFormatting, Style: streaming.Link, LinkURL: "https://example.com"},
		{Kind: streaming.Text, TextContent: "Link text"},
		{Kind: streaming.EndFormatting, Style: streaming.Link},
		{Kind: streaming.EndParagraph},

		{Kind: streaming.EndList},
		{Kind: streaming.EndDoc},
	}

	for _, event := range events {
		writer.Handle(event)
	}

	result := writer.Result()

	// Check that complex formatting works correctly on content but not markers
	// Note: bold+italic uses **_text_** format
	if !containsSubstring(result, "- **_Bold and italic_**") {
		t.Errorf("Expected bold+italic formatting not found in result:\n%s", result)
	}
	if !containsSubstring(result, "- [Link text](https://example.com)") {
		t.Errorf("Expected link formatting not found in result:\n%s", result)
	}

	if containsFormattedMarker(result) {
		t.Error("REGRESSION: Complex formatting should not affect list markers")
	}
}

// Helper functions

func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func containsFormattedMarker(text string) bool {
	// Check for common patterns that indicate formatting is applied to markers
	badPatterns := []string{
		"**- ", "**1. ", "*- ", "*1. ", // Bold/italic on markers
		"[- ", "[1. ", // Link formatting on markers
		"~~- ", "~~1. ", // Strikethrough on markers
		"<mark>- ", "<mark>1. ", // Highlight on markers
	}

	for _, pattern := range badPatterns {
		if containsSubstring(text, pattern) {
			return true
		}
	}
	return false
}

func containsFormattedOrderedMarker(text string) bool {
	// Specifically check for formatted numbered markers
	badPatterns := []string{
		"**1. ", "**2. ", "*1. ", "*2. ",
		"[1. ", "[2. ",
		"~~1. ", "~~2. ",
	}

	for _, pattern := range badPatterns {
		if containsSubstring(text, pattern) {
			return true
		}
	}
	return false
}

func containsSubstring(text, substr string) bool {
	return len(text) >= len(substr) && findSubstring(text, substr) >= 0
}

func findSubstring(text, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(text) < len(substr) {
		return -1
	}

	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
