# Development Guide

This document provides comprehensive information for developers working on the SlimAcademy project.

## Quick Start

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd slimacademy
   ./scripts/dev-setup.sh
   ```

2. **Open in VS Code:**
   ```bash
   code slimacademy.code-workspace
   ```

3. **Build and test:**
   ```bash
   go build -o bin/slim ./cmd/slim
   go test ./...
   ```

## Development Environment

### Prerequisites

- **Go 1.22+** - Primary development language
- **Git** - Version control
- **VS Code** - Recommended IDE (with Go extension)
- **pre-commit** - Git hooks for code quality
- **Docker** - Optional, for containerized builds

### Setup Scripts

| Script | Purpose |
|--------|---------|
| `scripts/dev-setup.sh` | Complete development environment setup |
| `scripts/install-dev-tools.sh` | Install Go tools and development utilities |
| `scripts/build-release.sh` | Cross-platform release builds |

### Development Tools

The following tools are automatically installed and configured:

- **golangci-lint** - Comprehensive Go linter
- **gosec** - Security scanner
- **govulncheck** - Vulnerability scanner
- **staticcheck** - Static analysis
- **gofumpt** - Code formatter
- **pre-commit** - Git hooks

## Project Architecture

### Directory Structure

```
slimacademy/
├── cmd/slim/                 # CLI entry point
├── internal/                 # Internal packages
│   ├── client/              # API client
│   ├── config/              # Configuration management
│   ├── models/              # Data models
│   ├── parser/              # JSON parsing
│   ├── sanitizer/           # Content sanitization
│   ├── streaming/           # Event streaming
│   └── writers/             # Output format writers
├── scripts/                 # Development scripts
├── test/                    # Test fixtures
├── .github/                 # GitHub Actions & templates
├── .vscode/                 # VS Code configuration
└── docs/                    # Documentation
```

### Key Components

- **CLI (`cmd/slim/`)** - Command-line interface with subcommands
- **Streaming (`internal/streaming/`)** - Memory-efficient event processing
- **Writers (`internal/writers/`)** - Pluggable output format handlers
- **Config (`internal/config/`)** - Configuration management with validation
- **Client (`internal/client/`)** - API client for Slim Academy

## Development Workflow

### Code Quality

```bash
# Format code
gofumpt -w .
goimports -w .

# Lint code
golangci-lint run --fix

# Run all quality checks
pre-commit run --all-files
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...

# Update golden test files
UPDATE_GOLDEN=1 go test ./...
```

### Security

```bash
# Security scan
gosec ./...

# Vulnerability check
govulncheck ./...

# Dependency audit
go list -json -deps ./... | nancy sleuth
```

### Building

```bash
# Build for current platform
go build -o bin/slim ./cmd/slim

# Cross-platform build
./scripts/build-release.sh

# Build with custom version
VERSION=v1.0.0 ./scripts/build-release.sh
```

## CI/CD Pipeline

### GitHub Actions Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push/PR | Main CI pipeline with tests, linting, builds |
| `security.yml` | Push/PR/Schedule | Security scanning and vulnerability checks |
| `quality.yml` | Push/PR | Code quality metrics and documentation checks |

### Release Process

1. **Create release tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Automated release:**
   - GitHub Actions builds for all platforms
   - Creates GitHub release with binaries
   - Generates checksums and archives

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch
- `feature/*` - Feature branches
- `hotfix/*` - Critical fixes

## VS Code Setup

### Recommended Extensions

The workspace includes recommendations for:
- Go extension with full language server support
- GitHub integration
- YAML/JSON support
- Code quality tools

### Debug Configurations

Pre-configured debug setups for:
- Main application with various commands
- Individual tests
- Benchmark testing
- Remote debugging

### Tasks

Available VS Code tasks:
- Build CLI
- Run tests with coverage
- Lint code
- Security scans
- Clean build artifacts

## Configuration

### Environment Variables

```bash
# Copy example configuration
cp .env.example .env

# Edit configuration
vim .env
```

### Config Files

- `.golangci.yml` - Linting configuration
- `.pre-commit-config.yaml` - Git hooks
- `.editorconfig` - Editor settings
- `.vscode/settings.json` - VS Code preferences

## Testing Strategy

### Test Types

1. **Unit Tests** - Individual function testing
2. **Integration Tests** - Component interaction testing
3. **Property Tests** - Generative testing for edge cases
4. **Golden Tests** - Output comparison testing
5. **Benchmark Tests** - Performance testing

### Test Organization

```
test/
├── fixtures/           # Test data
│   ├── valid_books/   # Valid test inputs
│   └── invalid_data/  # Invalid test inputs
└── utils/             # Test utilities
```

### Writing Tests

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Code Style Guidelines

### Go Conventions

- Follow standard Go formatting (`gofumpt`)
- Use meaningful variable names
- Keep functions small and focused
- Write comprehensive tests
- Document public APIs

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new output format support
fix: resolve memory leak in streaming
docs: update API documentation
test: add integration tests for parser
```

### Code Review

- All changes require PR review
- CI checks must pass
- Test coverage should not decrease
- Documentation updates when needed

## Performance Considerations

### Memory Usage

- Use streaming for large files
- Implement proper cleanup in defer blocks
- Monitor memory allocation in benchmarks

### Concurrency

- Use context for cancellation
- Properly handle goroutine cleanup
- Avoid shared mutable state

### Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Profile memory
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# Profile CPU
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof
```

## Debugging

### Local Debugging

1. **VS Code Debugger:**
   - Set breakpoints in code
   - Use debug configurations
   - Inspect variables and call stack

2. **Command Line:**
   ```bash
   # Build with debug symbols
   go build -gcflags="all=-N -l" -o bin/slim-debug ./cmd/slim

   # Run with delve
   dlv exec bin/slim-debug -- convert --help
   ```

### Production Debugging

```bash
# Enable verbose logging
SLIM_LOG_LEVEL=debug ./bin/slim convert input.json

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Troubleshooting

### Common Issues

1. **Module Download Failures:**
   ```bash
   go clean -modcache
   go mod download
   ```

2. **Pre-commit Hook Failures:**
   ```bash
   pre-commit clean
   pre-commit install
   ```

3. **Build Failures:**
   ```bash
   go mod tidy
   go mod verify
   ```

### Getting Help

- Check existing [GitHub Issues](https://github.com/kjanat/slimacademy/issues)
- Review [documentation](README.md)
- Run `./bin/slim --help` for CLI usage

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.
