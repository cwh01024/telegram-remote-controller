package controller

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// ClipboardMonitor monitors clipboard for changes
type ClipboardMonitor struct {
	pollInterval time.Duration
	timeout      time.Duration
}

// NewClipboardMonitor creates a new clipboard monitor
func NewClipboardMonitor() *ClipboardMonitor {
	return &ClipboardMonitor{
		pollInterval: 500 * time.Millisecond,
		timeout:      60 * time.Second, // Wait up to 60 seconds for response
	}
}

// GetClipboard reads the current clipboard content
func (m *ClipboardMonitor) GetClipboard() (string, error) {
	cmd := exec.Command("pbpaste")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard: %w", err)
	}
	return string(output), nil
}

// SetClipboard sets the clipboard content
func (m *ClipboardMonitor) SetClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// WaitForChange waits for clipboard content to change from the initial value
// Returns the new clipboard content when it changes
func (m *ClipboardMonitor) WaitForChange(initialContent string) (string, error) {
	log.Printf("Monitoring clipboard for changes (timeout: %v)...", m.timeout)

	startTime := time.Now()
	lastContent := initialContent

	for {
		// Check timeout
		if time.Since(startTime) > m.timeout {
			return "", fmt.Errorf("clipboard monitoring timed out after %v", m.timeout)
		}

		// Read current clipboard
		currentContent, err := m.GetClipboard()
		if err != nil {
			log.Printf("Warning: failed to read clipboard: %v", err)
			time.Sleep(m.pollInterval)
			continue
		}

		// Check if content changed
		if currentContent != lastContent && currentContent != "" {
			log.Printf("Clipboard changed! New content length: %d chars", len(currentContent))
			return currentContent, nil
		}

		time.Sleep(m.pollInterval)
	}
}

// WaitForNewContent clears the clipboard and waits for new content
func (m *ClipboardMonitor) WaitForNewContent() (string, error) {
	// Clear clipboard first
	log.Println("Clearing clipboard...")
	if err := m.SetClipboard(""); err != nil {
		log.Printf("Warning: failed to clear clipboard: %v", err)
	}

	// Wait for new content
	return m.WaitForChange("")
}
