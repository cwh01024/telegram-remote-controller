package ocr

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// LocalOCR uses macOS Vision framework for text recognition
type LocalOCR struct {
	scriptPath string
}

// NewLocalOCR creates a new local OCR instance
func NewLocalOCR() *LocalOCR {
	// Path to the Swift OCR script
	scriptPath := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/scripts/ocr.swift"
	return &LocalOCR{
		scriptPath: scriptPath,
	}
}

// ExtractText extracts text from an image using macOS Vision
func (o *LocalOCR) ExtractText(imagePath string) (string, error) {
	log.Printf("Running local OCR on: %s", imagePath)

	// Get absolute path
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		absPath = imagePath
	}

	// Run the Swift OCR script
	cmd := exec.Command("swift", o.scriptPath, absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("OCR failed: %w, output: %s", err, string(output))
	}

	text := strings.TrimSpace(string(output))
	log.Printf("OCR extracted %d characters", len(text))

	return text, nil
}

// ExtractResponseFromScreenshot extracts AI response from IDE screenshot
// Filters out UI elements and extracts actual response content
func (o *LocalOCR) ExtractResponseFromScreenshot(imagePath string) (string, error) {
	// First, get all text from the image
	fullText, err := o.ExtractText(imagePath)
	if err != nil {
		return "", err
	}

	// Filter out common UI elements (menu bar, file names, etc.)
	lines := strings.Split(fullText, "\n")
	var responseLines []string

	// Common UI elements to skip
	skipPrefixes := []string{
		"Antigravity", "File", "Edit", "Selection", "View", "Go", "Run",
		"Terminal", "Window", "Help", "Extensions", "GitHub",
		"applejobs", "package ", "import", "func ", "type ", "const ",
		"//", "/*", "*/",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip very short lines (likely UI fragments)
		if len(line) < 3 {
			continue
		}

		// Skip lines that look like UI elements
		skip := false
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(line, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		responseLines = append(responseLines, line)
	}

	if len(responseLines) == 0 {
		return fullText, nil // Return full text if filtering removes everything
	}

	return strings.Join(responseLines, "\n"), nil
}

// IsAvailable checks if local OCR is available
func (o *LocalOCR) IsAvailable() bool {
	// Check if swift is available
	_, err := exec.LookPath("swift")
	return err == nil
}
