# Installation Guide

This guide covers various installation methods for GitLab CI Linter.

## Prerequisites

- Go 1.24+ (if building from source)
- Git (if building from source)

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

1. Download the latest binary for your platform from [Releases](https://github.com/InkyQuill/gitlab-ci-lint/releases)
2. Extract the archive
3. Move the binary to your PATH

```bash
# Example for Linux
wget https://github.com/InkyQuill/gitlab-ci-lint/releases/latest/download/gitlab-ci-lint-linux-amd64.tar.gz
tar -xzf gitlab-ci-lint-linux-amd64.tar.gz
sudo mv gitlab-ci-lint /usr/local/bin/
```

### Method 2: Go Install

If you have Go installed:

```bash
go install github.com/InkyQuill/gitlab-ci-lint/cmd/gitlab-ci-lint@latest
```

The binary will be installed to `$GOPATH/bin` (usually `~/go/bin/`).

### Method 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint

# Build for your current platform
make build

# Or build for all supported platforms
make build-all

# The binaries will be in ./build/
ls -lh build/
```

#### Build for Specific Platform

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o gitlab-ci-lint-linux-amd64 cmd/gitlab-ci-lint/main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gitlab-ci-lint-darwin-arm64 cmd/gitlab-ci-lint/main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o gitlab-ci-lint-windows-amd64.exe cmd/gitlab-ci-lint/main.go
```

### Method 4: Using Makefile Install

```bash
# Clone the repository
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint

# Install to GOPATH/bin
make install
```

## Verification

After installation, verify it works:

```bash
gitlab-ci-lint version
```

You should see version information.

## Configuration

After installation, you can create a configuration file:

```bash
# Create config directory
mkdir -p ~/.tool-configs/.gitlab-ci-lint

# Create config file
cat > ~/.tool-configs/.gitlab-ci-lint/config.yaml <<EOF
gitlab:
  instance: "https://gitlab.com"
  timeout: 30

auth:
  token: ""
  netrc: false

validation:
  skip_api: false
  strict: true

output:
  format: "text"
  verbose: false
  color: "auto"
EOF
```

## GitLab Personal Access Token

For API validation, you'll need a GitLab personal access token:

1. Go to your GitLab instance
2. Navigate to **User Settings** > **Access Tokens**
3. Create a new token with `api` scope
4. Set the `GCL_TOKEN` environment variable:

```bash
export GCL_TOKEN=glpat-xxxxxxxxxxxx
```

Or add it to your configuration file (see [CONFIG.md](CONFIG.md)).

## Shell Completion

### Bash

```bash
# Generate completion script
gitlab-ci-lint completion bash > /etc/bash_completion.d/gitlab-ci-lint

# Or add to ~/.bash_completion
gitlab-ci-lint completion bash >> ~/.bash_completion
source ~/.bash_completion
```

### Zsh

```bash
# Generate completion script
gitlab-ci-lint completion zsh > /usr/local/share/zsh/site-functions/_gitlab-ci-lint

# Or add to ~/.zshrc
gitlab-ci-lint completion zsh >> ~/.zshrc
```

### Fish

```bash
gitlab-ci-lint completion fish > ~/.config/fish/completions/gitlab-ci-lint.fish
```

## Uninstallation

### Remove Binary

```bash
# If installed to /usr/local/bin
sudo rm /usr/local/bin/gitlab-ci-lint

# If installed via go install
rm $(go env GOPATH)/bin/gitlab-ci-lint
```

### Remove Configuration

```bash
rm -rf ~/.tool-configs/.gitlab-ci-lint
```

### Remove from Shell Completion

Remove the completion scripts added during installation.

## Troubleshooting

### "command not found: gitlab-ci-lint"

Ensure the binary is in your PATH:

```bash
echo $PATH
which gitlab-ci-lint
```

Add to PATH if needed:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:/path/to/gitlab-ci-lint
```

### "permission denied"

Make the binary executable:

```bash
chmod +x gitlab-ci-lint
```

### Build Errors

If building from source fails:

```bash
# Ensure you have Go 1.24+
go version

# Update dependencies
cd gitlab-ci-lint
go mod tidy

# Clean and rebuild
make clean
make build
```

## Next Steps

- Read the [Configuration Reference](CONFIG.md)
- Check out the [README.md](README.md) for usage examples
