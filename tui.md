# gimage TUI (Terminal User Interface) Development Plan

**Vision**: Transform gimage into a delightful, interactive CLI where users can generate and process images without memorizing flags or subcommands. ASCII art + Bubbletea = fun productivity.

---

## Architecture Overview

### Core Principle
- **Separate concerns**: Core image generation/processing logic remains unchanged
- **TUI as wrapper**: Interactive UI layer calls existing core functions
- **Shared infrastructure**: Config, auth, API clients shared between CLI and TUI
- **Dual-mode support**: Both traditional CLI (`gimage generate`) and TUI (`gimage tui` or interactive `gimage --interactive`)

### Technology Stack
- **Framework**: Charmbracelet Bubbletea v0.27+
- **Components**: Charmbracelet Bubbles (textinput, spinner, list, table)
- **Styling**: Charmbracelet Lipgloss
- **Status**: Charmbracelet Log for formatted output

---

## Phase 1: Foundation (MVP) — Weeks 1-2

### 1.1 Project Structure
```
internal/
├── tui/
│   ├── tui.go              # Main TUI entrypoint
│   ├── models.go           # State models for each screen
│   ├── screens/
│   │   ├── main_menu.go
│   │   ├── generate_flow.go
│   │   ├── process_menu.go
│   │   └── settings.go
│   ├── components/
│   │   ├── file_picker.go
│   │   ├── model_selector.go
│   │   └── progress_display.go
│   └── styles.go           # Unified styling + ASCII art
├── progress/               # NEW: Progress reporting system
│   └── reporter.go         # Progress callback mechanism
└── cli/
    └── tui.go              # New 'tui' command

cmd/gimage/
└── main.go                 # Add 'tui' command registration
```

### 1.2 Core App Features Needed (BLOCKERS)

These must be implemented before TUI can be fully functional:

**A. Progress Reporting System** [HIGH PRIORITY]
- [ ] Create `internal/progress/reporter.go`
  - Interface: `ProgressReporter` with callbacks (Start, Update, Complete, Error)
  - Used in: generate, resize, crop, compress, batch operations
  - Allows TUI to show spinners + progress bars
  - CLI usage: silent by default, verbose shows updates

**B. Operation Context** [HIGH PRIORITY]
- [ ] All image generation/processing functions accept context
- [ ] Allows cancellation mid-operation
- [ ] Supports timeouts for long-running tasks

**C. File Selection Helper** [MEDIUM PRIORITY]
- [ ] `internal/tui/file_picker.go` - interactive file browser
- [ ] List files in directory with preview
- [ ] Filter by extension (.png, .jpg, etc.)

**D. Batch Operation History** [MEDIUM PRIORITY]
- [ ] Track operation results (success/fail, input, output paths)
- [ ] Allows replay/undo functionality
- [ ] Stored in memory during session (optional: persist to file)

---

## Phase 2: Main Menu & Navigation — Week 2-3

### 2.1 Main Menu Screen
```
┌─────────────────────────────────────────┐
│                                         │
│         🎨 gimage - Image Magic 🎨      │
│                                         │
│         What would you like to do?      │
│                                         │
│         > Generate Image                │
│           Process Image                 │
│           Batch Operations              │
│           Settings                      │
│           Exit                          │
│                                         │
└─────────────────────────────────────────┘
```

**Tasks**:
- [ ] Create `MainMenu` Bubbletea model
- [ ] Implement navigation to other screens
- [ ] ASCII art logo/title
- [ ] Keyboard navigation (arrow keys, enter)
- [ ] Display current config status (API key set? Model selected?)

---

## Phase 3: Generate Image Workflow — Week 3-4

### 3.1 Multi-Step Flow
**Screen 1: Prompt Input**
```
┌─────────────────────────────────────────┐
│ Describe the image you want:            │
│                                         │
│ [_________________________]             │
│ [Type your prompt here...]              │
│                                         │
│ Press Tab to continue                   │
└─────────────────────────────────────────┘
```

**Screen 2: Model Selection**
```
┌─────────────────────────────────────────┐
│ Which model do you want to use?         │
│                                         │
│ > gemini-2.5-flash    (Fast, Free)      │
│   imagen-4            (Best Quality)    │
│   nova-canvas         (AWS Bedrock)     │
│                                         │
│ [More info]                             │
└─────────────────────────────────────────┘
```

