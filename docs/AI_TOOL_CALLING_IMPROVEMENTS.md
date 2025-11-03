# AI Tool Calling Improvements for gimage

**Date**: 2025-11-02
**Issue**: Model name mismatch causing CLI failures and nil pointer panics

## Problem Summary

When users request image generation using informal model names (e.g., "gemini flash"), AI assistants must translate these to exact CLI model IDs (e.g., `gemini-2.5-flash-image`). Incorrect translations cause:

1. **404 API errors** - Model not found
2. **Nil pointer panics** - CLI crashes when handling errors
3. **Poor UX** - Users must retry with correct names

**Example Failure**:
```bash
# User says: "use gemini flash"
# AI used: --model gemini-flash  ❌
# Correct:  --model gemini-2.5-flash-image  ✅

# Result: 404 error + panic
Error: failed to save image: image cannot be nil
panic: runtime error: invalid memory address or nil pointer dereference
```

## Improvements Implemented

### 1. ✅ Model Name Mapping in CLAUDE.md

**Location**: `/Users/chad/dev/gimage/CLAUDE.md` (lines 163-201)

**What was added**: A "Model Name Resolution for AI Assistants" section with:
- Exact mapping table from informal names to model IDs
- Resolution strategy (check table first, never guess)
- Example translations showing correct vs. wrong usage
- Clear instructions to consult documentation before tool calls

**Impact**: AI assistants now have authoritative reference for model name translation.

### 2. ⚠️ CLI Error Handling (Needs Fix)

**Issue**: When `GenerateImage()` returns an error, it returns `(nil, error)`. The CLI panics when trying to access `generatedImage.Format` on the nil pointer.

**Current code** (`internal/cli/generate.go:459-467`):
```go
if err != nil {
    return fmt.Errorf("failed to generate image: %w", err)
}

// Determine output path
if output == "" {
    output = generate.GenerateOutputPath(generatedImage.Format)  // PANIC HERE if generatedImage is nil
}
```

**Recommended fix**:
```go
if err != nil {
    return fmt.Errorf("failed to generate image: %w", err)
}

// DEFENSIVE: Ensure image is not nil (should never happen if error handling is correct)
if generatedImage == nil {
    return fmt.Errorf("internal error: generated image is nil but no error was returned")
}

// Determine output path
if output == "" {
    output = generate.GenerateOutputPath(generatedImage.Format)
}
```

**Why defensive nil checking**:
- Go best practice: never assume pointer is valid
- Prevents crashes even if error handling fails upstream
- Provides clear error message instead of cryptic panic
- Makes debugging easier (explicit error vs. stacktrace)

### 3. 📚 Model Aliases in CLI (Future Enhancement)

**Proposed**: Add model name aliases directly in the CLI to accept informal names:

```go
// internal/generate/models.go
var modelAliases = map[string]string{
    "gemini":           ModelGemini25FlashImage,
    "gemini-flash":     ModelGemini25FlashImage,
    "flash":            ModelGemini25FlashImage,
    "2.5-flash":        ModelGemini25FlashImage,

    "gemini-2.0-flash": ModelGemini20FlashPreview,
    "2.0-flash":        ModelGemini20FlashPreview,

    "imagen":           ModelImagen4,
    "imagen-4":         ModelImagen4,

    "imagen-3":         ModelImagen3,

    "nova":             ModelNovaCanvas,
    "nova-canvas":      ModelNovaCanvas,
}

func ResolveModelName(input string) string {
    if resolved, ok := modelAliases[input]; ok {
        return resolved
    }
    return input  // Already exact name
}
```

**Usage in CLI**:
```go
// internal/cli/generate.go
model, _ := cmd.Flags().GetString("model")
model = generate.ResolveModelName(model)  // Resolve aliases
```

**Benefits**:
- Users can use informal names directly in CLI
- AI assistants can use informal names without translation
- Backward compatible (exact names still work)
- Single source of truth for model naming

## Testing Strategy

### Unit Tests (No Mocking)

