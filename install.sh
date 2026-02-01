#!/usr/bin/env bash
set -e

# GitLab CI Lint Installation Script
# This script downloads and installs the latest release of gitlab-ci-lint

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect platform
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        echo "Supported: x86_64, aarch64, arm64, i386, i686"
        exit 1
        ;;
esac

# Get latest version
echo -e "${GREEN}Fetching latest version...${NC}"
VERSION=$(curl -s https://api.github.com/repos/InkyQuill/gitlab-ci-lint/releases/latest | grep -E '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not determine latest version${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: v${VERSION}${NC}"
echo -e "${GREEN}Platform: ${OS}_${ARCH}${NC}"

# Download and extract ONLY the binary
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="gitlab-ci-lint"
TEMP_DIR=$(mktemp -d)

echo -e "${GREEN}Downloading gitlab-ci-lint v${VERSION}...${NC}"

# Download tarball and extract only the binary
curl -sL "https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v${VERSION}/gitlab-ci-lint_${VERSION}_${OS}_${ARCH}.tar.gz" | \
    tar xz -O "${BINARY_NAME}" > "${TEMP_DIR}/${BINARY_NAME}"

# Make executable
chmod +x "${TEMP_DIR}/${BINARY_NAME}"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Move binary
mv "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
rm -rf "$TEMP_DIR"

echo -e "${GREEN}Binary installed to: ${INSTALL_DIR}/${BINARY_NAME}${NC}"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo -e "${YELLOW}Adding ~/.local/bin to PATH...${NC}"

    # Detect shell and add to appropriate config file
    if [ -n "$ZSH_VERSION" ]; then
        CONFIG_FILE="$HOME/.zshrc"
        echo 'export PATH=$HOME/.local/bin:$PATH' >> "$CONFIG_FILE"
        echo -e "${GREEN}Added to $CONFIG_FILE${NC}"
    elif [ -n "$BASH_VERSION" ]; then
        CONFIG_FILE="$HOME/.bashrc"
        echo 'export PATH=$HOME/.local/bin:$PATH' >> "$CONFIG_FILE"
        echo -e "${GREEN}Added to $CONFIG_FILE${NC}"
    else
        echo -e "${YELLOW}Could not detect shell. Please manually add:${NC}"
        echo -e "${YELLOW}export PATH=\$HOME/.local/bin:\$PATH${NC}"
        echo -e "${YELLOW}to your shell configuration file${NC}"
    fi

    echo -e "${YELLOW}Run: source ~/.bashrc (or ~/.zshrc) to update your current session${NC}"
fi

# Verify installation
echo ""
echo -e "${GREEN}Verifying installation...${NC}"
if "$INSTALL_DIR/$BINARY_NAME" version; then
    echo ""
    echo -e "${GREEN}✓ Installation successful!${NC}"
    echo -e "${GREEN}Run: ${BINARY_NAME} setup${NC} to configure GitLab instances."
else
    echo -e "${RED}✗ Installation verification failed${NC}"
    exit 1
fi