**Screen 3: Image Size**
```
┌─────────────────────────────────────────┐
│ What size do you want?                  │
│                                         │
│ > 1024x1024 (square)                    │
│   1792x1024 (landscape 16:9)            │
│   1024x1792 (portrait 9:16)             │
│   Custom: [____] x [____]               │
│                                         │
└─────────────────────────────────────────┘
```

**Screen 4: Style (Optional)**
```
┌─────────────────────────────────────────┐
│ Image style (optional):                 │
│                                         │
│ > None (default)                        │
│   Photorealistic                        │
│   Artistic                              │
│   Anime                                 │
│                                         │
└─────────────────────────────────────────┘
```

**Screen 5: Output Path**
```
┌─────────────────────────────────────────┐
│ Where to save the image?                │
│                                         │
│ [~/Desktop/generated_image.png]         │
│                                         │
│ [Browse] or type path...                │
│                                         │
└─────────────────────────────────────────┘
```

**Screen 6: Progress & Result**
```
┌─────────────────────────────────────────┐
│ Generating image...                     │
│                                         │
│ ⣾ [████████░░░░░░░░░░] 45%              │
│                                         │
│ Using: gemini-2.5-flash                 │
│ Size: 1024x1024                         │
│                                         │
└─────────────────────────────────────────┘
```

**Tasks**:
- [ ] Create flow state machine (track current step)
- [ ] Implement prompt input with multiline support
- [ ] Model selector with validation
- [ ] Size selector with custom option
- [ ] Style selector
- [ ] Output path input with file browser
- [ ] Progress display with animated spinner
- [ ] Result confirmation + option to continue/exit
- [ ] Store recent prompts for quick access

---

## Phase 4: Image Processing Workflow — Week 4-5

### 4.1 Process Menu
```
┌─────────────────────────────────────────┐
│ Select an image to process:             │
│                                         │
│ [Browse Files...]                       │
│ ~/Desktop/photo.jpg (2MB, 3840x2160)    │
│ ~/Pictures/art.png (4MB, 2048x2048)     │
│                                         │
│ What would you like to do?              │
│                                         │
│ > Resize                                │
│   Scale                                 │
│   Crop                                  │
│   Compress                              │
│   Convert Format                        │
│   Chain Multiple Ops                    │
│                                         │
└─────────────────────────────────────────┘
```

### 4.2 Individual Operations

**Resize**:
- [ ] Input: target width/height
- [ ] Preview aspect ratio change
- [ ] Output path selection

**Scale**:
- [ ] Input: scale factor (0.5 = 50%, 2.0 = 200%)
- [ ] Visual indicator of result size
- [ ] Output path selection

**Crop**:
- [ ] Interactive crop selection (show grid + coordinates)
- [ ] X, Y, Width, Height inputs
- [ ] Visual preview of crop area
- [ ] Output path selection

**Compress**:
- [ ] Quality slider (1-100)
- [ ] Show estimated file size reduction
- [ ] Preview before/after
- [ ] Output path selection

**Convert**:
- [ ] Format selector (PNG, JPG, WebP, GIF, TIFF, BMP)
- [ ] Format-specific options (JPG quality, WebP compression)
- [ ] Output path selection

**Chain Operations** [NICE-TO-HAVE]:
- [ ] Apply multiple operations in sequence
- [ ] Visual pipeline display
- [ ] Preview after each step

**Tasks**:
- [ ] File picker with preview capability
- [ ] Individual screens for each operation
- [ ] Real-time preview where possible (resize, crop)
- [ ] Operation preview before confirmation
- [ ] Progress tracking during operation
- [ ] Result confirmation with option to chain more ops

---

## Phase 5: Batch Operations — Week 5-6

### 5.1 Batch Menu
```
┌─────────────────────────────────────────┐
│ Batch Processing                        │
│                                         │
│ > Resize Multiple Images                │
│   Compress Batch                        │
│   Convert Formats                       │
│   Custom Pipeline                       │
│   Load Batch Script                     │
│                                         │
└─────────────────────────────────────────┘
```

