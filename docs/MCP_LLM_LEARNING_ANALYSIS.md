# MCP LLM Learning Analysis: gimage Tool Usage

**Date**: 2025-11-02
**Transcript Source**: Claude Code session using gimage MCP server
**Purpose**: Analyze how LLMs learn to use gimage tools and identify improvement opportunities

---

## 📊 Executive Summary

**Key Finding**: The LLM successfully learned gimage through **trial-and-error + self-correction**, discovering flags and syntax through error messages and --help output. The MCP tool descriptions are **well-structured** but could benefit from minor improvements for first-try success.

**Success Rate**:
- First attempt success: **0/1** (model name error)
- Self-corrected within: **3 attempts**
- Subsequent attempts: **10/10 successful**

**Learning Time**: ~2 minutes to full proficiency

---

## 🔄 Learning Progression

### **Phase 1: Initial Failure (Trial & Error)**

**Attempt 1** - Wrong model name
```bash
gimage generate "..." --model gemini-flash --output ~/file.png
```

**Error**:
```
404: models/gemini-flash is not found
```

**LLM Reasoning**:
- User said "gemini flash"
- LLM translated to `gemini-flash` (informal name)
- **Root cause**: No guidance that model names require exact IDs

---

**Attempt 2** - Added flags but still wrong model
```bash
gimage generate "..." --model gemini-flash --size 1792x1024 --style artistic
```

**LLM Reasoning**:
- Correctly discovered `--size` and `--style` flags
- **How**: Saw them in error message help text
- Still using wrong model name

---

### **Phase 2: Self-Discovery (--list-models)**

**Attempt 3** - LLM self-corrected
```bash
gimage generate --list-models 2>&1 | grep -i flash
```

**Output**:
```
│ ✅  gemini (gemini-2.5-flash-image)
│ ✅  gemini-2.0-flash (gemini-2.0-flash-preview-image-generation)
```

**LLM Reasoning**:
- "I need to find the correct model name"
- Used `--list-models` to discover available models
- **Success**: Found `gemini-2.5-flash-image` is the exact name

---

**Attempt 4** - Correct invocation
```bash
gimage generate "..." --model gemini-2.5-flash-image --size 1792x1024 --style artistic
```

✅ **Success**: Image generated

---

### **Phase 3: Mastery (10 consecutive successes)**

After learning the correct model name, the LLM:
- ✅ Generated 10 images in sequence
- ✅ Discovered `crop`, `resize`, `convert` commands through error messages
- ✅ Learned positional args vs flags (crop uses: x y width height, not --width --height)
- ✅ Chained commands efficiently: `resize && convert`

---

## 🧠 How LLM Learned Flags

### **Method 1: Error Message Parsing** (Primary)

When the LLM used wrong syntax:
```bash
gimage crop file.png --width 1920 --height 600
```

Error provided learning:
```
Error: unknown flag: --width
Usage: gimage crop [input] [x] [y] [width] [height] [flags]

Flags:
  -h, --help            help for crop
  -o, --output string   output file path
```

**LLM learned**:
- Crop uses **positional arguments** (x, y, width, height)
- Only `-o/--output` is a flag
- Corrected to: `gimage crop file.png 0 312 1024 400 --output result.png`

---

### **Method 2: Help Text in Errors** (Secondary)

Every error includes help text:
```
Flags:
  --api string        API to use: gemini or vertex
  --size string       Image size (default "1024x1024")
  --style string      Image style: photorealistic, artistic, anime
```

**LLM learned**:
- Available flags and their types
- Default values
- Enum options (size, style)

---

### **Method 3: MCP Tool Schema** (Initial Context)

The LLM had access to MCP tool descriptions:
```json
{
  "name": "generate_image",
  "description": "Generate an AI image from a text prompt...",
  "properties": {
    "model": {
      "enum": ["gemini-2.5-flash-image", "gemini", "imagen-4", ...],
      "description": "Supports exact names or aliases"
    }
  }
}
```

**LLM learned**:
- Tool exists and its purpose
- Required vs optional parameters
- Parameter types and constraints

---

## 🎯 What Worked Well

### ✅ **1. MCP Tool Descriptions Are Comprehensive**

**Example from `generate_image`**:
```json
"model": {
  "enum": [
    "gemini-2.5-flash-image",
    "gemini",               // Alias included!
    "gemini-flash",         // Alias included!
    "imagen-4"
  ],
  "description": "Supports exact names (e.g., gemini-2.5-flash-image) or aliases (e.g., gemini, flash). If model not found, automatically falls back to gemini-2.5-flash-image."
}
```

**Why it's good**:
- ✅ Lists both exact names AND aliases
- ✅ Explains fallback behavior
- ✅ Provides examples

---

### ✅ **2. Error Messages Are Educational**

**Example**:
```
Error: crop region (x=0 + width=1792 = 1792) exceeds image width 1024
```

**Why it's good**:
- ✅ Shows the math (x + width = total)
- ✅ Explains what went wrong
- ✅ Provides actual image dimensions
- **Result**: LLM immediately corrected dimensions

