# MCP Tool Description Improvements - Changelog

**Date**: 2025-11-02
**Based on**: LLM Learning Analysis (MCP_LLM_LEARNING_ANALYSIS.md)

## Overview

Enhanced MCP tool descriptions based on analysis of how LLMs learn to use gimage. The analysis showed that while the current implementation is excellent (8.5/10), adding concrete examples and improved guidance would improve first-try success rates.

## Changes Made

### 1. generate_image Tool (`internal/mcp/tools/generate.go`)

#### Main Description Enhancement
**Before**:
```
Generate an AI image from a text prompt using Gemini or Vertex AI. Supports multiple models...
```

**After**:
```
Generate an AI image from a text prompt using Gemini, Vertex AI, or AWS Bedrock.
Quick start: generate_image(prompt='sunset over mountains', output='~/Desktop/sunset.png')
uses the default free model (Gemini 2.5 Flash, 1024x1024). For higher quality, use
model='imagen-4' (paid, requires Vertex AI). Supports various sizes up to 2048x2048...
```

**Why**: Provides immediate "quick start" example showing simplest usage pattern, which was the #1 recommendation from the analysis.

---

#### Model Parameter Enhancement
**Before**:
```json
"description": "AI model to use. Supports exact names (e.g., gemini-2.5-flash-image) or aliases (e.g., gemini, flash). Default: gemini-2.5-flash-image (free, fast). imagen-4 offers highest quality but requires Vertex AI. If model not found, automatically falls back to gemini-2.5-flash-image."
```

**After**:
```json
"description": "AI model to use. Supports exact names or aliases. Common aliases: 'gemini' or 'gemini-flash' for gemini-2.5-flash-image (default, FREE up to 1500/day, supports up to 1024x1024), 'imagen' or 'imagen-4' for imagen-4.0-generate-001 (paid $0.02-0.04/image, highest quality, supports up to 2048x2048), 'nova-canvas' for amazon.nova-canvas-v1:0 (paid $0.04-0.08/image, supports up to 1408x1408). Invalid model names automatically fall back to default. Examples: 'gemini' (quick iterations), 'imagen-4' (final high-quality output)."
```

**Why**:
- Adds cost information (FREE vs paid, exact pricing)
- Documents model-specific size limitations
- Provides concrete usage examples
- Addresses LLM confusion about model names in transcript

---

#### Size Parameter Enhancement
**Before**:
```json
"description": "Image dimensions. Default is 1024x1024. Larger sizes available with Vertex AI."
```

**After**:
```json
"description": "Image dimensions (WIDTHxHEIGHT). Default: 1024x1024. Gemini supports up to 1024x1024. Larger sizes (1792x1024, 2048x2048) require Vertex AI with imagen-4. Examples: '1024x1024' (square), '1792x1024' (16:9 landscape), '1024x1792' (9:16 portrait), '2048x2048' (ultra HD)."
```

**Why**:
- Clarifies format (WIDTHxHEIGHT)
- Documents model-specific limitations
- Provides use case examples (16:9, portrait, etc.)
- Addresses LLM confusion when images came out wrong size

---

### 2. crop_image Tool (`internal/mcp/tools/crop.go`)

#### Main Description Enhancement
**Before**:
```
Crop an image to a specific rectangular region. Specify the top-left corner coordinates (x, y) and the width and height of the region to extract. Useful for removing unwanted borders, focusing on specific areas, or extracting thumbnails from larger images.
```

**After**:
```
Crop an image to a specific rectangular region. Specify coordinates and dimensions to extract.
Example: crop_image(input='photo.png', x=0, y=100, width=800, height=600, output='cropped.png')
extracts an 800x600 region starting at position (0,100). IMPORTANT: All parameters (x, y, width, height)
are positional integers, not flags. Coordinates start at (0,0) in the top-left corner. Useful for creating
hero images, removing borders, or focusing on specific areas. TIP: Use get_image_info first to check actual
image dimensions before cropping.
```

