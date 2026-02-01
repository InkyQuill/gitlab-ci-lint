# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitLab CI Lint is a Go CLI tool that validates `.gitlab-ci.yml` files in two stages:
1. **Local validation**: Fast YAML syntax checking (no API required)
2. **API validation**: GitLab instance validation via GitLab API (optional)

## Development Commands

### Build
```bash
make build              # Build for current platform (output: ./build/gitlab-ci-lint)
make build-all          # Build for all platforms (Linux, macOS, Windows)
make install            # Install to GOPATH/bin
```

### Test
```bash
make test               # Run tests with race detector
make test-coverage      # Run tests and generate coverage.html
```

### Code Quality
```bash
make lint               # Run golangci-lint
make fmt                # Format with gofmt and goimports
make tidy               # Tidy go.mod dependencies
```

### Running
```bash
make run ARGS='--help'  # Build and run with arguments
./build/gitlab-ci-lint --skip-api .gitlab-ci.yml  # Direct execution
```

## Architecture

The codebase follows a clean interface-based design with clear separation of concerns:

### Core Components

- **`cmd/gitlab-ci-lint/main.go`**: Entry point using Cobra CLI. Handles flag parsing, config loading, and orchestrates the two-stage validation flow (local → API).

- **`internal/config/`**: Configuration management with priority-based loading:
  - Priority: Defaults → Config file → Environment variables (`GCL_*` prefix) → CLI flags
  - `config.go`: Config structs, `loader.go`: Multi-source loading, `defaults.go`: Default values

- **`internal/validator/`**: Validation logic via `Validator` interface:
  - `validator.go`: Defines `Validator` interface and `Result`/`Error`/`Warning` types
  - `local.go`: YAML syntax validation using `gopkg.in/yaml.v3`
  - `api.go`: GitLab API validation integration

- **`internal/gitlab/`**: GitLab API client:
  - `client.go`: HTTP client with configurable timeout, handles endpoints
  - `models.go`: API response models
  - Endpoints: `/api/v4/ci/lint` (global) and `/api/v4/projects/:id/ci/lint` (project-specific)

- **`internal/output/`**: Output formatting:
  - `formatter.go`: Multi-format output (text, JSON, YAML)
  - `colors.go`: Terminal color handling (auto/always/never)

- **`internal/exit/codes.go`**: Exit codes (0=success, 1=general error, 10=validation error)

### Configuration File

Default location: `~/.tool-configs/.gitlab-ci-lint/config.yaml`

### Key Design Patterns

1. **Interface-based validation**: `Validator` interface allows adding new validation types
2. **Graceful degradation**: Can run with local validation only (`--skip-api`)
3. **Instance normalization**: Handles various GitLab URL formats via `gitlab.NormalizeInstanceURL()`
4. **Structured errors**: Line numbers, column positions, and context from YAML parsing

## Exit Codes

- `0`: All validations passed
- `1`: Runtime error (file not found, auth failed, etc.)
- `10`: CI configuration invalid

## Adding Features

### New validation type
Implement `Validator` interface in `internal/validator/` and wire up in `main.go`

### New output format
Extend `Formatter` in `internal/output/formatter.go`, update config schema

### New CLI flag
Add to `ConfigFlags` struct, bind in Cobra command setup in `main.go`
