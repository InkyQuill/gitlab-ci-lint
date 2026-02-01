# GitLab CI Lint - Project Specification

## Purpose & Scope

GitLab CI Lint is a command-line tool that validates `.gitlab-ci.yml` configuration files. It performs comprehensive validation through a two-stage architecture:

1. **Local Validation**: Fast YAML syntax and structure checking without requiring GitLab API access
2. **API Validation**: Full GitLab instance validation including job verification and semantic warnings

The tool is designed for:
- Developers who want fast feedback on CI configuration syntax
- CI/CD pipelines that need to validate configurations before deployment
- Teams managing multiple GitLab instances (gitlab.com, self-hosted)
- Automation workflows requiring structured validation output (JSON/YAML)

## Validation Stages

### Stage 1: Local Validation

**Purpose**: Provide immediate feedback on YAML syntax and basic structure errors.

**Implementation**: `internal/validator/local.go`

**Capabilities**:
- YAML syntax validation using `gopkg.in/yaml.v3`
- Error detection with line and column numbers
- Error context preservation (showing the problematic line)
- Configurable strict/lenient parsing modes

**Output**:
- Returns `validator.Result` with `Valid` boolean and `Errors` array
- Each error includes:
  - `Message`: Human-readable error description
  - `Line`: Line number where error occurred
  - `Column`: Column position within line
  - `Content`: The actual line content for context

**Exit Behavior**: If local validation fails, the tool exits immediately with code `10` (validation error) without proceeding to API validation. This fails fast and saves time.

### Stage 2: API Validation

**Purpose**: Validate configuration against a real GitLab instance for semantic correctness.

**Implementation**: `internal/validator/api.go`

**Capabilities**:
- Full GitLab CI/CD lint API integration
- Job definition verification
- Job dependency validation
- Script syntax checking
- Variable reference validation
- Warning reporting for non-critical issues

**Endpoints**:
- Global: `/api/v4/ci/lint` - Validates against instance-wide CI configuration schema
- Project-specific: `/api/v4/projects/:id/ci/lint` - Validates with project context (includes job references, variables, etc.)

**Output**:
- Returns `validator.Result` with `Valid` boolean, `Warnings` array, and potential `Errors`
- Warnings include:
  - `Message`: Warning description
  - `Line`: Related line number (if applicable)

**Authentication**: Requires personal access token with `read_api` scope

**Skip Option**: Can be disabled via `--skip-api` flag or `GCL_SKIP_API` environment variable for offline operation

## Configuration System

The configuration system uses a priority-based loading mechanism with four sources (low to high priority):

### 1. Defaults
**Location**: `internal/config/defaults.go`

```yaml
gitlab:
  instance: "https://gitlab.com"
  timeout: 30s
auth:
  token: ""
  netrc: false
validation:
  skip_api: false
  strict: true
  project: ""
output:
  format: "text"
  verbose: false
  color: "auto"
```

### 2. Configuration File
**Default Location**: `~/.tools-config/.gitlab-ci-lint/config.yaml`

**Custom Location**: Specified via:
- `--config` / `-c` flag
- `GCL_CONFIG` environment variable

**Example**:
```yaml
gitlab:
  instance: "https://gitlab.example.com"
  timeout: 60s
auth:
  token: "glpat-xxxxxxxxxxxx"
  netrc: false
validation:
  skip_api: false
  strict: true
  project: "group/project"
output:
  format: "json"
  verbose: true
  color: "never"
```

### 3. Environment Variables
**Prefix**: `GCL_`

| Variable | Config Path | Example |
|----------|-------------|---------|
| `GCL_CONFIG` | Config file path | `~/.custom/config.yaml` |
| `GCL_TOKEN` | `auth.token` | `glpat-xxxxxxxxxxxx` |
| `GCL_INSTANCE` | `gitlab.instance` | `https://gitlab.com` |
| `GCL_TIMEOUT` | `gitlab.timeout` | `30s` |
| `GCL_PROJECT` | `validation.project` | `123` or `group/project` |
| `GCL_SKIP_API` | `validation.skip_api` | `true` |
| `GCL_STRICT` | `validation.strict` | `false` |
| `GCL_OUTPUT_FORMAT` | `output.format` | `json` |
| `GCL_VERBOSE` | `output.verbose` | `true` |
| `GCL_COLOR` | `output.color` | `never` |

### 4. CLI Flags
**All flags override previous sources**