**Why**:
- Adds concrete example with actual values
- Explicitly clarifies positional parameters (not flags) - addresses LLM error in transcript
- Provides workflow tip (check dimensions first)
- Addresses LLM confusion about crop syntax

---

## Learning from the Transcript

### Key Issues Identified:

1. **Model Name Confusion**: LLM tried `gemini-flash` instead of `gemini-2.5-flash-image`
   - **Fixed**: Enhanced description now shows common aliases with exact mappings
   - **Result**: LLM can see that 'gemini' or 'gemini-flash' both map to the exact name

2. **Size Constraints**: LLM assumed 1792x1024 would work but got 1024x1024
   - **Fixed**: Size description now documents model-specific limitations
   - **Result**: LLM knows Gemini max is 1024x1024, larger sizes need Imagen

3. **Crop Positional Args**: LLM tried `--width` and `--height` flags
   - **Fixed**: Description explicitly states "positional integers, not flags"
   - **Result**: LLM knows to use `crop_image(input, x, y, width, height)` not flags

4. **Missing Workflow Tips**: LLM didn't check image dimensions before cropping
   - **Fixed**: Added tip to use get_image_info first
   - **Result**: LLM has guidance on proper workflow

---

## Impact Assessment

### Expected Improvements:

1. **First-Try Success Rate**:
   - Before: 0/1 (model name error on first attempt)
   - Expected After: ~80% (with quick-start examples and clear aliases)

2. **Learning Time**:
   - Before: 3 attempts (~2 minutes) to proficiency
   - Expected After: 1-2 attempts (~1 minute)

3. **Error Types Reduced**:
   - ✅ Model name mismatches (gemini-flash → gemini-2.5-flash-image)
   - ✅ Size constraint violations (requesting unsupported sizes)
   - ✅ Crop syntax errors (using flags instead of positional args)

---

## What Was NOT Changed

The analysis identified these aspects as already excellent:

- ✅ **Error messages** - Clear, actionable, educational
- ✅ **Self-discovery tools** - `--list-models`, `--help` work great
- ✅ **Alias support** - Model name aliases already implemented in code
- ✅ **Fallback behavior** - Automatic fallback to default model on errors
- ✅ **MCP schema structure** - Well-organized with proper types and constraints

---

## Future Enhancements

From the analysis, these improvements could be added later:

### Priority 2: Common Workflows
Add workflow examples to tool descriptions:
```json
"common_workflows": [
  {
    "name": "Generate and optimize for web",
    "steps": [
      "generate_image(prompt='...', output='raw.png')",
      "resize_image(input='raw.png', width=800, height=600, output='resized.png')",
      "convert_image(input='resized.png', format='webp', output='optimized.webp')"
    ]
  }
]
```

### Priority 3: Troubleshooting Hints
Add troubleshooting section to tool descriptions:
```json
"troubleshooting": {
  "model_not_found": "Use --list-models to see available models. Try 'gemini' (free) or 'imagen-4' (paid).",
  "image_too_large": "Gemini supports up to 1024x1024. For larger sizes, use model='imagen-4'.",
  "invalid_dimensions": "Use get_image_info to check actual image dimensions before cropping."
}
```

---

## Testing

All changes have been verified:

```bash
✅ Build successful: make build
✅ All tests pass: 128/128 unit tests + 6/6 CLI E2E tests + 4/4 Generate E2E tests
✅ Code compiles without errors
✅ No breaking changes to existing functionality
```

---

## References

- **Analysis Document**: `docs/MCP_LLM_LEARNING_ANALYSIS.md`
- **Transcript Source**: Claude Code session using gimage MCP server
- **Code Changes**:
  - `internal/mcp/tools/generate.go:19` (description)
  - `internal/mcp/tools/generate.go:39` (size parameter)
  - `internal/mcp/tools/generate.go:55` (model parameter)
  - `internal/mcp/tools/crop.go:16` (description)

---

**Status**: ✅ Complete
**Next Steps**: Monitor LLM usage to validate improvements reduce first-try errors
