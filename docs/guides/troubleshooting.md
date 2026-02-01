# Troubleshooting Guide

This guide helps you resolve common issues when using GitLab CI Lint.

## Authentication Issues

### "authentication failed: Invalid or expired token"

**Problem**: Your personal access token is invalid or has expired.

**Solutions**:
1. Generate a new token in GitLab (User Settings → Access Tokens)
2. Ensure the token has `read_api` scope
3. Update your configuration:
   ```bash
   gitlab-ci-lint setup  # Re-run setup wizard
   # or manually update ~/.tools-config/.gitlab-ci-lint/config.yaml
   ```
4. Or override with environment variable:
   ```bash
   export GCL_TOKEN="glpat-new-token"
   ```

### "Token validation failed with status 403"

**Problem**: Token lacks required permissions.

**Solution**: Ensure your token has the `read_api` scope.

### Token works in browser but fails with gitlab-ci-lint

**Problem**: Instance URL mismatch.

**Solution**: Verify you're connecting to the correct instance:
```bash
gitlab-ci-lint --instance https://gitlab.com .gitlab-ci.yml
# or for self-hosted:
gitlab-ci-lint --instance https://gitlab.example.com .gitlab-ci.yml
```

## Network Issues

### "Failed to connect to instance"

**Problem**: Network connectivity or firewall issues.

**Solutions**:
1. Verify connectivity:
   ```bash
   curl https://gitlab.com/api/v4/user
   ```
2. Check if proxy is required:
   ```bash
   export HTTP_PROXY=http://proxy.example.com:8080
   export HTTPS_PROXY=http://proxy.example.com:8080
   ```
3. Increase timeout:
   ```bash
   gitlab-ci-lint --timeout 60s .gitlab-ci.yml
   ```

### "Timeout exceeded"

**Problem**: Slow network or large CI file.

**Solution**: Increase timeout:
```bash
gitlab-ci-lint --timeout 120s .gitlab-ci.yml
```

Or in config file:
```yaml
gitlab:
  timeout: 120s
```

## YAML Parsing Errors

### "invalid YAML syntax"

**Problem**: Syntax error in your `.gitlab-ci.yml`.

**Common Issues**:
1. **Indentation errors** (YAML is strict about spaces, no tabs)
2. **Unquoted special characters** (`:`, `{`, `}`, `[`, `]`, `,`, `#`)
3. **Mismatched brackets/quotes**

**Example**:
```yaml
# ❌ Wrong (mixed tabs and spaces)
build:
	script: echo "test"

# ✅ Correct (consistent spaces)
build:
  script: echo "test"
```

