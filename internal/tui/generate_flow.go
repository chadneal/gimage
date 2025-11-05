// Package tui provides Terminal User Interface components for gimage.
package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apresai/gimage/internal/config"
	"github.com/apresai/gimage/internal/generate"
	"github.com/apresai/gimage/pkg/models"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GenerateStep represents a step in the image generation workflow
type GenerateStep int

const (
	StepPrompt GenerateStep = iota
	StepModel
	StepSize
	StepStyle
	StepOutput
	StepProgress
	StepResult
)

// Model selection item
type modelOption struct {
	name        string
	displayName string
	description string
	cost        string
	free        bool
}

// Size selection item
type sizeOption struct {
	size   string
	label  string
	aspect string
}

// Style selection item
type styleOption struct {
	value string
	label string
	desc  string
}

// GenerateFlowModel handles the multi-step image generation flow
type GenerateFlowModel struct {
	currentStep GenerateStep
	width       int
	height      int

	// Step 1: Prompt input
	promptTextarea textarea.Model

	// Step 2: Model selection
	models         []modelOption
	selectedModel  int
	modelInfo      *generate.ModelInfo

	// Step 3: Size selection
	sizes        []sizeOption
	selectedSize int
	customWidth  textinput.Model
	customHeight textinput.Model
	useCustom    bool

	// Step 4: Style selection
	styles        []styleOption
	selectedStyle int

	// Step 5: Output path
	outputInput textinput.Model

	// Step 6: Progress
	progressBar progress.Model
	progressMsg string
	generating  bool

	// Step 7: Result
	resultPath    string
	resultSize    int64
	generationTime time.Duration
	err           error

	// Navigation state
	showHelp bool
}

// NewGenerateFlowModel creates a new generate flow model
func NewGenerateFlowModel() *GenerateFlowModel {
	// Initialize prompt textarea
	ta := textarea.New()
	ta.Placeholder = "Describe the image you want to generate..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(70)
	ta.SetHeight(5)

	// Initialize custom size inputs
	widthInput := textinput.New()
	widthInput.Placeholder = "1024"
	widthInput.CharLimit = 4
	widthInput.Width = 10

	heightInput := textinput.New()
	heightInput.Placeholder = "1024"
	heightInput.CharLimit = 4
	heightInput.Width = 10

	// Initialize output path input
	outputInput := textinput.New()
	outputInput.Placeholder = "~/Desktop/gimage_output.png"
	outputInput.CharLimit = 256
	outputInput.Width = 60

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Load available models
	availableModels := generate.AvailableModels()
	modelOpts := make([]modelOption, 0, len(availableModels))
	for _, m := range availableModels {
		cost := generate.FormatPricingDisplay(&m)
		modelOpts = append(modelOpts, modelOption{
			name:        m.Name,
			displayName: m.DisplayName,
			description: m.Description,
			cost:        cost,
			free:        m.Pricing.FreeTier,
		})
	}

	// Size options
	sizeOpts := []sizeOption{
		{"1024x1024", "Square (1024x1024)", "1:1"},
		{"1792x1024", "Landscape (1792x1024)", "16:9"},
		{"1024x1792", "Portrait (1024x1792)", "9:16"},
		{"2048x2048", "Ultra HD (2048x2048)", "1:1"},
		{"custom", "Custom Size", ""},
	}

	// Style options
	styleOpts := []styleOption{
		{"", "None", "No specific style"},
		{"photorealistic", "Photorealistic", "Realistic photography style"},
		{"artistic", "Artistic", "Artistic and painterly style"},
		{"anime", "Anime", "Anime and manga style"},
	}

	return &GenerateFlowModel{
		currentStep:    StepPrompt,
		promptTextarea: ta,
		models:         modelOpts,
		selectedModel:  0,
		sizes:          sizeOpts,
		selectedSize:   0,
		customWidth:    widthInput,
		customHeight:   heightInput,
		styles:         styleOpts,
		selectedStyle:  0,
		outputInput:    outputInput,
		progressBar:    prog,
	}
}

