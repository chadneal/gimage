# GIMAGE CLI IMPLEMENTATION ANALYSIS

## Executive Summary

The gimage CLI implementation shows a mature, well-documented foundation with clear architectural patterns but contains several inconsistencies in command implementation, flag naming, and user experience. The codebase is undergoing a transition from a legacy provider system to a new unified provider registry system, which creates some redundancy and confusion.

**Key Issues Found:**
1. Incomplete command implementations (batch, compress, config)
2. Inconsistent output formatting and status messages
3. Redundant flag naming and configuration hierarchy complexity
4. UX friction in authentication flows
5. Mixed legacy and new provider systems coexisting

---

## 1. COMMAND STRUCTURE AND ORGANIZATION

### Directory Layout
- **Entry point**: `cmd/gimage/main.go` → calls `cli.Execute()`
- **Command definitions**: `internal/cli/*.go` (14 command files)
- **Root command**: `internal/cli/root.go` (orchestrates subcommands)

### Command Inventory

#### Fully Implemented Commands
1. **generate** - AI image generation (950+ lines, highly complex)
2. **resize** - Image dimension changes (70 lines)
3. **scale** - Image scaling by factor (70 lines)
4. **crop** - Image region extraction (90 lines)
5. **convert** - Image format conversion (70 lines)
6. **auth** + subcommands - Authentication setup (600+ lines)
7. **serve** - MCP server for Claude (100+ lines, partial)
8. **auth list** - Provider authentication status (245 lines)
9. **auth setup** - Provider-based auth wizard (270+ lines)
10. **auth test** - Authentication verification (170+ lines)

#### Stubbed/Incomplete Commands
- **batch** - Marked "TODO: Implement batch functionality" (empty RunE)
- **compress** - Marked "TODO: Implement compress functionality" (empty RunE)
- **config** - Marked "TODO: Implement config functionality" (empty RunE)
- **tui** - Terminal UI command file exists but minimal content

### Critical Finding: Two Auth Systems Coexist

The codebase has **two parallel authentication mechanisms**:

**Legacy System** (`auth.go`):
```go
authCmd.AddCommand(authGeminiCmd)    // gimage auth gemini
authCmd.AddCommand(authVertexCmd)    // gimage auth vertex
authCmd.AddCommand(authBedrockCmd)   // gimage auth bedrock
```

**New Provider-Based System** (`auth_setup.go`, `auth_list.go`, `auth_testcmd.go`):
```go
authCmd.AddCommand(authListCmd)      // gimage auth list
authCmd.AddCommand(authTestCmd)      // gimage auth test
authCmd.AddCommand(authSetupCmd)     // gimage auth setup
```

**UX Problem**: Users can run both `gimage auth gemini` (legacy) AND `gimage auth setup gemini` (new). This is confusing and creates maintenance burden.

---

## 2. FLAG DEFINITION AND USAGE PATTERNS

### Inconsistencies Across Commands

#### Pattern 1: Output Flag
| Command | Flag | Behavior |
|---------|------|----------|
| generate | `-o, --output` | Custom file path, auto-generates if missing |
| resize | `-o, --output` | Custom file path, auto-generates if missing |
| scale | `-o, --output` | Custom file path, auto-generates if missing |
| crop | `-o, --output` | Custom file path, auto-generates if missing |
| convert | `-o, --output` | Custom file path, auto-generates if missing |
| compress | `-o, --output` | Custom file path, auto-generates if missing |
| batch | `-o, --output` | Directory path (inconsistent with others) |

**Issue**: The `batch` command uses `--output` for a directory, while others use it for a file. This is confusing.

#### Pattern 2: Model/Provider Selection
| Command | Approach | Flag Names |
|---------|----------|-----------|
| generate (old) | Model-based | `--api`, `--model`, `--api-key` |
| generate (new) | Provider-based | `--provider` |
| All others | N/A | No model selection |

**Issue**: The `generate` command has **both** `--api`/`--model` (legacy) AND `--provider` (new) flags. The code path is:
```go
if providerID != "" {
    return runGenerateWithProvider(...)  // New system
}
// Falls back to old system
```

#### Pattern 3: Size/Dimension Arguments
| Command | Style | Format |
|---------|-------|--------|
| generate | Flag | `--size 1024x1024` |
| resize | Positional args | `resize input.jpg 800 600` |
| batch resize | Flags | `--width 800 --height 600` |

**Issue**: Inconsistent argument passing style makes learning curve steep.

