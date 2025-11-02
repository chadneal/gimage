# Release Guide for gimage

This document explains how to create and publish a new release of gimage.

## Table of Contents
1. [First Release Setup (One-Time Only)](#first-release-setup-one-time-only)
2. [Prerequisites](#prerequisites)
3. [Release Process Overview](#release-process-overview)
4. [Step-by-Step Release Instructions](#step-by-step-release-instructions)
5. [What Happens During Release](#what-happens-during-release)
6. [Post-Release Tasks](#post-release-tasks)
7. [Troubleshooting](#troubleshooting)

## First Release Setup (One-Time Only)

If this is your first release, you need to configure GitHub secrets and ensure repositories are set up. **This only needs to be done once.**

### Step 1: Create GitHub Personal Access Token

This token allows GoReleaser to update your Homebrew tap automatically.

1. Go to: https://github.com/settings/tokens/new
2. Token name: `GoReleaser Homebrew Tap`
3. Expiration: `No expiration` (or custom period)
4. Select scopes:
   - ✅ **repo** (Full control of private repositories)
5. Click **"Generate token"**
6. **Copy the token immediately** (starts with `ghp_...`)
7. Save it securely - you'll need it in the next step

### Step 2: Create npm Authentication Token

This token allows GitHub Actions to publish your MCP server to npm.

1. Log in to npm: `npm login` (if not already logged in)
2. Go to: https://www.npmjs.com/settings/YOUR_USERNAME/tokens
3. Click **"Generate New Token"**
4. Choose **"Automation"** token type
5. **Copy the token immediately**
6. Save it securely - you'll need it in the next step

### Step 3: Add Secrets to GitHub Repository

Now add both tokens to your GitHub repository:

1. Go to: https://github.com/apresai/gimage/settings/secrets/actions
2. Click **"New repository secret"**
3. Add first secret:
   - Name: `HOMEBREW_TAP_TOKEN`
   - Secret: Paste your GitHub Personal Access Token from Step 1
   - Click **"Add secret"**
4. Click **"New repository secret"** again
5. Add second secret:
   - Name: `NPM_TOKEN`
   - Secret: Paste your npm token from Step 2
   - Click **"Add secret"**

### Step 4: Verify Setup

Check that everything is configured:

```bash
# 1. Verify homebrew-tap repository exists
gh repo view apresai/homebrew-tap

# 2. Verify you're logged into npm
npm whoami

# 3. Verify tests pass
make test

# 4. Test GoReleaser locally (won't publish)
goreleaser release --snapshot --clean
```

**✅ You're ready to release!** This setup only needs to be done once. For subsequent releases, skip to "Step-by-Step Release Instructions" below.

---

## Prerequisites

Before creating a release, ensure you have:

### Required Tools
- [ ] Go 1.22+ installed (`go version`)
- [ ] Git installed and configured (`git --version`)
- [ ] GoReleaser installed (`brew install goreleaser` or https://goreleaser.com/install/)
- [ ] GitHub CLI installed (optional): `brew install gh`

### Required Access
- [ ] Write access to the `apresai/gimage` repository
- [ ] GitHub Personal Access Token with `repo` scope (for Homebrew tap)
  - Go to: https://github.com/settings/tokens/new
  - Select scopes: `repo` (full control of private repositories)
  - Save token as `HOMEBREW_TAP_TOKEN` repository secret
- [ ] npm account and token (for MCP server distribution)
  - Create token at: https://www.npmjs.com/settings/yourname/tokens
  - Save as `NPM_TOKEN` repository secret

### Repository Setup
- [x] Homebrew tap repository already exists: https://github.com/apresai/homebrew-tap
  - GoReleaser will automatically push formula updates here
- [ ] All tests passing: `make test`
- [ ] Code linted: `make lint` (or install golangci-lint)

## Release Process Overview

```
┌─────────────────┐
│ 1. Update Code  │  Update CHANGELOG, verify tests
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 2. Create Tag   │  git tag v0.x.x
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  3. Push Tag    │  git push origin v0.x.x
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 4. GitHub       │  Automatically:
│    Actions      │  - Runs tests
│    Triggered    │  - Builds binaries
└────────┬────────┘  - Creates GitHub release
         │           - Updates Homebrew tap
         │           - Publishes to npm
         ▼
┌─────────────────┐
│ 5. Verify       │  Check release page
│    Release      │  Test installations
└─────────────────┘
```

## Step-by-Step Release Instructions

### Automated Release (Recommended)

**The entire release process is now automated!** Just run:

```bash
make release
```

This single command will:
1. **Update `CHANGELOG.md`** with the current version (auto-calculated from git commits)
   - Uses Claude Code AI to generate intelligent changelog entries from git commits
   - Falls back to commit messages if Claude is not available
   - Can be customized with a changes file if needed
2. Sync version to `package.json` and `npm/package.json`
3. Commit changes to git
4. Create and push git tag
5. Trigger GoReleaser to build and publish binaries
6. Publish npm package

**That's it!** The version is automatically calculated as `1.1.[commit_count]`.

### AI-Powered Changelog Generation

The script uses Claude Code (if available) to intelligently analyze git commits and generate structured changelog entries:

```bash
# Claude analyzes commits like:
a6b24a6 Migrate from chadneal to apresai organization
19bae35 Add make release target to automate the release process

# And generates:
### Changed
- Migrated repository from chadneal to apresai organization
- Updated all import paths and documentation references

### Added
- Automated release target in Makefile
```

**Fallback Levels:**
1. Custom changes file (manual override)
2. Claude Code AI generation (intelligent summaries)
3. Raw commit messages (simple list)
4. Default entry (minimal)

### Manual Release (Advanced)

If you need more control over the version number or changelog content:

#### 1.1 Decide on the Version Number

**Note**: Version is now auto-calculated as `1.1.[git-commit-count]`. To use a custom version, set it explicitly:

```bash
VERSION=1.2.0 make release
```

Use [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)

Increment:
- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

Examples:
- `1.1.17` → `1.1.18`: Automatic (commit count increases)
- `1.1.18` → `1.2.0`: New features (manual override)
- `1.2.0` → `2.0.0`: Breaking changes (manual override)

#### 1.2 Optionally Customize the CHANGELOG

The changelog is automatically updated with a default entry. To customize it:

```bash
# Create a changes file
cat > /tmp/changes.txt <<EOF
### Added
- WebP support via nativewebp library
- CLI convert command for format conversion

### Changed
- Improved error handling

### Fixed
- Fixed bug in resize operation
EOF

# Update changelog with custom content
./scripts/update-changelog.sh 1.1.18 /tmp/changes.txt
```

Or manually edit `CHANGELOG.md` before running `make release`.

#### 1.3 Run Tests

```bash
# Run all tests
make test

# Run linter (fix any issues)
make lint

# Build for all platforms to ensure it compiles
make build-all
```

#### 1.4 Run Release

```bash
# Run automated release
make release

# Or with custom version
VERSION=1.2.0 make release
```

### Step 3: Monitor the Release

#### 3.1 Watch GitHub Actions

Go to: https://github.com/apresai/gimage/actions

You should see:
1. **CI workflow** (green check): Tests passed
2. **Release workflow** (running): Creating release

Click on the Release workflow to watch progress:
- Test job: Runs all tests
- Release job: Builds binaries, creates release
- npm-publish job: Publishes to npm

#### 3.2 Check for Errors

If the workflow fails:
1. Click on the failed step to see error
2. Fix the issue in your code
3. Delete the tag: `git tag -d v0.2.0 && git push origin :refs/tags/v0.2.0`
4. Fix, commit, and create tag again

### Step 4: Verify the Release

#### 4.1 Check GitHub Release Page

Go to: https://github.com/apresai/gimage/releases

Verify:
- [ ] Release is published (not draft)
- [ ] Release notes are generated
- [ ] Binaries are attached (6 files):
  - `gimage_0.2.0_Darwin_arm64.tar.gz` (macOS Apple Silicon)
  - `gimage_0.2.0_Darwin_x86_64.tar.gz` (macOS Intel)
  - `gimage_0.2.0_Linux_arm64.tar.gz`
  - `gimage_0.2.0_Linux_x86_64.tar.gz`
  - `gimage_0.2.0_Windows_x86_64.zip`
  - `checksums.txt`

#### 4.2 Test Installation

**Test Homebrew (macOS/Linux):**
```bash
# Update Homebrew
brew update

# Install/upgrade
brew install apresai/tap/gimage
# or
brew upgrade apresai/tap/gimage

# Verify version
gimage --version
# Should show: gimage version 0.2.0
```

**Test Direct Download (Linux):**
```bash
# Download
curl -L https://github.com/apresai/gimage/releases/download/v0.2.0/gimage_0.2.0_Linux_x86_64.tar.gz -o gimage.tar.gz

# Extract
tar -xzf gimage.tar.gz

# Test
./gimage --version
```

**Test npm (MCP server):**
```bash
# Install globally
npm install -g @apresai/gimage-mcp

# Or update
npm update -g @apresai/gimage-mcp

# Verify
gimage-mcp --version
```

#### 4.3 Smoke Test

```bash
# Test basic functionality
gimage --help
gimage --version

# Test convert command
echo "Testing convert functionality..."
gimage convert test.png webp

# Test generation (requires API key)
export GEMINI_API_KEY=your_key
gimage generate "test image" --dry-run
```

## What Happens During Release

When you push a git tag, the automated release process distributes gimage through three channels:

### 1. GitHub Releases (Direct Downloads)
- **Binary files** for all platforms are uploaded
- Users can download and install manually
- Platforms: macOS (Intel/ARM), Linux (x64/ARM64), Windows (x64)

### 2. Homebrew Tap (macOS/Linux Package Manager)
- **Homebrew formula** is automatically updated in `apresai/homebrew-tap`
- Users install with: `brew install apresai/tap/gimage`
- Formula includes:
  - Binary URL pointing to GitHub release
  - SHA256 checksum for verification
  - Installation and test commands

### 3. npm Registry (MCP Server Distribution)
- **npm package** `@apresai/gimage-mcp` is published
- Users install with: `npm install -g @apresai/gimage-mcp`
- Package includes:
  - Postinstall script that downloads correct binary
  - Platform detection (macOS/Linux/Windows, x64/ARM64)
  - Automatic extraction and setup

### Distribution Timeline

After pushing a tag (e.g., `git push origin v0.2.0`):

| Time | Event |
|------|-------|
| +0s | GitHub Actions triggered |
| +30s | Tests pass |
| +2min | Binaries built for all platforms |
| +3min | GitHub Release created |
| +3min | Homebrew formula updated |
| +4min | npm package published |
| +5min | **Users can install!** |

Users can verify the release:
- **GitHub**: https://github.com/apresai/gimage/releases
- **Homebrew**: `brew info apresai/tap/gimage`
- **npm**: `npm view @apresai/gimage-mcp`

---

## Post-Release Tasks

### Announce the Release

Consider announcing on:
- [ ] GitHub Discussions
- [ ] Twitter/X
- [ ] Reddit (r/golang, r/programming)
- [ ] Your blog/website
- [ ] Product Hunt (for major releases)

### Update Documentation

If you have documentation sites:
- [ ] Update version numbers
- [ ] Update installation instructions
- [ ] Add new features to docs

### Monitor Issues

Watch for:
- Installation problems
- Platform-specific bugs
- User feedback

## Troubleshooting

### Tag Already Exists

```bash
# Delete local tag
git tag -d v0.2.0

# Delete remote tag
git push origin :refs/tags/v0.2.0

# Create new tag
git tag v0.2.0
git push origin v0.2.0
```

### GoReleaser Fails

```bash
# Test locally before pushing tag
goreleaser release --snapshot --clean

# This builds everything but doesn't publish
# Check the dist/ folder for outputs
```

### Homebrew Formula Not Updated

Check:
1. Is `HOMEBREW_TAP_TOKEN` secret configured?
2. Does the `homebrew-tap` repository exist?
3. Check GoReleaser logs for errors

Manual formula update:
```bash
# Clone the tap
git clone https://github.com/apresai/homebrew-tap
cd homebrew-tap

# Edit gimage.rb manually
# Update version, SHA256, URL

# Commit and push
git add gimage.rb
git commit -m "Update gimage to v0.2.0"
git push
```

### npm Publish Fails

1. Check `NPM_TOKEN` is configured correctly
2. Ensure `package.json` version matches release
3. Manually publish:
```bash
npm login
npm publish --access public
```

## Quick Reference

### Create a Patch Release (0.1.1 → 0.1.2)

```bash
# Update CHANGELOG.md
# Update Makefile VERSION
git add CHANGELOG.md Makefile
git commit -m "chore: prepare release v0.1.2"
git push
git tag v0.1.2
git push origin v0.1.2
```

### Create a Minor Release (0.1.2 → 0.2.0)

```bash
# Update CHANGELOG.md with new features
# Update Makefile VERSION
git add CHANGELOG.md Makefile
git commit -m "chore: prepare release v0.2.0"
git push
git tag v0.2.0
git push origin v0.2.0
```

### Create a Major Release (0.x.x → 1.0.0)

```bash
# Update CHANGELOG.md highlighting breaking changes
# Update Makefile VERSION
# Update README noting stability
git add CHANGELOG.md Makefile README.md
git commit -m "chore: prepare release v1.0.0 - first stable release!"
git push
git tag v1.0.0
git push origin v1.0.0
# 🎉 Celebrate!
```

## Tools and Resources

- GoReleaser Docs: https://goreleaser.com/
- Semantic Versioning: https://semver.org/
- Keep a Changelog: https://keepachangelog.com/
- GitHub Actions: https://docs.github.com/en/actions
- Homebrew Formula: https://docs.brew.sh/Formula-Cookbook

---

## Summary: Complete Release Checklist

Use this checklist for every release:

### Before First Release (One-Time Setup)
- [ ] Create GitHub Personal Access Token (repo scope)
- [ ] Create npm Automation Token
- [ ] Add `HOMEBREW_TAP_TOKEN` to GitHub Secrets
- [ ] Add `NPM_TOKEN` to GitHub Secrets
- [ ] Verify `apresai/homebrew-tap` repository exists
- [ ] Run `goreleaser release --snapshot --clean` to test locally

### For Every Release (Automated Process)

**1. Preparation** (2 minutes)
- [ ] Run `make test` (all tests must pass)
- [ ] Run `make lint` (no errors)
- [ ] Optionally customize CHANGELOG entry (see "Manual Release" section)

**2. Create Release** (1 command!)
- [ ] Run: `make release`
- [ ] **Done!** Everything is automated:
  - ✅ CHANGELOG.md updated
  - ✅ package.json files synced
  - ✅ Changes committed to git
  - ✅ Git tag created and pushed
  - ✅ GoReleaser builds binaries
  - ✅ npm package published

**3. Monitor** (5 minutes)
- [ ] Watch GitHub Actions: https://github.com/apresai/gimage/actions
- [ ] Verify tests pass
- [ ] Verify release job completes
- [ ] Verify npm-publish job completes

**4. Verify Release** (10 minutes)
- [ ] Check GitHub release: https://github.com/apresai/gimage/releases
- [ ] Test Homebrew: `brew install apresai/tap/gimage`
- [ ] Test npm: `npm view @apresai/gimage-mcp`
- [ ] Run smoke tests: `gimage --version`, `gimage --help`

**5. Post-Release** (Optional)
- [ ] Announce on social media
- [ ] Update documentation sites
- [ ] Monitor for issues

### Distribution Channels Summary

After a successful release, users can install gimage through:

| Method | Command | Users |
|--------|---------|-------|
| **Homebrew** | `brew install apresai/tap/gimage` | macOS/Linux developers |
| **npm** | `npm install -g @apresai/gimage-mcp` | Claude Desktop MCP users |
| **Direct** | Download from GitHub releases | All platforms |

### Version Numbering Guide

- **0.1.0 → 0.1.1**: Bug fixes (PATCH)
- **0.1.1 → 0.2.0**: New features or breaking changes (MINOR)
- **0.9.0 → 1.0.0**: First stable release! (MAJOR)

### Common Release Commands

```bash
# Automated release (recommended)
make release

# Release with custom version
VERSION=1.2.0 make release

# Update changelog only
make update-changelog

# Sync version to package.json files only
make sync-version

# Check current version
make version

# Delete tag if needed (before running make release again)
git tag -d v1.1.18
git push origin :refs/tags/v1.1.18

# Test GoReleaser locally without publishing
goreleaser release --snapshot --clean
```

**Total time for a release**: ~5 minutes (fully automated!)

---

**Questions?** Check the [Troubleshooting](#troubleshooting) section above or review the [Tools and Resources](#tools-and-resources).
