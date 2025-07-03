package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kjanat/slimacademy/internal/client"
)

// Test helper to create temporary directory
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "slim-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// Test helper to create test book structure
func createTestBook(t *testing.T, dir, bookID, title string) string {
	bookDir := filepath.Join(dir, fmt.Sprintf("test-book-%s", bookID))
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("Failed to create book directory: %v", err)
	}

	// Create summary JSON
	bookIDInt, _ := strconv.ParseInt(bookID, 10, 64)
	summary := map[string]any{
		"id":          bookIDInt,
		"title":       title,
		"description": "Test book description",
	}
	summaryJSON, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(bookDir, fmt.Sprintf("%s.json", bookID)), summaryJSON, 0644); err != nil {
		t.Fatalf("Failed to create summary file: %v", err)
	}

	// Create chapters JSON - needs to match Chapter model structure
	chapters := []map[string]any{
		{
			"id":              1,
			"summaryId":       bookIDInt,
			"title":           "Chapter 1",
			"isFree":          1,
			"isSupplement":    0,
			"isLocked":        0,
			"isVisible":       1,
			"parentChapterId": nil,
			"gDocsChapterId":  "chapter1-gdocs-id",
			"sortIndex":       1,
			"subChapters":     []any{},
		},
		{
			"id":              2,
			"summaryId":       bookIDInt,
			"title":           "Chapter 2",
			"isFree":          1,
			"isSupplement":    0,
			"isLocked":        0,
			"isVisible":       1,
			"parentChapterId": nil,
			"gDocsChapterId":  "chapter2-gdocs-id",
			"sortIndex":       2,
			"subChapters":     []any{},
		},
	}
	chaptersJSON, _ := json.Marshal(chapters)
	if err := os.WriteFile(filepath.Join(bookDir, "chapters.json"), chaptersJSON, 0644); err != nil {
		t.Fatalf("Failed to create chapters file: %v", err)
	}

	// Create content JSON
	content := map[string]any{
		"chapters": chapters,
		"metadata": map[string]any{
			"total_chapters": 2,
		},
	}
	contentJSON, _ := json.Marshal(content)
	if err := os.WriteFile(filepath.Join(bookDir, "content.json"), contentJSON, 0644); err != nil {
		t.Fatalf("Failed to create content file: %v", err)
	}

	return bookDir
}

// Test helper to capture stdout
func captureOutput(t *testing.T, fn func()) string {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = oldStdout
	return <-outC
}

// Test helper to capture stderr
func captureStderr(t *testing.T, fn func()) string {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stderr = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()

	w.Close()
	os.Stderr = oldStderr
	return <-outC
}

func TestPrintUsage(t *testing.T) {
	output := captureOutput(t, func() {
		printUsage()
	})

	expectedStrings := []string{
		"slim - Document transformation tool",
		"slim convert",
		"slim check",
		"slim list",
		"slim fetch",
		"--all",
		"--format",
		"--output",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected usage output to contain %q, but it didn't.\nOutput: %s", expected, output)
		}
	}
}

func TestParseConvertOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expected    *ConvertOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "simple input path",
			args: []string{"book1"},
			expected: &ConvertOptions{
				All:       false,
				Formats:   []string{"markdown"},
				InputPath: "book1",
			},
		},
		{
			name: "all flag",
			args: []string{"--all"},
			expected: &ConvertOptions{
				All:     true,
				Formats: []string{"markdown"},
			},
		},
		{
			name: "single format",
			args: []string{"--format", "html", "book1"},
			expected: &ConvertOptions{
				All:       false,
				Formats:   []string{"html"},
				InputPath: "book1",
			},
		},
		{
			name: "multiple formats",
			args: []string{"--formats", "html,latex,epub", "book1"},
			expected: &ConvertOptions{
				All:       false,
				Formats:   []string{"html", "latex", "epub"},
				InputPath: "book1",
			},
		},
		{
			name: "with output path",
			args: []string{"--output", "/tmp/output", "book1"},
			expected: &ConvertOptions{
				All:        false,
				Formats:    []string{"markdown"},
				InputPath:  "book1",
				OutputPath: "/tmp/output",
			},
		},
		{
			name: "with config path",
			args: []string{"--config", "/tmp/config.yaml", "book1"},
			expected: &ConvertOptions{
				All:        false,
				Formats:    []string{"markdown"},
				InputPath:  "book1",
				ConfigPath: "/tmp/config.yaml",
			},
		},
		{
			name: "all options combined",
			args: []string{"--formats", "html,latex", "--output", "/tmp/out", "--config", "/tmp/cfg", "book1"},
			expected: &ConvertOptions{
				All:        false,
				Formats:    []string{"html", "latex"},
				InputPath:  "book1",
				OutputPath: "/tmp/out",
				ConfigPath: "/tmp/cfg",
			},
		},
		{
			name:        "no input path and no all",
			args:        []string{"--format", "html"},
			expectError: true,
			errorMsg:    "input path is required",
		},
		{
			name:        "format without value",
			args:        []string{"--format"},
			expectError: true,
			errorMsg:    "--format requires a value",
		},
		{
			name:        "formats without value",
			args:        []string{"--formats"},
			expectError: true,
			errorMsg:    "--formats requires a value",
		},
		{
			name:        "output without value",
			args:        []string{"--output"},
			expectError: true,
			errorMsg:    "--output requires a value",
		},
		{
			name:        "config without value",
			args:        []string{"--config"},
			expectError: true,
			errorMsg:    "--config requires a value",
		},
		{
			name:        "unknown option",
			args:        []string{"--unknown", "book1"},
			expectError: true,
			errorMsg:    "unknown option: --unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseConvertOptions(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.All != tt.expected.All {
				t.Errorf("Expected All=%v, got %v", tt.expected.All, result.All)
			}

			if len(result.Formats) != len(tt.expected.Formats) {
				t.Errorf("Expected %d formats, got %d", len(tt.expected.Formats), len(result.Formats))
			} else {
				for i, format := range tt.expected.Formats {
					if result.Formats[i] != format {
						t.Errorf("Expected format[%d]=%q, got %q", i, format, result.Formats[i])
					}
				}
			}

			if result.InputPath != tt.expected.InputPath {
				t.Errorf("Expected InputPath=%q, got %q", tt.expected.InputPath, result.InputPath)
			}

			if result.OutputPath != tt.expected.OutputPath {
				t.Errorf("Expected OutputPath=%q, got %q", tt.expected.OutputPath, result.OutputPath)
			}

			if result.ConfigPath != tt.expected.ConfigPath {
				t.Errorf("Expected ConfigPath=%q, got %q", tt.expected.ConfigPath, result.ConfigPath)
			}
		})
	}
}

func TestParseFetchOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expected    *FetchOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "login only",
			args: []string{"--login"},
			expected: &FetchOptions{
				LoginOnly: true,
				OutputDir: "source",
			},
		},
		{
			name: "fetch all",
			args: []string{"--all"},
			expected: &FetchOptions{
				All:       true,
				OutputDir: "source",
			},
		},
		{
			name: "fetch by ID",
			args: []string{"--id", "3631"},
			expected: &FetchOptions{
				BookID:    "3631",
				OutputDir: "source",
			},
		},
		{
			name: "with custom output dir",
			args: []string{"--all", "--output", "/tmp/data"},
			expected: &FetchOptions{
				All:       true,
				OutputDir: "/tmp/data",
			},
		},
		{
			name: "with clean flag",
			args: []string{"--all", "--clean"},
			expected: &FetchOptions{
				All:       true,
				Clean:     true,
				OutputDir: "source",
			},
		},
		{
			name: "all options combined",
			args: []string{"--id", "1234", "--output", "/tmp/out", "--clean"},
			expected: &FetchOptions{
				BookID:    "1234",
				OutputDir: "/tmp/out",
				Clean:     true,
			},
		},
		{
			name:        "no action specified",
			args:        []string{},
			expectError: true,
			errorMsg:    "specify --login, --all, or --id <id>",
		},
		{
			name:        "login with other options",
			args:        []string{"--login", "--all"},
			expectError: true,
			errorMsg:    "--login cannot be combined with other fetch options",
		},
		{
			name:        "login with ID",
			args:        []string{"--login", "--id", "123"},
			expectError: true,
			errorMsg:    "--login cannot be combined with other fetch options",
		},
		{
			name:        "all with ID",
			args:        []string{"--all", "--id", "123"},
			expectError: true,
			errorMsg:    "--all and --id cannot be used together",
		},
		{
			name:        "id without value",
			args:        []string{"--id"},
			expectError: true,
			errorMsg:    "--id requires a value",
		},
		{
			name:        "output without value",
			args:        []string{"--output"},
			expectError: true,
			errorMsg:    "--output requires a value",
		},
		{
			name:        "unknown option",
			args:        []string{"--unknown"},
			expectError: true,
			errorMsg:    "unknown option: --unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFetchOptions(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.LoginOnly != tt.expected.LoginOnly {
				t.Errorf("Expected LoginOnly=%v, got %v", tt.expected.LoginOnly, result.LoginOnly)
			}

			if result.All != tt.expected.All {
				t.Errorf("Expected All=%v, got %v", tt.expected.All, result.All)
			}

			if result.BookID != tt.expected.BookID {
				t.Errorf("Expected BookID=%q, got %q", tt.expected.BookID, result.BookID)
			}

			if result.OutputDir != tt.expected.OutputDir {
				t.Errorf("Expected OutputDir=%q, got %q", tt.expected.OutputDir, result.OutputDir)
			}

			if result.Clean != tt.expected.Clean {
				t.Errorf("Expected Clean=%v, got %v", tt.expected.Clean, result.Clean)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "Simple Title"},
		{"Title/With/Slashes", "Title-With-Slashes"},
		{"Title\\With\\Backslashes", "Title-With-Backslashes"},
		{"Title:With:Colons", "Title-With-Colons"},
		{"Title*With*Asterisks", "Title-With-Asterisks"},
		{"Title?With?Questions", "Title-With-Questions"},
		{"Title\"With\"Quotes", "Title-With-Quotes"},
		{"Title<With>Brackets", "Title-With-Brackets"},
		{"Title|With|Pipes", "Title-With-Pipes"},
		{"  Title With Spaces  ", "Title With Spaces"},
		{"", ""},
		{"All/\\:*?\"<>|Bad", "All---------Bad"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sanitize_%s", tt.input), func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"markdown", "md"},
		{"html", "html"},
		{"latex", "tex"},
		{"epub", "epub"},
		{"plaintext", "txt"},
		{"unknown", "txt"},
		{"", "txt"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := getExtension(tt.format)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	tempDir := createTempDir(t)
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create test books
	createTestBook(t, sourceDir, "1001", "First Book")
	createTestBook(t, sourceDir, "1002", "Second Book")

	ctx := context.Background()

	t.Run("list default directory", func(t *testing.T) {
		// Change to temp directory so "source" points to our test data
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		output := captureOutput(t, func() {
			err := handleList(ctx, []string{})
			if err != nil {
				t.Errorf("handleList failed: %v", err)
			}
		})

		if !strings.Contains(output, "Found 2 books:") {
			t.Errorf("Expected to find 2 books, got output: %s", output)
		}
		if !strings.Contains(output, "test-book-1001") {
			t.Errorf("Expected to find test-book-1001 in output: %s", output)
		}
		if !strings.Contains(output, "test-book-1002") {
			t.Errorf("Expected to find test-book-1002 in output: %s", output)
		}
	})

	t.Run("list specific directory", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := handleList(ctx, []string{sourceDir})
			if err != nil {
				t.Errorf("handleList failed: %v", err)
			}
		})

		if !strings.Contains(output, "Found 2 books:") {
			t.Errorf("Expected to find 2 books, got output: %s", output)
		}
	})

	t.Run("list empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		output := captureOutput(t, func() {
			err := handleList(ctx, []string{emptyDir})
			if err != nil {
				t.Errorf("handleList failed: %v", err)
			}
		})

		if !strings.Contains(output, "No books found") {
			t.Errorf("Expected 'No books found', got output: %s", output)
		}
	})

	t.Run("list nonexistent directory", func(t *testing.T) {
		err := handleList(ctx, []string{"/nonexistent/directory"})
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
		if !strings.Contains(err.Error(), "failed to find books") {
			t.Errorf("Expected 'failed to find books' error, got: %v", err)
		}
	})
}

func TestHandleCheck(t *testing.T) {
	tempDir := createTempDir(t)
	bookPath := createTestBook(t, tempDir, "1001", "Test Book")

	ctx := context.Background()

	t.Run("check valid book", func(t *testing.T) {
		output := captureOutput(t, func() {
			err := handleCheck(ctx, []string{bookPath})
			if err != nil {
				t.Errorf("handleCheck failed: %v", err)
			}
		})

		// The output depends on the sanitizer implementation
		// At minimum, it should not error and produce some output
		if len(output) == 0 {
			t.Error("Expected some output from check command")
		}
	})

	t.Run("check nonexistent book", func(t *testing.T) {
		err := handleCheck(ctx, []string{"/nonexistent/book"})
		if err == nil {
			t.Error("Expected error for nonexistent book")
		}
		if !strings.Contains(err.Error(), "failed to parse book") {
			t.Errorf("Expected 'failed to parse book' error, got: %v", err)
		}
	})

	t.Run("check missing argument", func(t *testing.T) {
		err := handleCheck(ctx, []string{})
		if err == nil {
			t.Error("Expected error for missing argument")
		}
		if !strings.Contains(err.Error(), "check command requires input path") {
			t.Errorf("Expected 'requires input path' error, got: %v", err)
		}
	})
}