| Flag | Short | Config Path | Type |
|------|-------|-------------|------|
| `--config` | `-c` | Config file path | string |
| `--token` | `-t` | `auth.token` | string |
| `--netrc` | | `auth.netrc` | bool |
| `--instance` | | `gitlab.instance` | string |
| `--timeout` | | `gitlab.timeout` | duration |
| `--project` | | `validation.project` | string |
| `--skip-api` | `-s` | `validation.skip_api` | bool |
| `--strict` | | `validation.strict` | bool |
| `--output` | `-o` | `output.format` | string |
| `--verbose` | `-v` | `output.verbose` | bool |
| `--color` | | `output.color` | string |

**Implementation**: `internal/config/loader.go`

## Exit Code Semantics

The tool uses specific exit codes to indicate different outcomes:

| Code | Constant | Description |
|------|----------|-------------|
| 0 | `exit.ExitSuccess` | All validations passed |
| 1 | `exit.ExitGeneralError` | Runtime error (file not found, authentication failed, network error, etc.) |
| 10 | `exit.ExitValidationError` | CI configuration is invalid (YAML syntax errors, semantic errors, etc.) |

**Implementation**: `internal/exit/codes.go`

**Usage in CI/CD**:
```bash
#!/bin/bash
gitlab-ci-lint .gitlab-ci.yml
case $? in
  0) echo "Valid configuration" ;;
  1) echo "Runtime error - check environment" ;;
  10) echo "Invalid CI configuration" ;;
esac
```

## Supported GitLab Instances

The tool supports any GitLab instance that implements the CI lint API:

### GitLab.com
**Default**: `https://gitlab.com`

**No configuration required** if using defaults

### Self-Hosted Instances
**Supported versions**: GitLab 11.0+ (CI Lint API introduced)

**Configuration**:
```bash
# Via flag
gitlab-ci-lint --instance https://gitlab.example.com .gitlab-ci.yml

# Via config file
echo 'gitlab:
  instance: "https://gitlab.example.com"' > ~/.tools-config/.gitlab-ci-lint/config.yaml

# Via environment
export GCL_INSTANCE=https://gitlab.example.com
```

### URL Normalization
**Implementation**: `internal/gitlab/client.go` - `NormalizeInstanceURL()`

**Rules**:
1. Adds `https://` scheme if missing
2. Removes trailing slashes
3. Trims whitespace

**Examples**:
| Input | Normalized |
|-------|------------|
| `gitlab.com` | `https://gitlab.com` |
| `https://gitlab.example.com/` | `https://gitlab.example.com` |
| `http://localhost:8080` | `http://localhost:8080` |

## Authentication Methods

### Personal Access Token (Recommended)
**Token Scope**: `read_api` (minimum required)

**Creation**: User Settings → Access Tokens → Create personal access token

**Configuration Methods**:

1. **CLI Flag**:
```bash
gitlab-ci-lint --token glpat-xxxxxxxxxxxx .gitlab-ci.yml
```

2. **Config File**:
```yaml
auth:
  token: "glpat-xxxxxxxxxxxx"
```

3. **Environment Variable**:
```bash
export GCL_TOKEN=glpat-xxxxxxxxxxxx
```

**Security**: Config file should have restricted permissions (0600)

### .netrc Support (Planned)
**Status**: Not yet implemented

**Planned Implementation**:
- Read credentials from `~/.netrc`
- Match machine to GitLab instance hostname
- Extract token as password

**Current**: `internal/gitlab/client.go` - `ExtractTokenFromNetrc()` returns "not yet implemented" error

### Job Token (CI/CD Context)
**Environment Variable**: `CI_JOB_TOKEN`

**Use Case**: Validating CI configurations within GitLab CI/CD pipelines

**Implementation**: Client sends both `JOB-TOKEN` and `PRIVATE-TOKEN` headers

## Output Formats

### Text Format (Default)
**Implementation**: `internal/output/formatter.go`

**Features**:
- Human-readable output
- Color coding (auto/always/never)
- Verbose mode for detailed information
- Error formatting with line numbers and context

**Example**:
```
✓ Local validation passed
✓ API validation passed

Configuration .gitlab-ci.yml is valid
```

**Error Example**:
```
✗ Local validation failed

Error: invalid YAML syntax
  Line: 10
  Column: 5
  Content:   script: echo "Hello

  Missing closing quote
```

