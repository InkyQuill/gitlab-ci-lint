# GitLab CI Lint - Roadmap

## Overview

GitLab CI Lint is a Go CLI tool that validates `.gitlab-ci.yml` files in two stages:
1. **Local validation**: Fast YAML syntax checking (no API required)
2. **API validation**: GitLab instance validation via GitLab API (optional)

Current status: **v2.0** - Multi-instance support with auto-detection

---

## Completed Features ✅

### v1.0 - Core Functionality
- [x] Local YAML syntax validation
- [x] GitLab API validation (single instance)
- [x] Single binary distribution
- [x] Interactive setup wizard
- [x] Multi-format output (text/JSON/YAML)
- [x] Configuration file support (`~/.tools-config/.gitlab-ci-lint/config.yaml`)
- [x] Environment variable overrides (`GCL_*`)
- [x] File discovery (auto, recursive, explicit)
- [x] Custom GitLab instance support
- [x] CLI flags override system
- [x] Comprehensive unit tests (~65% coverage)

### v2.0 - Multi-Instance with Auto-Detection (Current)
- [x] **Multi-instance support** - configure multiple GitLab instances
- [x] **Automatic instance detection** - detects instance from `.git/config` origin URL
- [x] **Per-file routing** - each CI file validates against its GitLab instance
- [x] **Setup wizard enhancement** - manage multiple instances interactively
- [x] **--list-instances flag** - show all configured instances
- [x] **Smart API validation** - gracefully skips API validation when:
  - File is outside git repository
  - No token configured for detected instance
  - Stdin input (no .git/config for detection)
- [x] **SSH URL support** - handles `git@gitlab.com:...` format
- [x] **Submodule detection** - correctly handles git submodules
- [x] **Worktree support** - works with git worktrees
- [x] **Legacy config migration** - auto-migrates v1.0 configs to v2.0 format
- [x] **Default instance removal** - clean design, no fallback complexity
- [x] **Comprehensive testing** - 70+ new unit tests
- [x] **golangci-lint passing** - code quality verified

---

## Planned Features 📋

### v2.1 - Polish and User Experience
**Status:** Pending
**Priority:** HIGH
**Effort:** ~3-5 days

#### User Experience
- [ ] Better error messages for common issues
  - "No token configured for instance X" → suggest running `gitlab-ci-lint setup`
  - "File outside git repository" → suggest creating git repo
  - "Invalid .git/config" → suggest `git init`
- [ ] Progress indicators for long operations
- [ ] Colored output for better readability (enhance existing color support)
- [ ] Machine-readable output format for JSON/YAML modes

#### Documentation
- [ ] Complete user guide with examples
- [ ] Multi-instance setup tutorial
- [ ] Troubleshooting guide
- [ ] Migration guide from v1.0 to v2.0
- [ ] API documentation for tool integration
- [ ] Contributing guide

#### CLI Enhancements
- [ ] `--instance` flag override (force specific instance for all files)
- [ ] `--dry-run` flag - show what would be validated without making API calls
- [ ] Bash completion support
- [ ] Man page generation

### v2.2 - Performance and Reliability
**Status:** Planned
**Priority:** MEDIUM
**Effort:** ~3-4 days

#### Performance
- [ ] Parallel validation for multiple files
- [ ] Connection pooling for GitLab API
- [ ] Cached validation results (with TTL)
- [ ] Rate limiting awareness
- [ ] Timeout per-instance configuration

#### Reliability
- [ ] Retry logic for transient API failures
- [ ] Graceful degradation when GitLab API is slow
- [ ] Better error recovery (network timeouts, 503 errors)
- [ ] Validation result caching
- [ ] Health check mode for CI/CD integration

### v2.3 - Integration Features
**Status:** Planned
**Priority:** MEDIUM
**Effort:** ~4-5 days

#### CI/CD Integration
- [ ] Pre-commit hook integration
- [ ] CI pipeline configuration validation
- [ ] Merge request API validation (requires MR number)
- [ ] Pipeline YAML validation (requires `gitlab-ci.yml` linting API)
- [ ] Badge generation for validation status

