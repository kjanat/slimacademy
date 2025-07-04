package utils

import (
	"fmt"
	"strings"
	"unicode"
)

// Slugify converts text to a URL-friendly slug using consistent rules across the application.
// It converts to lowercase, replaces spaces with hyphens, and keeps Unicode letters, numbers, and hyphens.
func Slugify(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, " ", "-")

	// Keep Unicode letters, numbers, and hyphens
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SlugifyWithCache generates unique slugs by appending numbers for duplicates.
// The cache maps base slug -> count of times seen.
func SlugifyWithCache(text string, cache map[string]int) string {
	base := Slugify(text)
	if count, exists := cache[base]; exists {
		cache[base] = count + 1
		return fmt.Sprintf("%s-%d", base, count+1)
	}
	cache[base] = 0
	return base
}
