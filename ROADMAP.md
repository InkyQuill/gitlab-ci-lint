# GitLab CI Lint Roadmap

## Completed ✅

### Core Functionality
- [x] Two-stage validation system (local + API)
- [x] Configuration management with priority-based loading
- [x] Multi-format output (text, JSON, YAML)
- [x] Cross-platform builds (Linux, macOS, Windows)
- [x] GitLab instance URL normalization
- [x] Exit code standardization (0, 1, 10)
- [x] CLI flag and environment variable configuration

### Documentation
- [x] README.md with usage examples
- [x] INSTALL.md with installation instructions
- [x] CONFIG.md with configuration details
- [x] CLAUDE.md with development guidance

### Build Infrastructure
- [x] Makefile with build targets
- [x] Version injection at build time
- [x] Multi-platform binary builds

---

## In Progress 🚧

### File Discovery System
- [x] Phase: Enhanced file discovery and multi-file validation
  - [x] `internal/discover/discoverer.go` - File discovery package
  - [x] `internal/discover/discoverer_test.go` - Comprehensive unit tests
  - [x] Auto-discovery from current/parent directories
  - [x] Multiple files support with `-f` flag
  - [x] Recursive directory scanning with `-d` flag
  - [x] Combined file and directory validation
  - [x] Batch validation with summary output
  - [x] Smart directory filtering (node_modules, .git, vendor, etc.)
  - [x] Integration tests for all discovery modes
  - [x] Backward compatibility preserved
  - [x] Stdin input support maintained

### Testing Suite
- [ ] Phase 3: Comprehensive testing (Target: 70%+ coverage)
  - [x] Unit tests
    - [x] `internal/validator/local_test.go`
    - [x] `internal/validator/api_test.go`
    - [x] `internal/config/loader_test.go`
    - [x] `internal/gitlab/client_test.go`
    - [x] `internal/output/formatter_test.go`
    - [x] `internal/discover/discoverer_test.go`
  - [x] Integration tests
    - [x] `tests/integration/setup_test.go`
    - [x] `tests/integration/validation_test.go`
    - [x] File discovery integration tests
  - [x] Update Makefile targets
    - [x] `make test-unit`
    - [x] `make test-integration`
    - [x] `make test-all`
  - [ ] Run integration tests and verify 70%+ coverage
  - [ ] Add tests for setup package (internal/setup)

---

## Completed ✅ Since Last Update

### Project Specification
- [x] SPEC.md - Complete feature specification

### Interactive Setup
- [x] Phase 2: Setup command implementation
  - [x] `cmd/gitlab-ci-lint/main.go` - Integrated setup command
  - [x] `internal/config/writer.go` - Safe config writing
  - [x] `internal/setup/validator.go` - Token validation
  - [x] Add setup command to main binary
  - [x] Add survey/v2 dependency
  - [x] Remove separate setup binary

### Bug Fixes
- [x] Fix config directory path (~/.tools-config/)

### Release Automation
- [x] Phase 4: Semantic release setup
  - [x] `.releaserc.json` - Release configuration
  - [x] `.commitlintrc.json` - Commit message linting
  - [x] `.versionrc.json` - Changelog formatting
  - [x] `package.json` - Node.js dependencies for semantic-release
  - [x] `.husky/commit-msg` - Commit message hook
  - [x] `.husky/pre-commit` - Pre-commit checks
  - [x] Initialize Husky

### CI/CD
- [x] Phase 5: GitHub Actions workflows
  - [x] `.github/workflows/ci.yml` - Testing pipeline
  - [x] `.github/workflows/release.yml` - Release automation
  - [x] Coverage reporting to Codecov
  - [x] Multi-version Go testing (1.23, 1.24)

---

## Planned 📋

### Testing Suite (Continued)
- [ ] Complete integration test verification
- [ ] Achieve 70%+ overall test coverage

---

## Future 🚀

### Major Features
- [ ] Web interface for interactive validation
- [ ] VS Code extension integration
- [ ] Pre-commit hook integration
- [ ] Docker image for containerized usage
- [ ] Configuration diffing (compare before/after)
- [ ] Batch file validation
- [ ] Custom validation rules engine
- [ ] Plugin system for extensibility

### Authentication Enhancements
- [ ] Complete .netrc support implementation
- [ ] OAuth2 authentication
- [ ] CI job token validation in pipelines
- [ ] Token management (list, revoke, rotate)

### Output Enhancements
- [ ] HTML output format
- [ ] Markdown output format
- [ ] SARIF format for security integrations
- [ ] JUnit XML for CI test reporting

### GitLab Integration
- [ ] Pipeline integration (validate within running pipelines)
- [ ] Merge request comment integration
- [ ] GitLab file system operations (edit files in-place)
- [ ] Multi-project validation
- [ ] Pipeline history analysis

