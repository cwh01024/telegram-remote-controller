package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/applejobs/telegram-remote-controller/internal/automation"
)

const (
	// AntigravityAppName is the name of the Antigravity application
	AntigravityAppName = "Antigravity"

	// DefaultInputDelay is the delay between input operations
	DefaultInputDelay = 100 * time.Millisecond
)

// IDEController controls the Antigravity IDE
type IDEController struct {
	appName    string
	inputDelay time.Duration
}

// NewIDEController creates a new IDE controller
func NewIDEController() *IDEController {
	return &IDEController{
		appName:    AntigravityAppName,
		inputDelay: DefaultInputDelay,
	}
}

// EnsureReady ensures the IDE is open and focused
func (c *IDEController) EnsureReady() error {
	log.Printf("Ensuring %s is ready...", c.appName)

	// Open and focus the app
	if err := automation.OpenApp(c.appName); err != nil {
		return fmt.Errorf("failed to open %s: %w", c.appName, err)
	}

	// Wait for app to be ready
	time.Sleep(500 * time.Millisecond)

	return nil
}

// InputPrompt inputs a prompt into the IDE
// Uses clipboard to paste long text for reliability
func (c *IDEController) InputPrompt(prompt string) error {
	log.Printf("Inputting prompt (%d chars)", len(prompt))

	// Ensure IDE is focused
	if err := c.EnsureReady(); err != nil {
		return err
	}

	// For long prompts, use clipboard
	if len(prompt) > 100 {
		return c.inputViaClipboard(prompt)
	}

	// For short prompts, type directly
	return c.inputViaTyping(prompt)
}

// inputViaClipboard inputs text by copying to clipboard and pasting
func (c *IDEController) inputViaClipboard(text string) error {
	// Save current clipboard (optional, for restoration)
	// oldClipboard, _ := automation.GetClipboard()

	// Set new clipboard content
	if err := automation.SetClipboard(text); err != nil {
		return fmt.Errorf("failed to set clipboard: %w", err)
	}

	time.Sleep(c.inputDelay)

	// Paste
	if err := automation.PasteFromClipboard(); err != nil {
		return fmt.Errorf("failed to paste: %w", err)
	}

	// Restore old clipboard (optional)
	// automation.SetClipboard(oldClipboard)

	return nil
}

// inputViaTyping inputs text by simulating keyboard typing
func (c *IDEController) inputViaTyping(text string) error {
	if err := automation.TypeText(text); err != nil {
		return fmt.Errorf("failed to type text: %w", err)
	}
	return nil
}

// Submit sends the current input (press Enter or Cmd+Enter)
func (c *IDEController) Submit() error {
	log.Println("Submitting prompt...")

	time.Sleep(c.inputDelay)

	// Try Cmd+Enter first (common for chat-style interfaces)
	if err := automation.PressCommandEnter(); err != nil {
		// Fallback to Enter
		return automation.PressEnter()
	}

	return nil
}

// SelectModel attempts to select a specific model in the IDE
// Note: Implementation depends on IDE's UI
func (c *IDEController) SelectModel(model string) error {
	log.Printf("Selecting model: %s", model)

	// This is IDE-specific and may need adjustment
	// For now, we'll log and return nil
	// TODO: Implement model selection based on Antigravity's UI

	return nil
}

// ClearInput clears the current input field
func (c *IDEController) ClearInput() error {
	// Select all and delete
	script := `
		tell application "System Events"
			keystroke "a" using command down
			key code 51
		end tell
	`
	_, err := automation.RunScript(script)
	return err
}

// TakeScreenshot takes a screenshot of the current state
func (c *IDEController) TakeScreenshot() (string, error) {
	screenshot := NewScreenshot()
	return screenshot.CaptureScreen()
}