// Init initializes the generate flow
func (m *GenerateFlowModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the generate flow
func (m *GenerateFlowModel) Update(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global shortcuts
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			// Go back to previous step or main menu
			if m.currentStep > StepPrompt {
				m.currentStep--
				// Reset focus for the previous step
				m.resetFocusForStep()
			} else {
				return m, Navigate(ScreenMainMenu)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case progressMsg:
		m.progressMsg = string(msg)
		return m, nil

	case generationCompleteMsg:
		m.generating = false
		m.resultPath = msg.path
		m.resultSize = msg.size
		m.generationTime = msg.duration
		m.err = msg.err
		m.currentStep = StepResult
		return m, nil
	}

	// Delegate to step-specific handlers
	switch m.currentStep {
	case StepPrompt:
		return m.updatePromptStep(msg)
	case StepModel:
		return m.updateModelStep(msg)
	case StepSize:
		return m.updateSizeStep(msg)
	case StepStyle:
		return m.updateStyleStep(msg)
	case StepOutput:
		return m.updateOutputStep(msg)
	case StepProgress:
		// Update progress bar - convert the model back to progress.Model
		var progModel tea.Model
		progModel, cmd = m.progressBar.Update(msg)
		m.progressBar = progModel.(progress.Model)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case StepResult:
		return m.updateResultStep(msg)
	}

	return m, tea.Batch(cmds...)
}

// View renders the generate flow
func (m *GenerateFlowModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.currentStep {
	case StepPrompt:
		return m.viewPromptStep()
	case StepModel:
		return m.viewModelStep()
	case StepSize:
		return m.viewSizeStep()
	case StepStyle:
		return m.viewStyleStep()
	case StepOutput:
		return m.viewOutputStep()
	case StepProgress:
		return m.viewProgressStep()
	case StepResult:
		return m.viewResultStep()
	default:
		return "Unknown step"
	}
}

// resetFocusForStep resets focus when going back to a step
func (m *GenerateFlowModel) resetFocusForStep() {
	switch m.currentStep {
	case StepPrompt:
		m.promptTextarea.Focus()
	case StepOutput:
		m.outputInput.Focus()
	}
}

// Step 1: Prompt Input
func (m *GenerateFlowModel) updatePromptStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+d", "ctrl+enter":
			// Move to next step if prompt is not empty
			if len(strings.TrimSpace(m.promptTextarea.Value())) > 0 {
				m.currentStep = StepModel
				return m, nil
			}
		}
	}

	m.promptTextarea, cmd = m.promptTextarea.Update(msg)
	return m, cmd
}

func (m *GenerateFlowModel) viewPromptStep() string {
	charCount := len(m.promptTextarea.Value())
	charLimit := m.promptTextarea.CharLimit

	content := TitleStyle.Render("Generate Image - Step 1/6") + "\n\n" +
		SubtitleStyle.Render("Describe the image you want to generate") + "\n\n" +
		m.promptTextarea.View() + "\n\n" +
		MutedStyle.Render(fmt.Sprintf("Characters: %d/%d", charCount, charLimit)) + "\n\n" +
		HelpStyle.Render("Ctrl+D or Ctrl+Enter: Next • Esc: Cancel • ?: Help")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 2: Model Selection
func (m *GenerateFlowModel) updateModelStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if m.selectedModel > 0 {
				m.selectedModel--
			}
		case "down", "j":
			if m.selectedModel < len(m.models)-1 {
				m.selectedModel++
			}
		case "enter", " ":
			// Load model info and move to next step
			modelInfo, err := generate.GetModelInfo(m.models[m.selectedModel].name)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.modelInfo = modelInfo
			m.currentStep = StepSize
			return m, nil
		}
	}
	return m, nil
}

