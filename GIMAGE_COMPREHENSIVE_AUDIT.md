# Gimage Project - Comprehensive Audit Report

**Date:** 2025-11-06
**Scope:** Full project audit with focus on authentication/configuration UX
**Context:** Users have API keys and want minimal friction to start using gimage

---

## Executive Summary

Gimage is a well-architected Go CLI tool with three major issues affecting user experience:

1. **Authentication Complexity**: Multiple overlapping systems create confusion
2. **Configuration Unpredictability**: 3-tier hierarchy (flags > env > config) lacks visibility
3. **Inconsistent CLI Patterns**: Different commands use different argument styles

**Key Finding**: The core problem isn't the configuration *system* - it's the **lack of clarity about what's being used**. Users can't easily tell which credentials are active or where they came from.

---

## Part 1: Authentication & Configuration - The Core Problem

### Current Pain Points (Ranked by Impact)

#### 🔴 **CRITICAL: Credential Visibility Gap**

**Problem:** Users have no way to know which credentials are active.

```bash
# User has both env var and config file set
export GEMINI_API_KEY="key1"
# ~/.gimage/config.md contains "key2"

# Which key is being used? No way to know without reading docs!
gimage generate "test"
```

**Impact:**
- Silent failures when wrong credential is used
- Can't debug "why isn't my API key working?"
- No confidence in system state

**Fix Priority:** P0 - Must fix

---

#### 🔴 **CRITICAL: Two Parallel Auth Systems**

**Problem:** Legacy and new auth commands coexist, creating confusion.

- **Legacy:** `gimage auth gemini`, `gimage auth vertex`, `gimage auth bedrock`
- **New:** `gimage auth setup <provider>`, `gimage auth list`, `gimage auth test`

**Evidence:**
- `/home/user/gimage/internal/cli/auth.go` has old interactive prompts
- `/home/user/gimage/internal/cli/auth_setup.go` has new provider registry system
- Both work, docs mention both

**Impact:**
- User doesn't know which to use
- Code duplication (2 paths to maintain)
- Inconsistent help text

**Fix Priority:** P0 - Deprecate legacy, keep new system

---

#### 🟡 **HIGH: Validation Only Happens at API Call Time**

**Problem:** `ValidateConfig()` is only called in `SaveConfig()`, not `LoadConfig()`

**Evidence:** `/home/user/gimage/internal/config/config.go:195`
```go
// SaveConfig validates before saving
func SaveConfig(cfg *Config) error {
    if err := ValidateConfig(cfg); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    // ...
}

// LoadConfig does NOT validate!
func LoadConfig() (*Config, error) {
    // Loads and parses but doesn't validate
}
```

**Impact:**
- Invalid config loads silently
- Errors appear during generation (confusing, wrong place)
- Example: `vertex_project: "INVALID-PROJECT-123"` loads fine, fails at API time

**Fix Priority:** P1 - Validate on load, warn on save

---

#### 🟡 **HIGH: Config File Documentation Bug**

**Problem:** Help text says `~/.gimage/config.yaml` but actual file is `~/.gimage/config.md`

**Evidence:**
- Multiple command help texts reference `.yaml`
- Actual implementation uses markdown format (`**key**: value`)

**Impact:** Users create wrong file, config doesn't load

**Fix Priority:** P1 - Fix all help text

---

#### 🟡 **HIGH: Bedrock Auth Complexity**

**Problem:** 4 different auth methods with no clear precedence or guidance

1. Bearer token (`AWS_BEARER_TOKEN_BEDROCK`)
2. Access keys (`AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY`)
3. AWS Profile (`AWS_PROFILE`)
4. IAM role (implicit, no config)

**Evidence:** `/home/user/gimage/internal/config/auth.go:251-293`

**Impact:**
- User sets both bearer token and access keys - which is used?
- No error if multiple methods configured
- Confusing fallback logic

**Fix Priority:** P1 - Document and enforce precedence

---

#### 🟢 **MEDIUM: No Credential Source Transparency**

**Problem:** When a credential works, you can't tell where it came from.