#### Pattern 4: Quality Parameter
| Command | Parameter | Range | Default |
|---------|-----------|-------|---------|
| compress | `--quality` | 1-100 | 90 |
| batch compress | `--quality` | 1-100 | 90 |
| convert | N/A | N/A | N/A |

**Issue**: No quality control for format conversion, inconsistent with batch operations.

### Flag Binding to Viper

In `root.go`:
```go
viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
```

But in `generate.go`:
```go
viper.BindPFlag("generate.api", generateCmd.Flags().Lookup("api"))
viper.BindPFlag("generate.model", generateCmd.Flags().Lookup("model"))
viper.BindPFlag("generate.size", generateCmd.Flags().Lookup("size"))
```

**Issue**: Only `generate` command flags are bound to viper, so config file doesn't apply to other commands.

---

## 3. USER-FACING ERROR MESSAGES

### Quality Examples - Clear and Helpful

**Good Error #1** (from resize.go):
```
"invalid width '%s': must be a positive integer"
```

**Good Error #2** (from generate.go):
```
"prompt is required (or use --list-models or --list-providers to see available options)"
```

**Good Error #3** (from auth.go - Gemini setup):
```
"✓ Configuration saved successfully!"
"You can now use Gemini API with:"
"  gimage generate \"your prompt here\""
```

### Quality Issues - Unclear or Generic

**Poor Error #1** (from generate.go):
```go
if len(args) == 0 {
    return fmt.Errorf("prompt is required (or use --list-models or --list-providers to see available options)")
}
```
**Problem**: If someone runs `gimage generate`, they get this generic message. Better: show usage example.

**Poor Error #2** (from generate.go):
```go
return fmt.Errorf("no API credentials found. Please set up credentials using:\n" +
    "  Gemini:  gimage auth gemini\n" +
    "  Vertex:  gimage auth vertex\n" +
    "  Bedrock: gimage auth bedrock")
```
**Problem**: This tells users about the legacy `auth gemini/vertex` commands, but the new system uses `auth setup gemini`. Contradictory guidance.

**Poor Error #3** (from generate.go, auth_setup.go):
```
# In generate.go
"failed to get Gemini API key: %w\nHint: Set GEMINI_API_KEY environment variable or use --api-key flag"

# But users might also try
gimage auth gemini  # (legacy)
gimage auth setup gemini  # (new)
```
**Problem**: Three different ways suggested but unclear which is correct.

**Poor Error #4** (from config validation):
```
"unsupported Vertex AI location: %s (supported: us-central1, us-east1, us-west1, ...)"
```
**Problem**: If someone misspells a location, error is helpful. But the default is hardcoded and there's no way to list supported locations.

### Missing Context Messages

From the `generate` command (~950 lines), the verbose output is extensive but:
- No progress indication for long-running operations
- No estimated time remaining
- No token/cost calculation in real-time

Example from line 189:
```go
printVerbose("Generating image with prompt: %s", prompt)
```
No indication of: API in use, model selected, pricing, estimated time.

---

## 4. CONFIGURATION HIERARCHY AND VALIDATION

### Priority Order (Documented vs Actual)

**Documented** (root.go comments):
```
1. Command-line flags (highest)
2. Environment variables
3. Config file
4. Default values (lowest)
```

**Actual Implementation**:

For Gemini API Key (auth.go):
```go
// Priority: flagKey > env var > config file
if flagKey != "" { return flagKey }
if os.Getenv("GEMINI_API_KEY") != "" { return cfgKey }
cfg, _ := LoadConfig()
if cfg.GeminiAPIKey != "" { return cfg.GeminiAPIKey }
```
✓ Matches documented order

For model selection (generate.go):
```go
// Priority: --model flag > auto-detect > config > default
if model != "" { /* use flag */ }
else {
    detectedAPI, err := generate.DetectAPIFromModel(model)
    // But if model is empty, falls to credential auto-detect
}
```
✓ Matches documented order

For project/location (generate.go):
```go
// Uses flags, but then loads from config
project, _ := cmd.Flags().GetString("project")
if project == "" {
    project = os.Getenv("VERTEX_PROJECT")
}
if project == "" {
    cfg, _ := config.LoadConfig()
    project = cfg.VertexProject
}
```
✓ Matches documented order

### Configuration File Issues

**Location**: `~/.gimage/config.md` (markdown format)

**Format**: 
```markdown
**gemini_api_key**: AIzaSy...
**vertex_project**: my-project
```