#### Advanced Features
- [ ] Config profiles (dev/staging/prod)
- [ ] Project-specific validation settings
- [ ] Custom rules engine for YAML validation
- [ ] Webhook integration for validation events

### v3.0 - Enterprise Features (Future)
**Status:** Exploratory
**Priority:** LOW
**Effort:** TBD

- [ ] Multi-tenancy support
- [] Centralized configuration server
- [ ] Audit logging
- [ ] Team management
- [] Usage analytics and reporting

---

## Testing Roadmap

### Unit Tests (Current: ~70% coverage)

#### Completed ✅
- [x] **internal/gitlab/registry_test.go** (18 tests)
  - ClientRegistry creation and management
  - Client retrieval by URL and name
  - Token validation
  - Instance listing

- [x] **internal/config/migrate_test.go** (22 tests)
  - Legacy config migration
  - Instance name extraction
  - Validation logic
  - Find operations

- [x] **internal/gitlab/detect_enhanced_test.go** (30 tests)
  - Instance URL extraction (HTTPS, HTTP, SSH)
  - Git repository detection
  - Submodule and worktree handling
  - File discovery edge cases

#### Planned 📋
- [ ] **internal/validator/**_test.go** - Expand coverage to 80%+
  - API validator error handling
  - Edge cases for YAML parsing
  - Error message formatting

- [ ] **internal/discover/**_test.go** - Add tests for discovery
  - Recursive directory scanning
  - Ignore pattern handling
  - File discovery edge cases

### Integration Tests

#### Created but pending implementation 📝
**File:** `tests/integration/multi_instance_test.go`

| Test | Description | Priority |
|------|-------------|----------|
| TestMultiInstance_ListInstances | `--list-instances` flag | HIGH |
| TestMultiInstance_AutoDetectionPerFile | Auto-detection for each file | HIGH |
| TestMultiInstance_GitLabComFile | File with gitlab.com | HIGH |
| TestMultiInstance_CustomInstanceFile | File with custom instance | HIGH |
| TestMultiInstance_MixedInstances | Mixed instances in one run | HIGH |
| TestMultiInstance_MissingTokenForInstance | No token for detected instance | MEDIUM |
| TestMultiInstance_OutsideGitRepo_SkipsAPI | File outside git repo | MEDIUM |
| TestMultiInstance_LegacyConfigMigration | Migration from v1.0 config | MEDIUM |
| TestMultiInstance_Stdin_SkipsAPI | stdin input handling | MEDIUM |
| TestMultiInstance_BackwardCompatibility | Old configs still work | HIGH |
| TestMultiInstance_SubmoduleDetection | Submodule scenarios | MEDIUM |
| TestMultiInstance_WorktreeDetection | Worktree scenarios | MEDIUM |
| TestMultiInstance_DepthLimit | --max-depth flag | LOW |
| TestMultiInstance_IgnorePatterns | Ignore patterns work | LOW |
| TestMultiInstance_ErrorHandling | Error messages are clear | MEDIUM |
| TestMultiInstance_Performance | Large file sets performance | LOW |
| TestMultiInstance_Concurrency | Parallel file processing | LOW |

**Total:** 15 integration tests

**Implementation effort:** ~2-3 days

### Test Coverage Goals

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| internal/gitlab | 75% | 80% | HIGH |
| internal/config | 75% | 80% | HIGH |
| internal/validator | 65% | 75% | MEDIUM |
| internal/discover | 60% | 70% | MEDIUM |
| internal/output | 70% | 75% | LOW |
| **Overall** | **70%** | **75%** | - |

---

## Bug Fixes 🔧

### Fixed in v2.0
- [x] Default instance fallback logic removed (complex and error-prone)
- [x] Instance detection not working for files outside git repositories
- [ [FS-22] Cannot determine project path for files in subdirectories
- [x] SSH URL parsing failures
- [x] Submodule .git/config not being read correctly
- [x] Worktree configuration not detected
- [x] Multiple instances marked as default causing confusion

### Known Limitations ⚠️

### v2.0 Limitations
1. **Files outside git repositories** - API validation is skipped
   - **Workaround:** Run `git init` and configure remote, or use `--skip-api`
   - **Reason:** No way to determine which GitLab instance to use

2. **No auto-detection for stdin input** - API validation is skipped
   - **Workaround:** Provide explicit file paths instead of stdin
   - **Reason:** Stdin has no file path for `.git/config` detection

3. **SSH URLs in legacy migration** - May not normalize correctly
   - **Workaround:** Use HTTPS URLs in config or re-run setup
   - **Reason:** `NormalizeInstanceURL()` doesn't handle SSH format

4. **No default instance fallback** - Intentional design decision
   - **Workaround:** Configure token for detected instance
   - **Reason:** Simpler, more predictable behavior

---

## Deprecations ⚠️

### v2.0 Deprecations
- **`InstanceConfig.Default` field** - Removed (use auto-detection instead)
- **`GitLabConfig.Default` field** - Removed (use auto-detection instead)
- **`GetDefaultClient()` method** - Removed (use `GetClient(instanceURL)` instead)
- **`GetDefaultInstance()` method** - Removed (use auto-detection instead)
- **`--instance` flag fallback** - Now requires explicit instance or auto-detection

### Migration from v1.0 to v2.0

**Breaking change:** Default instance concept removed

**Old behavior (v1.0):**
```yaml
# Config file ~/.tool-configs/.gitlab-ci-lint/config.yaml
gitlab:
  instance: "https://gitlab.com"
  default: "gitlab.com"  # REMOVED in v2.0
auth:
  token: "glpat-..."
```

**New behavior (v2.0):**
```yaml
# Config file ~/.tools-config/.gitlab-ci-lint/config.yaml
gitlab:
  instances:
    - name: gitlab.com
      url: "https://gitlab.com"
      token: "glpat-..."
```

**Auto-detection replaces defaults:**
- v2.0: Detects instance from `.git/config` origin URL
- v1.0: Fallback to configured default instance

---

## Technical Debt 💳

### Low Priority
- [ ] Refactor large functions in `cmd/gitlab-ci-lint/main.go`
- [ ] Extract validation result formatting to separate package
- [ ] Add benchmarks for performance testing
- [ ] Consider adding fuzzy matching for instance URL detection

### Medium Priority
- [ ] Improve test coverage for error paths
- [ ] Add integration tests for setup wizard
- [ ] Document public API with GoDoc
- [ ] Add example configs for common scenarios

---

## Contribution Guidelines

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Code Quality Standards
- All code must pass `golangci-lint` (see `.golangci.yml`)
- Unit tests required for new features
- Integration tests for multi-file workflows
- Documentation updates for user-facing changes
- Conventional commit messages (see `commitlint` config)

### Development Workflow
```bash
# 1. Make changes
git checkout -b feature/my-feature

# 2. Run tests
make test-unit
make test-integration

# 3. Run linter
make lint

# 4. Build
make build

# 5. Test manually
./build/gitlab-ci-lint --help

# 6. Commit (conventional commits)
git add .
git commit -m "feat: add my feature"

# 7. Push
git push
```

---

## Version History

### v2.0.0 (Current)
- Multi-instance support
- Automatic instance detection from `.git/config`
- Removed default instance concept
- 70+ new unit tests
- Comprehensive documentation updates

### v1.0.0
- Initial release
- Single instance support
- Local YAML validation
- GitLab API validation
- Interactive setup wizard
- File discovery (auto, recursive)
- Multiple output formats

---

## See Also

- [README.md](README.md) - User documentation
- [CLAUDE.md](CLAUDE.md) - Development guide for Claude Code/AI assistants
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [CHANGELOG.md](CHANGELOG.md) - Version history