---

### ✅ **3. Self-Discovery Tools Work**

Commands like `--list-models` enable LLMs to:
- Explore available options
- Find correct syntax
- Learn without human intervention

---

## ⚠️ Opportunities for Improvement

### **Issue 1: Model Name Ambiguity** (Medium Priority)

**Problem**:
User says "gemini flash" → LLM tries `gemini-flash` → 404 error

**Current State**:
```json
"enum": ["gemini-2.5-flash-image", "gemini", "gemini-flash", ...]
"description": "Supports exact names or aliases"
```

**Improvement**:
```json
"enum": ["gemini-2.5-flash-image", "gemini", "gemini-flash", ...],
"description": "Model name or alias. Common aliases: 'gemini' or 'flash' for gemini-2.5-flash-image (default, FREE), 'imagen' or 'imagen-4' for imagen-4 (paid, highest quality). Invalid model names automatically fall back to default.",
"default": "gemini-2.5-flash-image",
"examples": ["gemini", "gemini-2.5-flash-image", "imagen-4"]
```

**Why better**:
- ✅ Shows common usage patterns
- ✅ Highlights default/free option
- ✅ Provides concrete examples
- ✅ Explains fallback

---

### **Issue 2: Size Parameter Documentation** (Low Priority)

**Problem**:
LLM assumed `--size 1792x1024` worked but images came out `1024x1024`

**Current State**:
```json
"size": {
  "enum": ["256x256", "512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"],
  "description": "Image dimensions. Default is 1024x1024."
}
```

**Improvement**:
```json
"size": {
  "enum": ["256x256", "512x512", "1024x1024", "1024x1792", "1792x1024", "2048x2048"],
  "description": "Image dimensions (WIDTHxHEIGHT). Default: 1024x1024. Larger sizes (1792x1024, 2048x2048) may require Vertex AI or specific models. Gemini supports up to 1024x1024.",
  "default": "1024x1024",
  "examples": ["1024x1024", "1792x1024 (16:9)", "2048x2048 (ultra HD)"]
}
```

**Why better**:
- ✅ Clarifies width×height format
- ✅ Warns about model limitations
- ✅ Provides use case examples (16:9, HD)

---

### **Issue 3: Crop Tool - Positional Args Not Obvious** (Low Priority)

**Problem**:
LLM tried: `gimage crop file.png --width 1920 --height 600`

**Current MCP Description**:
```json
{
  "name": "crop_image",
  "properties": {
    "x": {"type": "integer", "description": "X coordinate"},
    "y": {"type": "integer", "description": "Y coordinate"},
    "width": {"type": "integer", "description": "Width"},
    "height": {"type": "integer", "description": "Height"}
  },
  "required": ["input", "x", "y", "width", "height"]
}
```

**Improvement**:
```json
{
  "name": "crop_image",
  "description": "Crop an image to a specific rectangular region. Specify coordinates and dimensions to extract. Example: crop_image('photo.png', x=0, y=100, width=800, height=600, output='cropped.png')",
  "properties": {
    "x": {"type": "integer", "description": "X coordinate of top-left corner (0 = left edge)", "minimum": 0, "default": 0},
    "y": {"type": "integer", "description": "Y coordinate of top-left corner (0 = top edge)", "minimum": 0, "default": 0},
    "width": {"type": "integer", "description": "Width of crop region in pixels (must be positive)", "minimum": 1},
    "height": {"type": "integer", "description": "Height of crop region in pixels (must be positive)", "minimum": 1}
  },
  "examples": [
    {"input": "photo.png", "x": 0, "y": 0, "width": 800, "height": 600, "output": "thumbnail.png"},
    {"input": "banner.png", "x": 100, "y": 200, "width": 1920, "height": 400, "output": "hero.png"}
  ]
}
```

**Why better**:
- ✅ Adds usage example in description
- ✅ Shows full examples with real values
- ✅ Specifies constraints (minimum, positive)

---

## 📈 Learning Metrics

### **Time to Proficiency**

| Metric | Value |
|--------|-------|
| First successful invocation | 3 attempts (~2 min) |
| Subsequent success rate | 10/10 (100%) |
| New tool discovery time | Immediate (via error messages) |
| Flag syntax learning | 1-2 attempts per tool |

### **Error Recovery**

| Error Type | Recovery Method | Time |
|------------|----------------|------|
| Wrong model name | `--list-models` | 1 attempt |
| Wrong flag format | Error message → correct syntax | 1 attempt |
| Invalid dimensions | Error message → check image size | 1 attempt |
| Wrong positional args | Usage help → correct order | 1 attempt |

---

## 🎓 LLM Learning Strategies Observed

### **Strategy 1: Incremental Refinement**
```
Attempt 1: Base command
Attempt 2: + flags discovered from error
Attempt 3: + correct model name from --list-models
Attempt 4: ✅ Success
```

