#!/bin/sh
# Tero CLI installer
# Usage: curl -sSfL https://sh.usetero.com | sh
# Or: curl -sSfL https://sh.usetero.com | sh -s -- --help

set -e

# Configuration
REPO="usetero/cli"
BINARY_NAME="tero"
INSTALL_DIR="${TERO_INSTALL_DIR:-$HOME/.tero/bin}"
GITHUB_API="https://api.github.com"
GITHUB_DOWNLOAD="https://github.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
SKIP_PROMPTS=0
REQUESTED_VERSION=""

while [ $# -gt 0 ]; do
  case "$1" in
    -y|--yes)
      SKIP_PROMPTS=1
      shift
      ;;
    --version=*)
      REQUESTED_VERSION="${1#*=}"
      shift
      ;;
    --prefix=*)
      INSTALL_DIR="${1#*=}"
      shift
      ;;
    --help|-h)
      cat << EOF
Tero CLI Installer

Usage:
  curl -sSfL https://sh.usetero.com | sh
  curl -sSfL https://sh.usetero.com | sh -s -- [OPTIONS]

Options:
  -y, --yes              Skip confirmation prompts
  --version=VERSION      Install specific version (e.g., 1.1.0)
  --prefix=PATH          Install to custom directory (default: ~/.tero/bin)
  -h, --help             Show this help message

Environment Variables:
  TERO_INSTALL_DIR       Installation directory (default: ~/.tero/bin)
  TERO_VERSION           Version to install

Examples:
  # Install latest version
  curl -sSfL https://sh.usetero.com | sh

  # Install specific version
  curl -sSfL https://sh.usetero.com | sh -s -- --version=1.1.0

  # Install to custom location
  curl -sSfL https://sh.usetero.com | sh -s -- --prefix=/usr/local/bin

EOF
      exit 0
      ;;
    *)
      echo "${RED}Unknown option: $1${NC}"
      echo "Run with --help for usage information"
      exit 1
      ;;
  esac
done

# Functions
log_info() {
  printf "${GREEN}==>${NC} %s\n" "$1"
}

log_warn() {
  printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

log_error() {
  printf "${RED}Error:${NC} %s\n" "$1" >&2
}

detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux*)
      OS="linux"
      ;;
    Darwin*)
      OS="darwin"
      ;;
    *)
      log_error "Unsupported operating system: $OS"
      log_error "Tero currently supports Linux and macOS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64|amd64)
      ARCH="amd64"
      ;;
    arm64|aarch64)
      ARCH="arm64"
      ;;
    *)
      log_error "Unsupported architecture: $ARCH"
      log_error "Tero currently supports amd64 and arm64"
      exit 1
      ;;
  esac

  log_info "Detected platform: ${OS}_${ARCH}"
}

check_dependencies() {
  for cmd in curl tar; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      log_error "Required command not found: $cmd"
      log_error "Please install $cmd and try again"
      exit 1
    fi
  done
}

get_latest_version() {
  log_info "Fetching latest version..."

  # Try to get latest release from GitHub API
  LATEST_VERSION=$(curl -sSfL "$GITHUB_API/repos/$REPO/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"v([^"]+)".*/\1/')

  if [ -z "$LATEST_VERSION" ]; then
    log_error "Failed to fetch latest version"
    exit 1
  fi

  echo "$LATEST_VERSION"
}

confirm_install() {
  if [ "$SKIP_PROMPTS" -eq 1 ]; then
    return 0
  fi

  printf "Install Tero CLI %s to %s? [y/N] " "$VERSION" "$INSTALL_DIR"
  read -r response
  case "$response" in
    [yY][eE][sS]|[yY])
      return 0
      ;;
    *)
      log_info "Installation cancelled"
      exit 0
      ;;
  esac
}

download_and_install() {
  ARCHIVE_NAME="tero_${VERSION}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="$GITHUB_DOWNLOAD/$REPO/releases/download/v${VERSION}/$ARCHIVE_NAME"

  log_info "Downloading Tero CLI v${VERSION}..."
  log_info "URL: $DOWNLOAD_URL"

  # Create temporary directory
  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT

  # Download archive
  if ! curl -sSfL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"; then
    log_error "Failed to download Tero CLI"
    log_error "Make sure version $VERSION exists at:"
    log_error "$DOWNLOAD_URL"
    exit 1
  fi

  log_info "Extracting archive..."
  tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

  # Create installation directory if it doesn't exist
  mkdir -p "$INSTALL_DIR"

  # Install binary
  log_info "Installing binary to $INSTALL_DIR..."
  mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  log_info "${GREEN}Successfully installed Tero CLI v${VERSION}!${NC}"
}

check_path() {
  if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    return 0
  fi
  return 1
}

print_path_instructions() {
  if check_path; then
    return 0
  fi

  echo ""
  log_warn "$INSTALL_DIR is not in your PATH"
  echo ""
  echo "Add it to your PATH by running:"
  echo ""

  # Detect shell and provide appropriate instructions
  SHELL_NAME="$(basename "$SHELL")"
  case "$SHELL_NAME" in
    bash)
      echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc"
      echo "  source ~/.bashrc"
      ;;
    zsh)
      echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc"
      echo "  source ~/.zshrc"
      ;;
    fish)
      echo "  fish_add_path $INSTALL_DIR"
      ;;
    *)
      echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
      echo ""
      echo "Add this to your shell's profile file to make it permanent"
      ;;
  esac

  echo ""
  echo "Or run Tero directly:"
  echo "  $INSTALL_DIR/$BINARY_NAME"
  echo ""
}

verify_installation() {
  if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    INSTALLED_VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
    log_info "Installed version: $INSTALLED_VERSION"

    if ! check_path; then
      print_path_instructions
    else
      echo ""
      log_info "Run 'tero' to get started!"
      echo ""
    fi
  else
    log_error "Installation verification failed"
    exit 1
  fi
}

# Main installation flow
main() {
  echo ""
  echo "Tero CLI Installer"
  echo "=================="
  echo ""

  detect_platform
  check_dependencies

  # Determine version to install
  if [ -n "$REQUESTED_VERSION" ]; then
    VERSION="$REQUESTED_VERSION"
    log_info "Installing requested version: $VERSION"
  elif [ -n "$TERO_VERSION" ]; then
    VERSION="$TERO_VERSION"
    log_info "Installing version from TERO_VERSION: $VERSION"
  else
    VERSION=$(get_latest_version)
    log_info "Latest version: $VERSION"
  fi

  confirm_install
  download_and_install
  verify_installation
}

main
