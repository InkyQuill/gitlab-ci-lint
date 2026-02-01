# Configuration Reference

This document provides detailed information about configuring GitLab CI Linter.

## Configuration Priority

Configuration is loaded from multiple sources in order of priority (low to high):

1. **Default values** - Built-in defaults
2. **Config file** - `~/.tool-configs/.gitlab-ci-lint/config.yaml` or custom path
3. **Environment variables** - `GCL_*` prefixed variables
4. **CLI flags** - Command-line flags (highest priority)

Higher-priority sources override lower-priority ones.

## Configuration File

### Default Location

The default configuration file location is:

```
~/.tool-configs/.gitlab-ci-lint/config.yaml
```

### Custom Location

You can specify a custom configuration file using:

- CLI flag: `-c /path/to/config.yaml` or `--config /path/to/config.yaml`
- Environment variable: `GCL_CONFIG=/path/to/config.yaml`

### Configuration File Format

```yaml
# GitLab instance configuration
gitlab:
  instance: "https://gitlab.com"  # GitLab instance URL
  timeout: 30s                    # API timeout (Go duration format)

# Authentication
auth:
  token: ""                        # Personal access token (recommended: use GCL_TOKEN instead)
  netrc: false                     # Use ~/.netrc for credentials

# Validation behavior
validation:
  skip_api: false                  # Skip API validation (YAML syntax only)
  strict: true                     # Use strict mode for YAML parsing
  project: ""                      # Project ID for context (e.g., "123" or "group/project")

# Output formatting
output:
  format: "text"                   # Output format: text, json, yaml
  verbose: false                   # Enable verbose output
  color: "auto"                    # Color output: auto, always, never
```

## Environment Variables

All environment variables use the `GCL_` prefix.

### Authentication

| Variable | Description | Example |
|----------|-------------|---------|
| `GCL_TOKEN` | GitLab personal access token | `glpat-xxxxxxxxxxxx` |

### GitLab Instance

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `GCL_INSTANCE` | GitLab instance URL | `https://gitlab.example.com` | `https://gitlab.com` |
| `GCL_TIMEOUT` | API timeout in seconds | `30` | `30` |

### Validation

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `GCL_PROJECT` | Project ID | `mygroup/myproject` or `123` | (empty) |
| `GCL_SKIP_API` | Skip API validation | `true` or `false` | `false` |
| `GCL_STRICT` | Strict YAML parsing | `true` or `false` | `true` |

### Output

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `GCL_OUTPUT_FORMAT` | Output format | `text`, `json`, or `yaml` | `text` |
| `GCL_VERBOSE` | Verbose output | `true` or `false` | `false` |
| `GCL_COLOR` | Color output | `auto`, `always`, or `never` | `auto` |

### Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `GCL_CONFIG` | Path to config file | `/path/to/config.yaml` |

## CLI Flags

### General

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--help` | `-h` | Show help | `gitlab-ci-lint --help` |
| `--version` | `-V` | Show version | `gitlab-ci-lint --version` |

### Configuration

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--config` | `-c` | Path to config file | `--config /path/to/config.yaml` |

### Authentication

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--token` | `-t` | GitLab personal access token | `--token glpat-xxxxxx` |
| `--netrc` | | Use .netrc for credentials | `--netrc` |

### GitLab Instance

| Flag | Description | Example | Default |
|------|-------------|---------|---------|
| `--instance` | GitLab instance URL | `--instance https://gitlab.example.com` | `https://gitlab.com` |
| `--timeout` | API timeout | `--timeout 30s` | `30s` |
| `--project` | Project ID | `--project mygroup/myproject` | (empty) |

### Validation

| Flag | Short | Description | Example | Default |
|------|-------|-------------|---------|---------|
| `--skip-api` | `-s` | Skip API validation | `--skip-api` | `false` |
| `--strict` | | Use strict YAML parsing | `--strict` | `true` |

