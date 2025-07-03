#!/bin/bash

# dev-setup.sh
# Complete development environment setup script for SlimAcademy

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

log_header() {
    echo
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN} $1${NC}"
    echo -e "${CYAN}================================${NC}"
}

# Check system requirements
check_requirements() {
    log_header "Checking System Requirements"

    local missing_tools=()

    # Check Go
    if ! command -v go >/dev/null 2>&1; then
        missing_tools+=("go")
    else
        local go_version=$(go version | cut -d' ' -f3)
        log_success "Go $go_version is installed"
    fi

    # Check Git
    if ! command -v git >/dev/null 2>&1; then
        missing_tools+=("git")
    else
        log_success "Git is installed"
    fi

    # Check curl
    if ! command -v curl >/dev/null 2>&1; then
        missing_tools+=("curl")
    fi

    # Optional tools
    local optional_tools=("docker" "make" "zip" "tar")
    for tool in "${optional_tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            log_success "$tool is available"
        else
            log_warning "$tool is not installed (optional)"
        fi
    done

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and run this script again"
        exit 1
    fi
}

# Initialize Git hooks and configuration
setup_git() {
    log_header "Setting up Git Configuration"

    cd "$PROJECT_ROOT"

    # Set up Git hooks directory
    if [ ! -d ".git/hooks" ]; then
        log_warning "Not a Git repository or hooks directory missing"
        return 1
    fi

    # Configure Git settings for the project
    git config --local core.autocrlf false
    git config --local core.eol lf
    git config --local pull.rebase true
    git config --local fetch.prune true
    git config --local branch.autosetupmerge always
    git config --local branch.autosetuprebase always

    log_success "Git configuration updated"
}

# Setup Go environment
setup_go_environment() {
    log_header "Setting up Go Environment"

    cd "$PROJECT_ROOT"

    # Download and verify dependencies
    log_step "Downloading Go dependencies..."
    go mod download
    go mod verify

    # Tidy modules
    log_step "Tidying Go modules..."
    go mod tidy

    # Generate any code if needed
    if grep -r "go:generate" . >/dev/null 2>&1; then
        log_step "Running go generate..."
        go generate ./...
    fi

    log_success "Go environment setup complete"
}

# Install development tools
install_development_tools() {
    log_header "Installing Development Tools"

    if [ -f "$SCRIPT_DIR/install-dev-tools.sh" ]; then
        log_step "Running development tools installation..."
        bash "$SCRIPT_DIR/install-dev-tools.sh"
    else
        log_warning "install-dev-tools.sh not found, skipping tool installation"
    fi
}

# Setup pre-commit hooks
setup_pre_commit() {
    log_header "Setting up Pre-commit Hooks"

    cd "$PROJECT_ROOT"

    if command -v pre-commit >/dev/null 2>&1; then
        log_step "Installing pre-commit hooks..."
        pre-commit install
        pre-commit install --hook-type commit-msg

        log_step "Running pre-commit on all files..."
        pre-commit run --all-files || log_warning "Some pre-commit checks failed"

        log_success "Pre-commit hooks setup complete"
    else
        log_warning "pre-commit not available"
    fi
}

# Create necessary directories
create_directories() {
    log_header "Creating Project Directories"

    cd "$PROJECT_ROOT"

    local dirs=("bin" "output" "test_output" "logs" "tmp")

    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_success "Created directory: $dir"
        else
            log_info "Directory already exists: $dir"
        fi
    done

    # Create .gitkeep files for empty directories that should be tracked
    touch bin/.gitkeep output/.gitkeep test_output/.gitkeep
}

# Setup VS Code workspace
setup_vscode_workspace() {
    log_header "Setting up VS Code Workspace"

    cd "$PROJECT_ROOT"

    # Create workspace file
    local workspace_file="slimacademy.code-workspace"

    cat > "$workspace_file" << 'EOF'
{
    "folders": [
        {
            "name": "SlimAcademy",
            "path": "."
        }
    ],
    "settings": {
        "files.exclude": {
            "**/bin": true,
            "**/output": true,
            "**/test_output": true,
            "**/.backup": true
        }
    },
    "extensions": {
        "recommendations": [
            "golang.go",
            "github.vscode-pull-request-github",
            "editorconfig.editorconfig",
            "redhat.vscode-yaml"
        ]
    },
    "tasks": {
        "version": "2.0.0",
        "tasks": [
            {
                "label": "Build CLI",
                "type": "shell",
                "command": "go build -o bin/slim ./cmd/slim",
                "group": {
                    "kind": "build",
                    "isDefault": true
                }
            },
            {
                "label": "Run Tests",
                "type": "shell",
                "command": "go test -v ./...",
                "group": {
                    "kind": "test",
                    "isDefault": true
                }
            }
        ]
    }
}
EOF

    log_success "VS Code workspace file created: $workspace_file"
}

# Create development environment file
create_env_template() {
    log_header "Creating Environment Template"

    cd "$PROJECT_ROOT"

    if [ ! -f ".env.example" ]; then
        cat > ".env.example" << 'EOF'
# SlimAcademy Development Environment Configuration
# Copy this file to .env and fill in your values

# API Configuration
# SLIM_API_BASE_URL=https://api.slimacademy.nl
# SLIM_API_CLIENT_ID=slim_api
# SLIM_API_DEVICE_ID=your-device-id

# Authentication (for development)
# USERNAME=your-username
# PASSWORD=your-password

# Development Settings
# LOG_LEVEL=debug
# OUTPUT_DIR=output
# SOURCE_DIR=source

# Optional: Custom configuration file path
# CONFIG_FILE=config.yaml
EOF
        log_success "Created .env.example template"
    else
        log_info ".env.example already exists"
    fi
}

