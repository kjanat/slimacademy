#!/bin/bash

# build-release.sh
# Cross-platform build script for SlimAcademy CLI

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="slim"
PACKAGE_PATH="./cmd/slim"
BUILD_DIR="bin"
DIST_DIR="dist"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-w -s -X 'main.Version=${VERSION}' -X 'main.GitCommit=${COMMIT_HASH}' -X 'main.BuildTime=${BUILD_TIME}'"

# Target platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "freebsd/amd64"
)

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

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$BUILD_DIR" "$DIST_DIR"
    log_success "Build directories cleaned"
}

# Build for specific platform
build_platform() {
    local platform=$1
    local goos=${platform%/*}
    local goarch=${platform#*/}
    local output_name="$BINARY_NAME-$goos-$goarch"

    # Add .exe extension for Windows
    if [ "$goos" = "windows" ]; then
        output_name="$output_name.exe"
    fi

    local output_path="$BUILD_DIR/$output_name"

    log_info "Building for $goos/$goarch..."

    # Set environment variables and build
    env GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
        go build -a -installsuffix cgo \
        -ldflags="$LDFLAGS" \
        -o "$output_path" \
        "$PACKAGE_PATH"

    if [ $? -eq 0 ]; then
        log_success "Built $output_name"

        # Get file size
        local size=$(du -h "$output_path" | cut -f1)
        log_info "Size: $size"

        return 0
    else
        log_error "Failed to build for $goos/$goarch"
        return 1
    fi
}

# Create archives
create_archives() {
    log_info "Creating distribution archives..."

    for platform in "${PLATFORMS[@]}"; do
        local goos=${platform%/*}
        local goarch=${platform#*/}
        local binary_name="$BINARY_NAME-$goos-$goarch"

        # Add .exe extension for Windows
        if [ "$goos" = "windows" ]; then
            binary_name="$binary_name.exe"
        fi

        local binary_path="$BUILD_DIR/$binary_name"

        if [ -f "$binary_path" ]; then
            # Create archive directory
            local archive_dir="$DIST_DIR/${BINARY_NAME}-${VERSION}-$goos-$goarch"
            mkdir -p "$archive_dir"

            # Copy binary
            cp "$binary_path" "$archive_dir/"

            # Copy documentation
            cp README.md "$archive_dir/" 2>/dev/null || true
            cp LICENSE "$archive_dir/" 2>/dev/null || true
            cp QUICK_REFERENCE.md "$archive_dir/" 2>/dev/null || true

            # Create install script for Unix systems
            if [ "$goos" != "windows" ]; then
                cat > "$archive_dir/install.sh" << EOF
#!/bin/bash
# Installation script for $BINARY_NAME

set -e

INSTALL_DIR="\${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="$binary_name"

echo "Installing $BINARY_NAME to \$INSTALL_DIR..."

# Check if install directory exists
if [ ! -d "\$INSTALL_DIR" ]; then
    echo "Creating install directory: \$INSTALL_DIR"
    sudo mkdir -p "\$INSTALL_DIR"
fi

# Copy binary
sudo cp "\$BINARY_NAME" "\$INSTALL_DIR/$BINARY_NAME"
sudo chmod +x "\$INSTALL_DIR/$BINARY_NAME"

echo "Installation completed! You can now use '$BINARY_NAME' command."
echo "Run '$BINARY_NAME --help' to get started."
EOF
                chmod +x "$archive_dir/install.sh"
            fi

            # Create archive
            cd "$DIST_DIR"
            if [ "$goos" = "windows" ]; then
                # Create ZIP for Windows
                zip -r "${BINARY_NAME}-${VERSION}-$goos-$goarch.zip" "${BINARY_NAME}-${VERSION}-$goos-$goarch/"
                log_success "Created ${BINARY_NAME}-${VERSION}-$goos-$goarch.zip"
            else
                # Create tar.gz for Unix systems
                tar -czf "${BINARY_NAME}-${VERSION}-$goos-$goarch.tar.gz" "${BINARY_NAME}-${VERSION}-$goos-$goarch/"
                log_success "Created ${BINARY_NAME}-${VERSION}-$goos-$goarch.tar.gz"
            fi
            cd - > /dev/null

            # Remove temporary directory
            rm -rf "$archive_dir"
        else
            log_warning "Binary not found for $goos/$goarch, skipping archive creation"
        fi
    done
}

