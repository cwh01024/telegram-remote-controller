package controller

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ResponseCapture handles capturing responses from the IDE
type ResponseCapture struct {
	screenshotDir string
}

// NewResponseCapture creates a new response capture instance
func NewResponseCapture() *ResponseCapture {
	dir := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/screenshots"
	os.MkdirAll(dir, 0755)
	return &ResponseCapture{
		screenshotDir: dir,
	}
}

// WaitAndCapture waits for a response and captures a screenshot
func (r *ResponseCapture) WaitAndCapture(waitDuration time.Duration) (string, error) {
	log.Printf("Waiting %v for response...", waitDuration)

	// Wait for the specified duration
	time.Sleep(waitDuration)

	// Take a screenshot of the result
	return r.captureScreen()
}

// CaptureText tries to extract text from the screen using OCR
// For now, this returns a screenshot path since OCR is complex
func (r *ResponseCapture) CaptureText() (string, error) {
	// Use screencapture and return the path
	// In the future, this could use OCR to extract text
	return r.captureScreen()
}

// captureScreen captures the current screen
func (r *ResponseCapture) captureScreen() (string, error) {
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("response_%d.png", timestamp)
	path := filepath.Join(r.screenshotDir, filename)

	log.Printf("Capturing response screenshot: %s", path)

	cmd := exec.Command("screencapture", "-x", "-C", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("screenshot file not created: %w", err)
	}

	log.Printf("Response captured: %s", path)
	return path, nil
}
