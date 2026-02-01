# Basic Usage Examples

This document provides practical examples of using GitLab CI Lint.

## Simple Validation

### Check a CI file

```bash
gitlab-ci-lint .gitlab-ci.yml
```

Output:
```
✓ Local validation passed
✓ API validation passed

Configuration .gitlab-ci.yml is valid
```

### Check with syntax errors

```bash
cat > invalid.yml <<EOF
image: alpine:latest

build:
  stage: build
    script: echo "test"  # Wrong indentation
EOF

gitlab-ci-lint invalid.yml
```

Output:
```
✗ Local validation failed

Error: mapping values are not allowed in this context
  Line: 5
  Column: 4
  Content:     script: echo "test"
```

## Output Formats

### Text (Default)

```bash
gitlab-ci-lint .gitlab-ci.yml
```

Human-readable with colors and symbols.

### JSON

```bash
gitlab-ci-lint --output json .gitlab-ci.yml
```

Output:
```json
{
  "file": ".gitlab-ci.yml",
  "valid": true,
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
      "warnings": []
    }
  ]
}
```

Use in scripts:
```bash
valid=$(gitlab-ci-lint --output json .gitlab-ci.yml | jq -r '.valid')
if [ "$valid" = "true" ]; then
  echo "Valid!"
fi
```

### YAML

```bash
gitlab-ci-lint --output yaml .gitlab-ci.yml
```

Output:
```yaml
file: .gitlab-ci.yml
valid: true
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

## Local-Only Validation

Skip API validation for faster feedback:

```bash
gitlab-ci-lint --skip-api .gitlab-ci.yml
```

Use cases:
- Development iterations
- Offline work
- Syntax-only validation
- Pre-commit hooks

## Verbose Mode

Show detailed information:

```bash
gitlab-ci-lint --verbose .gitlab-ci.yml
```

Shows:
- Validation steps
- API endpoint URLs
- Warnings (even if valid)
- Token validation

## Color Control

### Auto (Default)

Detects TTY, enables colors for terminals:
```bash
gitlab-ci-lint .gitlab-ci.yml
```

### Always

Force colors (useful for CI logs):
```bash
gitlab-ci-lint --color always .gitlab-ci.yml | less -R
```

### Never

Disable colors (useful for logging):
```bash
gitlab-ci-lint --color never .gitlab-ci.yml > validation.log
```

## Project-Specific Validation

Validate with project context (job references, variables):

```bash
# By project ID
gitlab-ci-lint --project 123 .gitlab-ci.yml

# By project path
gitlab-ci-lint --project mygroup/myproject .gitlab-ci.yml

# With nested groups
gitlab-ci-lint --project mygroup/subgroup/project .gitlab-ci.yml
```

Use when:
- CI file references other jobs
- Using project variables
- Need context-aware validation

## Self-Hosted GitLab

### Using environment variable

```bash
export GCL_INSTANCE="https://gitlab.example.com"
export GCL_TOKEN="glpat-xxxxxxxxxxxx"
gitlab-ci-lint .gitlab-ci.yml
```

### Using flags

```bash
gitlab-ci-lint \
  --instance https://gitlab.example.com \
  --token glpat-xxxxxxxxxxxx \
  .gitlab-ci.yml
```

### Using config file

```yaml
# ~/.tools-config/.gitlab-ci-lint/config.yaml
gitlab:
  instance: "https://gitlab.example.com"
  timeout: 60s
auth:
  token: "glpat-xxxxxxxxxxxx"
```

## Exit Codes in Scripts

### Check result

```bash
#!/bin/bash

gitlab-ci-lint .gitlab-ci.yml
case $? in
  0)
    echo "✓ Valid configuration"
    ;;
  1)
    echo "✗ Runtime error (check auth/network)"
    exit 1
    ;;
  10)
    echo "✗ Invalid CI configuration"
    exit 1
    ;;
esac
```

### In Makefile

```makefile
.PHONY: validate-ci
validate-ci:
	@echo "Validating CI configuration..."
	@if gitlab-ci-lint .gitlab-ci.yml; then \
		echo "✓ Valid"; \
	else \
		echo "✗ Invalid"; \
		exit 1; \
	fi
```

### In GitLab CI

```yaml
lint:
  stage: test
  script:
    - gitlab-ci-lint .gitlab-ci.yml
  allow_failure: false