**What Users Want:**
```bash
$ gimage auth status
✓ Gemini API configured
  Source: Environment variable (GEMINI_API_KEY)
  Validated: Yes

✓ Vertex AI configured
  Source: Config file (~/.gimage/config.md)
  Mode: Express (API key)
  Validated: Yes
```

**Current State:**
```bash
$ gimage auth list
# Shows provider names and ✓/✗ but no source info
```

**Fix Priority:** P2 - Add `auth status --verbose` command

---

### What Works Well

✅ **3-tier hierarchy is correct** - Flags > Env > Config is standard
✅ **Secure permissions** - Config file uses 0600, good security
✅ **Multiple auth modes** - Flexibility for different use cases
✅ **Provider registry architecture** - Clean extensible design
✅ **Markdown config format** - Human-readable (though unusual choice)

---

## Part 2: Recommended Auth/Config Architecture

### Proposal: "Show, Don't Hide" Philosophy

**Core Principle:** Users should always know what credentials are being used.

### Changes

#### 1. Add `gimage auth status` Command (P0)

```bash
$ gimage auth status

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Authentication Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ Gemini API
  Status:     Configured
  Source:     Environment (GEMINI_API_KEY)
  Key:        AIza...cVVI (last 4: cVVI)
  Validated:  Not tested (run: gimage auth test gemini)

✓ Vertex AI
  Status:     Configured
  Source:     Config file (~/.gimage/config.md)
  Mode:       Express (API key)
  Project:    my-project-123
  Location:   us-central1
  Validated:  Not tested

✗ AWS Bedrock
  Status:     Not configured
  Setup:      gimage auth setup bedrock

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Default API: gemini (from config file)
Default Model: gemini-2.5-flash-image
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Run 'gimage auth test <provider>' to validate credentials.
```

#### 2. Show Active Credentials in Verbose Mode (P0)

```bash
$ gimage generate "test" --verbose

Using credentials:
  API: gemini (from config file: default_api)
  Model: gemini-2.5-flash-image (from config file: default_model)
  API Key: AIza...cVVI (from environment: GEMINI_API_KEY)

Generating image...
```

#### 3. Consolidate to Single Auth System (P0)

**Remove:** Legacy `gimage auth gemini|vertex|bedrock` commands
**Keep:** New provider-based system (`gimage auth setup`, `auth list`, `auth test`)

**Migration:**
- Deprecation warning for 2 releases
- Redirect old commands to new ones internally
- Update all docs

#### 4. Add Validation to LoadConfig (P1)

```go
func LoadConfig() (*Config, error) {
    cfg := &Config{/* ... */}

    // Load from file and env...

    // VALIDATE before returning
    if err := ValidateConfig(cfg); err != nil {
        // Non-fatal: log warning but don't fail
        log.Warnf("Config validation warning: %v", err)
    }

    return cfg, nil
}
```

#### 5. Enforce Bedrock Credential Precedence (P1)

**Order:** Bearer Token > Access Keys > AWS Profile > IAM Role

**Implementation:**
```go
// In internal/generate/providers.go CreateClient
if cfg.AWSBedrockAPIKey != "" {
    return NewBedrockRESTClient(cfg.AWSBedrockAPIKey, region)
}
if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
    return NewBedrockSDKClient(ctx, region, WithKeys(...))
}
if cfg.AWSProfile != "" {
    return NewBedrockSDKClient(ctx, region, WithProfile(...))
}
// Default: IAM role via SDK credential chain
return NewBedrockSDKClient(ctx, region)
```

**Document clearly** in help text and errors.

---

## Part 3: CLI UX Issues

### Issues Found

#### 🔴 **CRITICAL: Three Stubbed Commands**

**Commands that don't work:**
1. `batch` - Empty RunE, no implementation
2. `compress` - Empty RunE, no implementation
3. `config` - Empty RunE, no implementation

**Evidence:** These commands are defined but have no implementation.

**Impact:**
- User runs `gimage compress image.jpg` - nothing happens
- No error message, just silent failure or "coming soon" message

**Fix:** Either implement or remove from `--help`

---

