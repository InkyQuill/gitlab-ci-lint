# API Reference

This document provides reference documentation for the gitlab-ci-lint API packages.

## Packages

### cmd/gitlab-ci-lint

Main CLI application entry point.

#### Type: Command

`rootCmd` - Main Cobra command for linting

**Usage**:
```bash
gitlab-ci-lint [flags] <file>
```

#### Functions

- `runLint()` - Main linting logic

### cmd/setup

Interactive setup wizard for configuration.

#### Type: Command

`rootCmd` - Setup Cobra command

**Usage**:
```bash
gitlab-ci-lint setup
```

---

## internal/config

Configuration management with multi-source loading.

### Types

#### Config

```go
type Config struct {
    GitLab     GitLabConfig
    Auth       AuthConfig
    Validation ValidationConfig
    Output     OutputConfig
}
```

**Configuration** for the application.

#### GitLabConfig

```go
type GitLabConfig struct {
    Instance string        // GitLab instance URL
    Timeout  time.Duration // API timeout
}
```

#### AuthConfig

```go
type AuthConfig struct {
    Token string // Personal access token
    Netrc bool   // Use .netrc for credentials
}
```

#### ValidationConfig

```go
type ValidationConfig struct {
    SkipAPI bool   // Skip API validation
    Strict  bool   // Use strict YAML parsing
    Project string // Default project (ID or path)
}
```

#### OutputConfig

```go
type OutputConfig struct {
    Format  string // Output format: text|json|yaml
    Verbose bool   // Verbose output
    Color   string // Color mode: auto|always|never
}
```

### Functions

#### GetDefaults

```go
func GetDefaults() Config
```

Returns default configuration values.

#### NewLoader

```go
func NewLoader(flags *ConfigFlags) *Loader
```

Creates a new configuration loader.

#### (*Loader) Load

```go
func (l *Loader) Load() (*Config, error)
```

Loads configuration from all sources with proper priority.

#### (*Config) WriteConfig

```go
func WriteConfig(path string, cfg *Config) error
```

Writes configuration to file with atomic operation and secure permissions.

---

## internal/gitlab

GitLab API client for CI linting.

### Types

#### Client

```go
type Client struct {
    instance   string
    token      string
    httpClient *http.Client
}
```

GitLab API client.

#### LintRequest

```go
type LintRequest struct {
    Content     string `json:"content"`
    DryRun      bool   `json:"dry_run,omitempty"`
    IncludeJobs bool   `json:"include_jobs,omitempty"`
}
```

Request body for CI lint API.

#### LintResponse

```go
type LintResponse struct {
    Valid      bool       `json:"valid"`
    Errors     []APIError `json:"errors,omitempty"`
    Warnings   []string   `json:"warnings,omitempty"`
    MergedYaml string     `json:"merged_yaml,omitempty"`
}
```

Response from CI lint API.

#### APIError

```go
type APIError struct {
    Message string `json:"message"`
}
```

Error from GitLab API.

### Functions

#### NewClient

```go
func NewClient(instance, token string, timeout time.Duration) *Client
```

Creates a new GitLab API client.

#### (*Client) Lint

```go
func (c *Client) Lint(ctx context.Context, content []byte, projectRef string) (*LintResponse, error)
```

Validates a CI configuration using the GitLab API.

**Parameters**:
- `ctx` - Context for cancellation
- `content` - CI configuration file content
- `projectRef` - Project reference (ID or path), empty for global validation

**Returns**: `LintResponse` or error

#### (*Client) ValidateToken

```go
func (c *Client) ValidateToken(ctx context.Context) error
```

Validates the current token by calling `/api/v4/user`.

#### (*Client) ExtractTokenFromNetrc

```go
func ExtractTokenFromNetrc(instance string) (string, error)
```

Attempts to find credentials in .netrc file.

**Status**: Not yet implemented

#### NormalizeInstanceURL

```go
func NormalizeInstanceURL(instance string) string
```

Normalizes a GitLab instance URL:
- Adds `https://` scheme if missing
- Removes trailing slashes
- Trims whitespace

---

## internal/validator

Validation logic for CI configurations.

### Types

#### Validator

```go
type Validator interface {
    Validate(content []byte) Result
}
```

Interface for validation operations.

#### Result

```go
type Result struct {
    Valid    bool
    Errors   []Error
    Warnings []Warning
    Stage    string // "local" or "api"
}
```

Result of a validation operation.

#### Error

```go
type Error struct {
    Message string
    Line    int
    Column  int
    Content string // Error line content
}
```

Validation error with context.

#### Warning

```go
type Warning struct {
    Message string
    Line    int
}
```

Validation warning.

### Implementations

#### LocalValidator

```go
type LocalValidator struct {
    strict bool
}
```

Validates YAML syntax locally.

**Methods**:
- `NewLocalValidator(strict bool) *LocalValidator`
- `Validate(content []byte) Result`

#### APIValidator

```go
type APIValidator struct {
    client    *gitlab.Client
    project   string
}
```

