# GitHub Wiki Migration Plan

This document outlines the plan for migrating documentation from markdown files to GitHub Wiki.

## Current Documentation Structure

### Files to Keep (Essential)
- `README.md` - Project overview, quick start, installation
- `LICENSE` - License file
- `CHANGELOG.md` - Auto-generated changelog
- `CONTRIBUTING.md` - Contribution guidelines
- `CODE_OF_CONDUCT.md` - Community guidelines
- `SECURITY.md` - Security policy
- `SUPPORT.md` - Support guide
- `ROADMAP.md` - Development roadmap

### Files to Migrate to Wiki

The following detailed documentation files should be moved to GitHub Wiki:

#### Configuration & Setup
- `CONFIG.md` → Wiki: [Configuration Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Configuration-Guide)
- `INSTALL.md` → Wiki: [Installation Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Installation-Guide)

#### Architecture
- `SPEC.md` → Wiki: [Technical Specification](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Technical-Specification)
- `IMPLEMENTATION_SUMMARY.md` → Wiki: [Implementation Status](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Implementation-Status)
- `docs/architecture/overview.md` → Wiki: [Architecture Overview](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Architecture-Overview)

#### Guides
- `docs/guides/quick-start.md` → Wiki: [Quick Start Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Quick-Start)
- `docs/guides/troubleshooting.md` → Wiki: [Troubleshooting](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Troubleshooting)

#### Examples
- `docs/examples/basic-usage.md` → Wiki: [Basic Usage Examples](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Basic-Usage)
- `docs/examples/ci-integration.md` → Wiki: [CI/CD Integration](https://github.com/InkyQuill/gitlab-ci-lint/wiki/CI-CD-Integration)

#### API Reference
- `docs/api/reference.md` → Wiki: [API Reference](https://github.com/InkyQuill/gitlab-ci-lint/wiki/API-Reference)

## Migration Steps

### 1. Prepare Wiki Repository

```bash
# Clone the wiki repository
git clone https://github.com/InkyQuill/gitlab-ci-lint.wiki.git
cd gitlab-ci-lint.wiki

# Create structure
mkdir -p configuration architecture guides examples api
```

### 2. Create Wiki Pages

For each document:

```bash
# Create page from markdown file
cp CONFIG.md "Configuration-Guide.md"
cp INSTALL.md "Installation-Guide.md"
cp SPEC.md "Technical-Specification.md"
# ... etc for all files
```

### 3. Update Internal Links

Replace internal links in wiki pages:
- `CONFIG.md` → `Configuration-Guide`
- `docs/architecture/overview.md` → `Architecture-Overview`
- etc.

Example:
```markdown
<!-- Before -->
See [CONFIG.md](CONFIG.md) for details.

<!-- After -->
See [Configuration Guide](Configuration-Guide) for details.
```

### 4. Create Wiki Home Page

Create `Home.md` as the wiki landing page:

```markdown
# GitLab CI Lint Wiki

Welcome to the GitLab CI Lint documentation wiki!

## Getting Started

- **[Quick Start](Quick-Start)** - Get started in 5 minutes
- **[Installation Guide](Installation-Guide)** - Detailed installation instructions
- **[Configuration Guide](Configuration-Guide)** - Configure gitlab-ci-lint

## Documentation

### User Guides
- [Basic Usage](Basic-Usage)
- [CI/CD Integration](CI-CD-Integration)
- [Troubleshooting](Troubleshooting)

### Technical Documentation
- [Technical Specification](Technical-Specification)
- [Architecture Overview](Architecture-Overview)
- [API Reference](API-Reference)

### Development
- [Implementation Status](Implementation-Status)
- [Contributing](https://github.com/InkyQuill/gitlab-ci-lint/blob/main/CONTRIBUTING.md)
- [Roadmap](https://github.com/InkyQuill/gitlab-ci-lint/blob/main/ROADMAP.md)

## Quick Links

- **[GitHub Repository](https://github.com/InkyQuill/gitlab-ci-lint)**
- **[Issues](https://github.com/InkyQuill/gitlab-ci-lint/issues)**
- **[Discussions](https://github.com/InkyQuill/gitlab-ci-lint/discussions)**
- **[Releases](https://github.com/InkyQuill/gitlab-ci-lint/releases)**
```

### 5. Update README.md

Update references in README to point to wiki:

```markdown
## Documentation

- 📚 [Wiki](https://github.com/InkyQuill/gitlab-ci-lint/wiki) - Full documentation
- 🚀 [Quick Start](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Quick-Start) - Get started in minutes
- 🔧 [Configuration Guide](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Configuration-Guide)
- 🏗️ [Architecture](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Architecture-Overview)
- 📖 [Examples](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Basic-Usage)
- 🐛 [Troubleshooting](https://github.com/InkyQuill/gitlab-ci-lint/wiki/Troubleshooting)
```

### 6. Commit and Push

```bash
git add .
git commit -m "docs: migrate documentation to GitHub Wiki"
git push origin main
```

### 7. Clean Up (Optional)

After migration and verification, you can remove the migrated docs from the repository:

```bash
# Keep only essential files
rm -rf docs/
rm CONFIG.md INSTALL.md SPEC.md IMPLEMENTATION_SUMMARY.md
git commit -am "chore: remove docs migrated to wiki"
```

## Post-Migration Checklist

- [ ] All wiki pages created
- [ ] Internal links updated
- [ ] README.md links updated to point to wiki
- [ ] Home page created with navigation
- [ ] Test all links work correctly
- [ ] Update .github/PULL_REQUEST_TEMPLATE.md to reference wiki
- [ ] Announce migration in an issue/discussion

## Benefits of Wiki Migration

1. **Separation of Concerns** - README focuses on essentials, wiki for detailed docs
2. **Easier Updates** - Non-committers can suggest wiki edits via pull requests to wiki repo
3. **Better Organization** - Wiki provides better navigation and search
4. **Version Independence** - Docs can be updated without touching main codebase
5. **Community Contributions** - Easier for community to improve documentation

## Maintenance

### Keeping Wiki in Sync

When updating code that affects documentation:
1. Update relevant wiki pages
2. Cross-link commit messages to wiki pages
3. Add "Documentation" label to doc-only PRs

### Wiki Best Practices

- Keep pages focused on single topics
- Use clear, descriptive titles
- Include table of contents for long pages
- Add examples and screenshots
- Keep content up-to-date with code changes
- Use consistent formatting and style

## Migration Timeline

- [ ] Phase 1: Create wiki structure and migrate core docs (Week 1)
- [ ] Phase 2: Update all links and navigation (Week 1)
- [ ] Phase 3: Remove old doc files from repo (Week 2)
- [ ] Phase 4: Community announcement and feedback (Week 2)