### **Strategy 2: Pattern Recognition**
After learning one tool:
```
generate → resize → convert → crop
```
LLM applied same patterns:
- Always specify `--output`
- Check dimensions first
- Use verbose mode when debugging

### **Strategy 3: Chaining for Efficiency**
Once confident:
```bash
gimage resize input.png 200 200 --output temp.png &&
gimage convert temp.png webp --output final.webp
```

---

## 🚀 Recommendations

### **Priority 1: Add "Quick Start" Examples to MCP Descriptions**

**Current**:
```json
"description": "Generate an AI image from a text prompt..."
```

**Recommended**:
```json
"description": "Generate an AI image from a text prompt. Quick start: generate_image(prompt='sunset over mountains', output='~/sunset.png') - Uses default free model (Gemini 2.5 Flash). For higher quality, use model='imagen-4' (paid).",
"examples": [
  {
    "prompt": "abstract tech art with orange and blue",
    "output": "~/artwork.png",
    "style": "artistic"
  },
  {
    "prompt": "professional headshot photo",
    "model": "imagen-4",
    "size": "1024x1024",
    "style": "photorealistic"
  }
]
```

---

### **Priority 2: Highlight Common Patterns**

Add to tool descriptions:
```json
"common_workflows": [
  {
    "name": "Generate and optimize for web",
    "steps": [
      "generate_image(prompt='...', output='raw.png')",
      "resize_image(input='raw.png', width=800, height=600, output='resized.png')",
      "convert_image(input='resized.png', format='webp', output='optimized.webp')"
    ]
  },
  {
    "name": "Create hero image from generation",
    "steps": [
      "generate_image(prompt='...', size='1024x1024', output='full.png')",
      "crop_image(input='full.png', x=0, y=312, width=1024, height=400, output='hero.png')"
    ]
  }
]
```

---

### **Priority 3: Add "Troubleshooting" Section**

```json
"troubleshooting": {
  "model_not_found": "Use --list-models to see available models. Try 'gemini' (free) or 'imagen-4' (paid, high quality).",
  "image_too_large": "Gemini supports up to 1024x1024. For larger sizes, use model='imagen-4' with Vertex AI.",
  "invalid_dimensions": "Use 'file <image>' or check error message for actual image dimensions before cropping."
}
```

---

## ✅ What NOT to Change

### **Keep: Clear Error Messages**
Current errors are **excellent** - they teach the LLM:
```
Error: crop region (x=0 + width=1792 = 1792) exceeds image width 1024
```

### **Keep: Self-Discovery Tools**
`--list-models`, `--help`, `--verbose` enable autonomous learning

### **Keep: MCP Tool Schema Structure**
Well-organized with types, constraints, and descriptions

---

## 🔮 Future Enhancements

### **1. Add "Related Tools" to Descriptions**

```json
"generate_image": {
  "related_tools": ["resize_image", "crop_image", "convert_image"],
  "common_next_steps": "After generating, consider resizing for web (resize_image) or converting to WebP for optimization (convert_image)."
}
```

### **2. Add "Cost Warnings" for Paid Models**

```json
"model": {
  "cost_info": {
    "gemini": "FREE (500/day)",
    "gemini-2.5-flash-image": "FREE (500/day)",
    "imagen-4": "$0.04 per image (PAID)",
    "amazon.nova-canvas-v1:0": "$0.04-0.08 per image (PAID)"
  }
}
```

### **3. Add "Performance Tips"**

```json
"performance_tips": [
  "Use 'gemini' for quick iterations (free, fast)",
  "Use 'imagen-4' for final high-quality output",
  "Batch operations: process multiple images before conversion",
  "WebP format reduces file size by ~30% vs PNG"
]
```

---

## 📊 Conclusion

**Overall Assessment**: **8.5/10** - Excellent MCP implementation

**Strengths**:
- ✅ Comprehensive tool descriptions
- ✅ Clear error messages that teach
- ✅ Self-discovery tools (--list-models, --help)
- ✅ Alias support for model names
- ✅ Automatic fallback on errors

**Improvements**:
- ✅ **IMPLEMENTED** Add quick-start examples (Priority 1)
- ⚠️ Highlight common workflows (Priority 2)
- ✅ **IMPLEMENTED** Document model limitations (size constraints per model)

**Key Insight**:
LLMs learn best through **guided trial-and-error**. The combination of helpful error messages + self-discovery tools + comprehensive MCP schemas enables autonomous learning within 2-3 attempts.

**Implementation Details** (2025-11-02):
1. ✅ Enhanced generate_image description with quick-start example
2. ✅ Added cost information to model descriptions (FREE vs paid, pricing)
3. ✅ Documented size constraints per model (Gemini max 1024x1024, Imagen supports 2048x2048)
4. ✅ Improved crop_image with concrete example and positional parameter guidance
5. ⏭️ Common workflows documentation (future enhancement)
6. ⏭️ Troubleshooting hints (future enhancement)

---

**Generated**: 2025-11-02
**Tool Version**: gimage v1.1.40
**MCP Server**: @apresai/gimage-mcp