**Issues**:
1. **Filename mismatch**: Flag says `~/.gimage/config.yaml` but actually reads `.md`
   - root.go line 179: `config file (default is $HOME/.gimage/config.yaml)`
   - root.go line 202: `viper.SetConfigName("config")`
   - config.go line 201: Expects markdown format

2. **Parsing flexibility**: Markdown format means manually editing, no structured validation during save

3. **Inconsistent key naming**:
   - Config uses: `gemini_api_key`, `vertex_project`, `aws_region`
   - But display messages use different terminology

4. **No config migration**: If format changes, old configs break silently

### Validation Issues

From auth.go, setup functions validate interactively:
```go
projectID := promptWithDefault(reader, "Google Cloud Project ID", existingCfg.VertexProject, false)
// Later, in setupVertexFullModeServiceAccount:
if _, err := os.Stat(credsPath); os.IsNotExist(err) {
    return fmt.Errorf("service account file not found: %s", credsPath)
}
```

**Problem**: 
- Vertex project validation happens only when setting up credentials
- If user manually edits config with invalid project, error only appears at generation time
- No `gimage validate` command to check config before attempting to generate

---

## 5. BACKEND SELECTION AND MODEL RESOLUTION

### Model Name Resolution Flow

The system has complex aliasing:

**Entry point** (generate.go, line 158-167):
```go
model, _ := cmd.Flags().GetString("model")
originalModel := model
if model != "" {
    model = generate.ResolveModelName(model)
    if originalModel != model {
        printVerbose("Resolved model '%s' to '%s'", originalModel, model)
    }
}
```

**ResolveModelName** (providers.go):
```go
func ResolveModelName(name string) string {
    registry := GetProviderRegistry()
    provider, err := registry.ResolveProvider(name)
    if err != nil {
        return name
    }
    return provider.ModelID
}
```

**Alias Resolution** (providers.go):
```go
aliases := map[string]string{
    "gemini":       "gemini/flash-2.5",
    "gemini-flash": "gemini/flash-2.5",
    "flash":        "gemini/flash-2.5",
    "imagen":       "vertex/imagen-4",
    "imagen-4":     "vertex/imagen-4",
    "nova":         "bedrock/nova-canvas",
    "nova-canvas":  "bedrock/nova-canvas",
}
```

### Supported Aliases

| User Input | Resolved To | Provider | API |
|-----------|-------------|----------|-----|
| `gemini` | `gemini/flash-2.5` | Gemini 2.5 Flash | gemini |
| `imagen` | `vertex/imagen-4` | Imagen 4 | vertex |
| `nova` | `bedrock/nova-canvas` | Nova Canvas | bedrock |

**UX Issue**: These aliases are not documented in help text or error messages. User discovers them by trial-and-error or reading code.

### API Auto-Detection Logic

From generate.go (lines 205-264):

```
Priority:
1. --api flag (explicit)
2. Auto-detect from model name
3. Check for available credentials:
   - Count how many APIs have credentials
   - If 1: use that one
   - If >1: check default_api from config, fall back to "gemini"
   - If 0: show error with setup instructions
```

**Issue**: If user has both Gemini and Vertex configured and doesn't specify `--api`, it silently uses "gemini". No warning that Vertex is available.

Example:
```bash
$ GEMINI_API_KEY=... VERTEX_PROJECT=... gimage generate "test"
# Uses Gemini silently, even though Vertex is available
# No message: "Both APIs available, using Gemini. Use --api vertex for Vertex AI"
```

---

## 6. UX FRICTION POINTS AND INCONSISTENCIES

### Friction Point 1: Authentication Setup Complexity

**Path A** (Legacy):
```bash
gimage auth gemini
gimage auth vertex
gimage auth bedrock
```

**Path B** (New):
```bash
gimage auth setup gemini/flash-2.5
gimage auth setup vertex/imagen-4
gimage auth setup bedrock/nova-canvas
```

**Problem**: 
- No indication which is "correct"
- Both exist in codebase, both work
- Documentation mentions legacy style
- Error messages suggest legacy style
- But new style is more flexible

### Friction Point 2: Size Format Inconsistency

Command inconsistency:
```bash
# Generate - uses flag with WxH format
gimage generate "sunset" --size 1024x1024

# Resize - uses positional args with separate dimensions
gimage resize image.jpg 1024 1024

# Batch - uses separate flags
gimage batch resize . --width 1024 --height 1024
```

**Impact**: Users can't reuse mental model across commands.

### Friction Point 3: Missing Required Positional Args

