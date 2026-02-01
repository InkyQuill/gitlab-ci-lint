# System Architecture

This document provides an overview of the GitLab CI Lint architecture.

## High-Level Architecture

GitLab CI Lint follows a modular, layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│                      (cmd/gitlab-ci-lint/)                   │
│  - Command parsing (Cobra)                                   │
│  - Flag handling                                            │
│  - Exit code management                                     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Configuration Layer                       │
│                     (internal/config/)                       │
│  - Multi-source loading (defaults, file, env, flags)        │
│  - Priority-based merging                                   │
│  - Validation                                               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Validation Layer                          │
│                   (internal/validator/)                      │
│  - Local YAML validation (validator/local.go)              │
│  - API validation (validator/api.go)                        │
│  - Result aggregation                                       │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Client Layer                             │
│                     (internal/gitlab/)                       │
│  - HTTP client                                              │
│  - GitLab API integration                                   │
│  - Response parsing                                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Output Layer                              │
│                    (internal/output/)                        │
│  - Text formatting                                          │
│  - JSON serialization                                        │
│  - YAML serialization                                        │
│  - Color handling                                           │
└─────────────────────────────────────────────────────────────┘
```

## Components

### CLI Layer (`cmd/gitlab-ci-lint/main.go`)

**Responsibility**: User interaction and orchestration

**Key Functions**:
- Parse command-line arguments using Cobra
- Load configuration from all sources
- Coordinate validation pipeline
- Handle exit codes

**Design Pattern**: Command Pattern (Cobra)

### Configuration Layer (`internal/config/`)

**Responsibility**: Configuration management and loading

**Components**:

1. **config.go**: Configuration data structures
   - `Config`: Root configuration
   - `GitLabConfig`: Instance settings
   - `AuthConfig`: Authentication settings
   - `ValidationConfig`: Validation behavior
   - `OutputConfig`: Output formatting

2. **loader.go**: Multi-source configuration loading
   - Implements priority chain: defaults → file → env → flags
   - Merges configurations with proper precedence
   - Validates final configuration

3. **defaults.go**: Default values
   - Provides sensible defaults for all settings
   - Allows operation without configuration

4. **writer.go**: Configuration persistence
   - Atomic write operations
   - Secure file permissions (0600)
   - Directory creation

**Design Pattern**: Builder Pattern (for configuration loading)

### Validation Layer (`internal/validator/`)

**Responsibility**: CI configuration validation

**Interface**:
```go
type Validator interface {
    Validate(content []byte) Result
}
```

**Implementations**:

1. **Local Validator** (`local.go`):
   - YAML syntax validation using `gopkg.in/yaml.v3`
   - Line and column error reporting
   - Error context preservation
   - Configurable strict/lenient mode

2. **API Validator** (`api.go`):
   - GitLab API integration
   - Job definition verification
   - Semantic validation
   - Warning reporting

**Design Pattern**: Strategy Pattern (Validator interface)

### Client Layer (`internal/gitlab/`)

**Responsibility**: GitLab API communication

**Components**:

1. **client.go**: HTTP client
   - Configurable timeout
   - Token authentication
   - URL normalization
   - Error handling

2. **models.go**: API data structures
   - Request/response types
   - Error models
   - Warning models

**Endpoints**:
- Global: `/api/v4/ci/lint`
- Project-specific: `/api/v4/projects/:id/ci/lint`

**Design Pattern**: Facade Pattern (simplifies GitLab API)

### Output Layer (`internal/output/`)

**Responsibility**: Result formatting and presentation

**Components**:

1. **formatter.go**: Multi-format output
   - Text: Human-readable, colorized
   - JSON: Machine-readable
   - YAML: Structured data

2. **colors.go**: Terminal color handling
   - Auto-detection (TTY detection)
   - Forced modes (always/never)
   - ANSI color codes

**Design Pattern**: Strategy Pattern (format selection)

## Data Flow

### Validation Flow

```
1. User executes: gitlab-ci-lint .gitlab-ci.yml
                    ↓
2. Configuration loading (4 sources with priority)
                    ↓
3. Read CI file content
                    ↓
4. Stage 1: Local YAML validation
                    ↓
         [If invalid] → Print errors → Exit 10
                    ↓
         [If valid and API enabled]
                    ↓
5. GitLab client creation
   - Token validation (if provided)
   - Connection test
                    ↓
6. Stage 2: API validation
   - Build endpoint (global or project-specific)
   - Send request
   - Parse response
                    ↓
7. Result aggregation
   - Combine local + API results
   - Determine overall validity
                    ↓
8. Output formatting
   - Apply selected format
   - Colorize if needed
                    ↓