### JSON Format
**Structure**:
```json
{
  "file": ".gitlab-ci.yml",
  "results": [
    {
      "stage": "local",
      "valid": true,
      "errors": [],
      "warnings": []
    },
    {
      "stage": "api",
      "valid": true,
      "errors": [],
      "warnings": [
        {
          "message": "Job 'test' uses deprecated image",
          "line": 15
        }
      ]
    }
  ]
}
```

**Use Cases**:
- Programmatic parsing
- Integration with other tools
- CI/CD pipeline logging

### YAML Format
**Structure**: Same as JSON, but in YAML format

**Example**:
```yaml
file: .gitlab-ci.yml
results:
  - stage: local
    valid: true
    errors: []
    warnings: []
  - stage: api
    valid: true
    errors: []
    warnings: []
```

### Color Output
**Modes**:
- `auto` (default): Detects TTY, enables colors for terminals
- `always`: Always use ANSI color codes
- `never`: Disable all colors

**Configuration**:
```bash
--color auto    # Default
--color always  # Force colors
--color never   # Disable colors
```

**Implementation**: `internal/output/colors.go`

## Architecture

### Component Overview

```
cmd/gitlab-ci-lint/main.go
├── Configuration Loading
│   └── internal/config/
│       ├── config.go       (Config structs)
│       ├── defaults.go     (Default values)
│       └── loader.go       (Multi-source loading)
│
├── Validation Pipeline
│   └── internal/validator/
│       ├── validator.go    (Interface definition)
│       ├── local.go        (YAML validation)
│       └── api.go          (GitLab API validation)
│
├── GitLab Client
│   └── internal/gitlab/
│       ├── client.go       (HTTP client, endpoints)
│       └── models.go       (API request/response types)
│
└── Output Formatting
    └── internal/output/
        ├── formatter.go    (Multi-format output)
        └── colors.go       (Terminal colors)
```

### Data Flow

```
[CLI Args/Config/Env] → Config Loader → Config
                                   ↓
[YAML File Content] → Local Validator → Result (stage: local)
                                   ↓
                            (if valid and API enabled)
                                   ↓
                    GitLab Client → API Validator → Result (stage: api)
                                   ↓
                            Formatter → Output (text/json/yaml)
                                   ↓
                            Exit Code (0/1/10)
```

## Design Patterns

### Interface-Based Validation
**Pattern**: `Validator` interface allows extensible validation types

**Definition**: `internal/validator/validator.go`
```go
type Validator interface {
    Validate(content []byte) Result
}
```

**Benefits**:
- Easy to add new validation types (e.g., security scanner, custom rules)
- Testable with mock implementations
- Clean separation of concerns

### Graceful Degradation
**Pattern**: Tool functions with reduced capabilities if API is unavailable

**Implementation**:
- `--skip-api` flag enables offline operation
- Local validation always works (no network required)
- API errors are clearly reported but don't crash

### Structured Errors
**Pattern**: All errors include context (line numbers, content)

**Benefits**:
- Easy to locate issues in large CI files
- IDE-friendly error parsing
- Better user experience

## Future Extensions

### Planned Features
- Web interface for interactive validation
- VS Code extension integration
- Pre-commit hook integration
- Docker image for containerized usage
- Custom validation rules engine
- Configuration diffing (before/after comparison)
- Batch file validation

### Extensibility Points
- Custom `Validator` implementations
- Additional output formats (HTML, Markdown)
- Plugin system for validation rules
- Custom authentication providers

## Version Information

**Implementation**: `pkg/version/version.go`

**Build-time Injection** (via Makefile):
- `Version`: Git tag or `dev`
- `Commit`: Git commit hash
- `BuildDate`: UTC timestamp

**Access**:
```bash
$ gitlab-ci-lint version
gitlab-ci-lint version v1.2.3
Commit: abc1234
Build date: 2025-01-15T10:30:00Z
```

## Dependencies

### Go Dependencies
```go
require (
    github.com/spf13/cobra v1.8.0    // CLI framework
    gopkg.in/yaml.v3 v3.0.1          // YAML parsing
)
```

### Build Tools
- Go 1.24 or later
- Make (build automation)
- golangci-lint (linting, optional)
- goimports (import formatting, optional)

## Platform Support

**Supported Platforms**:
- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64, 386)

**Build Output**: Single static binary, no external runtime dependencies

## References

- **GitLab CI Lint API**: https://docs.gitlab.com/ee/api/rest/lint.md
- **GitLab CI/CD Syntax**: https://docs.gitlab.com/ee/ci/yaml/
- **Personal Access Tokens**: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html
- **Project Documentation**: See `docs/` directory
