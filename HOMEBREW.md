# Homebrew Distribution Plan for gimage

This document outlines the complete strategy for distributing `gimage` via Homebrew, enabling users to install it with `brew install gimage`.

## Overview

Homebrew distribution involves two main approaches:
1. **Custom Tap** (Initial/Personal): `brew install chadneal/gimage/gimage`
2. **Homebrew Core** (Official/Long-term): `brew install gimage`

We'll start with a custom tap and eventually submit to homebrew-core for wider distribution.

---

## Phase 1: Create Custom Tap (homebrew-gimage)

### Prerequisites
- GitHub repository: `https://github.com/chadneal/gimage` (already exists)
- Repository must have tagged releases (v1.0.0 already exists)
- `gh` CLI tool installed: `brew install gh`

### Step 1: Create Tap Repository

```bash
# Create the tap locally
brew tap-new chadneal/gimage

# This creates:
# /usr/local/Homebrew/Library/Taps/chadneal/homebrew-gimage/
# (or /opt/homebrew/... on Apple Silicon)

# Navigate to tap directory
cd "$(brew --repository chadneal/gimage)"
```

**Expected Output:**
```
Initialized empty Git repository in /opt/homebrew/Library/Taps/chadneal/homebrew-gimage/.git/
==> Created chadneal/gimage
/opt/homebrew/Library/Taps/chadneal/homebrew-gimage
```

### Step 2: Create Formula

```bash
# Option A: Create formula from release URL
brew create https://github.com/chadneal/gimage/archive/refs/tags/v1.0.0.tar.gz \
  --tap chadneal/gimage \
  --set-name gimage

# Option B: Create formula manually
cd "$(brew --repository chadneal/gimage)"
mkdir -p Formula
touch Formula/gimage.rb
```

### Step 3: Write Formula Definition

Edit `Formula/gimage.rb`:

```ruby
class Gimage < Formula
  desc "AI-powered image generation and processing CLI"
  homepage "https://github.com/chadneal/gimage"
  url "https://github.com/chadneal/gimage/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "PLACEHOLDER_SHA256"  # Update with actual SHA
  license "MIT"  # Update based on your LICENSE file
  head "https://github.com/chadneal/gimage.git", branch: "main"

  depends_on "go" => :build

  def install
    # Build the binary
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/chadneal/gimage/internal/cli.version=#{version}"), "./cmd/gimage"

    # Generate shell completions
    generate_completions_from_executable(bin/"gimage", "completion")
  end

  test do
    # Test version output
    assert_match version.to_s, shell_output("#{bin}/gimage --version")

    # Test help output
    assert_match "AI-powered image generation", shell_output("#{bin}/gimage --help")
  end
end
```

### Step 4: Generate SHA256 Checksum

```bash
# Download and compute SHA256
curl -L https://github.com/chadneal/gimage/archive/refs/tags/v1.0.0.tar.gz | shasum -a 256

# Update the sha256 in Formula/gimage.rb with the output
```

### Step 5: Test Formula Locally

```bash
# Audit the formula
brew audit --new --formula chadneal/gimage/gimage

# Install from source to test
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source --verbose chadneal/gimage/gimage

# Test the installation
gimage --version
gimage --help

# Run formula tests
brew test chadneal/gimage/gimage

# Uninstall for further testing
brew uninstall gimage
```

### Step 6: Push Tap to GitHub

```bash
# Navigate to tap directory
cd "$(brew --repository chadneal/gimage)"

# Create GitHub repository and push
gh repo create chadneal/homebrew-gimage \
  --push \
  --public \
  --source "$(brew --repository chadneal/gimage)" \
  --description "Homebrew tap for gimage - AI-powered image generation and processing CLI"

# Verify
git remote -v
```

**Expected Output:**
```
✓ Created repository chadneal/homebrew-gimage on GitHub
✓ Added remote https://github.com/chadneal/homebrew-gimage.git
✓ Pushed commits to https://github.com/chadneal/homebrew-gimage.git
```

---

## Phase 2: Enable User Installation

### User Installation Process

Once the tap is published, users can install gimage:

```bash
# Method 1: Install from tap (automatically taps if needed)
brew install chadneal/gimage/gimage

# Method 2: Tap first, then install
brew tap chadneal/gimage
brew install gimage
```

### Update Formula for New Releases

When releasing a new version (e.g., v1.1.0):

```bash
# Navigate to tap
cd "$(brew --repository chadneal/gimage)"

# Use bump-formula-pr for automated updates
brew bump-formula-pr \
  --url=https://github.com/chadneal/gimage/archive/refs/tags/v1.1.0.tar.gz \
  --sha256=NEW_SHA256 \
  chadneal/gimage/gimage

# Or manually update Formula/gimage.rb and commit
git add Formula/gimage.rb
git commit -m "gimage 1.1.0"
git push
```

---

## Phase 3: Enhance Formula Features

### Add Shell Completions

If gimage supports shell completions:

```ruby
def install
  system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/gimage"

  # Generate completions
  generate_completions_from_executable(bin/"gimage", "completion")
end
```

### Add Caveats for Post-Install Instructions

```ruby
def caveats
  <<~EOS
    Before using gimage, set up your API credentials:
      $ gimage auth gemini    # For Gemini API
      $ gimage auth vertex    # For Vertex AI

    Get your Gemini API key from:
      https://aistudio.google.com/app/apikey

    Configuration is stored in:
      ~/.gimage/config.md
  EOS
end
```