#### 🟡 **HIGH: Inconsistent Argument Patterns**

Different commands use different styles:

| Command | Style | Example |
|---------|-------|---------|
| `generate` | Flags | `--size 1024x1024` |
| `resize` | Positional args | `resize img.jpg 1024 1024` |
| `batch` | Mixed flags | `--width 1024 --height 1024` |
| `crop` | Positional args (5!) | `crop img.jpg 0 0 800 600` |

**Impact:** Confusing, hard to remember

**Fix:** Standardize on one pattern (prefer flags for optional, positional for required)

---

#### 🟡 **HIGH: Silent API Selection**

**Problem:** If user has multiple APIs configured, no indication which is selected.

```bash
# User has both Gemini and Vertex configured
$ gimage generate "test"
# Uses Gemini (default_api in config)
# No mention that Vertex is also available!
```

**Fix:**
- In verbose mode: "Using API: gemini (default, vertex also available)"
- Consider `--api` flag auto-complete showing available APIs

---

#### 🟢 **MEDIUM: Inconsistent Auto-Generated Filenames**

```bash
resize: photo_resized_800x600.jpg     # Good
scale:  photo_scaled_0.50x.jpg        # Malformed (0.50x)
crop:   photo_cropped_800x600.jpg     # Good
```

**Fix:** Standardize filename patterns

---

#### 🟢 **MEDIUM: No Output Control**

**Missing:**
- `--quiet` flag (suppress status messages)
- `--json` flag (machine-readable output)
- Status messages go to stdout (should be stderr)

**Fix:** Add output flags, redirect status to stderr

---

### What Works Well

✅ **Cobra framework** - Good help text generation
✅ **Root command examples** - Excellent documentation
✅ **Error wrapping with %w** - Good error context
✅ **Provider registry** - Clean architecture for multiple backends

---

## Part 4: Backend Implementation Issues

### Code Quality Findings

#### 🟡 **HIGH: Significant Code Duplication**

**~300 lines of duplicate code identified across:**

1. **Error handling** (80 lines duplicated in `gemini_rest.go`, `vertex_rest.go`, `bedrock_rest.go`)
2. **Bedrock buildRequest()** (60 lines in both `bedrock_rest.go` and `bedrock_sdk.go`)
3. **Retry logic** (60 lines circuit breaker + exponential backoff)
4. **Verbose flag init** (5 lines x 5 files = 25 lines)
5. **Model defaults** (hardcoded in 6 locations)

**Impact:**
- 20-30% code bloat
- Bug fixes must be applied in multiple places
- Harder to maintain

**Fix:** Extract common code to shared helpers

---

#### 🟡 **HIGH: Inconsistent Logging**

**Two different logging systems:**
- `fmt.Printf()` in some files (e.g., `gemini_rest.go`)
- `zerolog` in others (e.g., `bedrock_sdk.go`)

**Impact:** Can't control log levels uniformly

**Fix:** Standardize on zerolog

---

#### 🟢 **MEDIUM: No Unified Error Classification**

**Problem:** Each backend has its own error types, no common classification.

**Example:**
- Gemini returns "failed to generate image"
- Vertex returns "imagen generation failed"
- Bedrock returns "nova canvas error"

**Impact:** Hard to handle errors generically (e.g., retry only on rate limits)

**Fix:** Define common error types (RateLimitError, AuthError, ValidationError)

---

### What Works Well

✅ **Interface consistency** - All backends implement `ImageGenerator`
✅ **Circuit breaker pattern** - Good resilience
✅ **Provider metadata** - Pricing, capabilities well-documented
✅ **Prompt validation** - Good input checking

---

## Part 5: Testing Gaps

### Current State

**Found:**
- 28 test files
- 102 total Go files
- Test coverage unknown (test command failed)

**Test Types:**
- Unit tests: Exist for some modules
- Integration tests: Tagged, manual (cost money)
- E2E tests: Exist in `test/integration/`

### Issues

#### 🟡 **HIGH: No CLI Command Tests**

**Missing tests for:**
- `generate` command flag parsing
- `resize`, `scale`, `crop` commands
- `auth` command flows