### 5.2 Batch Processing Flow
- [ ] Select input directory or files
- [ ] Choose operation(s)
- [ ] Configure operation parameters
- [ ] Preview command that would be run
- [ ] Show parallel worker count (default: CPU cores)
- [ ] Real-time progress table:
  ```
  File                 Status      Progress
  ────────────────────────────────────────
  photo1.jpg          ✓ Done       100%
  photo2.jpg          ⣾ Running     45%
  photo3.jpg          ⊙ Queued       0%
  ```
- [ ] Final summary (total processed, failed, time taken, space saved)
- [ ] Option to export batch report

**Tasks**:
- [ ] Batch file selector (directory or file list)
- [ ] Batch operation configurator
- [ ] Real-time progress table with per-file status
- [ ] Worker count selector
- [ ] Summary report generation

---

## Phase 6: Settings & Configuration — Week 6

### 6.1 Settings Menu
```
┌─────────────────────────────────────────┐
│ Settings                                │
│                                         │
│ > API Configuration                     │
│   Default Model                         │
│   Default API                           │
│   Batch Workers                         │
│   Output Directory                      │
│   Theme                                 │
│   About                                 │
│                                         │
└─────────────────────────────────────────┘
```

### 6.2 API Configuration
- [ ] Gemini API key setup (interactive input)
- [ ] Vertex AI configuration (mode selection: Express/Service Account/ADC)
- [ ] AWS Bedrock setup (mode selection: keys or credential chain)
- [ ] Test connection for each API
- [ ] Display currently configured APIs

### 6.3 Preferences
- [ ] Default model selection
- [ ] Default API fallback order
- [ ] Batch operation worker count
- [ ] Default output directory
- [ ] Theme selector (colors, ASCII art intensity)
- [ ] Verbose mode toggle
- [ ] Auto-save recent operations

**Tasks**:
- [ ] Settings screen implementation
- [ ] API configuration UI (reuse from `gimage auth` but interactive)
- [ ] Preference storage/reload
- [ ] Settings validation
- [ ] Connection testing UI

---

## Phase 7: Polish & Experience — Week 7

### 7.1 Visual Polish
- [ ] ASCII art title screen/logo
- [ ] Color scheme (use Lipgloss theming)
- [ ] Consistent spacing and borders
- [ ] Animated spinners for long operations
- [ ] Keyboard shortcuts help overlay
- [ ] Status bar showing current API, model, config

### 7.2 Keyboard Navigation
- [ ] Arrow keys for menu navigation
- [ ] Tab/Shift+Tab to move between fields
- [ ] Enter to confirm
- [ ] Esc to cancel/go back
- [ ] Ctrl+C to exit safely
- [ ] `?` for help overlay
- [ ] Ctrl+L to clear screen

### 7.3 Error Handling & Recovery
- [ ] Display error messages with context
- [ ] Suggestions for recovery (e.g., "API key not set, go to settings")
- [ ] Retry functionality for failed operations
- [ ] Graceful handling of API rate limits
- [ ] Better error messages (vs raw API errors)

### 7.4 Help System
- [ ] Context-sensitive help on each screen
- [ ] Hover tooltips for options
- [ ] `--help` within TUI for each operation
- [ ] Links to documentation

**Tasks**:
- [ ] Unified styling/theme system
- [ ] Comprehensive keyboard shortcut mapping
- [ ] Error message formatting + suggestions
- [ ] Help system integration
- [ ] Mouse support (list selection, button clicks)

---

## Phase 8: Advanced Features [OPTIONAL] — Week 8+

### 8.1 Image Preview
- [ ] Show thumbnail of selected image in file picker
- [ ] Show crop preview with grid overlay
- [ ] Side-by-side resize preview

**Implementation**:
- Extract image metadata (dimensions, size)
- Consider terminal image rendering (Sixel, iTerm2) for advanced terminals
- Fallback to ASCII representation for basic terminals

### 8.2 Operation History
- [ ] View recent operations (last 20)
- [ ] Replay operation with same params
- [ ] Undo last operation
- [ ] Export operation log

### 8.3 Favorites/Quick Actions
- [ ] Save favorite prompts
- [ ] Save favorite operation chains
- [ ] Quick-access buttons on main menu
- [ ] Customizable quick actions

### 8.4 Scripting
- [ ] Load batch operations from `.gimage-script` file
- [ ] YAML/JSON format for batch definitions
- [ ] Record and replay TUI interactions

---

## Implementation Checklist