### Developer Experience
- [ ] Shell auto-completion scripts (bash, zsh, fish)
- [ ] Interactive TUI mode
- [ ] Configuration wizard
- [ ] Profiling and debugging mode
- [ ] Performance benchmarks

### Platform Support
- [ ] Homebrew tap
- [ ] Snap package
- [ ] Scoop (Windows)
- [ ] AUR (Arch Linux)
- [ ] Debian/RPM packages
- [ ] Nix package

### Monitoring & Analytics
- [ ] Anonymous usage statistics
- [ ] Error reporting integration
- [ ] Performance metrics collection
- [ ] Validation analytics dashboard

---

## Implementation Timeline

### Sprint 1: Foundation ✅
**Status**: Completed

- [x] Core validation system
- [x] Configuration management
- [x] Basic documentation

### Sprint 2: Quality ✅
**Status**: Completed

**Goals**:
- Interactive setup command
- Bug fixes
- Foundation for testing

**Tasks**:
- [x] Implement setup command
- [x] Fix config directory path
- [x] Create test infrastructure
- [x] Write unit tests (67% coverage achieved)
- [x] Write integration tests
- [x] Update Makefile with new test targets

### Sprint 3: Automation ✅
**Status**: Completed

**Goals**:
- Comprehensive test coverage (70%+)
- Automated releases
- CI/CD pipeline

**Tasks**:
- [x] Complete unit tests
- [x] Complete integration tests
- [x] Setup semantic-release
- [x] Create GitHub Actions workflows
- [x] Add pre-commit hooks
- [ ] Verify 70%+ overall coverage (in progress)

### Sprint 4: Documentation
**Status**: Planned

**Goals**:
- Complete documentation suite
- Developer guides
- Architecture documentation

**Tasks**:
- [ ] Create all docs/ content
- [ ] Update root documentation
- [ ] Add code documentation
- [ ] Create examples and tutorials

### Sprint 5: Polish
**Status**: Planned

**Goals**:
- Final testing and validation
- Release first stable version
- Community preparation

**Tasks**:
- [ ] Verify 70%+ test coverage
- [ ] Test release automation
- [ ] Final documentation review
- [ ] Prepare v1.0.0 release
- [ ] Create contribution guidelines
- [ ] Setup issue/PR templates

---

## Version Strategy

### v0.x.x - Development
- Current phase
- Feature development
- Breaking changes allowed
- No stability guarantees

### v1.0.0 - Stable Release (Planned)
**Prerequisites**:
- [ ] 70%+ test coverage
- [ ] Complete documentation
- [ ] Automated releases
- [ ] CI/CD pipeline
- [ ] Interactive setup
- [ ] Comprehensive testing

**Criteria**:
- All core features implemented
- Backward compatibility commitment
- Stable API
- Production-ready

### v1.x.x - Maintenance & Features
- Feature additions (backward compatible)
- Bug fixes
- Performance improvements
- Documentation updates

### v2.0.0 - Major Revision (Future)
- Breaking changes only
- Significant architectural improvements
- Deprecated feature removal

---

## Contributing

We welcome contributions! See `CONTRIBUTING.md` (to be created) for guidelines.

### Priority Areas for Contribution
1. **Tests**: We need comprehensive test coverage
2. **Documentation**: Help improve guides and examples
3. **Bug Reports**: Report issues with detailed reproduction steps
4. **Feature Requests**: Propose new features with use cases
5. **Code Review**: Review and improve existing code

---

## Dependencies & Tech Debt

### Technical Debt
- [ ] .netrc implementation incomplete
- [ ] Error messages could be more user-friendly
- [ ] Limited test coverage currently
- [ ] No integration tests
- [ ] Manual release process

### Dependencies to Monitor
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- Future: `AlecAivazis/survey/v2` - Interactive prompts

---

## Performance Targets

### Current Performance
- Local validation: < 100ms for typical files
- API validation: 1-3 seconds (network dependent)

### Goals
- Local validation: < 50ms
- API validation: Optimize with retries, connection pooling
- Startup time: < 10ms
- Memory footprint: < 50MB

---

## Success Metrics

### Adoption
- Number of downloads
- GitHub stars
- Active contributors
- Issues/PRs activity

### Quality
- Test coverage percentage
- Bug report rate
- Resolution time
- User satisfaction

### Stability
- Crash rate
- Error rate
- API compatibility
- Breaking changes per release

---

## Notes

- This roadmap is a living document and will be updated as priorities shift
- Items may move between sprints based on community feedback and needs
- Contributions that align with roadmap priorities are especially welcome
- For questions or suggestions, please open a GitHub issue

---

**Last Updated**: 2025-02-01
**Next Review**: After Sprint 2 completion
