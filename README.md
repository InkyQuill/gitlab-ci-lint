# GitLab CI Lint

[![CI](https://github.com/InkyQuill/gitlab-ci-lint/workflows/CI/badge.svg)](https://github.com/InkyQuill/gitlab-ci-lint/actions/workflows/ci.yml)
[![Release](https://github.com/InkyQuill/gitlab-ci-lint/workflows/Release/badge.svg)](https://github.com/InkyQuill/gitlab-ci-lint/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/InkyQuill/gitlab-ci-lint/branch/main/graph/badge.svg)](https://codecov.io/gh/InkyQuill/gitlab-ci-lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/InkyQuill/gitlab-ci-lint)](https://goreportcard.com/report/github.com/InkyQuill/gitlab-ci-lint)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A fast GitLab CI/CD configuration linter for `.gitlab-ci.yml` files. Two-stage validation: local YAML parsing and optional GitLab API validation.

**This project is developed with AI assistance** (Claude, Z.ai).

## Features

- **Local validation** — YAML syntax checking without API calls
- **Optional API validation** — Full semantic validation against a GitLab instance
- **Smart file discovery** — Auto-discover `.gitlab-ci.yml` in current and parent directories, recursive directory scanning
- **Multiple output formats** — Text, JSON, YAML
- **Flexible configuration** — Config file, environment variables, CLI flags
- **Single binary** — No external dependencies (Go)
- **Cross-platform** — Linux, macOS, Windows
- **Interactive setup** — `setup` wizard to save instances and tokens

## Quick Start

### 1. Install

**Linux / macOS (install script):**

```bash
curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/raw/main/install.sh | bash
```

The script installs the binary to `~/.local/bin`. If that directory is not in your `PATH`, add `export PATH=$HOME/.local/bin:$PATH` to your shell profile and run `source ~/.bashrc` (or `~/.zshrc`).

**With Go installed:**

```bash
go install github.com/InkyQuill/gitlab-ci-lint/cmd/gitlab-ci-lint@latest
```

The binary will be in `$(go env GOPATH)/bin` (typically `~/go/bin`). Ensure that directory is in your `PATH`.

### 2. Configure (optional)

For API validation you need a GitLab instance URL and a personal access token. Run the setup wizard:

```bash
gitlab-ci-lint setup
```

Settings are saved to `~/.tools-config/.gitlab-ci-lint/config.yaml`. Without configuration, only local YAML validation is available (use `--skip-api` or run without a token).

### 3. Validate

```bash
# Local validation only (no API)
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Auto-discover file in current and parent directories
gitlab-ci-lint

# Full validation via API (requires instance + token from setup or env)
gitlab-ci-lint .gitlab-ci.yml
```

## Installation (detailed)

### Install script (Linux / macOS)

Recommended: downloads the latest release for your platform and installs to `~/.local/bin`.

```bash
curl -sSL https://github.com/InkyQuill/gitlab-ci-lint/raw/main/install.sh | bash
```

### Manual install (Linux / macOS)

Run from any writable directory (e.g. `~/Downloads` or `/tmp`).

```bash
VERSION=$(curl -s https://api.github.com/repos/InkyQuill/gitlab-ci-lint/releases/latest | grep -E '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in x86_64) ARCH=amd64;; aarch64|arm64) ARCH=arm64;; i386|i686) ARCH=386;; *) echo "Unsupported: $ARCH"; exit 1;; esac

curl -sL "https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v${VERSION}/gitlab-ci-lint_${VERSION}_${OS}_${ARCH}.tar.gz" | tar xz -O gitlab-ci-lint > gitlab-ci-lint
chmod +x gitlab-ci-lint
mkdir -p ~/.local/bin
mv gitlab-ci-lint ~/.local/bin/

# Add to PATH if needed
[[ ":$PATH:" != *":$HOME/.local/bin:"* ]] && echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc
export PATH=$HOME/.local/bin:$PATH
```

### Windows (PowerShell)

```powershell
$version = (Invoke-RestMethod -Uri "https://api.github.com/repos/InkyQuill/gitlab-ci-lint/releases/latest").tag_name -replace '^v', ''
$zip = "gitlab-ci-lint_${version}_windows_amd64.zip"
Invoke-WebRequest -Uri "https://github.com/InkyQuill/gitlab-ci-lint/releases/download/v$version/$zip" -OutFile $zip -UseBasicParsing
Expand-Archive -Path $zip -DestinationPath . -Force
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin" | Out-Null
Move-Item -Path "gitlab-ci-lint.exe" -Destination "$env:USERPROFILE\bin\" -Force
Remove-Item $zip

$binPath = "$env:USERPROFILE\bin"
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$binPath*") {
  [Environment]::SetEnvironmentVariable("Path", "$path;$binPath", "User")
  $env:Path += ";$binPath"
}
```

### Build from source

Requires Go 1.24+.

```bash
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint

make build    # binary: ./build/gitlab-ci-lint
make install  # install to $(go env GOPATH)/bin
```

### Go install

```bash
go install github.com/InkyQuill/gitlab-ci-lint/cmd/gitlab-ci-lint@latest
# Binary at $(go env GOPATH)/bin/gitlab-ci-lint
export PATH=$PATH:$(go env GOPATH)/bin
```

## Configuration

Priority (low to high): defaults → config file → `GCL_*` environment variables → CLI flags.

### Config file

Default path: `~/.tools-config/.gitlab-ci-lint/config.yaml`. Created by `gitlab-ci-lint setup`.

Example (single instance and shared options):

```yaml
gitlab:
  instance: "https://gitlab.com"
  timeout: 30s

auth:
  token: ""           # prefer GCL_TOKEN env var
  netrc: false

validation:
  skip_api: false
  strict: true

output:
  format: "text"       # text | json | yaml
  verbose: false
  color: "auto"       # auto | always | never

files:
  search_parent: true
  max_depth: 10
  ignore_patterns:
    - ".git"
    - "node_modules"
    - "vendor"
    - "build"
    - "dist"
    - "*.tar.gz"
```

Project for project-specific validation is set only via `GCL_PROJECT` or `--project`, not stored in the config file.

### Environment variables

| Variable | Description |
|----------|-------------|
| `GCL_INSTANCE` | GitLab instance URL |
| `GCL_TOKEN` | Personal access token (scope `api`) |
| `GCL_PROJECT` | Project for API validation (ID or path) |
| `GCL_OUTPUT_FORMAT` | Output format: text, json, yaml |
| `GCL_SKIP_API` | true = local validation only |
| `GCL_VERBOSE` | true = verbose output |
| `GCL_COLOR` | auto, always, never |
| `GCL_DEBUG` | true = debug output |
| `GCL_CONFIG` | Path to config file |

### GitLab token

For API validation you need a personal access token with `api` scope:

1. GitLab → User Settings → Access Tokens
2. Create a token with `api` scope
3. Provide via: `GCL_TOKEN`, `gitlab-ci-lint setup`, config `auth.token`, or `--token`

## Usage

```bash
gitlab-ci-lint [flags] [file]
```

**Commands:** `setup` (interactive config), `version`, `help`.

**Main flags:**

- `-c, --config` — config file path
- `-t, --token` — GitLab token
- `--netrc` — use `~/.netrc`
- `--instance` — GitLab URL (default https://gitlab.com)
- `--project` — project for API validation (ID or group/project)
- `-s, --skip-api` — local validation only
- `-o, --output` — format: text | json | yaml
- `-v, --verbose` — verbose output
- `--color` — auto | always | never
- `-f, --file` — path(s) to file(s), repeatable
- `-d, --directory` — directory to scan recursively for CI files
- `--list-instances` — list configured instances and exit

**File discovery order** when no file argument is given: `-f` flags → `-d` flags → single positional argument → auto-discover in current and parent directories.

### Examples

```bash
# Auto-discover .gitlab-ci.yml in current and parent dirs
gitlab-ci-lint

# Multiple files
gitlab-ci-lint -f .gitlab-ci.yml -f ci/frontend.yml

# Recursive directory scan
gitlab-ci-lint -d ./monorepo

# Syntax only
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Custom instance
gitlab-ci-lint --instance https://gitlab.example.com .gitlab-ci.yml

# JSON for pipelines
gitlab-ci-lint --output json .gitlab-ci.yml | jq .valid

# Project-specific (extends, trigger, etc.)
gitlab-ci-lint --project mygroup/myproject .gitlab-ci.yml

# From stdin
cat .gitlab-ci.yml | gitlab-ci-lint -
```

## Exit codes

- `0` — all validations passed
- `1` — runtime error (file not found, network, auth)
- `10` — CI configuration invalid

For batch validation, exit code is `10` if any file is invalid.

## Documentation

- [Wiki](https://github.com/InkyQuill/gitlab-ci-lint/wiki)
- [Contributing](CONTRIBUTING.md)

## Development

```bash
make build             # build to ./build/gitlab-ci-lint
make test-unit         # unit tests
make test-integration  # integration tests (requires built binary)
make lint
make fmt
```

## License

MIT — see [LICENSE](LICENSE).

## Project status

- Two-stage validation (local + API)
- Interactive setup and multi-instance config
- Text/JSON/YAML output
- Unit and integration tests
- Semantic-release and goreleaser

Plans: [ROADMAP.md](ROADMAP.md).
