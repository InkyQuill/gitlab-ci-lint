# Contributing to GitLab CI Lint

Thank you for your interest in contributing to GitLab CI Lint! This document provides guidelines for contributing.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

- Go 1.25.6 or later
- Make (for build automation)
- Git

### Development Setup

```bash
# Fork the repository
# Clone your fork
git clone https://github.com/YOUR_USERNAME/gitlab-ci-lint.git
cd gitlab-ci-lint

# Add upstream remote
git remote add upstream https://github.com/InkyQuill/gitlab-ci-lint.git

# Install dependencies
go mod download

# Build
make build

# Run tests
make test-unit
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write clear, concise code
- Follow existing code style
- Add tests for new features
- Update documentation

### 3. Run Tests

```bash
# Unit tests
make test-unit

# Linting
make lint

# Formatting
make fmt

# All checks
make test-all
```

### 4. Commit Changes

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for custom validation rules
fix: handle timeout errors gracefully
docs: update installation instructions
test: add tests for API validator
refactor: simplify configuration loading
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Style

### Go Code

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `goimports` to organize imports
- Keep functions focused and small
- Add godoc comments for exported types/functions

```bash
make fmt
```

### Example

```go
// ValidateConfig validates the configuration and returns an error if invalid.
// It checks that all required fields are present and valid.
func ValidateConfig(cfg *Config) error {
    // implementation
}
```

### Documentation

- Use clear, descriptive language
- Include examples for complex features
- Update README.md for user-facing changes
- Update relevant docs/ files

## Testing

### Writing Tests

- Write tests for all new functionality
- Aim for >70% code coverage
- Use table-driven tests for multiple cases
- Mock external dependencies (HTTP, filesystem)

### Example Test

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: Config{
                Instance: "https://gitlab.com",
                Token:    "test-token",
            },
            wantErr: false,
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(&tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests

```bash
# Unit tests
make test-unit

# With coverage
make test-coverage

# Specific package
go test -v ./internal/validator
```

## Pull Request Guidelines

### PR Title

Use conventional commit format:

```
feat: add new validation rule for deprecated syntax
fix: handle empty project reference gracefully
docs: improve quick start guide
```

### PR Description

Include:
- **What**: Summary of changes
- **Why**: Reason for the change
- **How**: Implementation approach (if complex)
- **Testing**: How you tested the changes
- **Breaking Changes**: Note any breaking changes
- **Screenshots**: For UI changes (if applicable)

### Checklist

- [ ] Tests pass locally
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Commit messages follow conventional commits
- [ ] No merge conflicts with main
- [ ] Code follows project style
- [ ] PR description is clear and complete

## Project Structure

```
gitlab-ci-lint/
├── cmd/
│   ├── gitlab-ci-lint/     # Main CLI command
│   └── setup/              # Setup wizard
├── internal/
│   ├── config/             # Configuration management
│   ├── gitlab/             # GitLab API client
│   ├── output/             # Output formatting
│   ├── setup/              # Setup validation
│   └── validator/          # Validation logic
├── pkg/
│   └── version/            # Version information
├── docs/                   # Documentation
├── .github/workflows/      # CI/CD workflows
└── Makefile                # Build targets
```

## Adding Features

### 1. Design First

- Open an issue to discuss the feature
- Get feedback from maintainers
- Document the design

### 2. Implementation

- Follow existing patterns
- Keep changes focused
- Write tests as you go

### 3. Documentation

- Update relevant docs
- Add examples
- Update CHANGELOG (maintainers will handle)

## Reporting Issues

### Bug Reports

Include:
- Go version
- gitlab-ci-lint version
- Steps to reproduce
- Expected vs actual behavior
- Error messages
- Configuration (sanitized)

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternative approaches considered
- Examples of similar tools

## Release Process

Releases are automated using semantic-release:

1. Merge PR to main
2. CI runs tests
3. semantic-release analyzes commits
4. Version bumped automatically
5. Release created on GitHub
6. Binaries built and attached

### Version Bumping

Based on commit messages:

- `feat:` → Minor version bump (1.x → 1.y)
- `fix:` → Patch version bump (1.y.z → 1.y.(z+1))
- `BREAKING CHANGE:` → Major version bump (x → x+1)

## Community Guidelines

### Communication

- Use GitHub issues for questions
- Be patient with responses
- Search existing issues first

### Review Process

- Maintainers will review PRs as time allows
- Address review feedback promptly
- Ask questions if anything is unclear

### Recognition

Contributors are recognized in:
- Release notes
- CONTRIBUTORS file
- GitHub contributor statistics

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Project Specification](SPEC.md)
- [Architecture Overview](docs/architecture/overview.md)

## Getting Help

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and ideas
- Documentation: See docs/ directory
- Email: Open an issue for maintainer contact

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
