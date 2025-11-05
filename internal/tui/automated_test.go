package tui

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/apresai/gimage/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// TestAutomatedTUIFlow simulates the exact TUI flow programmatically
func TestAutomatedTUIFlow(t *testing.T) {
	// Load config to check if we have Gemini API key
	cfg, err := config.LoadConfig()
	if err != nil || cfg.GeminiAPIKey == "" {
		t.Skip("Skipping automated TUI test - Gemini API key not configured")
	}

	// Create a GenerateFlowModel
	model := NewGenerateFlowModel()

	// Set up a test program
	var buf bytes.Buffer
	program := tea.NewProgram(
		model,
		tea.WithInput(nil),
		tea.WithOutput(&buf),
		tea.WithoutRenderer(),
	)

	// Run the program in a goroutine
	done := make(chan error)
	go func() {
		_, err := program.Run()
		done <- err
	}()

	// Send commands to simulate user interaction
	go func() {
		time.Sleep(100 * time.Millisecond) // Let it initialize

		// Step 1: Enter prompt
		program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test image")})
		time.Sleep(50 * time.Millisecond)
		program.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(50 * time.Millisecond)

		// Step 2: Select model (default is already Gemini)
		program.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(50 * time.Millisecond)

		// Step 3: Select size (default is fine)
		program.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(50 * time.Millisecond)

		// Step 4: Select style (default is fine)
		program.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(50 * time.Millisecond)

		// Step 5: Output path (default is fine)
		program.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(50 * time.Millisecond)

		// Step 6: Command preview - verify before continuing
		// Let's just quit here for now
		program.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Logf("Program ended with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		program.Kill()
		t.Fatal("Test timed out")
	}

	t.Logf("Automated TUI test completed")
}

// TestDirectGenerationFlow tests the generation function directly
func TestDirectGenerationFlow(t *testing.T) {
	// Load config to check if we have Gemini API key
	cfg, err := config.LoadConfig()
	if err != nil || cfg.GeminiAPIKey == "" {
		t.Skip("Skipping direct generation test - Gemini API key not configured")
	}

	// Create a GenerateFlowModel
	model := NewGenerateFlowModel()

	// Set values directly
	model.promptTextarea.SetValue("test image")
	model.selectedModel = 0 // Gemini 2.5 Flash
	model.selectedSize = 0  // 1024x1024
	model.selectedStyle = 0 // None
	model.outputInput.SetValue("/tmp/test_tui.png")

	// Log what we're about to generate
	t.Logf("Model: %s", model.models[model.selectedModel].name)
	t.Logf("Prompt: %s", model.promptTextarea.Value())

	// Build the command to see what it would be
	model.buildCommand()
	t.Logf("Command that would be run: %s", model.commandStr)

	// Now simulate what happens when generation starts
	cmd := model.generateImageCmd()

	// Execute the command (this returns a tea.Msg)
	msg := cmd()

	// Check the result
	if completeMsg, ok := msg.(generationCompleteMsg); ok {
		if completeMsg.err != nil {
			t.Errorf("Generation failed: %v", completeMsg.err)
		} else {
			t.Logf("Generation succeeded: %s", completeMsg.path)
		}
	}
}