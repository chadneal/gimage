// Package tui provides Terminal User Interface components for gimage.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BatchMenuModel handles batch operations (simplified for MVP)
type BatchMenuModel struct {
	width    int
	height   int
	showHelp bool
}

// NewBatchMenuModel creates a new batch menu model
func NewBatchMenuModel() *BatchMenuModel {
	return &BatchMenuModel{}
}

// Init initializes the batch menu
func (m *BatchMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the batch menu
func (m *BatchMenuModel) Update(msg tea.Msg) (*BatchMenuModel, tea.Cmd) {
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
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the batch menu
func (m *BatchMenuModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	content := TitleStyle.Render("Batch Operations") + "\n\n" +
		SubtitleStyle.Render("Coming Soon") + "\n\n" +
		MutedStyle.Render("Batch operations will allow you to process") + "\n" +
		MutedStyle.Render("multiple images at once with:") + "\n\n" +
		FormatList([]string{
			"Batch Resize - Resize multiple images",
			"Batch Compress - Compress all images in a folder",
			"Batch Convert - Convert formats in bulk",
			"Custom Pipeline - Chain multiple operations",
		}) + "\n" +
		WarningStyle.Render("This feature is under development") + "\n\n" +
		HelpStyle.Render("m: Main menu • Esc: Back • ?: Help")

	box := FocusedBoxStyle.Width(76).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *BatchMenuModel) renderHelp() string {
	helpContent := TitleStyle.Render("Batch Operations Help") + "\n\n" +
		"Batch operations let you process multiple images" + "\n" +
		"simultaneously using parallel workers for speed." + "\n\n" +
		SubtitleStyle.Render("Planned Features") + "\n\n" +
		"• Directory selection and filtering\n" +
		"• Progress tracking for each file\n" +
		"• Worker count configuration\n" +
		"• Error handling and retry logic\n" +
		"• Summary reports\n\n" +
		HelpStyle.Render("Press Esc to close")

	box := FocusedBoxStyle.Width(70).Render(helpContent)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
