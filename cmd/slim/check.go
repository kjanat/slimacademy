package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/kjanat/slimacademy/internal/sanitizer"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [input]",
	Short: "Check document for issues (sanitizer only)",
	Long: `Validate and check a SlimAcademy book for common issues.

This command parses the book and runs it through the sanitizer to identify:
- Invalid UTF-8 sequences
- Dangerous HTML content
- Malformed URLs
- Empty headings
- Structural inconsistencies

The sanitizer provides warnings for issues that could affect conversion quality.

Examples:
  slim check book1                    # Check single book
  slim check --verbose book1          # Check with detailed output`,

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck(context.Background(), args[0])
	},
}

func runCheck(ctx context.Context, inputPath string) error {
	logger := slog.Default().With("command", "check", "input", inputPath)
	logger.Info("Starting book validation")

	// Parse book
	parser := parser.NewBookParser()
	book, err := parser.ParseBook(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse book: %w", err)
	}

	logger.Debug("Book parsed successfully", "title", book.Title, "chapters", len(book.Chapters))

	// Create sanitizer and check the book
	s := sanitizer.NewSanitizer()

	// Sanitize the book (this will generate warnings for issues)
	result := s.Sanitize(book)

	// Get warnings from result
	warnings := result.Warnings

	// Report results
	fmt.Printf("Book: %s\n", book.Title)
	fmt.Printf("Chapters: %d\n", len(book.Chapters))

	if len(warnings) == 0 {
		fmt.Println("✅ No issues found")
		logger.Info("Validation completed successfully", "warnings", 0)
	} else {
		fmt.Printf("⚠️  Found %d warnings:\n", len(warnings))

		for i, warning := range warnings {
			fmt.Printf("%d. %s\n", i+1, warning.Issue)
			if warning.Location != "" {
				fmt.Printf("   Location: %s\n", warning.Location)
			}
			if verbose && warning.Original != "" {
				fmt.Printf("   Original: %s\n", warning.Original)
				if warning.Fixed != "" {
					fmt.Printf("   Fixed: %s\n", warning.Fixed)
				}
			}
		}

		logger.Warn("Validation found issues", "warnings", len(warnings))
	}

	// Show differences if verbose
	if verbose && result.Book != nil {
		if book.Title != result.Book.Title {
			fmt.Printf("Title changed: %q → %q\n", book.Title, result.Book.Title)
		}
		if book.Description != result.Book.Description {
			fmt.Printf("Description changed: %q → %q\n", book.Description, result.Book.Description)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
