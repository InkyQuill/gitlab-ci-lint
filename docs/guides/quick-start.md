# Quick Start Guide

This guide will help you get started with GitLab CI Lint in just a few minutes.

## Installation

### From Source

```bash
git clone https://github.com/InkyQuill/gitlab-ci-lint.git
cd gitlab-ci-lint
make install
```

### From Binaries

Download the latest binary from the [releases page](https://github.com/InkyQuill/gitlab-ci-lint/releases) and add it to your PATH.

## Interactive Setup (Recommended)

The easiest way to configure gitlab-ci-lint is using the interactive setup wizard:

```bash
gitlab-ci-lint setup
```

The wizard will guide you through:
1. GitLab instance URL (default: https://gitlab.com)
2. Personal access token configuration
3. Default project (optional)
4. Output preferences

The configuration is saved to `~/.tools-config/.gitlab-ci-lint/config.yaml`.

## Manual Configuration

Alternatively, you can configure manually or use command-line flags.

### Option 1: Config File

Create `~/.tools-config/.gitlab-ci-lint/config.yaml`:

```yaml
gitlab:
  instance: "https://gitlab.com"
  timeout: 30s
auth:
  token: "glpat-your-token-here"
validation:
  skip_api: false
  strict: true
  project: ""
output:
  format: "text"
  verbose: false
  color: "auto"
```

### Option 2: Environment Variables

```bash
export GCL_TOKEN="glpat-your-token-here"
export GCL_INSTANCE="https://gitlab.com"
export GCL_OUTPUT_FORMAT="json"
```

### Option 3: Command-Line Flags

```bash
gitlab-ci-lint --token glpat-your-token --instance https://gitlab.com .gitlab-ci.yml
```

## Creating a GitLab Personal Access Token

1. Go to GitLab → User Settings → Access Tokens
2. Create a new token with `read_api` scope
3. Copy the token (starts with `glpat-`)

## Basic Usage

### Validate a CI Configuration File

```bash
gitlab-ci-lint .gitlab-ci.yml
```

### Skip API Validation (Local Only)

```bash
gitlab-ci-lint --skip-api .gitlab-ci.yml
```

### Output Formats

```bash
# Text (default, human-readable)
gitlab-ci-lint .gitlab-ci.yml

# JSON (for scripting)
gitlab-ci-lint --output json .gitlab-ci.yml

# YAML
gitlab-ci-lint --output yaml .gitlab-ci.yml
```

### Project-Specific Validation

For projects with job references or variables:

```bash
gitlab-ci-lint --project group/project .gitlab-ci.yml
# or
gitlab-ci-lint --project 123 .gitlab-ci.yml
```

## Understanding the Output

### Successful Validation

```
✓ Local validation passed
✓ API validation passed

Configuration .gitlab-ci.yml is valid
```

### Failed Validation

```
✗ Local validation failed

Error: invalid YAML syntax
  Line: 15
  Column: 3
  Content:   script: echo "test

  Missing closing quote
```

### With Warnings (Verbose Mode)

```bash
gitlab-ci-lint --verbose .gitlab-ci.yml
```

```
✓ Local validation passed
⚠ API validation passed with warnings

Warning: Deprecated image 'ruby:2.5' (line 8)
Warning: Job 'test' has no stage specified (line 12)
```

## Exit Codes

- `0`: All validations passed
- `1`: Runtime error (file not found, auth failed, etc.)
- `10`: CI configuration is invalid

Use in CI/CD:

```bash
#!/bin/bash
gitlab-ci-lint .gitlab-ci.yml
case $? in
  0) echo "✓ Valid CI config" ;;
  1) echo "✗ Runtime error"; exit 1 ;;
  10) echo "✗ Invalid CI config"; exit 1 ;;
esac
```

## Self-Hosted GitLab

```bash
gitlab-ci-lint setup
# When prompted, enter: https://gitlab.example.com
```

Or manually:

```bash
export GCL_INSTANCE="https://gitlab.example.com"
export GCL_TOKEN="glpat-your-token"
gitlab-ci-lint .gitlab-ci.yml
```

## Next Steps

- Read [Troubleshooting Guide](troubleshooting.md) for common issues
- See [Advanced Usage](advanced-usage.md) for more features
- Check [CI Integration](../examples/ci-integration.md) for pipeline usage
- Contribute! See [Contributing Guide](../development/contributing.md)

## Help

```bash
gitlab-ci-lint --help
gitlab-ci-lint version
```

For more information, visit:
- [Project Repository](https://github.com/InkyQuill/gitlab-ci-lint)
- [GitLab CI Documentation](https://docs.gitlab.com/ee/ci/yaml/)