### App Infrastructure Changes
- [ ] **progress/reporter.go** - Add progress callback system
  - Define `ProgressReporter` interface
  - Implement `NoOpReporter` (silent), `LogReporter` (CLI), `TUIReporter` (TUI)
  - Update `generate.GenerateImage()` to accept reporter
  - Update all imaging operations (Resize, Crop, Compress, etc.) to accept reporter
  - [ ] Add context.Context to all operations for cancellation

- [ ] **cli/tui.go** - New command registration
  - Command: `gimage tui` (optional: default when no subcommand given)
  - Flag: `gimage --interactive` (alias for tui mode)

### TUI Structure
- [ ] **tui/tui.go** - Main TUI model + event loop
  - `Model` struct with state machine
  - `Init()`, `Update()`, `View()` for Bubbletea
  - Navigation state (main menu → generate → progress → main menu)

- [ ] **tui/screens/** - Individual screen models
  - MainMenu (phase 2)
  - GenerateFlow with 6 steps (phase 3)
  - ProcessMenu + operations (phase 4)
  - BatchMenu + flow (phase 5)
  - SettingsMenu (phase 6)
  - ErrorScreen (all phases)
  - HelpScreen (all phases)

- [ ] **tui/components/** - Reusable components
  - FilePickerComponent
  - ModelSelectorComponent
  - SizePickerComponent
  - ProgressDisplayComponent
  - ConfirmationDialogComponent

- [ ] **tui/styles.go** - Unified styling
  - Color palette
  - Component styles (borders, padding, focus states)
  - ASCII art assets (logo, spinners, etc.)

### Testing
- [ ] Unit tests for state transitions
- [ ] Integration tests for full workflows (generate, process, batch)
- [ ] Manual testing of keyboard navigation
- [ ] Manual testing with different terminal sizes

### Documentation
- [ ] TUI user guide (docs/TUI.md)
- [ ] Keyboard shortcuts reference
- [ ] Troubleshooting guide
- [ ] Update main README with TUI section

---

## Dependencies to Add

```go
// go.mod additions
github.com/charmbracelet/bubbletea v0.27.0
github.com/charmbracelet/bubbles v0.18.0
github.com/charmbracelet/lipgloss v0.12.1
```

**No breaking changes** to existing dependencies.

---

## Success Criteria

1. **Phase 1-2**: User can launch `gimage tui` and navigate to all major screens
2. **Phase 3**: Full generate workflow, progress display, result saving
3. **Phase 4**: Image processing workflow with all operations functional
4. **Phase 5**: Batch operations with progress tracking
5. **Phase 6**: Settings accessible and saveable
6. **Phase 7**: Polished UI with good keyboard navigation
7. **Phase 8+**: Nice-to-have features for advanced users

---

## Known Constraints

- **Terminal compatibility**: Test on macOS (standard), Linux (gnome-terminal, xterm), Windows (Windows Terminal, WSL2)
- **Image preview**: Limited by terminal capabilities; graceful fallback to metadata display
- **Context cancellation**: Must handle API cancellation gracefully (not all APIs support it)
- **File I/O**: Need permission checks before writing to user-specified paths

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Bubbletea complexity learning curve | Medium | Medium | Study examples, pair programming, documentation |
| Terminal size changes | Low | Low | Handle resize events, reflow layout |
| API timeout during progress display | Medium | Medium | Implement timeout handling + retry UI |
| Memory usage with large batch ops | Low | High | Stream progress, avoid loading all files into memory |
| User confusion with many options | High | Medium | Sensible defaults, progressive disclosure, help system |

---

## Timeline Estimate

- **Phase 1 (Foundation)**: 5-7 days
- **Phase 2 (Main Menu)**: 3-4 days
- **Phase 3 (Generate)**: 5-6 days
- **Phase 4 (Process)**: 5-7 days
- **Phase 5 (Batch)**: 4-5 days
- **Phase 6 (Settings)**: 3-4 days
- **Phase 7 (Polish)**: 4-5 days
- **Phase 8 (Advanced)**: 5+ days (optional)

**Total (core, Phase 1-7)**: ~29-35 days

---

## Next Steps

1. Implement Phase 1 blockers (progress reporter, context support)
2. Set up TUI project structure
3. Create main menu with ASCII art
4. Implement generate workflow (most important, most used feature)
5. Implement process menu
6. Test end-to-end workflow
7. Polish and refine UX
