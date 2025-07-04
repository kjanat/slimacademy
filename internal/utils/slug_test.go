package utils

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic text",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with spaces and special chars",
			input:    "Hello, World! How are you?",
			expected: "hello-world-how-are-you",
		},
		{
			name:     "unicode characters",
			input:    "Café München Zürich",
			expected: "café-münchen-zürich",
		},
		{
			name:     "numbers and letters",
			input:    "Chapter 1 Section 2.3",
			expected: "chapter-1-section-23",
		},
		{
			name:     "already hyphenated",
			input:    "pre-formatted-text",
			expected: "pre-formatted-text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSlugifyWithCache(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string
		expected []string
	}{
		{
			name:     "unique inputs",
			inputs:   []string{"Hello", "World", "Test"},
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "duplicate inputs",
			inputs:   []string{"Hello", "Hello", "Hello"},
			expected: []string{"hello", "hello-1", "hello-2"},
		},
		{
			name:     "mixed duplicates",
			inputs:   []string{"Chapter", "Section", "Chapter", "Section", "Chapter"},
			expected: []string{"chapter", "section", "chapter-1", "section-1", "chapter-2"},
		},
		{
			name:     "empty and special cases",
			inputs:   []string{"", "Test", "", "Test"},
			expected: []string{"", "test", "-1", "test-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := make(map[string]int)
			for i, input := range tt.inputs {
				result := SlugifyWithCache(input, cache)
				if result != tt.expected[i] {
					t.Errorf("SlugifyWithCache(%q) = %q, want %q (iteration %d)",
						input, result, tt.expected[i], i)
				}
			}
		})
	}
}

func TestSlugifyWithCache_ConsistentCaching(t *testing.T) {
	cache := make(map[string]int)

	// First occurrence should not have suffix
	result1 := SlugifyWithCache("Test Heading", cache)
	if result1 != "test-heading" {
		t.Errorf("First occurrence: got %q, want %q", result1, "test-heading")
	}

	// Second occurrence should have -1 suffix
	result2 := SlugifyWithCache("Test Heading", cache)
	if result2 != "test-heading-1" {
		t.Errorf("Second occurrence: got %q, want %q", result2, "test-heading-1")
	}

	// Third occurrence should have -2 suffix
	result3 := SlugifyWithCache("Test Heading", cache)
	if result3 != "test-heading-2" {
		t.Errorf("Third occurrence: got %q, want %q", result3, "test-heading-2")
	}

	// Verify cache state
	if cache["test-heading"] != 2 {
		t.Errorf("Cache count: got %d, want %d", cache["test-heading"], 2)
	}
}