# Run initial tests
run_initial_tests() {
    log_header "Running Initial Tests"

    cd "$PROJECT_ROOT"

    log_step "Running Go tests..."
    if go test ./...; then
        log_success "All tests passed!"
    else
        log_warning "Some tests failed - this is normal for a fresh setup"
    fi

    log_step "Building CLI..."
    if go build -o bin/slim ./cmd/slim; then
        log_success "CLI built successfully"

        # Test basic functionality
        if ./bin/slim --help >/dev/null 2>&1; then
            log_success "CLI is functional"
        else
            log_warning "CLI build succeeded but --help failed"
        fi
    else
        log_error "CLI build failed"
    fi
}

# Create development documentation
create_dev_docs() {
    log_header "Creating Development Documentation"

    cd "$PROJECT_ROOT"

    if [ ! -f "CONTRIBUTING.md" ]; then
        cat > "CONTRIBUTING.md" << 'EOF'
# Contributing to SlimAcademy

Thank you for your interest in contributing to SlimAcademy!

## Development Setup

1. Clone the repository
2. Run the development setup script:
   ```bash
   ./scripts/dev-setup.sh
   ```
3. Install VS Code extensions (recommended)
4. Copy `.env.example` to `.env` and configure as needed

## Development Workflow

### Building

```bash
# Build the CLI
go build -o bin/slim ./cmd/slim

# Or use the VS Code task (Ctrl+Shift+P -> Tasks: Run Task -> Build CLI)
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Run pre-commit checks
pre-commit run --all-files

# Format code
gofumpt -w .
goimports -w .
```

### Security

```bash
# Security scan
gosec ./...

# Vulnerability check
govulncheck ./...
```

## Project Structure

- `cmd/slim/` - CLI entry point
- `internal/` - Internal packages
  - `client/` - API client
  - `config/` - Configuration management
  - `models/` - Data models
  - `parser/` - JSON parsing
  - `sanitizer/` - Content sanitization
  - `streaming/` - Event streaming
  - `writers/` - Output format writers
- `scripts/` - Development and build scripts
- `test/` - Test fixtures and utilities

## Submitting Changes

1. Create a feature branch
2. Make your changes
3. Run tests and linting
4. Commit with conventional commit messages
5. Submit a pull request

## Code Style

- Follow Go conventions
- Use gofumpt for formatting
- Write tests for new functionality
- Document public APIs
- Keep functions focused and testable

## Release Process

Releases are automated through GitHub Actions when tags are pushed.
EOF
        log_success "Created CONTRIBUTING.md"
    else
        log_info "CONTRIBUTING.md already exists"
    fi
}

# Show final summary
show_summary() {
    log_header "Development Environment Setup Complete"

    echo -e "${GREEN}✅ Development environment is ready!${NC}"
    echo
    echo -e "${CYAN}Next steps:${NC}"
    echo "1. Open the project in VS Code:"
    echo "   code slimacademy.code-workspace"
    echo
    echo "2. Copy and configure environment variables:"
    echo "   cp .env.example .env"
    echo "   # Edit .env with your configuration"
    echo
    echo "3. Build and test the CLI:"
    echo "   go build -o bin/slim ./cmd/slim"
    echo "   ./bin/slim --help"
    echo
    echo "4. Run the development aliases:"
    echo "   source ~/.dev_aliases"
    echo
    echo -e "${CYAN}Useful commands:${NC}"
    echo "  slim-build    # Build the CLI"
    echo "  slim-test     # Run tests"
    echo "  slim-lint     # Run linter"
    echo "  slim-check    # Run all checks"
    echo "  slim-clean    # Clean build artifacts"
    echo
    echo -e "${CYAN}Documentation:${NC}"
    echo "  README.md         # Project overview"
    echo "  USER_GUIDE.md     # User guide"
    echo "  QUICK_REFERENCE.md # Quick reference"
    echo "  CONTRIBUTING.md   # Development guide"
    echo
    echo -e "${GREEN}Happy coding! 🚀${NC}"
}

# Help function
show_help() {
    cat << EOF
SlimAcademy Development Environment Setup

Usage: $0 [options]

Options:
    -h, --help          Show this help message
    --skip-tools        Skip development tools installation
    --skip-tests        Skip initial test run
    --skip-git          Skip Git configuration
    --skip-vscode       Skip VS Code setup
    --quick             Quick setup (skip optional steps)

Examples:
    $0                  # Full setup
    $0 --quick          # Quick setup
    $0 --skip-tools     # Setup without installing tools

EOF
}

# Main setup function
main() {
    local skip_tools=false
    local skip_tests=false
    local skip_git=false
    local skip_vscode=false
    local quick_setup=false

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --skip-tools)
                skip_tools=true
                shift
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --skip-git)
                skip_git=true
                shift
                ;;
            --skip-vscode)
                skip_vscode=true
                shift
                ;;
            --quick)
                quick_setup=true
                skip_tests=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_header "SlimAcademy Development Environment Setup"

    # Core setup steps
    check_requirements
    create_directories
    setup_go_environment

    # Optional setup steps
    if [ "$skip_git" = false ]; then
        setup_git
    fi

    if [ "$skip_tools" = false ]; then
        install_development_tools
        setup_pre_commit
    fi

    if [ "$skip_vscode" = false ]; then
        setup_vscode_workspace
    fi

    create_env_template

    if [ "$quick_setup" = false ]; then
        create_dev_docs
    fi

    if [ "$skip_tests" = false ]; then
        run_initial_tests
    fi

    show_summary
}

# Make scripts executable
chmod +x "$SCRIPT_DIR"/*.sh 2>/dev/null || true

# Run main function
main "$@"