func TestHandleFetch_LoginOnly(t *testing.T) {
	ctx := context.Background()

	// Test login-only mode
	t.Run("login only argument parsing", func(t *testing.T) {
		// We'll test the argument parsing and basic flow
		// This will fail at the API call stage but that's expected
		err := handleFetch(ctx, []string{"--login"})
		// This should fail because it tries to make real API calls
		// But we can verify the error is not from argument parsing
		if err != nil && !strings.Contains(err.Error(), "login failed") && !strings.Contains(err.Error(), "failed to load credentials") {
			t.Errorf("Expected login or credential error, got: %v", err)
		}
	})
}

func TestHandleFetch_ArgumentValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		args     []string
		errorMsg string
	}{
		{
			name:     "no arguments",
			args:     []string{},
			errorMsg: "specify --login, --all, or --id <id>",
		},
		{
			name:     "login with all",
			args:     []string{"--login", "--all"},
			errorMsg: "--login cannot be combined with other fetch options",
		},
		{
			name:     "all with id",
			args:     []string{"--all", "--id", "123"},
			errorMsg: "--all and --id cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleFetch(ctx, tt.args)
			if err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

// Integration test for main function dispatch
func TestMainCommandDispatch(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "no arguments",
			args:           []string{"slim"},
			expectError:    true,
			expectedOutput: "slim - Document transformation tool",
		},
		{
			name:        "unknown command",
			args:        []string{"slim", "unknown"},
			expectError: true,
		},
		{
			name:        "convert missing args",
			args:        []string{"slim", "convert"},
			expectError: true,
		},
		{
			name:        "check missing args",
			args:        []string{"slim", "check"},
			expectError: true,
		},
		{
			name:        "fetch missing args",
			args:        []string{"slim", "fetch"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			// We can't easily test os.Exit, but we can test the error paths
			// by examining what gets written to stderr
			if tt.expectError {
				stderrOutput := captureStderr(t, func() {
					// Simulate the main function logic without os.Exit
					if len(os.Args) < 2 {
						printUsage()
						return
					}

					ctx := context.Background()
					var err error

					switch os.Args[1] {
					case "convert":
						err = handleConvert(ctx, os.Args[2:])
					case "check":
						err = handleCheck(ctx, os.Args[2:])
					case "list":
						err = handleList(ctx, os.Args[2:])
					case "fetch":
						err = handleFetch(ctx, os.Args[2:])
					default:
						fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
						return
					}

					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					}
				})

				if len(stderrOutput) == 0 && tt.name != "no arguments" {
					t.Error("Expected error output but got none")
				}

				if tt.expectedOutput != "" && !strings.Contains(stderrOutput, tt.expectedOutput) {
					// For "no arguments" case, check stdout instead
					if tt.name == "no arguments" {
						stdoutOutput := captureOutput(t, func() {
							printUsage()
						})
						if !strings.Contains(stdoutOutput, tt.expectedOutput) {
							t.Errorf("Expected output to contain %q, got: %s", tt.expectedOutput, stdoutOutput)
						}
					} else {
						t.Errorf("Expected output to contain %q, got: %s", tt.expectedOutput, stderrOutput)
					}
				}
			}
		})
	}
}

// Test helper functions for mocking API responses
func createMockAPIServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			resp := client.LoginResponse{
				AccessToken: "mock-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(resp)

		case "/api/library":
			resp := client.LibraryResponse{
				Summaries: []client.LibrarySummary{
					{ID: 1001, Title: "Mock Book 1", Description: "First mock book"},
					{ID: 1002, Title: "Mock Book 2", Description: "Second mock book"},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(resp)

		default:
			// Mock individual book endpoints
			if strings.HasPrefix(r.URL.Path, "/api/book/") {
				bookID := strings.TrimPrefix(r.URL.Path, "/api/book/")
				mockData := map[string]any{
					"id":          bookID,
					"title":       fmt.Sprintf("Mock Book %s", bookID),
					"description": fmt.Sprintf("Mock description for book %s", bookID),
				}
				json.NewEncoder(w).Encode(mockData)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
}

// Benchmark tests for CLI operations
func BenchmarkParseConvertOptions(b *testing.B) {
	args := []string{"--formats", "html,latex,epub", "--output", "/tmp/test", "book1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseConvertOptions(args)
		if err != nil {
			b.Fatalf("parseConvertOptions failed: %v", err)
		}
	}
}

func BenchmarkParseFetchOptions(b *testing.B) {
	args := []string{"--all", "--output", "/tmp/data", "--clean"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseFetchOptions(args)
		if err != nil {
			b.Fatalf("parseFetchOptions failed: %v", err)
		}
	}
}

func BenchmarkSanitizeFilename(b *testing.B) {
	testFilename := "Complex/File\\Name:With*Special?Characters\"<>|"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeFilename(testFilename)
	}
}
