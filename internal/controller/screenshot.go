package controller

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// Screenshot handles screen capture functionality
type Screenshot struct {
	outputDir string
}

// NewScreenshot creates a new Screenshot instance
func NewScreenshot() *Screenshot {
	return &Screenshot{
		outputDir: "/tmp",
	}
}

// CaptureScreen captures the entire screen
func (s *Screenshot) CaptureScreen() (string, error) {
	filename := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	path := filepath.Join(s.outputDir, filename)

	cmd := exec.Command("screencapture", "-x", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	return path, nil
}

// CaptureWindow captures a specific window (interactive)
func (s *Screenshot) CaptureWindow() (string, error) {
	filename := fmt.Sprintf("window_%d.png", time.Now().Unix())
	path := filepath.Join(s.outputDir, filename)

	// -w flag captures a window (requires user to click on it)
	cmd := exec.Command("screencapture", "-x", "-w", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	return path, nil
}

// CaptureRect captures a specific screen rectangle
func (s *Screenshot) CaptureRect(x, y, width, height int) (string, error) {
	filename := fmt.Sprintf("rect_%d.png", time.Now().Unix())
	path := filepath.Join(s.outputDir, filename)

	// -R flag captures a specific rectangle
	rect := fmt.Sprintf("%d,%d,%d,%d", x, y, width, height)
	cmd := exec.Command("screencapture", "-x", "-R", rect, path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	return path, nil
}

// CleanupOld removes screenshots older than the specified duration
func (s *Screenshot) CleanupOld(maxAge time.Duration) error {
	pattern := filepath.Join(s.outputDir, "screenshot_*.png")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	_ = maxAge // TODO: Implement age-based cleanup
	for _, f := range files {
		// For simplicity, just log files found
		_ = f
	}

	return nil
}