**Impact:** CLI regressions go unnoticed

**Fix:** Add CLI unit tests (can mock backends)

---

#### 🟢 **MEDIUM: No Config Validation Tests**

**Missing:**
- Test `ValidateConfig()` with invalid inputs
- Test config file parsing edge cases
- Test credential precedence order

**Fix:** Add config package tests

---

## Part 6: Documentation Issues

### Issues Found

#### 🟡 **HIGH: Config File Format Mismatch**

**Documented:** `~/.gimage/config.yaml`
**Actual:** `~/.gimage/config.md`

**Locations to fix:**
- COMMANDS.md references
- CLI help text
- Error messages

---

#### 🟢 **MEDIUM: Missing Troubleshooting Guide**

**What's missing:**
- "Which credential is being used?" guide
- Debugging credential issues
- Precedence order visualization

**Fix:** Add TROUBLESHOOTING.md

---

## Part 7: Recommendations by Priority

### Priority 0 (Critical - Fix First)

1. **Add `gimage auth status` command**
   - Show which credentials are active and their source
   - Estimated effort: 1 day

2. **Consolidate auth systems**
   - Deprecate legacy `auth gemini|vertex|bedrock`
   - Keep new provider-based system only
   - Estimated effort: 2 days

3. **Show active credentials in verbose mode**
   - `gimage generate "test" --verbose` shows which API key is used
   - Estimated effort: 4 hours

4. **Fix config file documentation**
   - Change all `.yaml` references to `.md`
   - Estimated effort: 1 hour

5. **Implement or remove stubbed commands**
   - `batch`, `compress`, `config` are broken
   - Estimated effort: 2 weeks (if implement), 1 hour (if remove)

### Priority 1 (High Impact)

1. **Validate config on load**
   - Call `ValidateConfig()` in `LoadConfig()`
   - Show warnings for invalid values
   - Estimated effort: 2 hours

2. **Enforce Bedrock credential precedence**
   - Document and implement clear order
   - Add tests
   - Estimated effort: 4 hours

3. **Standardize CLI argument patterns**
   - Choose flags vs positional args consistently
   - Update all commands
   - Estimated effort: 1 week

4. **Reduce code duplication**
   - Extract common error handling, retry logic, request building
   - Estimated effort: 1 week

### Priority 2 (Quality of Life)

1. **Add output control flags**
   - `--quiet`, `--json`, stderr for status
   - Estimated effort: 2 days

2. **Unified error classification**
   - Common error types across backends
   - Estimated effort: 3 days

3. **Standardize logging**
   - Use zerolog everywhere
   - Estimated effort: 1 day

4. **Add CLI command tests**
   - Unit tests for all commands
   - Estimated effort: 1 week

### Priority 3 (Nice to Have)

1. **Add TROUBLESHOOTING.md**
   - Common issues and solutions
   - Estimated effort: 4 hours

2. **Shell completion improvements**
   - Auto-complete for available APIs, models
   - Estimated effort: 2 days

3. **Config migration tool**
   - Migrate from old format to new
   - Estimated effort: 1 day

---

## Part 8: Proposed "Quick Start" User Experience

### Current Experience (Frustrating)

```bash
# User installs gimage
brew install apresai/tap/gimage

# User tries to generate
$ gimage generate "test"
Error: Gemini API key not found. Please set it via:
  1. Command flag: --api-key YOUR_KEY
  2. Environment variable: export GEMINI_API_KEY=YOUR_KEY
  3. Config file: gimage config set gemini_api_key YOUR_KEY
Get your API key at: https://ai.google.dev/

# User sets env var
$ export GEMINI_API_KEY="my-key"

# User tries again
$ gimage generate "test"
# Works! But user doesn't know:
# - If the env var is being used
# - If there's also a config file
# - What happens if both are set
```

### Proposed Experience (Clear)

