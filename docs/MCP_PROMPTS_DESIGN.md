# MCP Prompts Design for gimage

**Date**: 2025-11-02
**Purpose**: Design prompts to teach LLMs how to use gimage MCP tools through interactive examples

## What Are MCP Prompts?

According to the MCP specification, **Prompts** are:
- Pre-defined message templates that servers expose to clients
- Structured instructions for interacting with language models
- A way to teach LLMs common workflows and patterns
- Customizable with arguments for different use cases

**Key Benefits**:
- LLMs learn by seeing **actual working examples**
- Reduces trial-and-error learning time
- Shows common workflows and patterns
- Provides context about when to use each tool

## Current State

gimage MCP server currently returns empty prompts list:
```go
// internal/mcp/handler.go:167-179
func (s *MCPServer) handleListPrompts(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
    return &JSONRPCResponse{
        Result: map[string]interface{}{
            "prompts": []interface{}{},  // Empty!
        },
    }
}
```

## Proposed Prompts

### 1. Quick Start Prompt (`gimage_quick_start`)

**Purpose**: Teach the absolute basics - generate your first image

**Arguments**: None

**Template**:
```markdown
I want to generate an AI image using gimage. Here's how to get started:

STEP 1: Generate a simple image using the free Gemini model
Call generate_image with:
- prompt: "a sunset over mountains"
- output: "~/Desktop/sunset.png"
- (model defaults to gemini-2.5-flash-image - FREE!)

STEP 2: Check the result
The tool will return the output path and cost information.
For Gemini, it's FREE (up to 1500 images/day).

EXAMPLE:
generate_image(
  prompt="a sunset over mountains",
  output="~/Desktop/sunset.png"
)

That's it! The image will be saved to your Desktop.
```

**Why this helps**: Shows the simplest possible workflow with minimal parameters

---

### 2. Generate with Style Prompt (`generate_with_style`)

**Purpose**: Teach how to use style parameters

**Arguments**:
- `subject` (required): What to generate (e.g., "a cat", "a landscape")
- `style` (optional): Style preference (photorealistic, artistic, anime)

**Template**:
```markdown
I want to generate {{subject}} with a specific artistic style.

RECOMMENDED APPROACH:
generate_image(
  prompt="{{subject}}, {{style}} style, high quality, detailed",
  style="{{style}}",
  output="~/Desktop/{{subject}}_{{style}}.png"
)

STYLE OPTIONS:
- photorealistic: For realistic photos
- artistic: For artistic/painterly renders
- anime: For anime/manga style

EXAMPLE:
generate_image(
  prompt="a futuristic city, photorealistic style, high quality, detailed",
  style="photorealistic",
  output="~/Desktop/city_photorealistic.png"
)

NOTE: Gemini is free and works great for most use cases.
For highest quality, use model="imagen-4" (paid, $0.02-0.04 per image).
```

**Why this helps**: Shows how to combine prompt engineering with style parameters

---

### 3. Generate and Crop Workflow (`generate_and_crop`)

**Purpose**: Teach a common workflow - generate then crop to specific dimensions

**Arguments**:
- `description` (required): What to generate
- `crop_width` (required): Desired crop width
- `crop_height` (required): Desired crop height

**Template**:
```markdown
I want to generate {{description}} and crop it to {{crop_width}}x{{crop_height}} for a specific use case (like a hero image or banner).

WORKFLOW:
1. Generate the full image first
2. Crop to desired dimensions

STEP 1: Generate
generate_image(
  prompt="{{description}}",
  size="1024x1024",
  output="~/Desktop/temp_full.png"
)

STEP 2: Crop to {{crop_width}}x{{crop_height}}
crop_image(
  input="~/Desktop/temp_full.png",
  x=0,
  y={{(1024 - crop_height) / 2}},  # Center vertically
  width={{crop_width}},
  height={{crop_height}},
  output="~/Desktop/final_cropped.png"
)

EXAMPLE (Hero Image 1024x400):
1. generate_image(prompt="abstract tech art", size="1024x1024", output="~/Desktop/temp.png")
2. crop_image(input="~/Desktop/temp.png", x=0, y=312, width=1024, height=400, output="~/Desktop/hero.png")

TIP: Use get_image_info first to verify actual image dimensions before cropping.
```

