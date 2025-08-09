#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
REPO_OWNER="nathabonfim59"
REPO_NAME="cerebras-code-monitor"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="cerebras-monitor"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS and architecture
detect_platform() {
    local os arch
    
    case "$OSTYPE" in
        linux*)   os="linux" ;;
        darwin*)  os="darwin" ;;
        *)        
            print_error "Unsupported OS: $OSTYPE"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64)   arch="amd64" ;;
        arm64)    arch="arm64" ;;
        aarch64)  arch="arm64" ;;
        *)        
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Function to get latest release info
get_latest_release() {
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s "$api_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$api_url"
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
}

# Function to extract tag name from JSON response
extract_tag_name() {
    local json="$1"
    echo "$json" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# Function to download and extract release
download_and_install() {
    local tag_name="$1"
    local platform="$2"
    local archive_name="${REPO_NAME}-${platform}.tar.gz"
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag_name}/${archive_name}"
    local temp_dir
    
    temp_dir=$(mktemp -d)
    trap "rm -rf '$temp_dir'" EXIT
    
    print_status "Downloading ${archive_name}..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_dir/$archive_name" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_dir/$archive_name" "$download_url"
    else
        print_error "Neither curl nor wget is available."
        exit 1
    fi
    
    if [ ! -f "$temp_dir/$archive_name" ]; then
        print_error "Download failed: $archive_name not found"
        exit 1
    fi
    
    print_status "Extracting archive..."
    cd "$temp_dir"
    tar -xzf "$archive_name"
    
    # Find the binary (it has platform-specific name in archive)
    local binary_path
    local archive_binary_name="${REPO_NAME}-${platform}"
    binary_path=$(find . -name "$archive_binary_name" -type f | head -1)
    
    if [ -z "$binary_path" ]; then
        print_error "Binary $archive_binary_name not found in archive"
        exit 1
    fi
    
    print_status "Installing to $INSTALL_DIR..."
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary and make it executable
    cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    print_success "$BINARY_NAME installed successfully to $INSTALL_DIR"
}

# Function to check if binary is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_warning "$INSTALL_DIR is not in your PATH."
        print_warning "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# Main function
main() {
    print_status "Installing $REPO_NAME..."
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_status "Detected platform: $platform"
    
    # Get latest release info
    print_status "Fetching latest release information..."
    local release_json
    release_json=$(get_latest_release)
    
    if [ -z "$release_json" ]; then
        print_error "Failed to fetch release information"
        exit 1
    fi
    
    # Extract tag name
    local tag_name
    tag_name=$(extract_tag_name "$release_json")
    
    if [ -z "$tag_name" ]; then
        print_error "Failed to extract tag name from release information"
        exit 1
    fi
    
    print_status "Latest release: $tag_name"
    
    # Check if already installed and up to date
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local current_version
        current_version=$("$BINARY_NAME" --version 2>/dev/null | head -1 || echo "unknown")
        print_status "Current version: $current_version"
        
        if [[ "$current_version" == *"$tag_name"* ]]; then
            print_success "$BINARY_NAME is already up to date ($tag_name)"
            exit 0
        fi
    fi
    
    # Download and install
    download_and_install "$tag_name" "$platform"
    
    # Check PATH
    check_path
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version
        installed_version=$("$BINARY_NAME" --version 2>/dev/null | head -1 || echo "unknown")
        print_success "Installation complete! Version: $installed_version"
        print_status "Run '$BINARY_NAME --help' to get started."
    else
        print_warning "Installation complete, but $BINARY_NAME is not in PATH."
        print_status "You can run it directly: $INSTALL_DIR/$BINARY_NAME --help"
    fi
}

# Run main function
main "$@"