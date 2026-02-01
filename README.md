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
- 🎨 **Multiple output formats** - Text, JSON, YAML
- 🔧 **Flexible configuration** - File, environment variables, CLI flags
- 🎯 **Single binary** - No dependencies, written in Go
- 🌍 **Cross-platform** - Linux, macOS, Windows
- 🚀 **Interactive setup** - Easy configuration wizard

## Quick Start

### Interactive Setup (Recommended)

```bash
gitlab-ci-lint setup
```

The wizard will guide you through configuring your GitLab instance, token, and preferences.

### Basic Usage

```bash
# Local validation only (no API required)
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Full validation with GitLab API
export GCL_TOKEN=glpat-your-token
gitlab-ci-lint .gitlab-ci.yml

# Project-specific validation
gitlab-ci-lint --project group/project .gitlab-ci.yml
```

## Installation

### From Binaries

Download the latest release from [GitHub Releases](https://github.com/InkyQuill/gitlab-ci-lint/releases).

### Build from Source

```bash
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint
make build && make install
```

### Go Install

```bash
go install github.com/InkyQuill/gitlab-ci-lint/cmd/gitlab-ci-lint@latest
```

## Documentation

- 📚 [Quick Start Guide](docs/guides/quick-start.md) - Get started in minutes
- 🔧 [Configuration Reference](CONFIG.md) - Detailed configuration options
- 🏗️ [Architecture Overview](docs/architecture/overview.md) - System design
- 📖 [Examples](docs/examples/) - Usage examples
- 🐛 [Troubleshooting](docs/guides/troubleshooting.md) - Common issues and solutions
- 💻 [API Reference](docs/api/reference.md) - Developer documentation
- 🤝 [Contributing](CONTRIBUTING.md) - Contribution guidelines

## Usage

```bash
gitlab-ci-lint [flags] <file>

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
```

## Examples

```bash
# Quick syntax check during development
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Full validation with custom instance
gitlab-ci-lint --instance https://gitlab.example.com .gitlab-ci.yml

# JSON output for CI/CD pipelines
gitlab-ci-lint --output json .gitlab-ci.yml | jq .valid

# Project-specific validation with job references
gitlab-ci-lint --project mygroup/myproject .gitlab-ci.yml
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
