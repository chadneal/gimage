// Package tui provides Terminal User Interface components for gimage.
package tui

import (
	"github.com/apresai/gimage/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsMenuModel handles settings and configuration
type SettingsMenuModel struct {
	width      int
	height     int
	showHelp   bool
	cfg        *config.Config
	selectedOp int
	options    []string
}

// NewSettingsMenuModel creates a new settings menu model
func NewSettingsMenuModel() *SettingsMenuModel {
	// Try to load config
	cfg, _ := config.LoadConfig()

	return &SettingsMenuModel{
		cfg: cfg,
		options: []string{
			"View Configuration",
			"API Keys Status",
			"About gimage",
			"Keyboard Shortcuts",
		},
	}
}

// Init initializes the settings menu
func (m *SettingsMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the settings menu
func (m *SettingsMenuModel) Update(msg tea.Msg) (*SettingsMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc", "m":
			return m, Navigate(ScreenMainMenu)
		case "up", "k":
			if m.selectedOp > 0 {
				m.selectedOp--
			}
		case "down", "j":
			if m.selectedOp < len(m.options)-1 {
				m.selectedOp++
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the settings menu
func (m *SettingsMenuModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.selectedOp {
	case 0:
		return m.viewConfig()
	case 1:
		return m.viewAPIStatus()
	case 2:
		return m.viewAbout()
	case 3:
		return m.viewShortcuts()
	default:
		return m.viewMainSettings()
	}
}

func (m *SettingsMenuModel) viewMainSettings() string {
	var items []string
	for i, opt := range m.options {
		cursor := "  "
		style := MenuItemStyle
		if i == m.selectedOp {
			cursor = "> "
			style = SelectedMenuItemStyle
		}
		items = append(items, style.Render(cursor+opt))
	}

	content := TitleStyle.Render("Settings") + "\n\n" +
		SubtitleStyle.Render("Configuration & Information") + "\n\n" +
		lipgloss.JoinVertical(lipgloss.Left, items...) + "\n\n" +
		HelpStyle.Render("↑/↓: Navigate • Enter: Select • m: Main menu • Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *SettingsMenuModel) viewConfig() string {
	var configLines []string

	if m.cfg != nil {
		configLines = []string{
			FormatKeyValue("Config File", "~/.gimage/config.md"),
			"",
			SubtitleStyle.Render("Current Configuration"),
			"",
		}

		if m.cfg.GeminiAPIKey != "" {
			configLines = append(configLines, FormatKeyValue("Gemini API Key", "Set ("+maskKey(m.cfg.GeminiAPIKey)+")"))
		} else {
			configLines = append(configLines, FormatKeyValue("Gemini API Key", ErrorStyle.Render("Not set")))
		}

		if m.cfg.VertexAPIKey != "" {
			configLines = append(configLines, FormatKeyValue("Vertex API Key", "Set ("+maskKey(m.cfg.VertexAPIKey)+")"))
			configLines = append(configLines, FormatKeyValue("Vertex Project", m.cfg.VertexProject))
		} else {
			configLines = append(configLines, FormatKeyValue("Vertex API Key", ErrorStyle.Render("Not set")))
		}

		if m.cfg.AWSAccessKeyID != "" {
			configLines = append(configLines, FormatKeyValue("AWS Access Key", "Set ("+maskKey(m.cfg.AWSAccessKeyID)+")"))
			configLines = append(configLines, FormatKeyValue("AWS Region", m.cfg.AWSRegion))
		} else {
			configLines = append(configLines, FormatKeyValue("AWS Access Key", ErrorStyle.Render("Not set")))
		}

		configLines = append(configLines, "", FormatKeyValue("Default API", m.cfg.DefaultAPI))
		configLines = append(configLines, FormatKeyValue("Default Model", m.cfg.DefaultModel))
	} else {
		configLines = []string{
			ErrorStyle.Render("Configuration file not found"),
			"",
			MutedStyle.Render("Run 'gimage auth gemini' to set up credentials"),
		}
	}

	content := TitleStyle.Render("Configuration") + "\n\n" +
		lipgloss.JoinVertical(lipgloss.Left, configLines...) + "\n\n" +
		WarningStyle.Render("Use 'gimage auth' command to update credentials") + "\n\n" +
		HelpStyle.Render("Esc: Back")

	box := FocusedBoxStyle.Width(80).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *SettingsMenuModel) viewAPIStatus() string {
	hasGemini := config.HasGeminiCredentials()
	hasVertex := config.HasVertexCredentials()
	hasBedrock := config.HasBedrockCredentials()

	geminiStatus := redNo()
	vertexStatus := redNo()
	bedrockStatus := redNo()

	if hasGemini {
		geminiStatus = greenYes()
	}
	if hasVertex {
		vertexStatus = greenYes()
	}
	if hasBedrock {
		bedrockStatus = greenYes()
	}

	content := TitleStyle.Render("API Keys Status") + "\n\n" +
		SubtitleStyle.Render("Authentication Status") + "\n\n" +
		FormatKeyValue("Gemini API", geminiStatus+" "+(map[bool]string{true: "Configured", false: "Not configured"}[hasGemini])) + "\n" +
		MutedStyle.Render("  Free tier: 500 images/day") + "\n\n" +
		FormatKeyValue("Vertex AI", vertexStatus+" "+(map[bool]string{true: "Configured", false: "Not configured"}[hasVertex])) + "\n" +
		MutedStyle.Render("  Paid: $0.02-0.06 per image") + "\n\n" +
		FormatKeyValue("AWS Bedrock", bedrockStatus+" "+(map[bool]string{true: "Configured", false: "Not configured"}[hasBedrock])) + "\n" +
		MutedStyle.Render("  Paid: $0.04-0.08 per image") + "\n\n" +
		WarningStyle.Render("Setup: gimage auth <api>") + "\n\n" +
		HelpStyle.Render("Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *SettingsMenuModel) viewAbout() string {
	content := TitleStyle.Render("About gimage") + "\n\n" +
		SubtitleStyle.Render("AI Image Generation & Processing Tool") + "\n\n" +
		"gimage is a powerful CLI and TUI for generating\n" +
		"AI images and processing existing images.\n\n" +
		SubtitleStyle.Render("Supported AI Models") + "\n\n" +
		"• Gemini 2.5 Flash (Free tier)\n" +
		"• Imagen 4 (Premium quality)\n" +
		"• AWS Bedrock Nova Canvas\n\n" +
		SubtitleStyle.Render("Image Processing") + "\n\n" +
		"• Resize, Scale, Crop\n" +
		"• Compress (JPEG quality)\n" +
		"• Convert (PNG, JPG, WebP, etc)\n\n" +
		MutedStyle.Render("Built with Go • Pure Go (zero C dependencies)") + "\n\n" +
		HelpStyle.Render("Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *SettingsMenuModel) viewShortcuts() string {
	content := TitleStyle.Render("Keyboard Shortcuts") + "\n\n" +
		SubtitleStyle.Render("Global Shortcuts") + "\n\n" +
		FormatKeyValue("Ctrl+C, q", "Quit application") + "\n" +
		FormatKeyValue("Esc", "Go back / Cancel") + "\n" +
		FormatKeyValue("?", "Toggle help") + "\n\n" +
		SubtitleStyle.Render("Navigation") + "\n\n" +
		FormatKeyValue("↑/↓ or k/j", "Move up/down") + "\n" +
		FormatKeyValue("Enter/Space", "Select item") + "\n" +
		FormatKeyValue("Tab", "Next input field") + "\n" +
		FormatKeyValue("Shift+Tab", "Previous input field") + "\n\n" +
		SubtitleStyle.Render("Special Keys") + "\n\n" +
		FormatKeyValue("Ctrl+D", "Submit prompt (generate)") + "\n" +
		FormatKeyValue("Ctrl+L", "Clear screen") + "\n" +
		FormatKeyValue("m", "Return to main menu") + "\n\n" +
		HelpStyle.Render("Esc: Back")

	box := FocusedBoxStyle.Width(76).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *SettingsMenuModel) renderHelp() string {
	helpContent := TitleStyle.Render("Settings Help") + "\n\n" +
		"View and manage gimage configuration." + "\n\n" +
		SubtitleStyle.Render("Configuration Location") + "\n\n" +
		"Config file: ~/.gimage/config.md\n" +
		"Format: Markdown with key-value pairs\n\n" +
		SubtitleStyle.Render("Setting Up API Keys") + "\n\n" +
		"Use the CLI commands:\n" +
		"  gimage auth gemini\n" +
		"  gimage auth vertex\n" +
		"  gimage auth bedrock\n\n" +
		HelpStyle.Render("Press Esc to close")

	box := FocusedBoxStyle.Width(70).Render(helpContent)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func greenYes() string {
	return SuccessStyle.Render("✓")
}

func redNo() string {
	return ErrorStyle.Render("✗")
}
