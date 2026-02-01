# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitLab CI Lint is a Go CLI tool that validates `.gitlab-ci.yml` files in two stages:
1. **Local validation**: Fast YAML syntax checking (no API required)
2. **API validation**: GitLab instance validation via GitLab API (optional)

Key features: smart file discovery (auto-discovery, recursive scanning), multi-format output (text/JSON/YAML), interactive setup wizard, single binary distribution.

## Development Commands

### Build
```bash
make build              # Build for current platform (output: ./build/gitlab-ci-lint)
make build-all          # Build for all platforms (Linux, macOS, Windows)
make install            # Install to GOPATH/bin or ./build/
```

### Test
```bash
make test-unit          # Run unit tests for internal/ and pkg/
make test-integration   # Run integration tests (requires built binary)
make test-all           # Run both with combined coverage report
make test-coverage      # Alias for test-all

# Run single test
go test -v ./internal/validator -run TestLocalValidator_ValidYAML

# Run specific package tests
go test -v ./internal/config
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

# Test setup command
./build/gitlab-ci-lint setup --help

# Validate with auto-discovery
./build/gitlab-ci-lint

# Environment variables for testing
export GCL_INSTANCE=https://gitlab.com
export GCL_TOKEN=glpat-test
export GCL_SKIP_API=false
```

## Architecture

The codebase follows a clean interface-based design with clear separation of concerns:

### Entry Point & CLI

- **`cmd/gitlab-ci-lint/main.go`**: Cobra CLI entry point with integrated setup command
  - Flag parsing via `ConfigFlags` struct
  - Config loading via `config.NewLoader()`
  - Orchestrates two-stage validation (local → API)
  - File discovery orchestration
  - Contains `runSetup()` for interactive configuration wizard

### Core Packages

- **`internal/config/`**: Multi-source configuration loading
  - Priority: Defaults → Config file (`~/.tools-config/.gitlab-ci-lint/config.yaml`) → Environment vars (`GCL_*`) → CLI flags
  - `loader.go`: Merges all config sources
  - `defaults.go`: Built-in defaults
  - `writer.go`: Atomic config writes with 0600 permissions

- **`internal/validator/`**: Validation pipeline via `Validator` interface
  - `validator.go`: Interface definition, `Result`/`Error`/`Warning` types
  - `local.go`: YAML syntax validation using `gopkg.in/yaml.v3`
  - `api.go`: GitLab API validation (requires token)
  - Returns structured errors with line/column numbers from YAML parsing

- **`internal/gitlab/`**: GitLab API client
  - `client.go`: HTTP client with timeout, token validation
  - `models.go`: API response types
  - `NormalizeInstanceURL()`: Handles various URL formats (with/without protocol, trailing slash)
  - Endpoints: `/api/v4/ci/lint` (global) and `/api/v4/projects/:id/ci/lint` (project-specific)
  - URL encoding for project paths with special characters

- **`internal/discover/`**: Smart file discovery system
  - `discoverer.go`: `Discoverer` interface with multiple strategies
  - Strategies: auto-discovery (current + parent dirs), recursive directory scanning, explicit files
  - `FindInCurrentAndParents()`: Searches upward for `.gitlab-ci.yml`
  - `FindInDirectoryTree()`: Recursive scan with max depth and ignore patterns
  - Built-in ignores: `.git`, `node_modules`, `vendor`, `build`, `dist`, `*.tar.gz`

- **`internal/output/`**: Multi-format output
  - `formatter.go`: Text/JSON/YAML formatters
  - `colors.go`: Terminal color handling (auto/always/never)
  - `FormatResult()`: Outputs validation results with source context

- **`internal/setup/`**: Setup validation utilities
  - `validator.go`: Token validation, instance type detection, connection testing
  - Used by both setup command and can be used independently

- **`internal/exit/codes.go`**: Exit code constants (0=success, 1=error, 10=validation failed)

### Configuration Priority

