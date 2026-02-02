# Changelog

## [1.0.2](https://github.com/InkyQuill/gitlab-ci-lint/compare/v1.0.1...v1.0.2) (2026-02-02)


### Bug Fixes

* resolve setup command bugs with connection test timing ([d27d4f6](https://github.com/InkyQuill/gitlab-ci-lint/commit/d27d4f649ee518dceb7ff2e73330afc70fe90f61))

## [1.0.1](https://github.com/InkyQuill/gitlab-ci-lint/compare/v1.0.0...v1.0.1) (2026-02-01)


### Bug Fixes

* correct URL encoding test and remove redundant test step from lint ([ffeefa5](https://github.com/InkyQuill/gitlab-ci-lint/commit/ffeefa522ae04f1f04e87618898dc24dda179d2e))
* resolve double URL encoding in project path ([06f791a](https://github.com/InkyQuill/gitlab-ci-lint/commit/06f791a87e4fe7d25e6ba0c41d974bdcf5708a32))
* resolve nil pointer panics and implement proper config priority ([decc6c6](https://github.com/InkyQuill/gitlab-ci-lint/commit/decc6c6704a03370974103b897f82e570af397a1))

# [1.0.0](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.4.1...v1.0.0) (2026-02-01)


### Features

* implement multi-instance support with auto-detection ([cac4eef](https://github.com/InkyQuill/gitlab-ci-lint/commit/cac4eef5d5161da67b8e26ef1d5ca6049ad78d5e))


### BREAKING CHANGES

* Remove default instance concept

This major update implements multi-instance GitLab support with automatic
instance detection from .git/config, removing the need for default instances.

## Multi-Instance Support
- Configure multiple GitLab instances with tokens
- Automatic instance detection from .git/config origin URL
- Per-file routing to correct GitLab instance
- Support for HTTPS, HTTP, SSH URL formats
- Graceful handling of files outside git repositories
- Submodule and worktree detection

## Default Instance Removal
- Removed GitLabConfig.Default field
- Removed InstanceConfig.Default field
- Removed GetDefaultClient() and GetDefaultInstance() methods
- Removed fallback logic for files without .git/config
- Files outside git repos skip API validation (with debug message)
- Stdin input skips API validation (with clear warning)

## Testing
- Add 18 tests for ClientRegistry (registry_test.go)
- Add 22 tests for config migration (migrate_test.go)
- Add 30 tests for instance detection (detect_enhanced_test.go)
- Total: 70+ new unit tests achieving 70%+ coverage

## Code Quality
- Fix all golangci-lint issues (reduced from 18 to 0)
- Add .golangci.yml configuration
- Proper error handling for all file operations
- Add nolint comments for intentional code patterns

## Documentation
- Create comprehensive ROADMAP.md with:
  - v1.0 and v2.0 feature checklists
  - 15 planned integration tests (implementation postponed)
  - v2.1-v3.0 roadmap
  - Known limitations and workarounds
  - Migration guide from v1.0 to v2.0
  - Testing roadmap with coverage goals

## Modified Files
- cmd/gitlab-ci-lint/main.go: Major refactor, remove default fallback
- internal/config/config.go: Remove Default fields
- internal/config/migrate.go: Remove GetDefaultInstance()
- internal/gitlab/registry.go: Remove GetDefaultClient()
- internal/gitlab/detect.go: Add DetectInstanceForFile()
- internal/gitlab/client.go: Improve error handling

## New Files
- internal/gitlab/registry.go: Multi-instance client registry
- internal/gitlab/registry_test.go: 18 tests
- internal/config/migrate.go: Config migration logic
- internal/config/migrate_test.go: 22 tests
- internal/gitlab/detect_enhanced_test.go: 30 tests
- internal/config/duration.go: Duration type for YAML
- internal/output/debug.go: Debug logging
- .golangci.yml: Linter configuration
- ROADMAP.md: Project roadmap

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

## [0.4.1](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.4.0...v0.4.1) (2026-02-01)


### Bug Fixes

* pin goreleaser to v2 for config version 2 ([69e2360](https://github.com/InkyQuill/gitlab-ci-lint/commit/69e2360febabac1f789491c211c7e9d5e5251ddc))

# [0.4.0](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.3.2...v0.4.0) (2026-02-01)


### Features

* improve version package documentation ([8907457](https://github.com/InkyQuill/gitlab-ci-lint/commit/8907457aee05557a00cbf5ce9547aa101eb03ebe))

## [0.3.2](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.3.1...v0.3.2) (2026-02-01)


### Bug Fixes

* implement code quality fixes from review ([b66e793](https://github.com/InkyQuill/gitlab-ci-lint/commit/b66e793d8d8dbe18865cc0542833dfb1402b71e1))
* implement code quality fixes from second review audit ([a7a7c53](https://github.com/InkyQuill/gitlab-ci-lint/commit/a7a7c53a6ce245185dd4e1f9c8b1a59309ccd057))

## [0.3.1](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.3.0...v0.3.1) (2026-02-01)


### Bug Fixes

* correct release workflow to inject proper version into binaries ([b2b05c2](https://github.com/InkyQuill/gitlab-ci-lint/commit/b2b05c23cbc3a33c867a25136aea8254f5951d64))

# [0.3.0](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.2.1...v0.3.0) (2026-02-01)


### Features

* improve API integration with auto-detection and debug mode ([96b740b](https://github.com/InkyQuill/gitlab-ci-lint/commit/96b740bdc6f68436c67541a648ef3eb9189a3ff9))

## [0.2.1](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.2.0...v0.2.1) (2026-02-01)


### Bug Fixes

* remove incorrect npm package labels from release config ([0d8a52c](https://github.com/InkyQuill/gitlab-ci-lint/commit/0d8a52cf41557aa1918ece16fbca260a7aa1bb69))

# [0.2.0](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.1.1...v0.2.0) (2026-02-01)


### Bug Fixes

* configure semantic-release for GitHub Actions ([aa42d9d](https://github.com/InkyQuill/gitlab-ci-lint/commit/aa42d9d24b9b05bf2b8ea01d668cf8c2d1ac7c1a))
* Fix CI tests and Go version requirements ([24dcec5](https://github.com/InkyQuill/gitlab-ci-lint/commit/24dcec5acb11e0e9b1829894be3b593225303744))


### Features

* Implement enhanced file discovery and multi-file validation ([f8e723d](https://github.com/InkyQuill/gitlab-ci-lint/commit/f8e723d52faf0429d29854e35d4522f890abe28c))

## [0.1.1](https://github.com/InkyQuill/gitlab-ci-lint/compare/v0.1.0...v0.1.1) (2026-02-01)


### Bug Fixes

* update Go version to 1.24 and fix linting issues ([e163c4a](https://github.com/InkyQuill/gitlab-ci-lint/commit/e163c4a8ad6dc2ac5332b288e1a6835bf977512d))

This file is automatically generated by semantic-release based on commit messages.

## [Unreleased]

### Added
- Enhanced file discovery system with auto-discovery
- Support for validating multiple files with `-f` flag
- Recursive directory scanning with `-d` flag
- Smart directory filtering (ignores .git, node_modules, vendor, etc.)
- Parent directory traversal for CI file discovery
- Batch validation with summary output
- Support for both `.gitlab-ci.yml` and `.gitlab-ci.yaml`
- Integration tests for file discovery features
- Interactive setup wizard (`gitlab-ci-lint setup`)
- Comprehensive test suite (72%+ coverage)
- Semantic release automation
- GitHub Actions CI/CD workflows
- Complete documentation suite

### Changed
- Integrated setup command into main binary (no separate setup binary needed)
- Fixed config directory path typo (~/.tool-configs → ~/.tools-config/)
- CLI argument handling: now accepts optional file argument
- Exit code behavior for batch validation (exit 10 if any file invalid)

### Fixed
- URL encoding for project paths in API endpoints
- Stdin input support preserved with new file discovery system