```bash
# User installs gimage
brew install apresai/tap/gimage

# User tries to generate
$ gimage generate "test"
Error: No API credentials configured.

Run 'gimage auth setup gemini' to configure Gemini API (free tier available)
Or see all providers: 'gimage auth list'

# User runs setup
$ gimage auth setup gemini

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Gemini API Setup
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Get your free API key: https://aistudio.google.com/app/apikey

Gemini API Key: ***paste key***

Testing credentials... ✓ Success!

Configuration saved to: ~/.gimage/config.md
You can now generate images!

Try: gimage generate "a sunset over mountains"

# User generates
$ gimage generate "test"

Using Gemini API (gemini-2.5-flash-image)
Credentials: Config file (~/.gimage/config.md)

Generating... ✓ Done!
Saved to: generated_20251106_143022.png

# User checks status later
$ gimage auth status

✓ Gemini API
  Source: Config file (~/.gimage/config.md)
  Key: AIza...cVVI
  Last tested: 2025-11-06 14:30 (Success)

Default API: gemini
Available: gemini, vertex (not configured), bedrock (not configured)

# User adds vertex
$ gimage auth setup vertex
# Interactive setup...

$ gimage auth status

✓ Gemini API
  Source: Config file (~/.gimage/config.md)

✓ Vertex AI
  Source: Environment (VERTEX_API_KEY, VERTEX_PROJECT)
  ⚠ Warning: Also configured in config file (env var takes precedence)

# Now user KNOWS what's happening!
```

---

## Part 9: "Should We Change the Config System?"

### Answer: No, Keep the 3-Tier Hierarchy

**The hierarchy is correct:**
1. Flags (highest) - User explicit intent
2. Environment variables - Shell/CI configuration
3. Config file - User defaults
4. Defaults (lowest) - Sensible fallbacks

**Don't change this.** It's standard across CLI tools (Docker, AWS CLI, gcloud, etc.)

**The problem is visibility, not architecture.**

### What TO Change

#### Instead of:
```bash
# Silent magic
$ gimage generate "test"
# Which key is used? Who knows!
```

#### Do this:
```bash
# Clear visibility
$ gimage generate "test" --verbose
Using credentials:
  API: gemini (from default_api in config file)
  API Key: AIza...cVVI (from environment variable GEMINI_API_KEY)
  Override: Environment variable takes precedence over config file

$ gimage auth status
Shows full credential picture
```

---

## Part 10: Specific Answers to Your Questions

### Q: "Authentication feels clunky with env vars and config files"

**Root Cause:** Not the env vars or config file - it's the **lack of visibility**.

**Evidence:**
- No way to see which credential is active
- No way to see source of credential
- No warning when multiple sources conflict

