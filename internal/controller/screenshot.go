package controller

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Screenshot handles screen capture functionality
type Screenshot struct {
	outputDir string
}

// NewScreenshot creates a new Screenshot instance
func NewScreenshot() *Screenshot {
	// Use absolute path for screenshots directory
	dir := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/screenshots"

	// Ensure directory exists
	os.MkdirAll(dir, 0755)

	return &Screenshot{
		outputDir: dir,
	}
}

// CaptureScreen captures the screen using Cmd+Shift+3 keyboard shortcut
func (s *Screenshot) CaptureScreen() (string, error) {
	log.Println("=== SCREENSHOT START ===")

	// Get the user's home directory for Desktop path
	homeDir, _ := os.UserHomeDir()
	desktopPath := filepath.Join(homeDir, "Desktop")

	// Record existing screenshot files before taking new one
	existingFiles := s.getDesktopScreenshots(desktopPath)
	log.Printf("Found %d existing screenshots on Desktop", len(existingFiles))

	// Use AppleScript to press Cmd+Shift+3
	log.Println("Pressing Cmd+Shift+3...")
	script := `
		tell application "System Events"
			key code 20 using {command down, shift down}
		end tell
	`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Cmd+Shift+3 failed: %v, output: %s", err, string(output))
		// Fall back to screencapture
		return s.fallbackCapture()
	}

	// Wait for screenshot to be saved (macOS saves to Desktop)
	log.Println("Waiting for screenshot file to appear on Desktop...")
	time.Sleep(1500 * time.Millisecond)

	// Find the new screenshot file
	newFile := s.findNewScreenshot(desktopPath, existingFiles)
	if newFile == "" {
		log.Println("No new screenshot found on Desktop, using fallback")
		return s.fallbackCapture()
	}

	log.Printf("Found new screenshot: %s", newFile)

	// Move to our screenshots directory with a new name
	timestamp := time.Now().UnixNano()
	destName := fmt.Sprintf("screen_%d.png", timestamp)
	destPath := filepath.Join(s.outputDir, destName)

	// Copy the file (keeping original on Desktop)
	input, err := os.ReadFile(newFile)
	if err != nil {
		log.Printf("Failed to read screenshot: %v", err)
		return newFile, nil // Return original path
	}

	if err := os.WriteFile(destPath, input, 0644); err != nil {
		log.Printf("Failed to copy screenshot: %v", err)
		return newFile, nil // Return original path
	}

	// Remove the original from Desktop
	os.Remove(newFile)

	log.Printf("=== SCREENSHOT SUCCESS: %s ===", destPath)
	return destPath, nil
}

// getDesktopScreenshots returns current screenshot files on desktop
func (s *Screenshot) getDesktopScreenshots(desktopPath string) map[string]bool {
	files := make(map[string]bool)
	entries, err := os.ReadDir(desktopPath)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		name := entry.Name()
		// macOS screenshot files: "Screenshot YYYY-MM-DD at HH.MM.SS.png" (English)
		// or "螢幕截圖" (Chinese) or "截屏" (Simplified Chinese)
		if strings.HasPrefix(name, "Screenshot") ||
			strings.HasPrefix(name, "螢幕截圖") ||
			strings.HasPrefix(name, "截屏") ||
			strings.HasPrefix(name, "スクリーンショット") {
			files[filepath.Join(desktopPath, name)] = true
		}
	}
	return files
}

// findNewScreenshot finds a screenshot file that wasn't there before
func (s *Screenshot) findNewScreenshot(desktopPath string, existing map[string]bool) string {
	entries, err := os.ReadDir(desktopPath)
	if err != nil {
		return ""
	}

	var newFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "Screenshot") ||
			strings.HasPrefix(name, "螢幕截圖") ||
			strings.HasPrefix(name, "截屏") ||
			strings.HasPrefix(name, "スクリーンショット") {
			fullPath := filepath.Join(desktopPath, name)
			if !existing[fullPath] {
				newFiles = append(newFiles, fullPath)
			}
		}
	}

	if len(newFiles) == 0 {
		return ""
	}

	// Sort and return the newest (last in alphabetical order which includes timestamp)
	sort.Strings(newFiles)
	return newFiles[len(newFiles)-1]
}

// fallbackCapture uses screencapture as fallback
func (s *Screenshot) fallbackCapture() (string, error) {
	log.Println("Using fallback screencapture...")
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("screen_%d.png", timestamp)
	path := filepath.Join(s.outputDir, filename)

	cmd := exec.Command("screencapture", "-x", "-C", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	log.Printf("Fallback screenshot saved: %s", path)
	return path, nil
}

// CaptureAllDisplays captures all displays
func (s *Screenshot) CaptureAllDisplays() (string, error) {
	return s.CaptureScreen()
}
