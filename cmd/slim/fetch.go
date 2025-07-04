package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kjanat/slimacademy/internal/client"
	"github.com/spf13/cobra"
)

var (
	// Fetch command flags
	fetchAll    bool
	fetchLogin  bool
	fetchBookID string
	fetchOutput string
	fetchClean  bool
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch book data from Slim Academy API",
	Long: `Fetch book data from the Slim Academy API.

This command can authenticate with the API, fetch individual books by ID,
or download all available books from your library.

Examples:
  slim fetch --login                          # Login and save authentication
  slim fetch --all                            # Fetch all books to source/
  slim fetch --id 3631                        # Fetch specific book by ID
  slim fetch --all --output data/ --clean     # Fetch all books to data/, clean first`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return runFetch(context.Background())
	},
}

func runFetch(ctx context.Context) error {
	logger := slog.Default().With("command", "fetch")

	// Validate options
	if !fetchLogin && !fetchAll && fetchBookID == "" {
		return fmt.Errorf("specify --login, --all, or --id <id>")
	}

	if fetchLogin && (fetchAll || fetchBookID != "") {
		return fmt.Errorf("--login cannot be combined with other fetch options")
	}

	if fetchAll && fetchBookID != "" {
		return fmt.Errorf("--all and --id cannot be used together")
	}

	// Set default output directory
	outputDir := fetchOutput
	if outputDir == "" {
		outputDir = "source"
	}

	// Create API client
	apiClient := client.NewSlimClient(outputDir)

	// Handle login-only mode
	if fetchLogin {
		logger.Info("Authenticating with Slim Academy API")
		if err := apiClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Get token info for confirmation
		tokenInfo, err := apiClient.GetTokenInfo()
		if err != nil {
			logger.Warn("Login succeeded but could not retrieve token info", "error", err)
		} else {
			logger.Info("Login successful", "username", tokenInfo.Username, "expires_at", tokenInfo.ExpiresAt)
			fmt.Printf("✅ Successfully logged in as %s\n", tokenInfo.Username)
			fmt.Printf("Token expires at: %s\n", tokenInfo.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	}

	// Clean output directory if requested
	if fetchClean {
		logger.Info("Cleaning output directory", "dir", outputDir)
		if err := os.RemoveAll(outputDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Handle fetch all books
	if fetchAll {
		logger.Info("Fetching all books from library")
		books, err := apiClient.FetchLibraryBooks(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch library books: %w", err)
		}

		logger.Info("Found books in library", "count", len(books))
		fmt.Printf("Found %d books in library\n", len(books))

		for i, book := range books {
			logger.Info("Processing book", "index", i+1, "total", len(books), "id", book.ID, "title", book.Title)
			fmt.Printf("(%d/%d) Fetching book: %s\n", i+1, len(books), book.Title)

			if err := writeBookFiles(outputDir, book, logger); err != nil {
				logger.Warn("Failed to write book files", "id", book.ID, "title", book.Title, "error", err)
				fmt.Printf("  ⚠️  Failed to write files for book %s: %v\n", book.Title, err)
				continue
			}

			fmt.Printf("  ✅ Successfully saved book %s\n", book.Title)
		}

		fmt.Printf("\n✅ Successfully fetched %d books to %s/\n", len(books), outputDir)
		return nil
	}

	// Handle fetch single book by ID
	if fetchBookID != "" {
		logger.Info("Fetching single book", "id", fetchBookID)
		fmt.Printf("Fetching book ID: %s\n", fetchBookID)

		book, err := apiClient.FetchAllBookData(ctx, fetchBookID)
		if err != nil {
			return fmt.Errorf("failed to fetch book %s: %w", fetchBookID, err)
		}

		logger.Info("Book fetched successfully", "id", book.ID, "title", book.Title)

		if err := writeBookFiles(outputDir, book, logger); err != nil {
			return fmt.Errorf("failed to write book files: %w", err)
		}

		fmt.Printf("✅ Successfully saved book: %s\n", book.Title)
		fmt.Printf("Files written to: %s/%s/\n", outputDir, book.ID)
		return nil
	}

	return fmt.Errorf("internal error: no valid fetch mode selected")
}

// writeBookFiles writes the book data to the expected file structure
func writeBookFiles(outputDir string, book *client.BookData, logger *slog.Logger) error {
	// Create book directory
	bookDir := filepath.Join(outputDir, book.ID)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("failed to create book directory: %w", err)
	}

	// Write summary file ({id}.json)
	summaryFile := filepath.Join(bookDir, book.ID+".json")
	summaryData, err := json.MarshalIndent(book.Summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}
	if err := os.WriteFile(summaryFile, summaryData, 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}
	logger.Debug("Wrote summary file", "path", summaryFile, "size", len(summaryData))

	// Write chapters file
	chaptersFile := filepath.Join(bookDir, "chapters.json")
	chaptersData, err := json.MarshalIndent(book.Chapters, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chapters: %w", err)
	}
	if err := os.WriteFile(chaptersFile, chaptersData, 0644); err != nil {
		return fmt.Errorf("failed to write chapters file: %w", err)
	}
	logger.Debug("Wrote chapters file", "path", chaptersFile, "size", len(chaptersData))

	// Write content file
	contentFile := filepath.Join(bookDir, "content.json")
	contentData, err := json.MarshalIndent(book.Content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}
	if err := os.WriteFile(contentFile, contentData, 0644); err != nil {
		return fmt.Errorf("failed to write content file: %w", err)
	}
	logger.Debug("Wrote content file", "path", contentFile, "size", len(contentData))

	// Write notes file (optional, may be empty)
	if book.Notes != nil {
		notesFile := filepath.Join(bookDir, "notes.json")
		notesData, err := json.MarshalIndent(book.Notes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal notes: %w", err)
		}
		if err := os.WriteFile(notesFile, notesData, 0644); err != nil {
			return fmt.Errorf("failed to write notes file: %w", err)
		}
		logger.Debug("Wrote notes file", "path", notesFile, "size", len(notesData))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Fetch-specific flags
	fetchCmd.Flags().BoolVar(&fetchLogin, "login", false, "Login only (authenticate and save token)")
	fetchCmd.Flags().BoolVar(&fetchAll, "all", false, "Fetch all books from library")
	fetchCmd.Flags().StringVar(&fetchBookID, "id", "", "Fetch specific book by ID")
	fetchCmd.Flags().StringVarP(&fetchOutput, "output", "o", "source", "Output directory")
	fetchCmd.Flags().BoolVar(&fetchClean, "clean", false, "Clean output directory before fetching")
}
