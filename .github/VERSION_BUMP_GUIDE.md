# Version Bump Guide

This guide explains how semantic versioning works in this project and how to properly bump the version for your PR.

## Semantic Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/). Version numbers have the format: `MAJOR.MINOR.PATCH`

### When to Bump Each Number

#### MAJOR (X.0.0) - Breaking Changes
Increment when you make incompatible API changes:
- Removing or renaming public functions, types, or fields
- Changing function signatures
- Removing support for older Go versions
- Changing behavior that breaks existing code

**Example:** `1.5.3` → `2.0.0`

#### MINOR (0.X.0) - New Features
Increment when you add functionality in a backwards compatible manner:
- Adding new public functions or methods
- Adding new configuration options
- Adding new features
- Deprecating functionality (but not removing it)

**Example:** `1.5.3` → `1.6.0`

#### PATCH (0.0.X) - Bug Fixes
Increment when you make backwards compatible bug fixes:
- Fixing bugs
- Improving documentation
- Refactoring internal code
- Updating dependencies (without breaking changes)
- Performance improvements

**Example:** `1.5.3` → `1.5.4`

## How to Bump the Version

### Step 1: Edit `version.go`

Update the `Version` constant in the `version.go` file:

```go
package ziptax

const Version = "1.6.0"  // ← Change this
```

### Step 2: Commit the Change

```bash
git add version.go
git commit -m "chore: bump version to 1.6.0"
git push
```

### Step 3: Create/Update Your PR

The GitHub Action will automatically:
1. ✅ Check that the version was bumped
2. ✅ Validate the semantic version format
3. ✅ Determine the bump type (major/minor/patch)
4. ✅ Post a comment on your PR with the version change details

## GitHub Action Workflow

### What It Checks

The `version-bump-check.yml` workflow runs on every PR to `main` and verifies:

1. **Version Changed**: Ensures version was modified
2. **Valid Format**: Checks semantic versioning format (e.g., `1.2.3`)
3. **Version Increased**: Ensures version went up, not down
4. **Bump Type**: Determines if it's a major, minor, or patch bump

### Workflow Output

When successful, you'll see a comment on your PR:

```
🚀 Version Bump Detected

Type: MINOR (New Features)
From: 1.5.3 → To: 1.6.0

✅ Semantic version has been properly bumped.
```

### If Check Fails

If you forget to bump the version, the workflow will fail with:

```
❌ Version must be bumped for this PR!

Current version: 1.5.3
PR version: 1.5.3

Semantic Versioning Guidelines:
  • MAJOR (X.0.0): Breaking changes
  • MINOR (0.X.0): New features (backward compatible)
  • PATCH (0.0.X): Bug fixes (backward compatible)

To skip this check, add the 'skip-version-check' label to the PR.
```

## Skipping the Version Check

In rare cases, you may need to skip the version check (e.g., documentation-only changes, CI updates):

### Option 1: Add a Label

Add the `skip-version-check` label to your PR via the GitHub UI.

### Option 2: Create the Label (First Time)

If the label doesn't exist yet:

1. Go to your repo → Issues → Labels
2. Click "New label"
3. Name: `skip-version-check`
4. Description: "Skip the semantic version bump check"
5. Color: `#d4c5f9` (optional)

## Examples

### Example 1: Adding a New Feature

You're adding TaxCloud order management to the SDK:

```go
// Before
const Version = "1.0.0"

// After
const Version = "1.1.0"  // MINOR bump (new features)
```

**Reasoning:** New functionality added, but existing code still works.

### Example 2: Fixing a Bug

You're fixing a bug in error handling:

```go
// Before
const Version = "1.1.0"

// After
const Version = "1.1.1"  // PATCH bump (bug fix)
```

**Reasoning:** No new features, just fixing existing functionality.

### Example 3: Breaking Change

You're removing a deprecated function:

```go
// Before
const Version = "1.1.1"

// After
const Version = "2.0.0"  // MAJOR bump (breaking change)
```

**Reasoning:** Removing public API breaks existing code.

### Example 4: Multiple Changes

If your PR includes multiple types of changes, use the **highest level** bump needed:

- Bug fix + new feature = **MINOR** (1.1.1 → 1.2.0)
- New feature + breaking change = **MAJOR** (1.2.0 → 2.0.0)
- Multiple bug fixes = **PATCH** (1.2.0 → 1.2.1)

## Best Practices

### 1. One Version Bump Per PR
- Each PR should only bump the version once
- Don't bump multiple times in a single PR

### 2. Bump at the End
- Add your changes first
- Bump version as the final commit
- This keeps the version history clean

### 3. Update CHANGELOG
- When bumping version, update `CHANGELOG.md` (if it exists)
- Document what changed in the new version

### 4. Coordinate Major Bumps
- Discuss major version bumps with the team
- Major bumps indicate breaking changes affecting users

### 5. Pre-release Versions
For beta/alpha releases, use pre-release identifiers:

```go
const Version = "2.0.0-beta.1"  // Beta release
const Version = "2.0.0-rc.2"     // Release candidate
const Version = "2.0.0"          // Stable release
```

## Troubleshooting

### "Version format is invalid"

**Problem:** Version doesn't match semantic versioning format.

**Solution:** Ensure version follows `X.Y.Z` format:
```go
// ❌ Wrong
const Version = "1.1"
const Version = "v1.1.0"

// ✅ Correct
const Version = "1.1.0"
const Version = "1.1.0-beta.1"
```

### "Version was decreased"

**Problem:** New version is lower than base version.

**Solution:** Check that you increased the version correctly:
```go
// Base: 1.5.3
const Version = "1.6.0"  // ✅ Correct (increased)
const Version = "1.5.2"  // ❌ Wrong (decreased)
```

### "Version was not changed"

**Problem:** Version wasn't modified in your PR.

**Solution:** Update `version.go` and commit the change.

## Questions?

If you're unsure which version number to bump:

1. **Check existing PRs** - See how similar changes were versioned
2. **Ask in PR comments** - Tag a maintainer for guidance
3. **Review semver.org** - Official semantic versioning specification

## Related Files

- [`version.go`](../version.go) - Version constant
- [`.github/workflows/version-bump-check.yml`](workflows/version-bump-check.yml) - GitHub Action
- [`CLAUDE.md`](../CLAUDE.md) - AI development documentation
- [`README.md`](../README.md) - Project documentation
