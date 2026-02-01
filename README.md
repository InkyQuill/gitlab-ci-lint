# GitLab CI Lint

[![CI](https://github.com/InkyQuill/gitlab-ci-lint/workflows/CI/badge.svg)](https://github.com/InkyQuill/gitlab-ci-lint/actions/workflows/ci.yml)
[![Release](https://github.com/InkyQuill/gitlab-ci-lint/workflows/Release/badge.svg)](https://github.com/InkyQuill/gitlab-ci-lint/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/InkyQuill/gitlab-ci-lint/branch/main/graph/badge.svg)](https://codecov.io/gh/InkyQuill/gitlab-ci-lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/InkyQuill/gitlab-ci-lint)](https://goreportcard.com/report/github.com/InkyQuill/gitlab-ci-lint)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A fast, flexible GitLab CI/CD configuration linter that validates `.gitlab-ci.yml` files.

## Features

- ⚡ **Fast local validation** - YAML syntax checking without API calls
- 🔌 **Optional API validation** - Complete semantic validation with GitLab instance
- 🔍 **Smart file discovery** - Auto-discover CI files, batch validation, directory scanning
- 🎨 **Multiple output formats** - Text, JSON, YAML
- 🔧 **Flexible configuration** - File, environment variables, CLI flags
- 🎯 **Single binary** - No dependencies, written in Go
- 🌍 **Cross-platform** - Linux, macOS, Windows
- 🚀 **Interactive setup** - Easy configuration wizard

## Quick Start

### 1. Install

```bash
# macOS / Linux (one-line install)
curl -s https://api.github.com/repos/InkyQuill/gitlab-ci-lint/releases/latest | \
  grep "browser_download_url.*$(uname | tr '[:upper:]' '[:lower:]')-$(uname -m)" | \
  cut -d '"' -f 4 | xargs -I {} curl -L -o gitlab-ci-lint {} && \
  chmod +x gitlab-ci-lint && sudo mv gitlab-ci-lint /usr/local/bin/
```

### 2. Configure

```bash
# Run interactive setup
gitlab-ci-lint setup
```

The wizard will help you configure:
- GitLab instance URL (e.g., `https://gitlab.com` or self-hosted)
- Personal access token for API validation
- Default project (optional)
- Output preferences

### 3. Validate

```bash
# Local validation only (no API required)
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Full validation with GitLab API
export GCL_TOKEN=glpat-your-token
gitlab-ci-lint .gitlab-ci.yml

# Project-specific validation (validates job references)
gitlab-ci-lint --project group/project .gitlab-ci.yml

# Auto-discovery (searches current and parent directories)
gitlab-ci-lint
```

## Installation

### Quick Install (Latest Release)

#### macOS / Linux

Download and install the latest binary for your platform:

```bash
# Detect platform and download latest binary
curl -s https://api.github.com/repos/InkyQuill/gitlab-ci-lint/releases/latest | \
  grep "browser_download_url.*$(uname | tr '[:upper:]' '[:lower:]')-$(uname -m)" | \
  cut -d '"' -f 4 | \
  xargs -I {} curl -L -o gitlab-ci-lint {}

# Make executable and install to ~/.local/bin
chmod +x gitlab-ci-lint
mkdir -p ~/.local/bin
mv gitlab-ci-lint ~/.local/bin/

# Add ~/.local/bin to PATH if not already present
if [[ :$PATH: != *":$HOME/.local/bin:"* ]]; then
  echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc  # or ~/.zshrc
  export PATH=$HOME/.local/bin:$PATH
fi

# Verify installation
gitlab-ci-lint version
```

#### Linux (AMD64)

```bash
# Download latest binary
curl -L -o gitlab-ci-lint https://github.com/InkyQuill/gitlab-ci-lint/releases/latest/download/gitlab-ci-lint-linux-amd64

# Install
chmod +x gitlab-ci-lint
sudo mv gitlab-ci-lint /usr/local/bin/

# Verify
gitlab-ci-lint version
```

#### macOS (Apple Silicon)

```bash
# Download latest binary
curl -L -o gitlab-ci-lint https://github.com/InkyQuill/gitlab-ci-lint/releases/latest/download/gitlab-ci-lint-darwin-arm64

# Install
chmod +x gitlab-ci-lint
sudo mv gitlab-ci-lint /usr/local/bin/

# Verify
gitlab-ci-lint version
```

#### Windows (PowerShell)

```powershell
# Download latest binary
Invoke-WebRequest -Uri "https://github.com/InkyQuill/gitlab-ci-lint/releases/latest/download/gitlab-ci-lint-windows-amd64.exe" -OutFile "gitlab-ci-lint.exe"

# Create directory and move binary
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"
Move-Item -Path gitlab-ci-lint.exe -Destination "$env:USERPROFILE\bin\"

# Add to PATH (if not already present)
$binPath = "$env:USERPROFILE\bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binPath*") {
  [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binPath", "User")
  $env:Path += ";$binPath"
}

# Verify installation (restart terminal first)
gitlab-ci-lint version
```

### Build from Source

If you have Go 1.24 or later installed:

```bash
# Clone the repository
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint

# Build for your current platform
make build

# Install to GOPATH/bin (or ~/.local/bin)
make install

# The binary will be available as:
# - $(go env GOPATH)/bin/gitlab-ci-lint
# - or: ./build/gitlab-ci-lint
```

#### Build for Specific Platform

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o gitlab-ci-lint cmd/gitlab-ci-lint/main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gitlab-ci-lint cmd/gitlab-ci-lint/main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o gitlab-ci-lint.exe cmd/gitlab-ci-lint/main.go
```

### Go Install

If you have Go installed:

```bash
go install github.com/InkyQuill/gitlab-ci-lint/cmd/gitlab-ci-lint@latest

# The binary will be installed to $(go env GOPATH)/bin (usually ~/go/bin/)
# Make sure $GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

## Configuration

### Quick Setup (Recommended)

The interactive setup wizard will guide you through configuration:

```bash
gitlab-ci-lint setup
```

The wizard will help you configure:
- GitLab instance URL (e.g., `https://gitlab.com` or your self-hosted instance)
- Personal access token for API validation
- Default project (optional)
- Output preferences

Configuration is saved to `~/.tools-config/.gitlab-ci-lint/config.yaml`

### Environment Variables

You can configure gitlab-ci-lint using environment variables (prefix `GCL_`):

```bash
# GitLab instance URL
export GCL_INSTANCE=https://gitlab.example.com

# Personal access token (requires 'api' scope)
export GCL_TOKEN=glpat-xxxxxxxxxxxxxxx

# Default project (for project-specific validation)
export GCL_PROJECT=mygroup/myproject

# Output format
export GCL_OUTPUT=json

# Skip API validation (local YAML syntax only)
export GCL_SKIP_API=true

# Verbose output
export GCL_VERBOSE=true

# Disable colored output
export GCL_COLOR=never
```

Add these to your `~/.bashrc`, `~/.zshrc`, or shell profile to make them persistent.

### Configuration File

Configuration is loaded from multiple sources (low to high priority):

1. **Defaults** - Built-in sensible defaults
2. **Config file** - `~/.tools-config/.gitlab-ci-lint/config.yaml`
3. **Environment variables** - Prefix `GCL_`
4. **CLI flags** - Command-line arguments

#### Config File Example

```yaml
# ~/.tools-config/.gitlab-ci-lint/config.yaml
gitlab:
  instance: "https://gitlab.com"  # Your GitLab instance URL
  timeout: 30s                    # API request timeout

auth:
  token: "glpat-xxxxxxxxxxxx"     # Or use GCL_TOKEN environment variable
  netrc: false                    # Use .netrc for credentials

validation:
  skip_api: false                 # Skip API validation (local only)
  strict: true                    # Use strict YAML parsing
  project: "group/project"        # Default project ID or path

output:
  format: "text"                  # Output format: text|json|yaml
  verbose: false                  # Verbose output
  color: "auto"                   # Color output: auto|always|never

files:
  max_depth: 5                    # Max depth for directory scanning
  search_parent: true             # Search parent directories
  ignore_patterns:                # Patterns to ignore during discovery
    - ".git"
    - "node_modules"
    - "vendor"
    - "build"
    - "dist"
    - "*.tar.gz"
```

### Getting Your GitLab Token

For API validation, you need a GitLab personal access token:

1. Go to your GitLab instance
2. Navigate to **User Settings** → **Access Tokens**
3. Create a new token with `api` scope
4. Use the token via:
   - Environment variable: `export GCL_TOKEN=glpat-xxxxx`
   - Setup wizard: `gitlab-ci-lint setup`
   - Config file: Add `auth.token` field
   - CLI flag: `gitlab-ci-lint --token glpat-xxxxx`

## Documentation

- 📚 [Wiki](https://github.com/InkyQuill/gitlab-ci-lint/wiki) - Full documentation
- 🚀 [Quick Start](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Quick-Start) - Get started in minutes
- 🔧 [Configuration Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Configuration-Guide) - Detailed configuration options
- 🏗️ [Architecture Overview](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Architecture-Overview) - System design
- 📖 [Basic Usage](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Basic-Usage) - Usage examples
- 🐛 [Troubleshooting](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Troubleshooting) - Common issues and solutions
- 🔌 [CI/CD Integration](https://github.com/InkyQuill/gitlab-ci-lint/wiki/CI-CD-Integration) - Integration examples
- 💻 [API Reference](https://github.com/InkyQuill/gitlab-ci-lint/wiki/API-Reference) - Developer documentation
- 🤝 [Contributing](CONTRIBUTING.md) - Contribution guidelines

## Usage

```bash
gitlab-ci-lint [flags] [file]

Available Commands:
  setup       Interactive configuration wizard
  version     Show version information
  help        Help about any command

Flags:
  -c, --config string       Path to config file
  -t, --token string        GitLab personal access token
      --instance string     GitLab instance URL (default: https://gitlab.com)
      --project string      Project ID (e.g., "123" or "group/project")
  -s, --skip-api            Skip API validation (local only)
  -o, --output string       Output format: text|json|yaml (default: text)
  -v, --verbose             Verbose output
      --color string        Color output: auto|always|never
  -f, --file strings        Path(s) to CI file(s). Can be specified multiple times.
  -d, --directory strings   Directory path(s) to recursively search for CI files.
```

## Examples

```bash
# Auto-discovery (searches current and parent directories)
gitlab-ci-lint

# Validate multiple files
gitlab-ci-lint -f .gitlab-ci.yml -f ci/frontend.yml

# Recursively validate all CI files in a directory
gitlab-ci-lint -d ./monorepo

# Combined: specific files and directory scanning
gitlab-ci-lint -f .gitlab-ci.yml -d ./services

# Quick syntax check during development
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Full validation with custom instance
gitlab-ci-lint --instance https://gitlab.example.com .gitlab-ci.yml

# JSON output for CI/CD pipelines
gitlab-ci-lint --output json .gitlab-ci.yml | jq .valid

# Project-specific validation with job references
gitlab-ci-lint --project mygroup/myproject .gitlab-ci.yml

# Validate from stdin
cat .gitlab-ci.yml | gitlab-ci-lint -
```

## Exit Codes

- `0` - All validations passed
- `1` - Runtime error (file not found, auth failed, etc)
- `10` - CI configuration is invalid

Use in scripts:

```bash
#!/bin/bash
gitlab-ci-lint .gitlab-ci.yml
case $? in
  0) echo "✓ Valid" ;;
  1) echo "✗ Error" ;;
  10) echo "✗ Invalid configuration" ;;
esac
```

## Configuration

Configuration is loaded from multiple sources (low to high priority):

1. **Defaults** - Built-in sensible defaults
2. **Config file** - `~/.tools-config/.gitlab-ci-lint/config.yaml`
3. **Environment variables** - Prefix `GCL_`
4. **CLI flags** - Command-line arguments

### Config File Example

```yaml
# ~/.tools-config/.gitlab-ci-lint/config.yaml
gitlab:
  instance: "https://gitlab.com"
  timeout: 30s

auth:
  token: "glpat-xxxxxxxxxxxx"  # Or use GCL_TOKEN

validation:
  skip_api: false
  strict: true
  project: "group/project"

output:
  format: "text"
  verbose: false
  color: "auto"
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development

```bash
# Run tests
make test-unit

# Run linters
make lint

# Format code
make fmt

# Build
make build
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Project Status

- [x] Two-stage validation system
- [x] Interactive setup wizard
- [x] Multi-format output
- [x] Comprehensive tests (72%+ coverage)
- [x] Semantic release automation
- [x] CI/CD workflows

See [ROADMAP.md](ROADMAP.md) for future plans.