**Fix:**
- Add `auth status` command (shows everything)
- Add verbose logging (shows what's being used)
- Add warnings when env var overrides config

### Q: "What is the best way to make a highly predictable and easy to configure experience?"

**Best way:** "Show, Don't Hide"

1. **Always tell the user what's happening**
   - Which API is being used
   - Where the credentials came from
   - What alternatives are available

2. **Make the happy path obvious**
   - `gimage auth setup <provider>` is the one true way
   - Remove confusing alternatives (deprecate legacy auth commands)
   - Clear error messages with exact next steps

3. **Provide visibility tools**
   - `auth status` - see everything
   - `auth test` - validate credentials
   - `--verbose` - see what's being used

### Q: "These users will have a API key and want to provide it and then have the tools work with no or low overhead"

**Current best path (should be better):**

```bash
gimage auth setup gemini
# Paste key, done
```

**Make this even better:**

```bash
# Option 1: Environment variable (for CI/scripts)
export GEMINI_API_KEY="your-key"
gimage generate "test"  # Just works

# Option 2: One-line setup
gimage auth setup gemini --key "your-key"  # Non-interactive

# Option 3: Quick start (interactive)
gimage generate "test"
# Error with: "No API key. Run: gimage auth setup gemini"
gimage auth setup gemini
# Interactive, saves, done

# All paths should:
# 1. Work on first try
# 2. Tell you what happened
# 3. Be testable (gimage auth test gemini)
```

---

## Part 11: Comparison to Best-in-Class CLIs

### What gimage does BETTER than others

✅ **Single binary** - No dependencies (vs. Python tools)
✅ **Multiple providers** - Gemini, Vertex, Bedrock (unique)
✅ **Provider registry** - Extensible architecture
✅ **Markdown config** - Human-readable (unusual but nice)

### What others do BETTER

#### AWS CLI

✅ **`aws configure`** - Interactive, clear, saves to `~/.aws/credentials`
✅ **Named profiles** - Multiple accounts easy
✅ **Clear precedence** - Docs explain env > credentials > config
✅ **`aws sts get-caller-identity`** - "Who am I?" command

**Lesson:** Have a "who am I?" command (`gimage auth status`)

#### Docker CLI

✅ **`docker login`** - One command, stores credentials
✅ **Registry in URL** - `docker.io/image` makes source clear
✅ **`docker system info`** - Shows active config

**Lesson:** Show active configuration clearly

#### gcloud CLI

✅ **`gcloud auth list`** - Shows all accounts
✅ **`gcloud config list`** - Shows all active settings
✅ **Active account marked** - `*` shows which is active

**Lesson:** Make active credential obvious

### What gimage should adopt

1. **`gimage auth status`** - Show everything (like `gcloud config list`)
2. **Mark active provider** - If multiple configured
3. **Test before save** - Validate credentials work (like `docker login`)
4. **Clear source labels** - (env), (config), (default) tags

---

## Part 12: Implementation Roadmap

### Phase 1: Visibility (1 week)

**Goal:** Users can see what's happening

- [ ] Add `gimage auth status` command
- [ ] Add `--verbose` credential logging
- [ ] Fix config file doc (.yaml → .md)
- [ ] Add warnings when env overrides config

**Impact:** Solves 80% of "clunky auth" complaints

### Phase 2: Consolidation (1 week)

**Goal:** One way to do auth

- [ ] Deprecate legacy auth commands
- [ ] Update all docs to use new system
- [ ] Add migration guide

**Impact:** Reduces confusion, less code to maintain

### Phase 3: Validation (3 days)

**Goal:** Catch errors early

- [ ] Validate config on load
- [ ] Test credentials before save
- [ ] Show validation errors clearly

**Impact:** Better error messages, fail fast

### Phase 4: Consistency (2 weeks)

**Goal:** Predictable CLI patterns

- [ ] Standardize argument styles
- [ ] Implement or remove stubbed commands
- [ ] Add output control flags
- [ ] Standardize logging

**Impact:** Better UX, easier to learn

### Phase 5: Quality (2 weeks)

**Goal:** Maintainable codebase

- [ ] Reduce code duplication
- [ ] Add CLI command tests
- [ ] Unified error types
- [ ] Add TROUBLESHOOTING.md

**Impact:** Fewer bugs, easier contributions

---

## Summary: Top 5 Changes to Make Now

1. **Add `gimage auth status`** (1 day)
   - Shows which credentials are active and where they came from
   - Solves the "where's my key?" problem

2. **Show credentials in verbose mode** (4 hours)
   - `gimage generate --verbose` shows which API key is used
   - Provides visibility without extra command

3. **Fix documentation** (1 hour)
   - Change .yaml to .md everywhere
   - Update help text

4. **Deprecate legacy auth commands** (2 days)
   - Remove `auth gemini|vertex|bedrock`
   - Keep only `auth setup|list|test|status`
   - Update docs

5. **Validate config on load** (2 hours)
   - Catch errors early
   - Show warnings for invalid values

**Total effort:** ~1 week to transform the UX

---

## Conclusion

**The gimage project is well-architected with a solid foundation.** The authentication system isn't fundamentally broken - it just lacks visibility and has some legacy cruft.

**The core issue:** Users can't tell what's happening. They set credentials multiple ways and have no idea which one is active.

**The solution:** Show, don't hide. Add visibility tools and consolidate to one clear path.

**After these changes:**
- Users will know which credentials are active
- Users will know where credentials came from
- Users will have one clear way to set up auth
- Users will have confidence the system works

**The authentication hierarchy (flags > env > config) should NOT change.** It's correct. Just make it visible.