**Solution**: Use a YAML validator in your IDE:
- VS Code: YAML extension
- IntelliJ: Built-in YAML support
- Online: [YAML Lint](https://www.yamllint.com/)

### "mapping values are not allowed in this context"

**Problem**: Incorrect YAML syntax.

**Example**:
```yaml
# ❌ Wrong
job1:
  stage: build script: echo "test"

# ✅ Correct
job1:
  stage: build
  script:
    - echo "test"
```

## Configuration Issues

### "invalid output format"

**Problem**: Unsupported output format specified.

**Valid formats**: `text`, `json`, `yaml`

```bash
# ✅ Correct
gitlab-ci-lint --output text .gitlab-ci.yml
gitlab-ci-lint --output json .gitlab-ci.yml
gitlab-ci-lint --output yaml .gitlab-ci.yml
```

### "invalid color setting"

**Problem**: Invalid color mode.

**Valid modes**: `auto`, `always`, `never`

```bash
# ✅ Correct
gitlab-ci-lint --color auto .gitlab-ci.yml    # Default: detect TTY
gitlab-ci-lint --color always .gitlab-ci.yml  # Force colors
gitlab-ci-lint --color never .gitlab-ci.yml   # Disable colors
```

### Config file not being loaded

**Problem**: Configuration file path is incorrect.

**Solutions**:
1. Check default location:
   ```bash
   ls ~/.tools-config/.gitlab-ci-lint/config.yaml
   ```
2. Specify config file explicitly:
   ```bash
   gitlab-ci-lint --config /path/to/config.yaml .gitlab-ci.yml
   ```
3. Use environment variable:
   ```bash
   export GCL_CONFIG=/path/to/config.yaml
   ```

## GitLab Instance Issues

### "API returned status 404"

**Problem**: Invalid project path or ID.

**Solutions**:
1. Verify project exists in GitLab
2. Use correct format:
   ```bash
   # Project ID (number)
   gitlab-ci-lint --project 123 .gitlab-ci.yml

   # Project path (with namespace)
   gitlab-ci-lint --project group/subgroup/project .gitlab-ci.yml
   ```
3. Check URL encoding for special characters:
   ```bash
   # Project with dots or special chars
   gitlab-ci-lint --project "my.group/project" .gitlab-ci.yml
   ```

### Self-hosted GitLab returns 404

**Problem**: API endpoint path differs for older GitLab versions.

**Solution**: Ensure GitLab version is 11.0+ (when CI Lint API was introduced).

## File Issues

### "Failed to read file"

**Problem**: File doesn't exist or permission denied.

**Solutions**:
1. Check file exists:
   ```bash
   ls -la .gitlab-ci.yml
   ```
2. Check file path:
   ```bash
   gitlab-ci-lint /full/path/to/.gitlab-ci.yml
   ```
3. Verify read permissions:
   ```bash
   chmod +r .gitlab-ci.yml
   ```

## CI/CD Pipeline Issues

### GitLab CI Lint fails in pipeline but works locally

**Problem**: Environment differences.

**Common causes**:
1. Missing environment variables in CI config
2. Different GitLab instance (CI runner vs. local)
3. Token not available in CI environment

**Solution**: Use `CI_JOB_TOKEN` in GitLab CI:
```yaml
lint:
  stage: test
  script:
    - gitlab-ci-lint --token $CI_JOB_TOKEN .gitlab-ci.yml
```

See [CI Integration Guide](../examples/ci-integration.md) for more details.

## Performance Issues

### Slow validation on large files

**Problem**: Large CI configuration files take time to validate.

**Solutions**:
1. Skip API validation for syntax-only checks:
   ```bash
   gitlab-ci-lint --skip-api .gitlab-ci.yml
   ```
2. Increase timeout:
   ```bash
   gitlab-ci-lint --timeout 120s .gitlab-ci.yml
   ```

### Frequent validations are slow

**Problem**: Network overhead for repeated validations.

**Solution**: Use local-only validation during development:
```bash
# Development: Fast local checks
alias gitlab-ci-lint-dev='gitlab-ci-lint --skip-api'

# Pre-commit: Full validation
gitlab-ci-lint .gitlab-ci.yml
```

## Debug Mode

Enable verbose output to diagnose issues:

```bash
gitlab-ci-lint --verbose .gitlab-ci.yml
```

This shows:
- Token validation steps
- API endpoint URLs
- Network request details
- Warning messages

## Getting Help

If you're still stuck:

1. **Check the logs**: Run with `--verbose`
2. **Verify configuration**: Run `gitlab-ci-lint setup` to reconfigure
3. **Test connectivity**:
   ```bash
   curl -H "PRIVATE-TOKEN: $GCL_TOKEN" https://gitlab.com/api/v4/user
   ```
4. **Report issues**:
   - [GitHub Issues](https://github.com/InkyQuill/gitlab-ci-lint/issues)
   - Include: gitlab-ci-lint version, GitLab version, error message, and `-`verbose output

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `authentication failed` | Invalid/expired token | Generate new token |
| `connection refused` | Wrong instance URL or network issue | Check URL and network |
| `timeout` | Slow network or large file | Increase `--timeout` |
| `404 Not Found` | Project doesn't exist or wrong path | Verify project path/ID |
| `invalid YAML` | Syntax error in CI file | Fix YAML syntax |
| `mapping values not allowed` | YAML structure error | Check indentation/colons |
| `token validation failed` | Insufficient permissions | Add `read_api` scope to token |
