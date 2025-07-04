package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kjanat/slimacademy/internal/parser"
	"github.com/spf13/cobra"
)

var (
	// List command flags
	listAll bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List available books",
	Long: `List all SlimAcademy books found in the specified directory.

By default, searches the 'source' directory for books. You can specify
a different directory path as an argument.

Examples:
  slim list                     # List books in ./source/
  slim list /path/to/books/     # List books in specific directory
  slim list --all               # Show detailed information about each book`,

	Args: cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir := "source"
		if len(args) > 0 {
			rootDir = args[0]
		}

		return runList(context.Background(), rootDir)
	},
}

func runList(ctx context.Context, rootDir string) error {
	logger := slog.Default().With("command", "list", "directory", rootDir)
	logger.Info("Listing books")

	parser := parser.NewBookParser()
	books, err := parser.FindAllBooks(rootDir)
	if err != nil {
		return fmt.Errorf("failed to find books: %w", err)
	}

	if len(books) == 0 {
		fmt.Println("No books found")
		return nil
	}

	fmt.Printf("Found %d books:\n", len(books))

	if listAll {
		// Show detailed information
		for i, bookPath := range books {
			book, err := parser.ParseBook(bookPath)
			if err != nil {
				fmt.Printf("  %d. %s (error: %v)\n", i+1, bookPath, err)
				continue
			}

			fmt.Printf("  %d. %s\n", i+1, book.Title)
			fmt.Printf("     Path: %s\n", bookPath)
			if book.Description != "" {
				fmt.Printf("     Description: %s\n", book.Description)
			}
			fmt.Printf("     Chapters: %d\n", len(book.Chapters))
			if book.PageCount > 0 {
				fmt.Printf("     Pages: %d\n", book.PageCount)
			}
			if len(book.Periods) > 0 {
				fmt.Printf("     Periods: %v\n", book.Periods)
			}
			if book.BachelorYearNumber != "" {
				fmt.Printf("     Academic Year: %s\n", book.BachelorYearNumber)
			}
			fmt.Println()
		}
	} else {
		// Simple list
		for _, bookPath := range books {
			fmt.Printf("  %s\n", bookPath)
		}
	}

	logger.Info("Book listing completed", "count", len(books))
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)

	// List-specific flags
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show detailed information about each book")
}