Validates using GitLab API.

**Methods**:
- `NewAPIValidator(client *gitlab.Client, project string) *APIValidator`
- `Validate(content []byte) Result`

---

## internal/output

Output formatting for validation results.

### Types

#### Formatter

```go
type Formatter struct {
    colorizer *Colorizer
    verbose   bool
}
```

Handles output formatting.

#### Colorizer

```go
type Colorizer struct {
    mode string // "auto", "always", "never"
}
```

Terminal color handling.

### Functions

#### NewFormatter

```go
func NewFormatter(colorSetting string, verbose bool) *Formatter
```

Creates a new output formatter.

#### (*Formatter) FormatResult

```go
func (f *Formatter) FormatResult(w io.Writer, format string, results []validator.Result, filename string)
```

Formats and writes validation results.

**Parameters**:
- `w` - Writer for output
- `format` - Output format: "text", "json", "yaml"
- `results` - Validation results
- `filename` - File being validated

#### (*Formatter) FormatMessage

```go
func (f *Formatter) FormatMessage(w io.Writer, level, message string)
```

Formats a generic message.

**Parameters**:
- `level` - Message level: "error", "warning", "info"
- `message` - Message content

#### NewColorizer

```go
func NewColorizer(mode string) *Colorizer
```

Creates a new colorizer.

#### (*Colorizer) Color

```go
func (c *Colorizer) Color(color, text string) string
```

Applies color to text if enabled.

**Colors**: "red", "green", "yellow", "blue", "gray"

---

## internal/setup

Setup validation for the configuration wizard.

### Functions

#### ValidateToken

```go
func ValidateToken(instance, token string) (*TokenValidationResult, error)
```

Validates a GitLab personal access token.

**Returns**: `TokenValidationResult` with validity status and username (if valid)

#### DetectInstanceType

```go
func DetectInstanceType(instance string) (*InstanceInfo, error)
```

Detects the GitLab instance type and version.

**Returns**: `InstanceInfo` with version and instance name

#### TestConnection

```go
func TestConnection(ctx context.Context, instance string) error
```

Tests connectivity to a GitLab instance.

### Types

#### TokenValidationResult

```go
type TokenValidationResult struct {
    Valid    bool
    Username string
    Error    string
}
```

Result of token validation.

#### InstanceInfo

```go
type InstanceInfo struct {
    Version      string
    InstanceName string
    URL          string
}
```

Information about a GitLab instance.

---

## internal/exit

Exit code constants.

### Constants

```go
const (
    ExitSuccess       = 0
    ExitGeneralError  = 1
    ExitValidationError = 10
)
```

**ExitSuccess**: All validations passed

**ExitGeneralError**: Runtime error (file not found, auth failed, etc.)

**ExitValidationError**: CI configuration invalid

---

## pkg/version

Version information injected at build time.

### Variables

```go
var (
    Version    string // Version string (e.g., "v1.2.3")
    Commit     string // Git commit hash
    BuildDate  string // Build timestamp (UTC)
)
```

These are set via `-ldflags` during build:

```bash
go build \
  -ldflags "-X github.com/InkyQuill/gitlab-ci-lint/pkg/version.Version=v1.2.3" \
  -o gitlab-ci-lint
```

---

## Usage Examples

### Validate Configuration

```go
package main

import (
    "github.com/InkyQuill/gitlab-ci-lint/internal/config"
    "github.com/InkyQuill/gitlab-ci-lint/internal/validator"
)

func main() {
    // Load configuration
    loader := config.NewLoader(&config.ConfigFlags{})
    cfg, _ := loader.Load()

    // Create validator
    localValidator := validator.NewLocalValidator(cfg.Validation.Strict)

    // Read CI file
    content := []byte("image: alpine\nbuild:\n  script: echo test")

    // Validate
    result := localValidator.Validate(content)

    if !result.Valid {
        // Handle errors
        for _, err := range result.Errors {
            println(err.Message)
        }
    }
}
```

### Use GitLab Client

```go
package main

import (
    "context"
    "github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
)

func main() {
    client := gitlab.NewClient(
        "https://gitlab.com",
        "glpat-token",
        30*time.Second,
    )

    // Validate token
    err := client.ValidateToken(context.Background())
    if err != nil {
        panic(err)
    }

    // Lint CI configuration
    content := []byte("image: alpine")
    resp, err := client.Lint(context.Background(), content, "")
    if err != nil {
        panic(err)
    }

    println("Valid:", resp.Valid)
}
```

### Format Output

```go
package main

import (
    "os"
    "github.com/InkyQuill/gitlab-ci-lint/internal/output"
    "github.com/InkyQuill/gitlab-ci-lint/internal/validator"
)

func main() {
    formatter := output.NewFormatter("auto", true)

    results := []validator.Result{
        {Stage: "local", Valid: true},
    }

    formatter.FormatResult(os.Stdout, "text", results, ".gitlab-ci.yml")
}
```

For more examples, see the [examples directory](../examples/).