### Output

| Flag | Short | Description | Example | Default |
|------|-------|-------------|---------|---------|
| `--output` | `-o` | Output format | `--output json` | `text` |
| `--verbose` | `-v` | Verbose output | `--verbose` | `false` |
| `--color` | | Color output | `--color always` | `auto` |

## Configuration Examples

### Example 1: Local Validation Only

For quick local validation without API calls:

```yaml
# ~/.tool-configs/.gitlab-ci-lint/config.yaml
validation:
  skip_api: true
```

Or via CLI:

```bash
gitlab-ci-lint --skip-api .gitlab-ci.yml
```

Or via environment:

```bash
export GCL_SKIP_API=true
gitlab-ci-lint .gitlab-ci.yml
```

### Example 2: Self-Hosted GitLab Instance

```yaml
gitlab:
  instance: "https://gitlab.example.com"
```

Or:

```bash
export GCL_INSTANCE=https://gitlab.example.com
```

### Example 3: Project-Specific Validation

When you need project-specific context:

```yaml
validation:
  project: "mygroup/myproject"
```

### Example 4: JSON Output for CI/CD

```yaml
output:
  format: "json"
  verbose: false
  color: "never"
```

Or:

```bash
gitlab-ci-lint --output json --color never .gitlab-ci.yml
```

### Example 5: Using Environment Variables for Sensitive Data

Store only non-sensitive data in config file, use environment for token:

```yaml
# ~/.tool-configs/.gitlab-ci-lint/config.yaml
gitlab:
  instance: "https://gitlab.com"
  timeout: 30s

auth:
  token: ""  # Leave empty, use GCL_TOKEN instead

validation:
  skip_api: false
  strict: true

output:
  format: "text"
  color: "auto"
```

Then set token in your shell:

```bash
export GCL_TOKEN=glpat-xxxxxxxxxxxx
```

### Example 6: Multiple GitLab Instances

Use different configuration files for different instances:

```bash
# Default config
~/.tool-configs/.gitlab-ci-lint/config.yaml  # GitLab.com

# Custom config for self-hosted
~/.tool-configs/.gitlab-ci-lint/gitlab-ee.yaml
```

Use the custom config:

```bash
gitlab-ci-lint --config ~/.tool-configs/.gitlab-ci-lint/gitlab-ee.yaml .gitlab-ci.yml
```

## Duration Format

Timeout values use Go's duration format:

| Format | Description | Example |
|--------|-------------|---------|
| `ns` | Nanoseconds | `1000000000ns` |
| `us` or `µs` | Microseconds | `1000000us` |
| `ms` | Milliseconds | `1000ms` |
| `s` | Seconds | `30s` (default) |
| `m` | Minutes | `1m` |
| `h` | Hours | `1h` |

## Color Settings

| Value | Description |
|-------|-------------|
| `auto` | Enable colors if output is a terminal (default) |
| `always` | Always enable colors |
| `never` | Never enable colors |

## Token Permissions

Your GitLab personal access token needs the following scopes:

- `api` - Full API access (for CI lint endpoint)

For more information, see GitLab's [Personal Access Tokens documentation](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html).

## .netrc Support

You can store credentials in `~/.netrc`:

```
machine gitlab.com
login <username>
password <token>
```

Then enable with:

```bash
gitlab-ci-lint --netrc .gitlab-ci.yml
```

Or in config:

```yaml
auth:
  netrc: true
```

## Troubleshooting

### Configuration Not Loading

Check if the config file exists and is valid YAML:

```bash
cat ~/.tool-configs/.gitlab-ci-lint/config.yaml
```

### Environment Variables Not Working

Ensure variables are exported:

```bash
echo $GCL_TOKEN
```

### Wrong Configuration Priority Applied

Use verbose mode to see which configuration is being used:

```bash
gitlab-ci-lint --verbose .gitlab-ci.yml
```
