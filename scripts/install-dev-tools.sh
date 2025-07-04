#!/bin/bash

# install-dev-tools.sh
# Installs essential development tools for Go development

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Go tools
install_go_tools() {
    log_info "Installing Go development tools..."

    local tools=(
        "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        "honnef.co/go/tools/cmd/staticcheck@latest"
        "github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
        "golang.org/x/vuln/cmd/govulncheck@latest"
        "github.com/fzipp/gocyclo/cmd/gocyclo@latest"
        "github.com/gordonklaus/ineffassign@latest"
        "github.com/client9/misspell/cmd/misspell@latest"
        "github.com/google/go-licenses@latest"
        "github.com/sonatypecommunity/nancy@latest"
        "mvdan.cc/gofumpt@latest"
        "golang.org/x/tools/cmd/goimports@latest"
        "github.com/mgechev/revive@latest"
        "github.com/go-delve/delve/cmd/dlv@latest"
    )

    for tool in "${tools[@]}"; do
        log_info "Installing $tool..."
        if go install "$tool"; then
            log_success "Installed $tool"
        else
            log_error "Failed to install $tool"
        fi
    done
}

# Install pre-commit if not available
install_pre_commit() {
    if ! command_exists pre-commit; then
        log_info "Installing pre-commit..."

        if command_exists pip3; then
            pip3 install pre-commit
        elif command_exists pip; then
            pip install pre-commit
        elif command_exists brew; then
            brew install pre-commit
        elif command_exists apt-get; then
            sudo apt-get update && sudo apt-get install -y pre-commit
        else
            log_warning "Could not install pre-commit automatically. Please install manually."
            return 1
        fi

        log_success "Installed pre-commit"
    else
        log_info "pre-commit is already installed"
    fi
}

# Setup pre-commit hooks
setup_pre_commit() {
    if command_exists pre-commit; then
        log_info "Setting up pre-commit hooks..."

        if pre-commit install; then
            log_success "Pre-commit hooks installed"
        else
            log_error "Failed to install pre-commit hooks"
        fi

        # Install commit-msg hook for conventional commits
        if pre-commit install --hook-type commit-msg; then
            log_success "Commit-msg hook installed"
        else
            log_warning "Failed to install commit-msg hook"
        fi
    else
        log_warning "pre-commit not available, skipping hook setup"
    fi
}

# Install Docker tools (optional)
install_docker_tools() {
    if command_exists docker; then
        log_info "Docker is available, installing Docker-based tools..."

        # Pull useful Docker images for development
        docker pull hadolint/hadolint:latest || log_warning "Failed to pull hadolint image"
        docker pull aquasec/trivy:latest || log_warning "Failed to pull trivy image"

        log_success "Docker tools installed"
    else
        log_info "Docker not available, skipping Docker tools"
    fi
}

# Create useful aliases and functions
create_dev_aliases() {
    local alias_file="$HOME/.dev_aliases"

    log_info "Creating development aliases..."

    cat > "$alias_file" << 'EOF'
# SlimAcademy Development Aliases

# Go development
alias gob='go build ./...'
alias got='go test ./...'
alias gotv='go test -v ./...'
alias gotc='go test -coverprofile=coverage.out ./...'
alias gocov='go tool cover -html=coverage.out'
alias golint='golangci-lint run'
alias gofmt='gofumpt -w .'
alias goimp='goimports -w .'
alias gomod='go mod tidy'
alias gogen='go generate ./...'

# SlimAcademy specific
alias slim-build='go build -o bin/slim ./cmd/slim'
alias slim-test='go test -v ./...'
alias slim-lint='golangci-lint run --fix'
alias slim-check='pre-commit run --all-files'
alias slim-clean='rm -rf bin/ output/ test_output/ coverage.* *.prof'

# Security scanning
alias gosec-scan='gosec -fmt json -out gosec-report.json ./...'
alias vuln-check='govulncheck ./...'

# Git helpers
alias git-clean='git clean -fd && git reset --hard'
alias git-branches='git branch -a --sort=-committerdate'
alias git-tags='git tag -l --sort=-version:refname'

# Development server
alias dev-server='go run ./cmd/slim'
EOF

    log_success "Development aliases created in $alias_file"
    log_info "Add 'source $alias_file' to your shell profile to use them"
}

# Setup VS Code settings if not exists
setup_vscode() {
    local vscode_dir=".vscode"

    if [ -d "$vscode_dir" ]; then
        log_info "VS Code configuration already exists"
    else
        log_warning "VS Code configuration not found. Run the main dev-setup command to create it."
    fi
}

# Main installation function
main() {
    log_info "Starting development tools installation..."

    # Check if Go is installed
    if ! command_exists go; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi

    log_info "Go version: $(go version)"

    # Install tools
    install_go_tools
    install_pre_commit
    setup_pre_commit
    install_docker_tools
    create_dev_aliases
    setup_vscode

    log_success "Development tools installation completed!"

    echo
    log_info "Next steps:"
    echo "1. Source the aliases: source ~/.dev_aliases"
    echo "2. Restart your shell or VS Code"
    echo "3. Run 'slim-check' to verify everything works"
    echo "4. Happy coding! 🚀"
}

# Run main function
main "$@"