### Add Service Support (if applicable)

If gimage has a server mode:

```ruby
service do
  run [opt_bin/"gimage", "serve"]
  keep_alive true
  log_path var/"log/gimage.log"
  error_log_path var/"log/gimage.log"
end
```

---

## Phase 4: Submit to Homebrew Core (Optional)

### Prerequisites for Core Submission

1. **30-day history**: Formula must exist in a tap for 30+ days
2. **50+ forks/stars**: GitHub repo should have significant interest
3. **Regular updates**: Active maintenance
4. **CI/CD**: Automated testing
5. **Documentation**: Comprehensive README and help

### Submission Process

```bash
# Fork homebrew-core
gh repo fork homebrew/homebrew-core --clone

# Create formula in core
cd homebrew-core
brew create https://github.com/chadneal/gimage/archive/refs/tags/v1.0.0.tar.gz

# Test thoroughly
brew audit --new --online --formula gimage
brew install --build-from-source gimage
brew test gimage

# Submit PR
git checkout -b gimage
git add Formula/g/gimage.rb
git commit -m "gimage 1.0.0 (new formula)"
git push --set-upstream origin gimage
gh pr create --fill
```

### Core Formula Requirements

- Must be in `Formula/g/gimage.rb` (alphabetical subdirectory)
- No external tap dependencies
- Must build on macOS and Linux (if applicable)
- Clear, concise description
- Homepage must be HTTPS
- Well-tested with comprehensive test block

---

## Phase 5: Continuous Maintenance

### Automated Version Bumps

Create GitHub Actions in `homebrew-gimage` repo:

**.github/workflows/bump-formula.yml**
```yaml
name: Bump Formula
on:
  repository_dispatch:
    types: [new-release]

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Homebrew
        uses: Homebrew/actions/setup-homebrew@master

      - name: Bump formula
        env:
          HOMEBREW_GITHUB_API_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          brew bump-formula-pr \
            --url=https://github.com/chadneal/gimage/archive/refs/tags/${{ github.event.client_payload.version }}.tar.gz \
            --sha256=${{ github.event.client_payload.sha256 }} \
            chadneal/gimage/gimage
```

### Test Suite in Tap

Add comprehensive tests:

```ruby
test do
  # Version check
  assert_match version.to_s, shell_output("#{bin}/gimage --version")

  # Help text
  assert_match "AI-powered image generation", shell_output("#{bin}/gimage --help")

  # Subcommands exist
  assert_match "generate", shell_output("#{bin}/gimage --help")
  assert_match "resize", shell_output("#{bin}/gimage --help")

  # Config initialization
  system bin/"gimage", "config", "init"
  assert_predicate testpath/".gimage", :exist?
end
```

---

## Phase 6: Distribution Strategy

### Documentation Updates

1. **README.md** - Add Homebrew installation section:
   ```markdown
   ## Installation

   ### Homebrew (macOS/Linux)
   ```bash
   brew install chadneal/gimage/gimage
   ```

   ### Upgrade
   ```bash
   brew upgrade gimage
   ```
   ```

2. **COMMANDS.md** - Add Homebrew commands

3. **Website/Docs** - Prominent Homebrew installation instructions

### Release Process Integration

When creating new releases:

```bash
# 1. Update version and create git tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# 2. Create GitHub release (triggers brew formula update)
gh release create v1.1.0 \
  --title "v1.1.0" \
  --notes "Release notes here"

# 3. Update Homebrew formula
cd "$(brew --repository chadneal/gimage)"
brew bump-formula-pr --version=1.1.0 gimage
```

---

## Testing Checklist

Before publishing tap:

- [ ] Formula passes `brew audit --new --formula`
- [ ] Installation works: `brew install --build-from-source`
- [ ] Binary is executable and version correct
- [ ] Help output displays correctly
- [ ] Shell completions installed (if applicable)
- [ ] Test block passes: `brew test`
- [ ] Uninstall works: `brew uninstall`
- [ ] Reinstall works: `brew reinstall`
- [ ] Works on both Intel and Apple Silicon Macs
- [ ] Documentation updated

---

## Common Issues & Solutions

### Issue: SHA256 Mismatch
```bash
# Regenerate checksum
curl -L https://github.com/chadneal/gimage/archive/refs/tags/v1.0.0.tar.gz | shasum -a 256
```

### Issue: Build Failures
```bash
# Test interactively
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source --interactive --verbose gimage
```

### Issue: Formula Not Found
```bash
# Ensure tap is properly set up
brew tap chadneal/gimage
brew tap  # Verify tap is listed
```

---

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Go Formula Guidelines](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request#guidelines-for-go-formulae)

---

## Timeline Estimate

- **Phase 1** (Custom Tap): 2-3 hours
- **Phase 2** (User Testing): 1 week
- **Phase 3** (Enhancement): 1-2 hours
- **Phase 4** (Core Submission): 30+ days waiting period + 2-3 days for PR
- **Phase 5** (Automation): 2-3 hours

**Total to working tap**: ~4 hours
**Total to Homebrew Core**: ~30-60 days

---

## Success Metrics

- [ ] Users can install with single command
- [ ] Formula auto-updates with new releases
- [ ] Installation works on all supported platforms
- [ ] Clear error messages for common issues
- [ ] Comprehensive test coverage
- [ ] Active maintenance and updates
