# SlimAcademy CLI

## Overview

SlimAcademy CLI is a powerful document transformation tool built with Go that converts SlimAcademy books between various output formats including Markdown, HTML, LaTeX, EPUB, and plain text. The tool features a modern event-driven streaming architecture designed for memory efficiency and high performance.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Commands](#commands)
- [Configuration](#configuration)
- [Development](#development)
- [API Integration](#api-integration)
- [Testing](#testing)
- [Contributing](#contributing)

## Installation

### Prerequisites

- Go 1.23 or later
- Git

### Build from Source

```bash
git clone https://github.com/kjanat/slimacademy.git
cd slimacademy
go build -o bin/slim ./cmd/slim
```

### Install Dependencies

```bash
go mod tidy
```

## Quick Start

### Basic Usage

```bash
# Convert a single book to markdown
slim convert book1

# Convert to multiple formats
slim convert --formats html,epub book1

# Convert all books to ZIP archive
slim convert --all > all-books.zip

# Check a book for issues
slim check book1

# List available books
slim list source/

# Fetch books from API
slim fetch --login
slim fetch --all
slim fetch --id 3631
```

### Directory Structure

```
source/
├── book-id-1/
│   ├── 123.json      # Book metadata
│   ├── chapters.json # Chapter structure
│   ├── content.json  # Book content
│   └── notes.json    # User notes (optional)
└── book-id-2/
    └── ...
```

## Architecture

### Core Design Principles

- **Memory Efficient**: O(1) RAM usage with streaming iterators
- **Error Resilient**: Context cancellation prevents corrupt outputs
- **Maintainable**: Centralized validation, simple writers, shared stack logic
- **Observable**: Built-in stats collection and validation reporting
- **Extensible**: Plugin architecture via writer registry
- **Modern**: Leverages Go 1.23+ features for performance

### Event-Driven Processing

```
Raw JSON → Sanitizer → Streaming Events → MultiWriter → Concurrent Output
```

### Key Components

| Component | Purpose |
|-----------|---------|
| **Parser** | JSON parsing and validation |
| **Sanitizer** | Content cleaning and validation |
| **Streaming** | Memory-efficient event processing |
| **Writers** | Format-specific output generation |
| **Registry** | Plugin-based writer management |

## Commands

### convert

Convert documents between formats with support for batch processing and ZIP output.

```bash
# Basic conversion
slim convert book1                                    # Convert to markdown (default)
slim convert --format html book1                     # Convert to HTML
slim convert --formats html,epub book1               # Multiple formats
slim convert --all > all-books.zip                   # All books as ZIP
slim convert book1 --output /tmp/output.md           # Custom output path
slim convert --config config.yaml book1              # Custom configuration
```

**Flags:**
- `--all`: Convert all books to ZIP archive
- `--formats, -f`: Output formats (markdown,html,latex,epub,plaintext)
- `--output, -o`: Output file/directory path
- `--config`: Configuration file path

### check

Validate books and identify potential issues.

```bash
# Basic validation
slim check book1                    # Check single book
slim check --verbose book1          # Detailed output
```

**Features:**
- Invalid UTF-8 sequence detection
- Dangerous HTML content identification
- Malformed URL detection
- Empty heading detection
- Structural consistency validation

### list

List available books in a directory.

```bash
slim list source/                   # List books in source directory
slim list --detailed source/        # Detailed book information
```

### fetch

Download books from the SlimAcademy API.

```bash
# Authentication
slim fetch --login                          # Login and save token

# Download operations
slim fetch --all                            # Fetch all books
slim fetch --id 3631                        # Fetch specific book
slim fetch --all --output data/ --clean     # Custom output with cleanup
```

**Flags:**
- `--login`: Authenticate and save token
- `--all`: Fetch all books from library
- `--id`: Fetch specific book by ID
- `--output, -o`: Output directory (default: source)
- `--clean`: Clean output directory before fetching

## Configuration

### Configuration Files

Create configuration files to customize output formats:

```yaml
# config.yaml
markdown:
  headingStyle: "atx"

html:
  template: "academic"
  includeCSS: true

epub:
  author: "SlimAcademy"
  language: "en"
```

### Environment Variables

```bash
# API Authentication
export USERNAME=your@email.com
export PASSWORD=yourpassword

# Alternative: .env file
echo "USERNAME=your@email.com" > .env
echo "PASSWORD=yourpassword" >> .env
```

### Global Flags

- `--config`: Configuration file path
- `--debug`: Enable debug logging
- `--verbose, -v`: Enable verbose output

## Development

### Project Structure

```
cmd/slim/           # CLI application
├── main.go         # Entry point and utilities
├── root.go         # Cobra root command
├── convert.go      # Convert command
├── check.go        # Validation command
├── list.go         # List command
├── fetch.go        # API fetch command
└── main_test.go    # CLI tests

internal/
├── client/         # API client
├── config/         # Configuration management
├── models/         # Data models
├── parser/         # JSON parsing
├── sanitizer/      # Content sanitization
├── streaming/      # Event streaming
├── writers/        # Format writers
└── testing/        # Test utilities
```

### Build Commands

```bash
# Development build
go build -o bin/slim ./cmd/slim

# Run tests
go test ./...                      # All tests
go test -v ./internal/streaming/   # Specific package
go test -cover ./...               # With coverage

# Update golden test files
UPDATE_GOLDEN=1 go test ./...
```

### Modern Go Features

The project leverages Go 1.23+ features:

- **Streaming**: `iter.Seq[Event]` for memory efficiency
- **Deduplication**: `unique.Handle[string]` for O(1) operations
- **Text Processing**: `bytes.Lines()` for zero-allocation chunking
- **Concurrency**: Context-based cancellation with proper cleanup

## API Integration

### Authentication Flow

1. **Credential Setup**: Configure `.env` file or environment variables
2. **Login**: `slim fetch --login` authenticates and saves token
3. **Token Management**: Automatic token validation and refresh
4. **API Calls**: Authenticated requests to SlimAcademy API

### API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/api/auth/login` | Authentication |
| `/api/user/profile` | User information |
| `/api/summary/library` | Library listing |
| `/api/summary/{id}` | Book metadata |
| `/api/summary/{id}/chapters` | Chapter structure |
| `/api/summary/{id}/content` | Book content |
| `/api/summary/{id}/list-notes` | User notes |

### Data Flow

```
API Request → JSON Response → BookData Struct → File Writing → Directory Structure
```

## Testing

### Test Categories

- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component functionality
- **Golden Tests**: Output validation against reference files
- **Property Tests**: Invariant validation

### Test Execution

```bash
# Run all tests
go test ./...

# Test with coverage
go test -cover ./...

# Update golden files
UPDATE_GOLDEN=1 go test ./...

# Test specific functionality
go test ./cmd/slim/              # CLI tests
go test ./internal/client/       # API client tests
go test ./internal/streaming/    # Streaming tests
```

### Test Data

```
test/fixtures/
├── valid_books/
│   └── simple_book/
├── invalid_data/
│   ├── missing_files/
│   └── malformed_json/
└── golden/
    ├── expected_output.md
    └── expected_output.html
```

## Contributing

### Development Workflow

1. **Fork and Clone**: Create your development environment
2. **Feature Branch**: Create feature-specific branches
3. **Implementation**: Follow existing patterns and conventions
4. **Testing**: Ensure all tests pass
5. **Documentation**: Update relevant documentation
6. **Pull Request**: Submit for review

### Code Standards

- **Linting**: Follow Go standard formatting
- **Testing**: Maintain test coverage
- **Documentation**: Include package and function comments
- **Error Handling**: Use structured error handling
- **Logging**: Use structured logging with slog

### Architectural Guidelines

- **Memory Efficiency**: Use streaming for large data processing
- **Error Resilience**: Implement proper error handling and recovery
- **Maintainability**: Keep components loosely coupled
- **Extensibility**: Use plugin patterns for new features

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.

**Note**: The AGPL-3.0 is a copyleft license that requires anyone who distributes or provides this software as a service to also provide the source code under the same license terms. This ensures that modifications and improvements remain open source.

---

**Project Status**: Production-ready foundation with modern Go architecture
**Go Version**: 1.23+
**Maintainer**: kjanat
**Repository**: https://github.com/kjanat/slimacademy
