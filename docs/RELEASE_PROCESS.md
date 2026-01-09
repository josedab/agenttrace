# Release Process

This document describes how to create releases for AgentTrace.

## Overview

AgentTrace uses semantic versioning and automated release workflows. When a new version tag is pushed, GitHub Actions automatically:

1. Creates a GitHub Release with auto-generated changelog
2. Builds and pushes Docker images to Docker Hub and GHCR
3. Publishes Python SDK to PyPI
4. Publishes TypeScript SDK to npm
5. Builds CLI binaries for all platforms

## Version Numbering

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (X.0.0): Breaking API changes
- **MINOR** (0.X.0): New features, backwards-compatible
- **PATCH** (0.0.X): Bug fixes, documentation updates

### Pre-release Versions

For pre-releases, use suffixes:
- Alpha: `v1.0.0-alpha.1`
- Beta: `v1.0.0-beta.1`
- Release Candidate: `v1.0.0-rc.1`

## Creating a Release

### Prerequisites

- [ ] All CI checks passing on `main`
- [ ] CHANGELOG.md updated with release notes
- [ ] Version numbers updated in package files (if not automatic)

### Step 1: Update the Changelog

Move items from `[Unreleased]` to a new version section:

```markdown
## [Unreleased]

## [0.2.0] - 2024-02-01

### Added
- New feature X (#123)
- New feature Y (#124)

### Fixed
- Bug in Z (#125)
```

Update the comparison links at the bottom:

```markdown
[Unreleased]: https://github.com/agenttrace/agenttrace/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/agenttrace/agenttrace/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/agenttrace/agenttrace/releases/tag/v0.1.0
```

### Step 2: Update Version Numbers

Update version in these files:

```bash
# API version
api/cmd/server/main.go  # const Version = "0.2.0"

# Python SDK
sdk/python/pyproject.toml  # version = "0.2.0"

# TypeScript SDK
sdk/typescript/package.json  # "version": "0.2.0"

# Go SDK
sdk/go/version.go  # const Version = "0.2.0"

# CLI
sdk/cli/main.go  # var Version = "0.2.0"
```

### Step 3: Create Release Commit

```bash
git add -A
git commit -m "chore: release v0.2.0"
git push origin main
```

### Step 4: Create and Push Tag

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

### Step 5: Verify Release

1. Check [GitHub Actions](https://github.com/agenttrace/agenttrace/actions) for release workflow
2. Verify artifacts:
   - [ ] GitHub Release created with changelog
   - [ ] Docker images on [Docker Hub](https://hub.docker.com/u/agenttrace)
   - [ ] Python SDK on [PyPI](https://pypi.org/project/agenttrace/)
   - [ ] TypeScript SDK on [npm](https://www.npmjs.com/package/agenttrace)
   - [ ] CLI binaries attached to release

## Automated Changelog Generation

The release workflow uses [release-changelog-builder-action](https://github.com/mikepenz/release-changelog-builder-action) to generate changelogs from PR titles and labels.

### Label Categories

| Label | Category |
|-------|----------|
| `breaking-change`, `breaking` | Breaking Changes |
| `feature`, `enhancement` | New Features |
| `bug`, `fix`, `bugfix` | Bug Fixes |
| `performance`, `optimization` | Performance |
| `security` | Security |
| `documentation`, `docs` | Documentation |
| `dependencies`, `deps` | Dependencies |

### Conventional Commits

PR titles following [Conventional Commits](https://www.conventionalcommits.org/) are auto-labeled:

| Prefix | Label |
|--------|-------|
| `feat:` | feature |
| `fix:` | fix |
| `docs:` | documentation |
| `perf:` | performance |
| `security:` | security |
| `!:` or `BREAKING CHANGE` | breaking-change |

### Skipping Changelog

To exclude a PR from the changelog, add the `skip-changelog` label.

## Hotfix Releases

For urgent fixes to production:

1. Create branch from the release tag:
   ```bash
   git checkout -b hotfix/v0.2.1 v0.2.0
   ```

2. Apply fix and commit

3. Update CHANGELOG.md with hotfix notes

4. Tag and push:
   ```bash
   git tag -a v0.2.1 -m "Hotfix v0.2.1"
   git push origin v0.2.1
   ```

5. Cherry-pick fix to `main`:
   ```bash
   git checkout main
   git cherry-pick <commit-sha>
   ```

## Release Checklist

### Before Release
- [ ] All tests passing
- [ ] No critical security vulnerabilities
- [ ] Documentation updated for new features
- [ ] Migration guide written (if breaking changes)
- [ ] CHANGELOG.md updated

### After Release
- [ ] Verify all artifacts published
- [ ] Test installation from package managers
- [ ] Update website/docs with new version
- [ ] Announce on Discord/Twitter
- [ ] Close related GitHub milestones

## Rollback Procedure

If a release has critical issues:

1. **Docker**: Push previous version tag as `latest`
   ```bash
   docker pull agenttrace/api:0.1.0
   docker tag agenttrace/api:0.1.0 agenttrace/api:latest
   docker push agenttrace/api:latest
   ```

2. **PyPI**: Yank the broken version (cannot delete)
   ```bash
   pip install twine
   twine upload --skip-existing  # Re-upload previous
   ```

3. **npm**: Deprecate the broken version
   ```bash
   npm deprecate agenttrace@0.2.0 "Critical bug, use 0.1.0"
   ```

4. Create hotfix release (see above)

## SDK Version Compatibility

| AgentTrace Version | Python SDK | TypeScript SDK | Go SDK |
|--------------------|------------|----------------|--------|
| 0.2.x | 0.2.x | 0.2.x | 0.2.x |
| 0.1.x | 0.1.x | 0.1.x | 0.1.x |

SDKs are versioned in lockstep with the main platform.

## Contact

For release issues, contact:
- Release Manager: @maintainer
- Discord: #releases channel
