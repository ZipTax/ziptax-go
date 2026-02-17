# GitHub Actions Workflows

This document describes the automated workflows configured for this project.

## Workflows Overview

### 1. Version Bump Check (`version-bump-check.yml`)

**Trigger:** Pull requests to `main` branch

**Purpose:** Ensures semantic versioning is properly maintained across all PRs.

**What it does:**
- ✅ Verifies version was bumped in `version.go`
- ✅ Validates semantic version format (`MAJOR.MINOR.PATCH`)
- ✅ Ensures version increased (not decreased or unchanged)
- ✅ Determines bump type (major/minor/patch)
- ✅ Posts informative comment on PR with version details
- ✅ Can be skipped with `skip-version-check` label

**Example Success Output:**
```
🚀 Version Bump Detected

Type: MINOR (New Features)
From: 1.0.0 → To: 1.1.0

✅ Semantic version has been properly bumped.
```

**Example Failure Output:**
```
❌ Version must be bumped for this PR!

Current version: 1.0.0
PR version: 1.0.0

Semantic Versioning Guidelines:
  • MAJOR (X.0.0): Breaking changes
  • MINOR (0.X.0): New features (backward compatible)
  • PATCH (0.0.X): Bug fixes (backward compatible)
```

### 2. Release (`release.yml`)

**Trigger:** Push of version tags (e.g., `v1.0.0`, `v1.2.3-beta.1`)

**Purpose:** Automatically creates GitHub releases when version tags are pushed.

**What it does:**
- ✅ Verifies tag version matches `version.go`
- ✅ Runs all tests
- ✅ Builds the project
- ✅ Generates release notes from commits
- ✅ Creates GitHub release (marked as pre-release for versions containing `-`)
- ✅ Links to full changelog

**Tag Format:**
- Stable: `v1.0.0`, `v2.3.4`
- Pre-release: `v2.0.0-beta.1`, `v1.5.0-rc.2`

**How to trigger:**
```bash
# After merging PR with version bump
git tag v1.1.0
git push origin v1.1.0
```

## Workflow Files

| File | Description |
|------|-------------|
| `.github/workflows/version-bump-check.yml` | PR version validation |
| `.github/workflows/release.yml` | Automated releases |
| `.github/VERSION_BUMP_GUIDE.md` | Detailed versioning guide |
| `version.go` | Version constant |

## Common Scenarios

### Scenario 1: Regular Feature PR

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes
# ... code changes ...

# 3. Bump version (MINOR for new features)
# Edit version.go: "1.0.0" → "1.1.0"

# 4. Commit and push
git add version.go
git commit -m "chore: bump version to 1.1.0"
git push

# 5. Create PR
# Version bump check will automatically run and pass
```

### Scenario 2: Bug Fix PR

```bash
# 1. Create bugfix branch
git checkout -b fix/issue-123

# 2. Fix the bug
# ... code changes ...

# 3. Bump version (PATCH for bug fixes)
# Edit version.go: "1.1.0" → "1.1.1"

# 4. Commit and push
git add version.go
git commit -m "chore: bump version to 1.1.1"
git push

# 5. Create PR
# Version bump check will automatically run and pass
```

### Scenario 3: Documentation-Only PR

```bash
# 1. Create docs branch
git checkout -b docs/update-readme

# 2. Update documentation
# ... doc changes ...

# 3. Add skip label to PR
# In GitHub UI: Add "skip-version-check" label

# 4. Push without version bump
git push

# 5. Create PR
# Version bump check will be skipped
```

### Scenario 4: Creating a Release

```bash
# 1. Ensure version was bumped in merged PR
# 2. Checkout main and pull latest
git checkout main
git pull origin main

# 3. Create and push tag
git tag v1.1.0
git push origin v1.1.0

# 4. GitHub Action automatically creates release
# View at: https://github.com/your-org/ziptax-go/releases
```

### Scenario 5: Pre-release (Beta/RC)

```bash
# 1. Update version.go with pre-release suffix
# "1.1.0" → "2.0.0-beta.1"

# 2. Create PR and merge

# 3. Create pre-release tag
git tag v2.0.0-beta.1
git push origin v2.0.0-beta.1

# 4. GitHub Action creates pre-release
```

## Skipping Version Check

Use the `skip-version-check` label for:
- Documentation-only changes
- CI/CD configuration updates
- README updates
- Comment/whitespace changes
- Non-code changes

**To create the label:**
1. Go to repo → Issues → Labels
2. Click "New label"
3. Name: `skip-version-check`
4. Description: "Skip semantic version bump requirement"
5. Color: `#d4c5f9`

## Troubleshooting

### "Version format is invalid"

**Cause:** Version doesn't follow semantic versioning format.

**Fix:** Update `version.go` to match `X.Y.Z` format:
```go
// ❌ Wrong
const Version = "1.1"
const Version = "v1.1.0"  // No 'v' prefix in version.go

// ✅ Correct
const Version = "1.1.0"
const Version = "2.0.0-beta.1"  // Pre-release OK
```

### "Version was not changed"

**Cause:** Forgot to update `version.go`.

**Fix:** Edit `version.go` and commit:
```bash
# Edit version.go
vim version.go

# Commit
git add version.go
git commit -m "chore: bump version to X.Y.Z"
git push
```

### "Version was decreased"

**Cause:** New version is lower than base branch version.

**Fix:** Ensure you're increasing the version:
```go
// Base branch: 1.5.3
const Version = "1.6.0"  // ✅ Correct
const Version = "1.5.2"  // ❌ Wrong (decreased)
```

### Release workflow not triggered

**Cause:** Tag format incorrect or not pushed.

**Fix:** Ensure tag starts with 'v':
```bash
# ❌ Wrong
git tag 1.1.0

# ✅ Correct
git tag v1.1.0
git push origin v1.1.0
```

### Release creation failed

**Cause:** Tag version doesn't match `version.go`.

**Fix:** Ensure they match:
```go
// version.go
const Version = "1.1.0"

// Git tag must be
v1.1.0  // Not v1.1.1 or v1.0.0
```

## Workflow Permissions

The workflows require the following permissions:

### Version Bump Check
- `pull-requests: write` - To post comments on PRs
- `contents: read` - To read repository files

### Release
- `contents: write` - To create releases and tags

These are configured in the workflow files and don't require manual setup.

## Best Practices

1. **Bump version last** - Make your changes first, version bump as final commit
2. **One bump per PR** - Don't bump multiple times in a single PR
3. **Use skip label sparingly** - Only for non-code changes
4. **Coordinate major bumps** - Discuss breaking changes with team
5. **Test before tagging** - Ensure main branch is stable before creating release tag
6. **Write meaningful commits** - They become release notes

## Related Documentation

- [VERSION_BUMP_GUIDE.md](./VERSION_BUMP_GUIDE.md) - Detailed versioning guide
- [CHANGELOG.md](../CHANGELOG.md) - Release history and changes
- [Semantic Versioning](https://semver.org/) - Official SemVer specification
- [version.go](../version.go) - Current version constant
- [README.md](../README.md#versioning) - User-facing versioning info

## Questions?

If you have questions about the workflows:
1. Check [VERSION_BUMP_GUIDE.md](./VERSION_BUMP_GUIDE.md)
2. Review existing PRs for examples
3. Ask in PR comments