# Generate checksums
generate_checksums() {
    log_info "Generating checksums..."

    cd "$DIST_DIR"

    # Generate SHA256 checksums
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum *.tar.gz *.zip 2>/dev/null > "checksums-sha256.txt" || true
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 *.tar.gz *.zip 2>/dev/null > "checksums-sha256.txt" || true
    else
        log_warning "No SHA256 utility found, skipping checksum generation"
    fi

    # Generate MD5 checksums
    if command -v md5sum >/dev/null 2>&1; then
        md5sum *.tar.gz *.zip 2>/dev/null > "checksums-md5.txt" || true
    elif command -v md5 >/dev/null 2>&1; then
        md5 *.tar.gz *.zip 2>/dev/null > "checksums-md5.txt" || true
    else
        log_warning "No MD5 utility found, skipping MD5 checksum generation"
    fi

    cd - > /dev/null

    if [ -f "$DIST_DIR/checksums-sha256.txt" ]; then
        log_success "Generated checksums"
    fi
}

# Show build summary
show_summary() {
    log_info "Build Summary:"
    echo "Version: $VERSION"
    echo "Commit: $COMMIT_HASH"
    echo "Build Time: $BUILD_TIME"
    echo

    log_info "Built binaries:"
    if [ -d "$BUILD_DIR" ]; then
        find "$BUILD_DIR" -type f -name "${BINARY_NAME}-*" -exec basename {} \; | sort
    fi

    echo
    log_info "Distribution archives:"
    if [ -d "$DIST_DIR" ]; then
        find "$DIST_DIR" -type f \( -name "*.tar.gz" -o -name "*.zip" \) -exec basename {} \; | sort
    fi

    echo
    log_info "Total size:"
    if [ -d "$DIST_DIR" ]; then
        du -sh "$DIST_DIR" | cut -f1
    fi
}

# Validate build
validate_build() {
    log_info "Validating builds..."

    local failed=0

    for platform in "${PLATFORMS[@]}"; do
        local goos=${platform%/*}
        local goarch=${platform#*/}
        local binary_name="$BINARY_NAME-$goos-$goarch"

        if [ "$goos" = "windows" ]; then
            binary_name="$binary_name.exe"
        fi

        local binary_path="$BUILD_DIR/$binary_name"

        if [ -f "$binary_path" ]; then
            # Check if binary is executable (for current platform)
            if [[ "$OSTYPE" == "linux-gnu"* || "$OSTYPE" == "darwin"* ]] &&
               [[ "$goos" == "linux" || "$goos" == "darwin" ]] &&
               [[ "$(uname -m)" == "x86_64" && "$goarch" == "amd64" || "$(uname -m)" == "arm64" && "$goarch" == "arm64" ]]; then

                if "$binary_path" --help >/dev/null 2>&1; then
                    log_success "✓ $binary_name - executable and functional"
                else
                    log_error "✗ $binary_name - not functional"
                    failed=$((failed + 1))
                fi
            else
                log_info "✓ $binary_name - built successfully (cross-compiled)"
            fi
        else
            log_error "✗ $binary_name - missing"
            failed=$((failed + 1))
        fi
    done

    if [ $failed -eq 0 ]; then
        log_success "All builds validated successfully!"
        return 0
    else
        log_error "$failed builds failed validation"
        return 1
    fi
}

# Main build function
main() {
    log_info "Starting cross-platform build for SlimAcademy CLI"
    log_info "Version: $VERSION"
    log_info "Commit: $COMMIT_HASH"

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed"
        exit 1
    fi

    log_info "Go version: $(go version)"

    # Clean and prepare
    clean_build

    # Build for all platforms
    local failed_builds=0
    for platform in "${PLATFORMS[@]}"; do
        if ! build_platform "$platform"; then
            failed_builds=$((failed_builds + 1))
        fi
    done

    if [ $failed_builds -gt 0 ]; then
        log_warning "$failed_builds builds failed"
    fi

    # Create distribution archives
    create_archives

    # Generate checksums
    generate_checksums

    # Validate builds
    validate_build

    # Show summary
    show_summary

    log_success "Cross-platform build completed!"

    if [ $failed_builds -eq 0 ]; then
        log_info "All platforms built successfully! 🚀"
        exit 0
    else
        log_warning "Some builds failed, but proceeding with available builds"
        exit 1
    fi
}

# Help function
show_help() {
    cat << EOF
Usage: $0 [options]

Cross-platform build script for SlimAcademy CLI

Options:
    -h, --help          Show this help message
    -c, --clean-only    Only clean build directories
    -v, --version VER   Set version string (default: git describe)

Environment Variables:
    VERSION             Version string to embed in binaries
    BUILD_DIR           Directory for built binaries (default: bin)
    DIST_DIR            Directory for distribution archives (default: dist)

Examples:
    $0                  # Build all platforms
    $0 --clean-only     # Just clean build directories
    VERSION=v1.0.0 $0   # Build with specific version

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--clean-only)
            clean_build
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