9. Exit with appropriate code (0, 1, or 10)
```

### Configuration Loading Flow

```
1. Load defaults (hardcoded)
                    ↓
2. Load config file (if exists)
   - Check GCL_CONFIG env var
   - Check default path ~/.tools-config/.gitlab-ci-lint/config.yaml
   - Check --config flag
                    ↓
3. Merge defaults + config file
                    ↓
4. Load environment variables (GCL_*)
                    ↓
5. Merge with previous config
                    ↓
6. Load CLI flags
                    ↓
7. Merge with previous config
                    ↓
8. Validate final config
   - Check output format
   - Check color mode
   - Check instance URL
                    ↓
9. Return final configuration
```

## Key Design Decisions

### Interface-Based Validation

**Decision**: Use `Validator` interface instead of concrete types

**Benefits**:
- Easy to add new validation types
- Testable with mocks
- Clear separation of concerns

**Example**:
```go
type Validator interface {
    Validate(content []byte) Result
}

// Can easily add:
type SecurityValidator struct {}
type PerformanceValidator struct {}
```

### Priority-Based Configuration

**Decision**: 4-level priority chain (defaults → file → env → flags)

**Benefits**:
- Flexible configuration
- No configuration required for basic use
- Environment variables for CI/CD
- Flags for one-time overrides

**Precedence**:
```
CLI Flags (highest)
    ↓
Environment Variables
    ↓
Config File
    ↓
Defaults (lowest)
```

### Graceful Degradation

**Decision**: Tool works with reduced capabilities if API unavailable

**Benefits**:
- Works offline (local-only validation)
- Faster feedback for syntax errors
- No API rate limit concerns

**Implementation**:
```bash
# Explicit skip
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Fallback (if API unreachable)
# Still provides local validation
```

### Structured Error Reporting

**Decision**: All errors include context (line, column, content)

**Benefits**:
- Easy to locate issues
- IDE-friendly
- Better UX

**Structure**:
```go
type Error struct {
    Message string  // Human-readable description
    Line    int     // Line number
    Column  int     // Column position
    Content string  // Actual line content
}
```

## Extensibility Points

### Adding a New Validator

```go
// 1. Implement the interface
type CustomValidator struct {
    // dependencies
}

func (v *CustomValidator) Validate(content []byte) validator.Result {
    // validation logic
    return validator.Result{...}
}

// 2. Wire up in main.go
customValidator := &CustomValidator{...}
result := customValidator.Validate(content)
```

### Adding a New Output Format

```go
// 1. Add format function to formatter.go
func (f *Formatter) formatMarkdown(w io.Writer, results []validator.Result, filename string) {
    // implementation
}

// 2. Update FormatResult switch
switch format {
case "markdown":
    f.formatMarkdown(w, results, filename)
// ...
}
```

### Adding a New Configuration Source

```go
// 1. Add loader in loader.go
func (l *Loader) loadFromDatabase() Config {
    // implementation
}

// 2. Call in Load() before flags
cfg := l.loadFromDatabase()
cfg = l.mergeConfigs(cfg, flagCfg)
```

## Security Considerations

### Token Handling
- Tokens are never logged
- Config files use secure permissions (0600)
- Tokens are sent only over HTTPS

### API Communication
- All API requests use HTTPS
- Timeouts prevent hanging
- Token validation before use

### File Operations
- Atomic writes prevent corruption
- Config validation prevents injection
- No arbitrary code execution

## Performance Characteristics

### Memory
- Static binary, ~15-20 MB
- In-memory file processing
- No persistent state

### CPU
- YAML parsing: < 50ms for typical files
- API validation: 1-3 seconds (network bound)

### Network
- Single API request per validation
- Configurable timeout (default: 30s)
- No background processes

## Testing Architecture

### Unit Tests
- Package-level testing
- Mock HTTP servers for API tests
- Table-driven test cases

### Integration Tests
- End-to-end validation flows
- Real GitLab instances (when available)
- Multi-format output verification

### Test Coverage
- Current: ~72% overall
- Target: 70%+

## Dependencies

### Go
- `github.com/spf13/cobra`: CLI framework
- `gopkg.in/yaml.v3`: YAML parsing
- Standard library: HTTP, JSON, filesystem

### Node.js (Release Tooling)
- `semantic-release`: Automated releases
- `commitlint`: Commit message linting
- `husky`: Git hooks

## Future Architecture Considerations

### Potential Improvements
1. **Plugin System**: Allow custom validators
2. **Caching**: Cache API responses
3. **Parallel Validation**: Validate multiple files concurrently
4. **Streaming**: Process large files incrementally
5. **Web Interface**: Browser-based validation

### Scalability
- Current design supports single-file validation
- Batch validation would require minimal changes
- API client is stateless (ready for scaling)
