# SlimAcademy Makefile
# Leverages Go 1.24 tool directive for development tools

.PHONY: help build test lint security check-all clean install-tools

# Default target
help: ## Show this help message
	@echo "SlimAcademy - Enhanced Google Docs to HTML Converter"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Build targets
build: ## Build the slim binary
	go build -o bin/slim ./cmd/slim

build-all: ## Build all binaries
	go build -o bin/slim ./cmd/slim
	go build -o bin/transformer ./cmd/transformer

# Test targets
test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-cover: ## Run tests with coverage
	go test -cover ./...

test-race: ## Run tests with race detection
	go test -race ./...

test-update-golden: ## Update golden test files
	UPDATE_GOLDEN=1 go test ./...

# Fuzzing targets (Go 1.24 feature)
fuzz-sanitizer: ## Run fuzzing tests for sanitizer
	go test -fuzz=FuzzSanitize ./internal/sanitizer -fuzztime=30s

fuzz-all: ## Run all fuzzing tests
	@echo "Running all fuzzing tests..."
	go test -fuzz=. ./... -fuzztime=30s

# Static analysis using Go 1.24 tool directive
lint: ## Run golangci-lint
	go tool golangci-lint run ./...

staticcheck: ## Run staticcheck
	go tool staticcheck ./...

security: ## Run security analysis with gosec
	go tool gosec ./...

vulncheck: ## Check for known vulnerabilities
	go tool govulncheck ./...

# Quality checks
vet: ## Run go vet with enhanced tests analyzer
	go vet ./...

# Complete quality pipeline
check-all: vet lint staticcheck security vulncheck ## Run all quality checks
	@echo "✅ All quality checks passed!"

# Development workflow
dev-setup: ## Set up development environment
	@echo "Installing development tools using Go 1.24 tool directive..."
	go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	go get -tool honnef.co/go/tools/cmd/staticcheck@v0.4.7
	go get -tool github.com/securecodewarrior/gosec/v2/cmd/gosec@v2.18.2
	go get -tool golang.org/x/vuln/cmd/govulncheck@v1.0.1
	@echo "✅ Development environment ready!"

# Clean targets
clean: ## Clean build artifacts
	rm -rf bin/
	go clean ./...

clean-cache: ## Clean module and build cache
	go clean -modcache
	go clean -cache

# Release targets
release-test: check-all test ## Run full test suite for release
	@echo "🚀 Release tests completed successfully!"

# Performance targets
benchmark: ## Run benchmarks
	go test -bench=. ./...

profile-cpu: ## Run CPU profiling
	go test -cpuprofile=cpu.prof -bench=. ./...

profile-mem: ## Run memory profiling
	go test -memprofile=mem.prof -bench=. ./...

# Documentation
docs: ## Generate documentation
	go doc -all ./...

# Install the binary
install: build ## Install slim binary to GOPATH/bin
	go install ./cmd/slim

# Format code
fmt: ## Format Go code
	go fmt ./...

# Modern Go features demonstration
demo-iterators: ## Run demo showing Go 1.24 iterator features
	go run ./cmd/demo-iterators

# Show tool versions
tool-versions: ## Show versions of development tools
	@echo "Go version: $(shell go version)"
	@echo "golangci-lint: $(shell go tool golangci-lint version 2>/dev/null || echo 'not installed')"
	@echo "staticcheck: $(shell go tool staticcheck -version 2>/dev/null || echo 'not installed')"
	@echo "gosec: $(shell go tool gosec -version 2>/dev/null || echo 'not installed')"
	@echo "govulncheck: $(shell go tool govulncheck -version 2>/dev/null || echo 'not installed')"

# CI/CD targets
ci: dev-setup check-all test ## Complete CI pipeline
	@echo "🎉 CI pipeline completed successfully!"

# Export formats for testing
export-all: build ## Export all test books in all formats
	./scripts/export-all.sh

# Quick development cycle
quick: fmt vet test ## Quick development cycle (format, vet, test)
	@echo "✅ Quick checks passed!"