```

## Common Workflows

### Before Committing

```bash
# Quick syntax check
gitlab-ci-lint --skip-api .gitlab-ci.yml

# Full validation
gitlab-ci-lint .gitlab-ci.yml

# If valid, commit
git add .gitlab-ci.yml
git commit -m "Update CI configuration"
```

### Batch Validation

```bash
# Validate all CI files in a project
find . -name ".gitlab-ci.yml" -o -name "*.gitlab-ci.yml" | while read file; do
  echo "Validating $file..."
  gitlab-ci-lint "$file"
done
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
echo "Validating CI configuration..."

for file in $(git diff --cached --name-only | grep -E '\.(gitlab-ci|ci)\.yml$'); do
  if ! gitlab-ci-lint "$file"; then
    echo "✗ Invalid CI configuration in $file"
    exit 1
  fi
done

echo "✓ All CI files are valid"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Advanced Scenarios

### Validate with different instances

```bash
# Test against staging GitLab
gitlab-ci-lint --instance https://staging.gitlab.com .gitlab-ci.yml

# Test against production GitLab
gitlab-ci-lint --instance https://gitlab.com .gitlab-ci.yml
```

### Compare outputs

```bash
# Local validation only
gitlab-ci-lint --skip-api .gitlab-ci.yml > local.txt

# Full validation
gitlab-ci-lint .gitlab-ci.yml > full.txt

# Compare
diff local.txt full.txt
```

### Parse JSON output

```bash
# Get validity status
gitlab-ci-lint --output json .gitlab-ci.yml | jq -r '.valid')

# Count errors
errors=$(gitlab-ci-lint --output json .gitlab-ci.yml | \
  jq '[.results[].errors] | add | length')

echo "Found $errors errors"

# Extract error messages
gitlab-ci-lint --output json .gitlab-ci.yml | \
  jq -r '.results[].errors[]?.message'
```

### Debug mode

```bash
# Enable verbose output
gitlab-ci-lint --verbose .gitlab-ci.yml

# With colors forced
gitlab-ci-lint --verbose --color always .gitlab-ci.yml | less -R

# Save debug output
gitlab-ci-lint --verbose .gitlab-ci.yml 2>&1 | tee debug.log
```

## Examples by Use Case

### Quick Development Check

```bash
# Fast, syntax-only validation
alias gcl='gitlab-ci-lint --skip-api'
gcl .gitlab-ci.yml
```

### Pre-push Validation

```bash
# .git/hooks/pre-push
#!/bin/bash
gitlab-ci-lint --skip-api .gitlab-ci.yml || {
  echo "CI configuration is invalid. Commit aborted."
  exit 1
}
```

### CI/CD Pipeline Integration

```yaml
# In your GitLab CI
validate:
  stage: test
  script:
    - gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
  only:
    changes:
      - .gitlab-ci.yml
```

### Monitoring and Alerts

```bash
#!/bin/bash
# Periodic validation check

result=$(gitlab-ci-lint --output json .gitlab-ci.yml)
valid=$(echo "$result" | jq -r '.valid')

if [ "$valid" != "true" ]; then
  errors=$(echo "$result" | jq -r '.results[].errors[]?.message' | head -5)
  echo "Alert: CI configuration invalid!"
  echo "Errors: $errors"
  # Send alert (e.g., Slack, email)
fi
```

## Tips and Tricks

### Create an alias for common options

```bash
# ~/.bashrc or ~/.zshrc
alias gcl='gitlab-ci-lint'
alias gclv='gitlab-ci-lint --verbose'
alias gclj='gitlab-ci-lint --output json'
```

### Use with configuration files

```bash
# Project-specific config
cat > .gitlab-ci-lint.yaml <<EOF
gitlab:
  instance: "https://gitlab.example.com"
  project: "mygroup/myproject"
validation:
  strict: false
output:
  format: "text"
  verbose: true
EOF

gitlab-ci-lint --config .gitlab-ci-lint.yaml .gitlab-ci.yml
```

### Validate from stdin

```bash
cat .gitlab-ci.yml | gitlab-ci-lint /dev/stdin
```

### Shell completion

```bash
# Generate completion script
gitlab-ci-lint completion bash > /etc/bash_completion.d/gitlab-ci-lint
gitlab-ci-lint completion zsh > /usr/local/share/zsh/site-functions/_gitlab-ci-lint

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```