1. Built-in defaults
2. Config file: `~/.tools-config/.gitlab-ci-lint/config.yaml`
3. Environment variables: `GCL_INSTANCE`, `GCL_TOKEN`, `GCL_PROJECT`, `GCL_OUTPUT`, `GCL_SKIP_API`, `GCL_VERBOSE`, `GCL_COLOR`
4. CLI flags (highest priority)

### Key Design Patterns

1. **Interface-based validation**: `Validator` interface allows adding new validation types
2. **Graceful degradation**: Local validation works without API (`--skip-api` or no token)
3. **Instance normalization**: `gitlab.NormalizeInstanceURL()` standardizes URLs
4. **Structured errors**: YAML parser provides line/column for precise error reporting
5. **File discovery abstraction**: `Discoverer` interface allows multiple discovery strategies
6. **Batch validation**: Process multiple files with summary output

### File Discovery Priority

When no file argument provided:
1. `-f` flags: Explicit file paths
2. `-d` flags: Directory paths (recursive)
3. Positional arg: Single file path
4. Auto-discovery: Search current + parent directories

## Exit Codes

- `0`: All validations passed
- `1`: Runtime error (file not found, auth failed, network error)
- `10`: CI configuration invalid (any file failed validation)

For batch validation, exits with 10 if **any** file is invalid.

## GitLab API Integration

### Token Requirements
- Scope: `api` (full API access)
- Validation: Token validated on first use via `/api/v4/user` endpoint
- Sources: Config file, `GCL_TOKEN` env var, `--token` flag, or `.netrc` (if `--netrc` flag)

### Project-Specific Validation
- Use `--project group/project` or `GCL_PROJECT` env var
- Required for validating job references (`extends`, `trigger`, etc.)
- Project ID or path supported (URL-encoded if path contains special chars)

### Instance URL Normalization
```go
// All these are normalized to https://gitlab.com/
gitlab.com
https://gitlab.com
https://gitlab.com/
gitlab.com/
```

## Testing Strategy

### Unit Tests
- Located in each package: `internal/*/*_test.go`
- Use table-driven tests for multiple scenarios
- Mock HTTP responses for GitLab API tests
- Target: 70%+ coverage (currently 64.8% overall)

### Integration Tests
- Located in `tests/integration/`
- Require built binary: `make build && make test-integration`
- Test end-to-end CLI behavior
- Use `CI=true` env var to skip interactive tests
- Test file discovery, validation flows, exit codes

### Test Files
- `validation_test.go`: Core validation integration tests
- `setup_test.go`: Setup command tests (skipped in CI)
- `file_discovery_test.go`: Discovery system tests

## Adding Features

### New validation type
Implement `Validator` interface in `internal/validator/`:
```go
type Validator interface {
    Validate(content []byte) Result
}
```
Wire up in validation pipeline in `main.go`.

### New output format
Extend `Formatter` in `internal/output/formatter.go`, add format constant, update config schema.

### New CLI flag
1. Add field to `ConfigFlags` struct in `cmd/gitlab-ci-lint/main.go`
2. Bind with Cobra: `rootCmd.Flags().StringVar(&flags.YourFlag, "your-flag", ...)`
3. Add to config schema in `internal/config/config.go`
4. Update `defaults.go` for default value

### New file discovery strategy
Implement `Discoverer` interface in `internal/discover/`, add method to `Discoverer` struct.

## Important Notes

- **Single binary**: Setup command is integrated into main binary (no separate `gitlab-ci-lint-setup`)
- **Go version**: Requires Go 1.24 or later
- **Config directory**: Changed from `~/.tool-configs/` to `~/.tools-config/` (note the 's')
- **Semantic-release**: Automated versioning based on conventional commits (`feat:`, `fix:`, etc.)
- **Pre-commit hooks**: Configured to run fmt, lint, test-unit before commits
- **AI assistance**: Development uses Claude Code (Anthropic) and Z.ai (GLM 4.7) - disclose in PRs
- **License**: MIT - all contributions licensed under MIT