**Why this helps**: Shows a multi-step workflow that was observed in the learning analysis

---

### 4. High Quality Generation (`high_quality_image`)

**Purpose**: Teach when and how to use paid models for best results

**Arguments**:
- `subject` (required): What to generate

**Template**:
```markdown
I want to generate a high-quality image of {{subject}} for professional use.

RECOMMENDED: Use Imagen 4 for highest quality

generate_image(
  prompt="{{subject}}, ultra detailed, professional quality, 8k",
  model="imagen-4",
  size="2048x2048",
  output="~/Desktop/{{subject}}_hq.png"
)

COST BREAKDOWN:
- Gemini (free): Good quality, up to 1024x1024, FREE
- Imagen-4 (paid): Highest quality, up to 2048x2048, $0.02-0.04 per image

WHEN TO USE EACH:
- Gemini: Quick iterations, testing prompts, social media
- Imagen-4: Final production images, professional work, large sizes

EXAMPLE:
generate_image(
  prompt="professional headshot, studio lighting, ultra detailed, 8k",
  model="imagen-4",
  size="2048x2048",
  output="~/Desktop/headshot_hq.png"
)

NOTE: Requires Vertex AI setup. Run 'gimage auth vertex' first.
```

**Why this helps**: Clarifies when to use paid vs free models and cost implications

---

### 5. Batch Web Optimization Workflow (`optimize_for_web`)

**Purpose**: Teach how to generate and optimize images for web use

**Arguments**:
- `count` (required): Number of images to generate

**Template**:
```markdown
I want to generate {{count}} images and optimize them for web use (smaller file size, WebP format).

WORKFLOW:
1. Generate images with Gemini (fast and free)
2. Resize to web-friendly dimensions
3. Convert to WebP (30% smaller than PNG)

EXAMPLE (3 images):

# Generate 3 variations
generate_image(prompt="tech artwork 1", output="~/Desktop/raw1.png")
generate_image(prompt="tech artwork 2", output="~/Desktop/raw2.png")
generate_image(prompt="tech artwork 3", output="~/Desktop/raw3.png")

# Resize to web dimensions
resize_image(input="~/Desktop/raw1.png", width=800, height=600, output="~/Desktop/resized1.png")
resize_image(input="~/Desktop/raw2.png", width=800, height=600, output="~/Desktop/resized2.png")
resize_image(input="~/Desktop/raw3.png", width=800, height=600, output="~/Desktop/resized3.png")

# Convert to WebP for smaller file size
convert_image(input="~/Desktop/resized1.png", format="webp", output="~/Desktop/web1.webp")
convert_image(input="~/Desktop/resized2.png", format="webp", output="~/Desktop/web2.webp")
convert_image(input="~/Desktop/resized3.png", format="webp", output="~/Desktop/web3.webp")

TIP: You can also use batch_process_images for multiple files at once!
```

**Why this helps**: Shows advanced multi-tool workflow for real-world use case

---

### 6. Troubleshooting Guide (`troubleshooting`)

**Purpose**: Help LLMs recover from common errors

**Arguments**: None

