// Package main implements the SlimAcademy CLI application with Cobra framework.
// It provides commands for converting books between formats, validation, listing sources,
// and exporting content with ZIP archive support and concurrent processing.
package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	debug      bool
	configPath string
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slim",
	Short: "Document transformation tool",
	Long: `slim - Document transformation tool

A powerful CLI tool for converting SlimAcademy books between various formats
including Markdown, HTML, LaTeX, EPUB, and plain text.

Features:
- Memory-efficient streaming processing
- Concurrent multi-format conversion
- Configuration-driven output customization
- Input validation and sanitization
- ZIP archive generation for batch operations`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// setupLogging initializes structured logging based on flags
func setupLogging() {
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Store logger in context for global access
	slog.SetDefault(logger)
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