Resize command requires exactly 3 args:
```go
Args: cobra.ExactArgs(3),
```

Error if wrong:
```bash
$ gimage resize image.jpg 800
# Error: "accepts 3 arg(s), received 1"
```

Better error:
```
Error: resize requires 3 arguments: [input] [width] [height]
Usage: gimage resize input.jpg 800 600
```

### Friction Point 4: No Completion/Validation

Commands don't offer:
- Shell completion (`bash`, `zsh`, `fish`)
- Tab-completable values for flags
- Validation of image formats before processing

Batch command example:
```bash
$ gimage batch invalidoperation ./images
# No error until runtime discovers the operation type
```

### Friction Point 5: Output Path Behavior

Auto-generated filenames are inconsistent:

**resize.go**:
```go
outputPath = fmt.Sprintf("%s_resized_%dx%d%s", base, width, height, ext)
// Result: photo_resized_800x600.jpg
```

**scale.go**:
```go
outputPath = fmt.Sprintf("%s_scaled_%.2fx%s", base, factor, ext)
// Result: photo_scaled_0.50x.jpg (note: second 'x' is literal)
```

**convert.go**:
```go
outputPath = fmt.Sprintf("%s_converted.%s", base, targetFormat)
// Result: photo_converted.jpg
```

**crop.go**:
```go
outputPath = fmt.Sprintf("%s_cropped_%dx%d%s", base, width, height, ext)
// Result: photo_cropped_800x600.jpg
```

**Problem**: Different naming schemes, hard to batch process multiple operations.

### Friction Point 6: Verbose Mode Not Consistent

Only `printVerbose()` respects verbose flag in image processing commands, but detailed output is printed to stdout (not stderr):

```go
fmt.Printf("Resizing %s to %dx%d...\n", inputPath, width, height)
fmt.Printf("✓ Resized successfully!\n")
```

**Problem**: Can't suppress normal output, only enable verbose logging. Makes command output hard to script.

---

## 7. FLOW FROM CLI INVOCATION TO IMAGE GENERATION

### Generate Command Flow

```
User: gimage generate "sunset" --model imagen-4
  ↓
cli.Execute() [root.go]
  ↓
generateCmd.RunE = runGenerate [generate.go:150]
  ↓
Pre-validation checks:
  - If --list-models or --list-providers: show and exit ✓
  - If no prompt and not listing: error ✓
  - Validate size format: ✓
  ↓
Model name resolution:
  - "imagen-4" → ResolveModelName()
  - Lookup in providers registry → "vertex/imagen-4" → "imagen-4.0-generate-001"
  ↓
API selection logic (lines 205-264):
  - No --api flag provided
  - Model name implies vertex API
  - Set selectedAPI = "vertex" ✓
  ↓
Provider info lookup:
  - registry.ResolveProvider("imagen-4.0-generate-001")
  - Shows pricing: $0.04/image
  - Shows: "Using: Imagen 4 (vertex API)"
  ↓
Credential gathering:
  - Check for VERTEX_API_KEY env var
  - If not found, check config file
  - If found, use Express Mode (REST)
  - If not found, use Full Mode (SDK with service account)
  ↓
Client creation:
  - NewVertexRESTClient() or NewVertexSDKClient()
  ↓
Image generation:
  - client.GenerateImage(ctx, prompt, options)
  ↓
Output handling:
  - If --output not specified: auto-generate filename
  - Save image: generate.SaveImage()
  ↓
Success output:
  - printSuccess() - shows checkmark
  - File size, dimensions, cost
```

### Key Decision Points Missing

1. **No model capability check before generation**
   - Example: User requests negative prompts with Bedrock, but tries to use it with a model that doesn't support it
   - Error only appears after credentials are loaded and API called

2. **No size validation against model capabilities**
   - Example: Some models have max sizes (1024x1024), others (2048x2048)
   - Would be useful to validate before API call

3. **No remaining quota/rate limit checking**
   - Could prevent wasted API calls

---

## 8. CODE QUALITY OBSERVATIONS

### Strengths
1. **Error wrapping**: Uses `fmt.Errorf` with `%w` consistently
2. **Documentation**: Root command has extensive examples
3. **Configuration hierarchy**: Clearly documented and implemented
4. **Provider registry**: Clean, extensible architecture
5. **ANSI colors**: Proper use of colors for status output

### Weaknesses
1. **Code duplication**:
   - Auth logic duplicated in legacy (`auth.go`) and new (`auth_setup.go`) systems
   - Credential gathering duplicated across multiple functions
   - Model lookup duplicated

