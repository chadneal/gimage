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

1. Go to: https://github.com/chadneal/gimage/settings/secrets/actions
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
gh repo view chadneal/homebrew-tap

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
- [ ] Write access to the `chadneal/gimage` repository
- [ ] GitHub Personal Access Token with `repo` scope (for Homebrew tap)
  - Go to: https://github.com/settings/tokens/new
  - Select scopes: `repo` (full control of private repositories)
  - Save token as `HOMEBREW_TAP_TOKEN` repository secret
- [ ] npm account and token (for MCP server distribution)
  - Create token at: https://www.npmjs.com/settings/yourname/tokens
  - Save as `NPM_TOKEN` repository secret

### Repository Setup
- [x] Homebrew tap repository already exists: https://github.com/chadneal/homebrew-tap
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

### Step 1: Prepare the Release

#### 1.1 Decide on the Version Number

Use [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)

Increment:
- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

For pre-1.0 versions (0.x.x):
- **0.MINOR.PATCH**: Still in initial development
- Breaking changes can happen in MINOR versions

Examples:
- `0.1.1` → `0.1.2`: Bug fixes
- `0.1.1` → `0.2.0`: New features or breaking changes
- `0.9.0` → `1.0.0`: First stable release!

#### 1.2 Update the CHANGELOG

Edit `CHANGELOG.md`:

```bash
# Open CHANGELOG.md in your editor
vim CHANGELOG.md  # or use your preferred editor
```

**Before:**
```markdown
## [Unreleased]

### Added
- WebP support
- New convert command
```

**After:**
```markdown
## [Unreleased]

(empty - ready for next release)

## [0.2.0] - 2025-11-01

### Added
- WebP support via nativewebp library
- CLI convert command for format conversion
- ...
```

**Important**: Add the comparison link at the bottom:
```markdown
[Unreleased]: https://github.com/chadneal/gimage/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/chadneal/gimage/compare/v0.1.1...v0.2.0
```

#### 1.3 Update Version in Makefile (if needed)

```bash
# Edit Makefile
vim Makefile

# Change this line:
VERSION?=0.1.1

# To:
VERSION?=0.2.0
```

#### 1.4 Run Tests

```bash
# Run all tests
make test

# Run linter (fix any issues)
make lint

# Build for all platforms to ensure it compiles
make build-all
```

#### 1.5 Commit Changes

```bash
# Stage all changes
git add CHANGELOG.md Makefile

# Commit with conventional commit message
git commit -m "chore: prepare release v0.2.0"

# Push to main branch
git push origin main
```

### Step 2: Create and Push the Git Tag

```bash
# Create annotated tag (recommended)
git tag -a v0.2.0 -m "Release v0.2.0

Major changes:
- WebP support added
- CLI convert command
- Improved error handling

See CHANGELOG.md for full details."

# Verify tag was created
git tag -l "v0.2.*"

# Push the tag to GitHub
# This triggers the release workflow!
git push origin v0.2.0
```

**Alternative**: Lightweight tag (simpler)
```bash
git tag v0.2.0
git push origin v0.2.0
```

### Step 3: Monitor the Release

#### 3.1 Watch GitHub Actions

Go to: https://github.com/chadneal/gimage/actions

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

Go to: https://github.com/chadneal/gimage/releases

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
brew install chadneal/tap/gimage
# or
brew upgrade chadneal/tap/gimage

# Verify version
gimage --version
# Should show: gimage version 0.2.0
```

**Test Direct Download (Linux):**
```bash
# Download
curl -L https://github.com/chadneal/gimage/releases/download/v0.2.0/gimage_0.2.0_Linux_x86_64.tar.gz -o gimage.tar.gz

# Extract
tar -xzf gimage.tar.gz

# Test
./gimage --version
```

**Test npm (MCP server):**
```bash
# Install globally
npm install -g @chadneal/gimage-mcp

# Or update
npm update -g @chadneal/gimage-mcp

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
- **Homebrew formula** is automatically updated in `chadneal/homebrew-tap`
- Users install with: `brew install chadneal/tap/gimage`
- Formula includes:
  - Binary URL pointing to GitHub release
  - SHA256 checksum for verification
  - Installation and test commands

### 3. npm Registry (MCP Server Distribution)
- **npm package** `@chadneal/gimage-mcp` is published
- Users install with: `npm install -g @chadneal/gimage-mcp`
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
- **GitHub**: https://github.com/chadneal/gimage/releases
- **Homebrew**: `brew info chadneal/tap/gimage`
- **npm**: `npm view @chadneal/gimage-mcp`

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
git clone https://github.com/chadneal/homebrew-tap
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
- [ ] Verify `chadneal/homebrew-tap` repository exists
- [ ] Run `goreleaser release --snapshot --clean` to test locally

### For Every Release

**1. Preparation** (5-10 minutes)
- [ ] Get current date: `date +%Y-%m-%d`
- [ ] Update `CHANGELOG.md` (move Unreleased to new version)
- [ ] Update `VERSION` in `Makefile` (if needed)
- [ ] Run `make test` (all tests must pass)
- [ ] Run `make lint` (no errors)
- [ ] Commit: `git commit -m "chore: prepare release vX.Y.Z"`
- [ ] Push: `git push origin main`

**2. Create Release** (1 minute)
- [ ] Create tag: `git tag vX.Y.Z`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] **Done!** GitHub Actions handles the rest

**3. Monitor** (5 minutes)
- [ ] Watch GitHub Actions: https://github.com/chadneal/gimage/actions
- [ ] Verify tests pass
- [ ] Verify release job completes
- [ ] Verify npm-publish job completes

**4. Verify Release** (10 minutes)
- [ ] Check GitHub release: https://github.com/chadneal/gimage/releases
- [ ] Test Homebrew: `brew install chadneal/tap/gimage`
- [ ] Test npm: `npm view @chadneal/gimage-mcp`
- [ ] Run smoke tests: `gimage --version`, `gimage --help`

**5. Post-Release** (Optional)
- [ ] Announce on social media
- [ ] Update documentation sites
- [ ] Monitor for issues

### Distribution Channels Summary

After a successful release, users can install gimage through:

| Method | Command | Users |
|--------|---------|-------|
| **Homebrew** | `brew install chadneal/tap/gimage` | macOS/Linux developers |
| **npm** | `npm install -g @chadneal/gimage-mcp` | Claude Desktop MCP users |
| **Direct** | Download from GitHub releases | All platforms |

### Version Numbering Guide

- **0.1.0 → 0.1.1**: Bug fixes (PATCH)
- **0.1.1 → 0.2.0**: New features or breaking changes (MINOR)
- **0.9.0 → 1.0.0**: First stable release! (MAJOR)

### Common Release Commands

```bash
# Quick patch release
git add CHANGELOG.md Makefile
git commit -m "chore: prepare release v0.1.2"
git push
git tag v0.1.2
git push origin v0.1.2

# Delete tag if needed
git tag -d v0.1.2
git push origin :refs/tags/v0.1.2

# Test locally without publishing
goreleaser release --snapshot --clean
```

**Total time for a release**: ~20 minutes (mostly automated!)

---

**Questions?** Check the [Troubleshooting](#troubleshooting) section above or review the [Tools and Resources](#tools-and-resources).
