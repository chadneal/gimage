package mcp

// RegisterAllPrompts registers all gimage prompts with the MCP server
// These prompts teach LLMs how to use gimage tools through interactive examples
func RegisterAllPrompts(server *MCPServer) {
	// 1. Quick Start - Simplest possible workflow
	server.RegisterPrompt(Prompt{
		Name:        "gimage_quick_start",
		Title:       "Get Started with gimage",
		Description: "Learn how to generate your first AI image using the free Gemini model",
		Arguments:   []PromptArgument{},
		Template: `I want to generate an AI image using gimage. Here's how to get started:

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

That's it! The image will be saved to your Desktop.`,
	})

	// 2. Generate with Style - Teach style parameters
	server.RegisterPrompt(Prompt{
		Name:        "generate_with_style",
		Title:       "Generate Images with Artistic Styles",
		Description: "Learn how to use style parameters to control the artistic rendering of generated images",
		Arguments: []PromptArgument{
			{Name: "subject", Description: "What to generate (e.g., 'a cat', 'a landscape')", Required: true},
			{Name: "style", Description: "Style preference: photorealistic, artistic, or anime", Required: false},
		},
		Template: `I want to generate {{subject}} with a specific artistic style.

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
For highest quality, use model="imagen-4" (paid, $0.02-0.04 per image).`,
	})

	// 3. Generate and Crop Workflow - Multi-step workflow
	server.RegisterPrompt(Prompt{
		Name:        "generate_and_crop",
		Title:       "Generate and Crop for Hero Images",
		Description: "Learn the workflow to generate an image and crop it to specific dimensions (e.g., for hero images or banners)",
		Arguments: []PromptArgument{
			{Name: "description", Description: "What to generate", Required: true},
			{Name: "crop_width", Description: "Desired crop width in pixels", Required: true},
			{Name: "crop_height", Description: "Desired crop height in pixels", Required: true},
		},
		Template: `I want to generate {{description}} and crop it to {{crop_width}}x{{crop_height}} for a specific use case (like a hero image or banner).

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
  y=312,
  width={{crop_width}},
  height={{crop_height}},
  output="~/Desktop/final_cropped.png"
)

EXAMPLE (Hero Image 1024x400):
1. generate_image(prompt="abstract tech art", size="1024x1024", output="~/Desktop/temp.png")
2. crop_image(input="~/Desktop/temp.png", x=0, y=312, width=1024, height=400, output="~/Desktop/hero.png")

TIP: Use get_image_info first to verify actual image dimensions before cropping.`,
	})

	// 4. High Quality Generation - When to use paid models
	server.RegisterPrompt(Prompt{
		Name:        "high_quality_image",
		Title:       "Generate High-Quality Professional Images",
		Description: "Learn when and how to use paid models (Imagen 4) for professional-quality images",
		Arguments: []PromptArgument{
			{Name: "subject", Description: "What to generate", Required: true},
		},
		Template: `I want to generate a high-quality image of {{subject}} for professional use.

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

NOTE: Requires Vertex AI setup. Run 'gimage auth vertex' first.`,
	})

	// 5. Web Optimization Workflow - Multi-tool advanced workflow
	server.RegisterPrompt(Prompt{
		Name:        "optimize_for_web",
		Title:       "Generate and Optimize Images for Web",
		Description: "Learn how to generate images and optimize them for web use (smaller file size, WebP format)",
		Arguments: []PromptArgument{
			{Name: "count", Description: "Number of images to generate", Required: false},
		},
		Template: `I want to generate images and optimize them for web use (smaller file size, WebP format).

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

TIP: You can also use batch_process_images for multiple files at once!`,
	})

	// 6. Troubleshooting - Help with common errors
	server.RegisterPrompt(Prompt{
		Name:        "troubleshooting",
		Title:       "Troubleshoot Common gimage Errors",
		Description: "Learn how to fix common errors when using gimage tools",
		Arguments:   []PromptArgument{},
		Template: `I encountered an error when using gimage. Here are common issues and solutions:

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
4. Check get_image_info before cropping`,
	})
}