**Template**:
```markdown
I encountered an error when using gimage. Here are common issues and solutions:

ERROR: "Model not found: gemini-flash"
SOLUTION: Use exact model names or common aliases
✅ CORRECT: model="gemini" or model="gemini-2.5-flash-image"
❌ WRONG: model="gemini-flash" or model="flash"

ERROR: "Gemini API key not configured"
SOLUTION: Set up authentication first
Run: gimage auth gemini
Then retry your generate_image call

ERROR: "crop region (x=0 + width=1792 = 1792) exceeds image width 1024"
SOLUTION: Check actual image dimensions before cropping
1. Use get_image_info to check dimensions
2. Gemini max size is 1024x1024 (not 1792x1024)
3. For larger sizes, use model="imagen-4"

ERROR: "unknown flag: --width" (when using crop)
SOLUTION: crop_image uses positional arguments, not flags
✅ CORRECT: crop_image(input="file.png", x=0, y=100, width=800, height=600)
❌ WRONG: crop_image(input="file.png", --width=800, --height=600)

ERROR: Image dimensions are wrong (e.g., 1024x1024 instead of 1792x1024)
SOLUTION: Check model size limits
- Gemini supports up to 1024x1024
- Imagen-4 supports up to 2048x2048
- Use get_image_info to verify actual dimensions

GENERAL TIPS:
1. Always specify output path (e.g., ~/Desktop/image.png)
2. Start with Gemini (free) for testing
3. Use verbose mode for debugging: add --verbose flag
4. Check get_image_info before cropping
```

**Why this helps**: Addresses all the errors observed in the learning analysis transcript

---

## Implementation Plan

### Phase 1: Add Prompts Data Structure

Create `internal/mcp/prompts.go`:
```go
package mcp

type Prompt struct {
    Name        string
    Title       string
    Description string
    Arguments   []PromptArgument
    Template    string  // Message template with {{variable}} substitution
}

type PromptArgument struct {
    Name        string
    Description string
    Required    bool
}
```

### Phase 2: Register Prompts

Similar to how tools are registered:
```go
func (s *MCPServer) RegisterPrompt(prompt Prompt) {
    s.prompts[prompt.Name] = prompt
}
```

### Phase 3: Update Handlers

1. Update `handleInitialize` to declare prompts capability:
```go
"capabilities": map[string]interface{}{
    "prompts": map[string]interface{}{
        "listChanged": false,  // Prompts are static
    },
    "tools": map[string]interface{}{
        "listChanged": true,
    },
},
```

2. Update `handleListPrompts` to return actual prompts

3. Add `handleGetPrompt` to return specific prompt with arguments substituted

### Phase 4: Test with Claude

Verify that Claude Desktop can:
1. List available prompts
2. Use prompts to learn workflows
3. Apply patterns to new tasks

---

## Expected Impact

Based on the learning analysis findings:

**Before Prompts** (current state):
- First-try success: 0% (model name error)
- Learning time: 3 attempts (~2 minutes)
- Requires trial-and-error to discover flags

**After Prompts** (expected):
- First-try success: ~90% (following prompt examples)
- Learning time: Immediate (can reference prompts)
- Reduces exploratory --list-models calls

**Why this works**:
- LLMs learn best from **concrete examples** (not just descriptions)
- Prompts show **working code** not just documentation
- Multi-step workflows are pre-packaged
- Common errors are proactively addressed

---

## Alternative: Enhanced Tool Descriptions vs Prompts

| Feature | Tool Descriptions (Current) | Prompts (Proposed) |
|---------|---------------------------|-------------------|
| Location | Tool inputSchema | Separate prompt templates |
| Format | JSON schema | Markdown with examples |
| Length | Short (1-2 sentences) | Long (multi-step workflows) |
| Examples | Inline | Full working code |
| Workflows | Not supported | Multi-tool sequences |
| Troubleshooting | Not supported | Dedicated guides |

**Recommendation**: Use BOTH
- Tool descriptions: Quick reference for parameters
- Prompts: Learning workflows and patterns

---

## Next Steps

1. ✅ Design prompts (this document)
2. ⏭️ Implement prompts data structure
3. ⏭️ Register 6 prompts in MCP server
4. ⏭️ Update handlers to return prompts
5. ⏭️ Test with Claude Desktop
6. ⏭️ Measure impact on first-try success rate

---

## References

- MCP Specification: https://modelcontextprotocol.io/specification/2025-06-18/server/prompts
- Learning Analysis: `docs/MCP_LLM_LEARNING_ANALYSIS.md`
- Tool Improvements: `docs/MCP_TOOL_IMPROVEMENTS_CHANGELOG.md`
