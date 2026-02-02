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
	return &Screenshot{
		outputDir: "/tmp",
	}
}

// CaptureScreen captures the screen using Shift+Cmd+3 shortcut
// This saves to Desktop by default, then we move it
func (s *Screenshot) CaptureScreen() (string, error) {
	log.Println("Taking screenshot with Shift+Cmd+3...")

	// Get desktop path
	homeDir, _ := os.UserHomeDir()
	desktopPath := filepath.Join(homeDir, "Desktop")

	// Get existing screenshots before taking new one
	existingFiles := s.getScreenshotFiles(desktopPath)

	// Simulate Shift+Cmd+3 using AppleScript
	script := `
		tell application "System Events"
			key code 20 using {command down, shift down}
		end tell
	`
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("Shift+Cmd+3 failed: %v, falling back to screencapture", err)
		return s.fallbackCapture()
	}

	// Wait for screenshot to be saved
	time.Sleep(1 * time.Second)

	// Find the new screenshot file
	newFile := s.findNewScreenshot(desktopPath, existingFiles)
	if newFile == "" {
		log.Println("Could not find new screenshot, using fallback")
		return s.fallbackCapture()
	}

	// Move to our output dir
	destPath := filepath.Join(s.outputDir, filepath.Base(newFile))
	if err := os.Rename(newFile, destPath); err != nil {
		// If move fails, just use original location
		log.Printf("Could not move screenshot: %v", err)
		return newFile, nil
	}

	log.Printf("Screenshot saved to: %s", destPath)
	return destPath, nil
}

// getScreenshotFiles returns current screenshot files on desktop
func (s *Screenshot) getScreenshotFiles(desktopPath string) map[string]bool {
	files := make(map[string]bool)
	entries, err := os.ReadDir(desktopPath)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		name := entry.Name()
		// macOS screenshot files start with "Screenshot" or "螢幕截圖"
		if strings.HasPrefix(name, "Screenshot") ||
			strings.HasPrefix(name, "螢幕截圖") ||
			strings.HasPrefix(name, "截屏") {
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
			strings.HasPrefix(name, "截屏") {
			fullPath := filepath.Join(desktopPath, name)
			if !existing[fullPath] {
				newFiles = append(newFiles, fullPath)
			}
		}
	}

	if len(newFiles) == 0 {
		return ""
	}

	// Sort by name (which includes timestamp) and return newest
	sort.Strings(newFiles)
	return newFiles[len(newFiles)-1]
}

// fallbackCapture uses screencapture command as fallback
func (s *Screenshot) fallbackCapture() (string, error) {
	log.Println("Using fallback screencapture...")
	filename := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	path := filepath.Join(s.outputDir, filename)

	cmd := exec.Command("screencapture", "-x", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	log.Printf("Fallback screenshot saved to: %s", path)
	return path, nil
}

// CaptureAllDisplays captures all displays
func (s *Screenshot) CaptureAllDisplays() (string, error) {
	return s.CaptureScreen()
}