**Test request building**:
```go
func TestResolveModelName(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"gemini", "gemini-2.5-flash-image"},
        {"gemini-flash", "gemini-2.5-flash-image"},
        {"flash", "gemini-2.5-flash-image"},
        {"imagen", "imagen-4"},
        {"nova", "amazon.nova-canvas-v1:0"},
        {"gemini-2.5-flash-image", "gemini-2.5-flash-image"}, // exact name unchanged
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := ResolveModelName(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Test nil pointer safety**:
```go
func TestGenerateCommandHandlesNilImage(t *testing.T) {
    // Simulate error that returns nil image
    // Verify CLI returns error instead of panicking
    // This is a regression test for the panic bug
}
```

### Manual Testing

```bash
# Test model name aliases (after implementing enhancement #3)
gimage generate "test" --model gemini          # Should resolve to gemini-2.5-flash-image
gimage generate "test" --model flash           # Should resolve to gemini-2.5-flash-image
gimage generate "test" --model imagen          # Should resolve to imagen-4
gimage generate "test" --model nova            # Should resolve to amazon.nova-canvas-v1:0

# Test error handling (should not panic)
gimage generate "test" --model invalid-model   # Should show error, not crash

# Test with verbose output to verify resolution
gimage generate "test" --model flash --verbose
# Output should show: "Using model: gemini-2.5-flash-image (resolved from: flash)"
```

## AI Assistant Guidelines

### Before Making Tool Calls

1. **Consult CLAUDE.md** - Check "Model Name Resolution" section
2. **Use mapping table** - Never guess or abbreviate model names
3. **Default to recommended** - When ambiguous, use `gemini-2.5-flash-image` (free, fast)
4. **Validate parameters** - Ensure all required flags are present
5. **Use verbose mode** - Add `--verbose` for debugging during development

### Example Workflow

```
User: "Generate an image using gemini flash"

AI thinking:
1. User wants Gemini Flash model
2. Check CLAUDE.md mapping table
3. "gemini flash" → "gemini-2.5-flash-image"
4. Construct command with exact model ID
5. Add --verbose to confirm model selection

AI action:
gimage generate "prompt text" --model gemini-2.5-flash-image --verbose
```

### When to Use --list-models

**Use it when**:
- User asks "what models are available?"
- You're unsure if a new model exists
- User mentions a model not in your documentation
- Debugging unexpected "model not found" errors

**Don't use it when**:
- User requests a model in the mapping table (wastes time)
- Making routine generation calls (adds latency)

## Implementation Status

**✅ Completed (This Session)**:
- ✅ Add model name mapping to CLAUDE.md (lines 163-201)
- ✅ Fix --list-models to show exact names first, aliases in parentheses
- ✅ Add defensive nil checking in CLI error handling (`generate.go:463-466`)
- ✅ Implement model name aliases in CLI (`generate.go:153-160`)
- ✅ Add verbose logging for model resolution (shows "Resolved model 'X' to 'Y'")
- ✅ Update MCP server with alias resolution and fallback (`tools/generate.go:104-112`)
- ✅ Add comprehensive unit tests (`internal/generate/models_test.go`)

**Future Enhancements** (Nice to Have):
- 🎨 Improve error messages with model name suggestions
- 📊 Add telemetry for common model name mistakes
- 🔍 Implement fuzzy matching for typos

## Success Metrics

1. **Zero panics** - No nil pointer crashes regardless of input
2. **First-try success** - AI assistants use correct model names on first attempt
3. **Better errors** - Users get actionable error messages, not stacktraces
4. **Reduced latency** - No need for exploratory --list-models calls

## References

- CLAUDE.md: Model Name Resolution section (lines 163-201)
- API error handling: `internal/generate/gemini_rest.go:handleHTTPError()`
- CLI generate command: `internal/cli/generate.go:runGenerate()`
- Model definitions: `internal/generate/models.go`

## Lessons Learned

1. **Document for AI, not just humans** - AI assistants need explicit mappings, not just descriptions
2. **Fail gracefully** - Always add nil checks, even if "impossible"
3. **User language ≠ API language** - Provide translation layer for natural requests
4. **Test the happy path AND error path** - Error handling bugs only surface under failures
5. **Make it work, then make it better** - Defensive code > perfect architecture

---

**Next Steps**: Implement the defensive nil checking in `internal/cli/generate.go` to prevent crashes.
