package main

import (
	"testing"
)

// TestSanitizeFilename tests the sanitizeFilename function
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Title", "Normal_Title"},
		{"Title/With\\Slashes", "Title_With_Slashes"},
		{"Title:With*Special?Characters", "Title_With_Special_Characters"},
		{"Title\"With<Various>|Chars", "Title_With_Various_Chars"},
		{"Title__With__Multiple__Underscores", "Title_With_Multiple_Underscores"},
		{"__Title_With_Leading_And_Trailing__", "Title_With_Leading_And_Trailing"},
		{"", ""},
		{"_", ""},
		{"__", ""},
	}

	for _, test := range tests {
		result := sanitizeFilename(test.input)
		if result != test.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestGetFormatExtension tests the getFormatExtension function
func TestGetFormatExtension(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"markdown", "md"},
		{"html", "html"},
		{"latex", "tex"},
		{"epub", "epub"},
		{"plaintext", "txt"},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		result := getFormatExtension(test.format)
		if result != test.expected {
			t.Errorf("getFormatExtension(%q) = %q, expected %q", test.format, result, test.expected)
		}
	}
}
