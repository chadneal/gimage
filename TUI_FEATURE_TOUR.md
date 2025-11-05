# gimage TUI Feature Tour

Welcome to the interactive Terminal User Interface for gimage! This guide walks you through all the features and how to use them.

## 🚀 Getting Started

### Launch the TUI
```bash
# First time: set up your API key
gimage auth gemini

# Then launch the TUI
gimage tui
```

### Main Menu
When you launch, you'll see the beautiful ASCII art main menu:

```
┌──────────────────────────────────────────────────┐
│                                                  │
│  ╔═══════════════════════════════════════════╗  │
│  ║ 🎨 gimage - Image Magic 🎨               ║  │
│  ╚═══════════════════════════════════════════╝  │
│                                                  │
│  What would you like to do?                     │
│                                                  │
│  > Generate Image (with AI)                    │
│    Process Image (resize, crop, etc.)          │
│    Batch Operations (process multiple)         │
│    Settings & Configuration                   │
│    Help & Shortcuts                            │
│                                                  │
│  Navigation: ↑↓ to move, Enter to select       │
│  Press ? for help, q to quit                   │
│                                                  │
└──────────────────────────────────────────────────┘
```

---

## ✨ Feature 1: Generate Images

The most powerful feature - create AI images without remembering any flags!

### The 6-Step Workflow

#### Step 1: Describe Your Image
```
┌──────────────────────────────────────────────────┐
│ Describe the image you want to generate:         │
│                                                  │
│ ┌────────────────────────────────────────────┐  │
│ │ a serene landscape with mountains and a  │  │
│ │ sunset, oil painting style, vibrant      │  │
│ │ colors, cinematic lighting, 4k           │  │
│ └────────────────────────────────────────────┘  │
│                                                  │
│ Characters: 84 / 1000                           │
│                                                  │
│ [Ctrl+D or Ctrl+Enter to continue]             │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Multi-line input (write as much as you want)
- Character counter (soft limit of 1000)
- Real-time editing with backspace, arrows, etc.

#### Step 2: Choose Your AI Model
```
┌──────────────────────────────────────────────────┐
│ Which AI model would you like to use?            │
│                                                  │
│ > 🆓 gemini-2.5-flash    (FREE - Fastest)       │
│   💰 gemini-2.0-flash    (FREE - Good Quality)  │
│   💰 imagen-3            ($0.025/image)         │
│   💰 imagen-4            ($0.04/image - Best!)  │
│   💰 nova-canvas         ($0.06/image)          │
│   💻 vertex-aiexpress    ($0.02/image)          │
│   💻 bedrock-claude      ($0.08/image)          │
│                                                  │
│ [More info] [Select this model]                │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- See pricing for each model
- 🆓 badge for free models
- Pricing info helps you choose what's best

#### Step 3: Select Image Size
```
┌──────────────────────────────────────────────────┐
│ What size should the image be?                  │
│                                                  │
│   1024 × 1024 (Square)                          │
│   1792 × 1024 (Landscape 16:9)                  │
│ > 1024 × 1792 (Portrait 9:16)                   │
│   2048 × 2048 (Ultra HD)                        │
│   Custom Size: [1024] × [1024]                  │
│                                                  │
│ Selected: 1024 × 1792 (Portrait)               │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Preset sizes for quick selection
- Custom size option with validation
- Shows aspect ratio

#### Step 4: Choose a Style
```
┌──────────────────────────────────────────────────┐
│ What style do you prefer? (optional)             │
│                                                  │
│   None (default)                                │
│ > Photorealistic (realistic photography)        │
│   Artistic (painted, drawn, illustrated)        │
│   Anime (anime/manga style)                     │
│                                                  │
│ [Skip] [Use this style]                        │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Optional selection (can skip if not supported)
- Style descriptions help you understand each option

