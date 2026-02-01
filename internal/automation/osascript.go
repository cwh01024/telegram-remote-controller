package automation

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunScript executes an AppleScript and returns the output
func RunScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("osascript error: %w, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunScriptMultiLine executes a multi-line AppleScript
func RunScriptMultiLine(lines []string) (string, error) {
	args := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		args = append(args, "-e", line)
	}

	cmd := exec.Command("osascript", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("osascript error: %w, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// OpenApp opens an application by name
func OpenApp(appName string) error {
	script := fmt.Sprintf(`tell application "%s" to activate`, appName)
	_, err := RunScript(script)
	return err
}

// IsAppRunning checks if an application is running
func IsAppRunning(appName string) (bool, error) {
	script := fmt.Sprintf(`
		tell application "System Events"
			set appList to name of every process
			return appList contains "%s"
		end tell
	`, appName)
	result, err := RunScript(script)
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// TypeText types text using System Events
func TypeText(text string) error {
	// Escape special characters for AppleScript
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

	script := fmt.Sprintf(`
		tell application "System Events"
			keystroke "%s"
		end tell
	`, escaped)
	_, err := RunScript(script)
	return err
}

// TypeTextSlowly types text with a delay between characters (for stability)
func TypeTextSlowly(text string, delayMs int) error {
	for _, char := range text {
		escaped := strings.ReplaceAll(string(char), "\"", "\\\"")
		script := fmt.Sprintf(`
			tell application "System Events"
				keystroke "%s"
				delay %f
			end tell
		`, escaped, float64(delayMs)/1000.0)
		if _, err := RunScript(script); err != nil {
			return err
		}
	}
	return nil
}

// PressKey presses a key with optional modifiers
// key: the key to press (e.g., "return", "tab", "escape")
// modifiers: optional modifiers (e.g., "command", "shift", "option", "control")
func PressKey(key string, modifiers ...string) error {
	var script string
	if len(modifiers) > 0 {
		modList := strings.Join(modifiers, " down, ") + " down"
		script = fmt.Sprintf(`
			tell application "System Events"
				key code %s using {%s}
			end tell
		`, getKeyCode(key), modList)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				key code %s
			end tell
		`, getKeyCode(key))
	}
	_, err := RunScript(script)
	return err
}

// PressEnter presses the Enter/Return key
func PressEnter() error {
	script := `
		tell application "System Events"
			key code 36
		end tell
	`
	_, err := RunScript(script)
	return err
}

// PressCommandEnter presses Command+Enter
func PressCommandEnter() error {
	script := `
		tell application "System Events"
			key code 36 using command down
		end tell
	`
	_, err := RunScript(script)
	return err
}

// getKeyCode returns the key code for common keys
func getKeyCode(key string) string {
	codes := map[string]string{
		"return":    "36",
		"enter":     "36",
		"tab":       "48",
		"escape":    "53",
		"space":     "49",
		"delete":    "51",
		"backspace": "51",
		"up":        "126",
		"down":      "125",
		"left":      "123",
		"right":     "124",
	}
	if code, ok := codes[strings.ToLower(key)]; ok {
		return code
	}
	return key // Assume it's already a key code
}

// SetClipboard sets the system clipboard content
func SetClipboard(text string) error {
	escaped := strings.ReplaceAll(text, "\"", "\\\"")
	script := fmt.Sprintf(`set the clipboard to "%s"`, escaped)
	_, err := RunScript(script)
	return err
}

// GetClipboard gets the system clipboard content
func GetClipboard() (string, error) {
	script := `the clipboard as text`
	return RunScript(script)
}

// PasteFromClipboard pastes from clipboard
func PasteFromClipboard() error {
	script := `
		tell application "System Events"
			keystroke "v" using command down
		end tell
	`
	_, err := RunScript(script)
	return err
}
