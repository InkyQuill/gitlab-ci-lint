# Support Guide

## Getting Help

There are several ways to get help with gitlab-ci-lint:

### Documentation

- 📚 [Quick Start Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki) - Get started in minutes
- 🔧 [Configuration Reference](CONFIG.md) - Detailed configuration options
- 🐛 [Troubleshooting Guide](docs/guides/troubleshooting.md) - Common issues and solutions
- 📖 [Examples](docs/examples/) - Usage examples

### Quick Commands

```bash
# Get help
gitlab-ci-lint --help

# Get help for specific command
gitlab-ci-lint setup --help

# Check version
gitlab-ci-lint version

# Validate configuration
gitlab-ci-lint --verbose .gitlab-ci.yml
```

## Common Issues

### "command not found: gitlab-ci-lint"

**Problem**: The binary is not in your PATH.

**Solutions**:
```bash
# If installed to /usr/local/bin
export PATH=/usr/local/bin:$PATH

# If installed to ~/.local/bin
export PATH=~/.local/bin:$PATH

# If installed via go install
export PATH=$(go env GOPATH)/bin:$PATH

# Add to ~/.bashrc or ~/.zshrc to make it permanent
echo 'export PATH=/usr/local/bin:$PATH' >> ~/.bashrc
```

### "Token validation failed"

**Problem**: Invalid or expired GitLab token.

**Solutions**:
```bash
# Regenerate your token at GitLab User Settings → Access Tokens
# Ensure token has 'api' scope

# Test token manually
curl -H "PRIVATE-TOKEN: glpat-your-token" https://gitlab.com/api/v4/user

# Update token in config
gitlab-ci-lint setup

# Or use environment variable
export GCL_TOKEN=glpat-your-new-token
```

### "connection refused" / "timeout"

**Problem**: Cannot connect to GitLab instance.

**Solutions**:
```bash
# Check network connectivity
ping gitlab.com

# Verify instance URL
export GCL_INSTANCE=https://gitlab.example.com

# Increase timeout
gitlab-ci-lint --timeout 60s .gitlab-ci.yml

# Check for proxy settings
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
```

### "No CI configuration files found"

**Problem**: File discovery didn't find any `.gitlab-ci.yml` files.

**Solutions**:
```bash
# Specify file explicitly
gitlab-ci-lint -f .gitlab-ci.yml

# Search specific directory
gitlab-ci-lint -d ./ci

# Check current directory
ls -la .gitlab-ci.yml

# Validate from stdin
cat .gitlab-ci.yml | gitlab-ci-lint -
```

### "permission denied"

**Problem**: Binary is not executable.

**Solution**:
```bash
chmod +x gitlab-ci-lint
```

## Asking Questions

### Before Asking

1. **Check documentation** - Look for answers in existing docs
2. **Search issues** - Check if your question was already answered
3. **Try verbose mode** - Run with `-v` flag for more information
4. **Check version** - Ensure you're using the latest version

### How to Ask

**Good questions include**:
- What you're trying to accomplish
- What you've already tried
- Error messages or unexpected behavior
- Environment details (OS, gitlab-ci-lint version)
- Minimal reproduction case

**Example**:
```
Hi, I'm trying to validate my CI configuration but getting an error.

What I'm trying:
gitlab-ci-lint --project mygroup/myproject .gitlab-ci.yml

Error:
API validation failed: 404 Not Found

Environment:
- OS: Ubuntu 22.04
- gitlab-ci-lint: v0.1.0
- GitLab instance: https://gitlab.com

What I've tried:
- Verified token has api scope
- Project path is correct (works in GitLab UI)
- Local validation works fine with --skip-api
```

### Where to Ask

1. **GitHub Issues** - Bug reports and feature requests
2. **GitHub Discussions** - Questions and community discussions
3. **Documentation** - Wiki and docs for reference

## Reporting Bugs

See [SECURITY.md](SECURITY.md) for security vulnerabilities.

For non-security bugs:

1. **Search existing issues** first
2. **Use bug report template**
3. **Provide minimal reproduction**
4. **Include environment details**
5. **Add relevant logs**

## Feature Requests

We welcome feature requests!

1. **Check ROADMAP.md** for planned features
2. **Search existing requests**
3. **Use feature request template**
4. **Describe use case** - Why do you need this?
5. **Consider contributing** - PRs welcome!

## Contributing

We love contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Start

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/gitlab-ci-lint.git
cd gitlab-ci-lint

# Add upstream
git remote add upstream https://github.com/InkyQuill/gitlab-ci-lint.git

# Create branch
git checkout -b feature/your-feature

# Make changes and test
make test-unit
make lint

# Commit and push
git commit -m "feat: add your feature"
git push origin feature/your-feature

# Open PR from your fork
```

## Community Resources

### Documentation
- GitHub Wiki: https://github.com/InkyQuill/gitlab-ci-lint/wiki
- Examples: https://github.com/InkyQuill/gitlab-ci-lint/tree/main/docs/examples

### Development
- Roadmap: [ROADMAP.md](ROADMAP.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Architecture: [docs/architecture/overview.md](docs/architecture/overview.md)

### CI/CD Status
- GitHub Actions: https://github.com/InkyQuill/gitlab-ci-lint/actions
- Coverage: https://codecov.io/gh/InkyQuill/gitlab-ci-lint

## Professional Support

This is an open-source project maintained by volunteers.

**Current Status**: Community support only

For enterprise support or custom development, contact the maintainers.

## Acknowledgments

This project was developed with assistance from:
- **Claude Code** (Anthropic) - AI-powered development assistant
- **Z.ai** (GLM 4.7) - AI code analysis and generation

Their help was instrumental in:
- Code architecture and design
- Test coverage and quality assurance
- Documentation and best practices
- CI/CD automation setup

## Response Times

- **Bug reports**: We aim to respond within 7 days
- **Feature requests**: Community discussion, no fixed timeline
- **Security issues**: Within 48 hours (see SECURITY.md)
- **Questions**: Community-supported, best-effort basis

## License

MIT License - see [LICENSE](LICENSE) for details.