#### Step 5: Where to Save
```
┌──────────────────────────────────────────────────┐
│ Where should we save the image?                 │
│                                                  │
│ Output Path:                                    │
│ [/Users/chad/Desktop/gimage_landscape_2025.png]│
│                                                  │
│ [Browse Files]  [Use Default]  [Confirm]       │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Auto-generated default path
- Browse button to choose custom location
- Shows full path

#### Step 6: Watch it Generate
```
┌──────────────────────────────────────────────────┐
│ Generating your image...                         │
│                                                  │
│ ⠙ [████████░░░░░░░░░░░░░] 32%                  │
│                                                  │
│ Status: Processing image with Gemini API        │
│ Model: gemini-2.5-flash-image                   │
│ Size: 1024 × 1792                              │
│ Elapsed: 3.2 seconds                           │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Beautiful animated spinner (⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
- Progress bar showing percentage
- Real-time status updates
- Shows model, size, elapsed time

#### Step 7: Your Image is Ready!
```
┌──────────────────────────────────────────────────┐
│ ✓ Image generated successfully!                 │
│                                                  │
│ File: /Users/chad/Desktop/gimage_landscape.png  │
│ Size: 2.4 MB                                    │
│ Dimensions: 1024 × 1792                         │
│ Time: 4.5 seconds                              │
│                                                  │
│ [Generate Another] [Process This] [Main Menu]  │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Shows file info (path, size, dimensions, generation time)
- Quick options to continue or go back
- Can chain operations

---

## 🖼️ Feature 2: Process Images

Transform existing images with a beautiful workflow!

### Step 1: Choose an Image
```
┌──────────────────────────────────────────────────┐
│ Browse images on your Desktop:                   │
│                                                  │
│ > IMG_5432.jpg         (3.2 MB, 3840 × 2160)   │
│   vacation_photo.png   (1.8 MB, 2048 × 1536)   │
│   screenshot.png       (0.9 MB, 1920 × 1080)   │
│   sunset.jpg           (2.1 MB, 4000 × 3000)   │
│   portrait.png         (1.2 MB, 1024 × 1536)   │
│                                                  │
│ [Go to Desktop]  [Continue]                    │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Browse Desktop images
- Shows file size and dimensions
- Easy navigation with arrow keys

### Step 2: Choose an Operation
```
┌──────────────────────────────────────────────────┐
│ What would you like to do?                      │
│                                                  │
│ > Resize          Change width & height         │
│   Scale           Resize by percentage           │
│   Crop            Extract a rectangular region  │
│   Compress        Reduce file size              │
│   Convert         Change format (PNG/JPG/etc)  │
│                                                  │
│ [Select Operation]                             │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- 5 different operations
- Descriptions for each operation
- Easy selection interface

### Step 3: Configure the Operation

**If you chose RESIZE:**
```
┌──────────────────────────────────────────────────┐
│ Resize Image                                    │
│                                                  │
│ Original: 3840 × 2160 (3.2 MB)                 │
│                                                  │
│ New width:  [1920]                             │
│ New height: [1080]                             │
│                                                  │
│ Output: /Users/chad/Desktop/IMG_processed.jpg  │
│                                                  │
│ [Process]  [Change Output Path]                │
│                                                  │
└──────────────────────────────────────────────────┘
```

**If you chose COMPRESS:**
```
┌──────────────────────────────────────────────────┐
│ Compress Image                                  │
│                                                  │
│ Current size: 3.2 MB                           │
│ Format: JPEG                                    │
│                                                  │
│ Quality: [████████░░░░░░░░░░░░] 80%             │
│ Estimated new size: 1.2 MB                     │
│ Estimated savings: ~62%                        │
│                                                  │
│ [Process]  [Change Output Path]                │
│                                                  │
└──────────────────────────────────────────────────┘
```

**If you chose CONVERT:**
```
┌──────────────────────────────────────────────────┐
│ Convert Image Format                            │
│                                                  │
│ Current format: JPEG                            │
│ New format:                                     │
│                                                  │
│ > PNG   (Lossless, larger files)               │
│   JPG   (Lossy, smaller files)                 │
│   WebP  (Modern, best quality/size ratio)      │
│   GIF   (Animated or simple graphics)          │
│   TIFF  (Professional print)                   │
│   BMP   (Uncompressed bitmap)                  │
│                                                  │
└──────────────────────────────────────────────────┘
```

### Step 4: Watch It Process
```
┌──────────────────────────────────────────────────┐
│ Processing image...                             │
│                                                  │
│ ⠋ [████████████░░░░░░░] 65%                     │
│                                                  │
│ Status: Resizing image                         │
│ Elapsed: 1.2 seconds                           │
│                                                  │
└──────────────────────────────────────────────────┘
```

### Step 5: Done!
```
┌──────────────────────────────────────────────────┐
│ ✓ Image processed successfully!                │
│                                                  │
│ File: /Users/chad/Desktop/IMG_processed.jpg    │
│ Original: 3840 × 2160 (3.2 MB)                 │
│ New: 1920 × 1080 (1.8 MB)                      │
│ Saved: 1.4 MB (44%)                            │
│ Time: 0.8 seconds                              │
│                                                  │
│ [Process Another] [Generate New] [Main Menu]   │
│                                                  │
└──────────────────────────────────────────────────┘
```

---

## ⚙️ Feature 3: Settings

Configure everything visually without touching config files!

### Settings Menu
```
┌──────────────────────────────────────────────────┐
│ Settings                                        │
│                                                  │
│ > View Configuration                           │
│   Check API Keys Status                        │
│   About gimage                                 │
│   Keyboard Shortcuts                           │
│                                                  │
│ [Navigate]  [Back to Menu]                     │
│                                                  │
└──────────────────────────────────────────────────┘
```

### View Configuration
```
┌──────────────────────────────────────────────────┐
│ Configuration                                   │
│                                                  │
│ Config File: ~/.gimage/config.md                │
│                                                  │
│ Gemini API Key: AIzaSy...xxx (configured)      │
│ Default Model: gemini-2.5-flash-image          │
│ Default API: gemini                            │
│                                                  │
│ Vertex AI:                                      │
│   API Key: Not configured                      │
│   Project ID: Not set                          │
│   Location: Not set                            │
│                                                  │
│ AWS Bedrock:                                    │
│   Access Key: Not configured                   │
│   Region: Not set                              │
│                                                  │
│ [Back]                                         │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Features**:
- Shows all configuration
- API keys are masked for security
- Shows what's configured vs what's not

### API Keys Status
```
┌──────────────────────────────────────────────────┐
│ API Authentication Status                       │
│                                                  │
│ Gemini              ✓ Configured                │
│ Vertex AI           ✗ Not Configured            │
│ AWS Bedrock         ✗ Not Configured            │
│                                                  │
│ To set up new APIs:                            │
│   $ gimage auth gemini   # Google Gemini       │
│   $ gimage auth vertex   # Google Vertex AI    │
│   $ gimage auth bedrock  # AWS Bedrock         │
│                                                  │
│ [Back]                                         │
│                                                  │
└──────────────────────────────────────────────────┘
```

---

## 📚 Feature 4: Help System

Get help anytime by pressing `?` on any screen!

```
┌──────────────────────────────────────────────────┐
│ Keyboard Shortcuts                              │
│                                                  │
│ GLOBAL:                                         │
│   q, Ctrl+C       Quit TUI                     │
│   Esc             Go back / Cancel             │
│   ?               Toggle this help             │
│   m               Return to main menu          │
│                                                  │
│ NAVIGATION:                                     │
│   ↑/↓ or k/j      Move up/down in lists       │
│   Enter/Space     Select item                 │
│   Tab/Shift+Tab   Switch between fields       │
│                                                  │
│ GENERATE WORKFLOW:                             │
│   Ctrl+D/Enter    Submit prompt                │
│   Esc             Go back one step             │
│                                                  │
│ PROCESSING:                                     │
│   p               Process another image        │
│   g               Generate another image       │
│                                                  │
│ [Press ? again to close]                       │
│                                                  │
└──────────────────────────────────────────────────┘
```

---

## 🎯 Example Workflows

### Example 1: Generate a Sunset Image
```
1. Launch: gimage tui
2. Main Menu > Generate Image
3. Prompt: "a beautiful tropical sunset, palm trees silhouette,
            warm orange and pink colors, high quality photography"
4. Model: Select "gemini-2.5-flash" (it's free and fast!)
5. Size: Choose "1024x1024" (square)
6. Style: Choose "Photorealistic"
7. Output: Accept default ~/Desktop/
8. Wait... ✓ Done! Image saved!
```

### Example 2: Resize a Photo for Web
```
1. Launch: gimage tui
2. Main Menu > Process Image
3. Choose: IMG_5432.jpg (3840x2160)
4. Operation: Resize
5. Set: Width 1920, Height 1080
6. Click Process
7. ✓ Done! Resized image saved as IMG_5432_processed.jpg
```

### Example 3: Compress Photos for Email
```
1. Launch: gimage tui
2. Main Menu > Process Image
3. Choose: vacation.jpg (2.1 MB)
4. Operation: Compress
5. Set: Quality to 75
6. Process
7. ✓ Done! Compressed from 2.1 MB to ~800 KB
```

---

## ⌨️ Complete Keyboard Reference

### Movement & Selection
| Key | Action |
|-----|--------|
| `↑` or `k` | Move up in list |
| `↓` or `j` | Move down in list |
| `Enter` | Select item / Confirm |
| `Space` | Select item (in lists) |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |

### Navigation
| Key | Action |
|-----|--------|
| `Esc` | Go back / Cancel operation |
| `q` | Quit (from main menu) |
| `Ctrl+C` | Quit (anytime) |
| `m` | Return to main menu |
| `?` | Toggle help |

### In Generate Workflow
| Key | Action |
|-----|--------|
| `Ctrl+D` | Submit prompt and continue |
| `Ctrl+Enter` | Submit prompt and continue |
| `Ctrl+A` | Select all text |
| `Ctrl+U` | Clear current line |

### In Process Workflow
| Key | Action |
|-----|--------|
| `p` | Process another image |
| `g` | Generate another image |

---

## 💡 Pro Tips

1. **Press `?` on any screen** to see context-sensitive help
2. **Use arrow keys or `k/j`** for smooth navigation
3. **Tab between fields** when filling out forms
4. **Ctrl+D** to quickly submit your prompt in the generate workflow
5. **`m` key** returns you to main menu from result screens
6. **All operations are cancellable** with `Esc`
7. **File paths auto-generate** with smart defaults (_processed suffix for operations)
8. **API keys are masked** in settings for security (use `gimage auth` to change them)

---

## 🚀 Next Steps

- **Explore all the operations**: Generate images, process existing ones, check settings
- **Configure more API keys**: `gimage auth vertex` or `gimage auth bedrock`
- **Check out the CLI too**: `gimage --help` - everything still works from the command line!
- **Read the full docs**: Check `tui.md` for the original plan and design

---

## 📞 Feedback & Issues

Having fun with the TUI? Found a bug? Have a feature request?

The TUI is production-ready and actively maintained. All existing gimage CLI features still work, and nothing broke in the process!

Enjoy! 🎉