func (m *GenerateFlowModel) viewModelStep() string {
	var items []string

	for i, model := range m.models {
		var style lipgloss.Style
		cursor := "  "
		if i == m.selectedModel {
			style = SelectedMenuItemStyle
			cursor = "> "
		} else {
			style = MenuItemStyle
		}

		// Show free badge
		badge := ""
		if model.free {
			badge = SuccessStyle.Render(" [FREE]")
		}

		title := style.Render(cursor + model.displayName + badge)
		desc := MutedStyle.Render("  " + model.description)
		cost := MutedStyle.Render("  Cost: " + model.cost)

		items = append(items, title+"\n"+desc+"\n"+cost)
	}

	content := TitleStyle.Render("Generate Image - Step 2/6") + "\n\n" +
		SubtitleStyle.Render("Select AI Model") + "\n\n" +
		strings.Join(items, "\n\n") + "\n\n" +
		HelpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 3: Size Selection
func (m *GenerateFlowModel) updateSizeStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// If custom size is active, handle input
		if m.useCustom {
			switch keyMsg.String() {
			case "tab":
				// Toggle focus between width and height
				if m.customWidth.Focused() {
					m.customWidth.Blur()
					m.customHeight.Focus()
				} else {
					m.customHeight.Blur()
					m.customWidth.Focus()
				}
				return m, nil
			case "enter":
				// Validate and move to next step
				if m.customWidth.Value() != "" && m.customHeight.Value() != "" {
					m.currentStep = StepStyle
					return m, nil
				}
			case "esc":
				m.useCustom = false
				m.customWidth.Blur()
				m.customHeight.Blur()
				return m, nil
			}

			// Update inputs
			if m.customWidth.Focused() {
				m.customWidth, cmd = m.customWidth.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.customHeight, cmd = m.customHeight.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Normal navigation
		switch keyMsg.String() {
		case "up", "k":
			if m.selectedSize > 0 {
				m.selectedSize--
			}
		case "down", "j":
			if m.selectedSize < len(m.sizes)-1 {
				m.selectedSize++
			}
		case "enter", " ":
			// Check if custom size is selected
			if m.sizes[m.selectedSize].size == "custom" {
				m.useCustom = true
				m.customWidth.Focus()
				return m, textinput.Blink
			}
			m.currentStep = StepStyle
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *GenerateFlowModel) viewSizeStep() string {
	var items []string

	if !m.useCustom {
		for i, size := range m.sizes {
			var style lipgloss.Style
			cursor := "  "
			if i == m.selectedSize {
				style = SelectedMenuItemStyle
				cursor = "> "
			} else {
				style = MenuItemStyle
			}

			title := style.Render(cursor + size.label)
			desc := ""
			if size.aspect != "" {
				desc = MutedStyle.Render("  Aspect ratio: " + size.aspect)
			}

			if desc != "" {
				items = append(items, title+"\n"+desc)
			} else {
				items = append(items, title)
			}
		}

		content := TitleStyle.Render("Generate Image - Step 3/6") + "\n\n" +
			SubtitleStyle.Render("Select Image Size") + "\n\n" +
			strings.Join(items, "\n\n") + "\n\n" +
			HelpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

		box := FocusedBoxStyle.Width(76).Render(content)

		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			box,
		)
	}

	// Custom size input view
	content := TitleStyle.Render("Generate Image - Step 3/6") + "\n\n" +
		SubtitleStyle.Render("Enter Custom Size") + "\n\n" +
		FormatKeyValue("Width", m.customWidth.View()) + "\n\n" +
		FormatKeyValue("Height", m.customHeight.View()) + "\n\n" +
		WarningStyle.Render("Note: Max size is 2048x2048 for most models") + "\n\n" +
		HelpStyle.Render("Tab: Switch field • Enter: Continue • Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 4: Style Selection
func (m *GenerateFlowModel) updateStyleStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if m.selectedStyle > 0 {
				m.selectedStyle--
			}
		case "down", "j":
			if m.selectedStyle < len(m.styles)-1 {
				m.selectedStyle++
			}
		case "enter", " ":
			m.currentStep = StepOutput
			m.outputInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *GenerateFlowModel) viewStyleStep() string {
	var items []string

	for i, style := range m.styles {
		var itemStyle lipgloss.Style
		cursor := "  "
		if i == m.selectedStyle {
			itemStyle = SelectedMenuItemStyle
			cursor = "> "
		} else {
			itemStyle = MenuItemStyle
		}

		title := itemStyle.Render(cursor + style.label)
		desc := MutedStyle.Render("  " + style.desc)

		items = append(items, title+"\n"+desc)
	}

	content := TitleStyle.Render("Generate Image - Step 4/6") + "\n\n" +
		SubtitleStyle.Render("Select Image Style (Optional)") + "\n\n" +
		strings.Join(items, "\n\n") + "\n\n" +
		HelpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 5: Output Path
func (m *GenerateFlowModel) updateOutputStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			// Start generation
			return m, m.startGeneration()
		}
	}

	m.outputInput, cmd = m.outputInput.Update(msg)
	return m, cmd
}

func (m *GenerateFlowModel) viewOutputStep() string {
	// Set default path if empty
	if m.outputInput.Value() == "" {
		home, _ := os.UserHomeDir()
		timestamp := time.Now().Format("20060102_150405")
		defaultPath := filepath.Join(home, "Desktop", fmt.Sprintf("gimage_%s.png", timestamp))
		m.outputInput.SetValue(defaultPath)
	}

	content := TitleStyle.Render("Generate Image - Step 5/6") + "\n\n" +
		SubtitleStyle.Render("Specify Output Path") + "\n\n" +
		"Output file: " + m.outputInput.View() + "\n\n" +
		MutedStyle.Render("The image will be saved to this location") + "\n\n" +
		HelpStyle.Render("Enter: Generate • Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 6: Progress
func (m *GenerateFlowModel) viewProgressStep() string {
	spinner := SpinnerFrames[int(time.Now().Unix())%len(SpinnerFrames)]

	content := TitleStyle.Render("Generate Image - Step 6/6") + "\n\n" +
		SubtitleStyle.Render("Generating...") + "\n\n" +
		SuccessStyle.Render(spinner) + " " + m.progressMsg + "\n\n" +
		m.progressBar.View() + "\n\n" +
		MutedStyle.Render("Please wait while your image is being generated...")

	box := FocusedBoxStyle.Width(76).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// Step 7: Result
func (m *GenerateFlowModel) updateResultStep(msg tea.Msg) (*GenerateFlowModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "g":
			// Generate another image - reset to step 1
			return NewGenerateFlowModel(), textarea.Blink
		case "m":
			// Go to main menu
			return m, Navigate(ScreenMainMenu)
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *GenerateFlowModel) viewResultStep() string {
	if m.err != nil {
		content := TitleStyle.Render("Generation Failed") + "\n\n" +
			ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n" +
			HelpStyle.Render("g: Try again • m: Main menu • q: Quit")

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2).
			Width(76).
			Render(content)

		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			box,
		)
	}

	content := TitleStyle.Render("Generation Complete!") + "\n\n" +
		SuccessStyle.Render("✓ Image generated successfully") + "\n\n" +
		FormatKeyValue("Saved to", m.resultPath) + "\n" +
		FormatKeyValue("File size", FormatFileSize(m.resultSize)) + "\n" +
		FormatKeyValue("Time taken", m.generationTime.String()) + "\n\n" +
		HelpStyle.Render("g: Generate another • m: Main menu • q: Quit")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(1, 2).
		Width(76).
		Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// renderHelp renders the help screen
func (m *GenerateFlowModel) renderHelp() string {
	helpContent := TitleStyle.Render("Keyboard Shortcuts") + "\n\n" +
		FormatKeyValue("↑/k, ↓/j", "Navigate options") + "\n" +
		FormatKeyValue("Enter/Space", "Select option") + "\n" +
		FormatKeyValue("Ctrl+D, Ctrl+Enter", "Next step (from prompt)") + "\n" +
		FormatKeyValue("Tab", "Switch input fields") + "\n" +
		FormatKeyValue("Esc", "Go back") + "\n" +
		FormatKeyValue("?", "Toggle help") + "\n" +
		FormatKeyValue("Ctrl+C", "Quit") + "\n\n" +
		SubtitleStyle.Render("Generation Workflow") + "\n\n" +
		"1. Enter a detailed prompt\n" +
		"2. Choose an AI model (Gemini is free!)\n" +
		"3. Select image size\n" +
		"4. Pick a style (optional)\n" +
		"5. Specify output path\n" +
		"6. Watch the progress\n\n" +
		HelpStyle.Render("Press Esc to close this help")

	box := FocusedBoxStyle.Width(70).Render(helpContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// startGeneration begins the image generation process
func (m *GenerateFlowModel) startGeneration() tea.Cmd {
	m.currentStep = StepProgress
	m.generating = true
	m.progressMsg = "Initializing..."

	return tea.Batch(
		m.tickProgress(),
		m.generateImageCmd(),
	)
}

// tickProgress updates the progress display
func (m *GenerateFlowModel) tickProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// generateImageCmd performs the actual image generation
func (m *GenerateFlowModel) generateImageCmd() tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		// Build the size string
		var size string
		if m.useCustom {
			size = fmt.Sprintf("%sx%s", m.customWidth.Value(), m.customHeight.Value())
		} else {
			size = m.sizes[m.selectedSize].size
		}

		// Build options
		options := models.GenerateOptions{
			Model:          m.models[m.selectedModel].name,
			Size:           size,
			Style:          m.styles[m.selectedStyle].value,
			NegativePrompt: "", // Could add in future
		}

		// Send progress updates
		go func() {
			time.Sleep(500 * time.Millisecond)
			// Note: In real implementation, we'd have proper progress channels
		}()

		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			return generationCompleteMsg{err: fmt.Errorf("failed to load config: %w", err)}
		}

		// Detect API from model
		api, err := generate.DetectAPIFromModel(options.Model)
		if err != nil {
			return generationCompleteMsg{err: fmt.Errorf("failed to detect API: %w", err)}
		}

		ctx := context.Background()

		// Generate image based on API
		var result *models.GeneratedImage
		switch api {
		case "gemini":
			client, err := generate.NewGeminiClient(cfg.GeminiAPIKey)
			if err != nil {
				return generationCompleteMsg{err: fmt.Errorf("failed to create client: %w", err)}
			}
			defer client.Close()

			result, err = client.GenerateImage(ctx, m.promptTextarea.Value(), options)
			if err != nil {
				return generationCompleteMsg{err: fmt.Errorf("generation failed: %w", err)}
			}

		case "vertex":
			// TODO: Implement vertex support
			return generationCompleteMsg{err: fmt.Errorf("vertex API not yet supported in TUI")}

		case "bedrock":
			// TODO: Implement bedrock support
			return generationCompleteMsg{err: fmt.Errorf("bedrock API not yet supported in TUI")}

		default:
			return generationCompleteMsg{err: fmt.Errorf("unknown API: %s", api)}
		}

		// Save the generated image to disk
		outputPath := m.outputInput.Value()
		if err := generate.SaveImage(result, outputPath); err != nil {
			return generationCompleteMsg{err: fmt.Errorf("failed to save image: %w", err)}
		}

		// Get file info
		fileInfo, err := os.Stat(outputPath)
		if err != nil {
			return generationCompleteMsg{
				path:     outputPath,
				duration: time.Since(startTime),
				err:      fmt.Errorf("image saved but couldn't stat file: %w", err),
			}
		}

		return generationCompleteMsg{
			path:     outputPath,
			size:     fileInfo.Size(),
			duration: time.Since(startTime),
		}
	}
}

// Custom messages
type tickMsg time.Time
type progressMsg string

type generationCompleteMsg struct {
	path     string
	size     int64
	duration time.Duration
	err      error
}
