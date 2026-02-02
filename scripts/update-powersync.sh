#!/bin/bash
set -euo pipefail

# Update PowerSync SQLite extension to latest release
# Usage: ./scripts/update-powersync.sh [version]
# If version is omitted, fetches the latest release.

REPO="powersync-ja/powersync-sqlite-core"
DEST_DIR="internal/powersync/extensions"

# Get version (argument or latest)
if [ $# -ge 1 ]; then
    VERSION="$1"
else
    echo "Fetching latest release..."
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Error: Could not determine latest version"
        exit 1
    fi
fi

echo "Updating PowerSync extension to ${VERSION}"

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download each platform binary
download() {
    local asset_name="$1"
    local dest_name="$2"
    local url="https://github.com/${REPO}/releases/download/${VERSION}/${asset_name}"

    echo "Downloading ${asset_name}..."
    if ! curl -sL -o "${TEMP_DIR}/${dest_name}" "$url"; then
        echo "Error: Failed to download ${asset_name}"
        return 1
    fi

    # Verify it's not an error page
    if file "${TEMP_DIR}/${dest_name}" | grep -q "HTML"; then
        echo "Error: ${asset_name} not found in release ${VERSION}"
        return 1
    fi
}

# Map release assets to our naming convention
download "libpowersync_aarch64.dylib" "libpowersync_aarch64.macos.dylib"
download "libpowersync_x64.dylib" "libpowersync_x64.macos.dylib"
download "libpowersync_aarch64.so" "libpowersync_aarch64.linux.so"
download "libpowersync_x64.so" "libpowersync_x64.linux.so"

# Move to destination
echo "Installing to ${DEST_DIR}..."
mkdir -p "$DEST_DIR"
mv "${TEMP_DIR}"/*.dylib "${TEMP_DIR}"/*.so "$DEST_DIR/"

echo ""
echo "Updated PowerSync extension to ${VERSION}"
echo "Files:"
ls -lh "$DEST_DIR"/libpowersync*