2. **Mixed concerns**:
   - CLI commands (in `cli/`) also handle business logic (auth, model lookup)
   - Should be separated into separate packages

3. **Type-checking**:
   - Silent defaults for invalid values: `cmd.Flags().GetString()` returns empty string if flag not found, no error
   - Should validate flag types/values more strictly

4. **Testing**:
   - No unit tests found for CLI commands
   - No integration tests for CLI flows
   - Error messages not tested

---

## 9. RECOMMENDATIONS FOR IMPROVEMENT

### Priority 1: Fix Critical Issues

1. **Complete Stubbed Commands**
   - Implement `batch` command (currently empty TODO)
   - Implement `compress` command (currently empty TODO)
   - Implement `config` command (currently empty TODO)

2. **Consolidate Authentication Systems**
   - Keep new provider-based system (`auth setup`, `auth list`, `auth test`)
   - Remove legacy commands (`auth gemini`, `auth vertex`, `auth bedrock`)
   - Update all error messages and documentation to use new system only

3. **Fix Config File Documentation**
   - Update root.go to say `~/.gimage/config.md` instead of `.yaml`
   - Or rename actual file to `.yaml` if possible

### Priority 2: Improve User Experience

1. **Add Command-Line Validation**
   - Validate image format support before processing
   - Validate size requirements against model capabilities
   - Show helpful error for invalid/misspelled arguments

2. **Standardize Output Format**
   - All commands should output to stderr (not stdout)
   - Use consistent naming for auto-generated files
   - Add `--quiet` flag to suppress status messages
   - Add `--json` flag for machine-readable output

3. **Document All Aliases**
   - Add hidden `--list-models` equivalent for each command
   - Show alias examples in help text
   - Document supported model names in COMMANDS.md

4. **Add Progress Indication**
   - For long operations, show progress bars
   - Estimate time remaining
   - Show real-time token/cost calculation

### Priority 3: Consistency Improvements

1. **Standardize Argument Passing**
   - Use flags for all optional parameters
   - Use positional args only for essential input
   - Consider: `resize` should be `resize --input image.jpg --width 800 --height 600` for consistency

2. **Unify Size/Dimension Parameters**
   - All commands should accept `--size WxH` format
   - Convert to separate width/height internally if needed

3. **Add Configuration Validation**
   - Add `gimage config validate` command
   - Check config before attempting operations
   - Suggest fixes for common problems

4. **Improve Error Messages**
   - All errors should include: what went wrong, why, and how to fix
   - Show relevant documentation links
   - Suggest alternative commands if typo detected

---

## APPENDIX: FILE REFERENCE

### CLI Files
- `/home/user/gimage/cmd/gimage/main.go` - Entry point
- `/home/user/gimage/internal/cli/root.go` - Root command, flag setup
- `/home/user/gimage/internal/cli/generate.go` - Image generation (950 lines)
- `/home/user/gimage/internal/cli/resize.go` - Image resizing
- `/home/user/gimage/internal/cli/scale.go` - Image scaling
- `/home/user/gimage/internal/cli/crop.go` - Image cropping
- `/home/user/gimage/internal/cli/convert.go` - Image format conversion
- `/home/user/gimage/internal/cli/batch.go` - Batch operations (stubbed)
- `/home/user/gimage/internal/cli/compress.go` - Image compression (stubbed)
- `/home/user/gimage/internal/cli/config.go` - Config management (stubbed)
- `/home/user/gimage/internal/cli/auth.go` - Legacy auth (600+ lines)
- `/home/user/gimage/internal/cli/auth_list.go` - Provider list (245 lines)
- `/home/user/gimage/internal/cli/auth_setup.go` - Provider setup (270+ lines)
- `/home/user/gimage/internal/cli/auth_testcmd.go` - Auth testing (170 lines)
- `/home/user/gimage/internal/cli/serve.go` - MCP server (100+ lines)
- `/home/user/gimage/internal/cli/tui.go` - Terminal UI (minimal)

### Config/Provider Files
- `/home/user/gimage/internal/config/config.go` - Config loading/saving
- `/home/user/gimage/internal/config/auth.go` - Auth validation
- `/home/user/gimage/internal/generate/providers.go` - Provider registry (679 lines)

### Documentation
- `/home/user/gimage/COMMANDS.md` - Command reference (200+ lines)
- `/home/user/gimage/README.md` - Project overview
- `/home/user/gimage/CLAUDE.md` - Development guidelines

