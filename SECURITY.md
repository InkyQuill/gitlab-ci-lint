# Security Policy

## Supported Versions

Currently supported versions:
- Version 0.x.x (current development version)

Security updates are applied to the latest version.

## Reporting a Vulnerability

The GitLab CI Lint team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

If you discover a security vulnerability, please do **NOT** open a public issue.

### How to Report

**Private vulnerability disclosure:**
- Send an email to the project maintainer
- Use the subject prefix: `[Security]`
- Include detailed information about the vulnerability
- Provide steps to reproduce if applicable
- Include your suggested fix (if you have one)

**What to Include:**
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any relevant logs or screenshots
- Version of gitlab-ci-lint affected

### What Happens Next?

1. We will send a confirmation email within 48 hours
2. We will investigate the vulnerability
3. We will determine the severity and impact
4. We will develop a fix
5. We will release a security update
6. We will announce the security fix (crediting you if you wish)

### Response Time

We aim to respond to security reports within 48 hours and provide regular updates on our progress.

### Disclosure Policy

We will coordinate disclosure of security vulnerabilities:
- We will not disclose vulnerabilities publicly before a fix is released
- We will provide credit to reporters (unless you wish to remain anonymous)
- We will include details in security advisories after fixes are deployed

## Security Best Practices

### For Users

1. **Token Security**
   - Never commit GitLab tokens to version control
   - Use environment variables or config files with proper permissions (0600)
   - Rotate tokens regularly
   - Use tokens with minimum required scopes (`api` for this tool)

2. **Configuration File Permissions**
   ```bash
   # Ensure config file is readable only by you
   chmod 600 ~/.tools-config/.gitlab-ci-lint/config.yaml
   ```

3. **Network Security**
   - Use HTTPS for GitLab instances
   - Validate SSL certificates (tool does this by default)
   - Be cautious with self-signed certificates

4. **Updates**
   - Keep gitlab-ci-lint updated to the latest version
   - Review changelogs for security updates

### For Developers

1. **Dependency Management**
   - Regularly update Go dependencies
   - Review security advisories for dependencies
   - Use `go mod tidy` to clean up dependencies

2. **Code Review**
   - All code changes go through pull requests
   - At least one approval required
   - Automated security scanning in CI/CD

3. **Secrets Management**
   - Never log tokens or credentials
   - Use secure credential storage
   - Sanitize error messages (remove sensitive data)

## Security Features

### Built-in Protections

- **Token Validation**: Validates GitLab tokens before use
- **URL Sanitization**: Normalizes GitLab instance URLs
- **Secure Defaults**: Safe by default configuration
- **No Credential Logging**: Tokens are never logged
- **Permission Checks**: Validates file permissions

## Security Audits

This project has not yet undergone a formal security audit.

We welcome security researchers to review our code and report vulnerabilities responsibly.

## Contact

For security-related questions or concerns:
- Open a GitHub issue with the `[Security]` tag (non-sensitive matters only)
- Email the project maintainers for sensitive matters

## Thanks

We appreciate your help in keeping gitlab-ci-lint and our users safe!
