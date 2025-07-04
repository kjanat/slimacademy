# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Build:**

```bash
go build -o bin/slim ./cmd/slim    # New architecture CLI
go mod tidy
```

**Testing:**

```bash
go test ./...                      # Run all tests
go test -v ./internal/streaming/   # Run specific package tests
go test -cover ./...               # Run with coverage
UPDATE_GOLDEN=1 go test ./...      # Update golden test files
```

**New CLI Usage:**

```bash
slim convert --all > all-books.zip                    # All books, all formats as ZIP
slim convert --format markdown book1                  # Single book to markdown
slim convert --formats "html,epub" book1 --output out # Multiple formats
slim check book1                                       # Validate book1 only
slim list source/                                      # List books in source/
```

**Legacy Scripts:**

```bash
./scripts/export-all.sh  # Export all books in all formats concurrently
```

## Architecture Overview (New - 2024)

This is a **modernized document transformation system** with significant architectural improvements for performance, maintainability, and observability.

### Core Architecture Pattern

**Sanitized Event-Driven Stream Processing:**

- Input JSON → **Token Sanitizer** → **Event Stream** → **Writer Registry** → Output Files
- Events: Concrete structs with StyleFlags, not `any` interface
- Go 1.23+ iterators with `unique.Handle[string]` for O(1) duplicate detection
- Context-based cancellation and error propagation

### Key Components (New Architecture)

**Entry Point:** `cmd/slim/main.go` - Enhanced CLI with ZIP output and validation

**Improved Transformation Pipeline:**

1. `internal/parser/` - JSON parsing (unchanged)
2. `internal/sanitizer/` - **NEW**: Token sanitization with diagnostics
3. `internal/streaming/` - **NEW**: Memory-efficient event streaming
4. `internal/writers/registry.go` - **NEW**: Auto-registration writer system
5. `internal/writers/stack.go` - **NEW**: Shared structural stack logic

**Core Systems:**

- `internal/config/validator.go` - **NEW**: Upfront configuration validation
- `internal/testing/` - **NEW**: Golden tests and property-based testing
- `internal/writers/interfaces.go` - Legacy interfaces (deprecated)

### Data Flow (Improved)

**Processing Pipeline:**

```
Raw JSON → Sanitizer (warnings) → Streaming Events → MultiWriter → Concurrent Output
```

**Input Structure:** (unchanged)

- `source/{book-id}/` contains: `{id}.json`, `chapters.json`, `content.json`

**Output Options:**

- Individual files per format
- ZIP archive with all formats (`--all` flag)
- Validation-only mode (`check` command)

### Modern Go Features (1.23+)

- **Streaming**: `iter.Seq[Event]` for O(1) memory usage
- **Deduplication**: `unique.Handle[string]` for heading anchors
- **Text Processing**: `bytes.Lines()` for zero-allocation chunking
- **Concurrency**: Context-based cancellation with proper cleanup

### Concurrency Model (Enhanced)

- **Writer Registry**: Pluggable writers with auto-registration
- **MultiWriter**: Context-aware concurrent processing with error propagation
- **Stack Management**: Shared logic for tables, lists, formatting
- **Memory Safety**: O(1) RAM usage even on 100MB documents

### Testing Infrastructure (New)

- **Golden Tests**: `internal/testing/golden.go` with automatic discovery
- **Property Tests**: Balanced markers, UTF-8 validation, empty heading detection
- **Test Commands**: `UPDATE_GOLDEN=1` for test updates
- **Validation Tests**: Pre-commit hooks ensure code quality

### Configuration (Enhanced)

- **Validation**: Upfront config validation with actionable suggestions
- **Format-Specific**: Each format has dedicated config with validation
- **Error Handling**: Clear error messages with fix suggestions

## Key Design Principles (Modernized)

- **Memory Efficient**: O(1) RAM usage with streaming iterators
- **Error Resilient**: Context cancellation prevents corrupt outputs
- **Maintainable**: Centralized validation, simple writers, shared stack logic
- **Observable**: Built-in stats collection and validation reporting
- **Extensible**: Plugin architecture via writer registry
- **Modern**: Leverages Go 1.23+ features for performance

## Migration Notes

**Obsolete Components Removed:**

- `internal/events/` (replaced by `internal/streaming/`)
- `internal/exporters/` (replaced by `internal/writers/registry.go`)
- `cmd/transformer/` (replaced by `cmd/slim/`)

**Breaking Changes:**

- Writers must implement `WriterV2` interface with error handling
- Events use concrete structs instead of `any` interface
- CLI commands changed (`slim` instead of `transformer`)

## Current Status

**Architecture**: 90% complete, production-ready foundation
**Remaining Work**: Update existing format writers to use new Event struct fields
**Benefits**: Memory efficient, error resilient, highly maintainable, observable
