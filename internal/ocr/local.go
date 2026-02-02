package ocr

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
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
	log.Printf("OCR extracted %d characters (before cleanup)", len(text))

	// Clean up the extracted text
	cleaned := o.cleanupText(text)
	log.Printf("OCR cleaned to %d characters", len(cleaned))

	return cleaned, nil
}

// cleanupText removes noise and extracts meaningful content
func (o *LocalOCR) cleanupText(text string) string {
	lines := strings.Split(text, "\n")
	var cleanedLines []string

	// Patterns to skip (UI elements, file names, code artifacts)
	skipPatterns := []regexp.Regexp{
		*regexp.MustCompile(`^\d+月\d+日`),         // Date
		*regexp.MustCompile(`^[上下]午\d+:\d+`),     // Time
		*regexp.MustCompile(`^Open\s`),           // "Open ..."
		*regexp.MustCompile(`^S\s?Code`),         // "S Code"
		*regexp.MustCompile(`^f\d+\s`),           // "f2 Walkthrough"
		*regexp.MustCompile(`^@id:`),             // "@id:..."
		*regexp.MustCompile(`^import[（(]`),       // import statements
		*regexp.MustCompile(`^\"`),               // Quoted strings (code)
		*regexp.MustCompile(`^\d+$`),             // Just numbers (line numbers)
		*regexp.MustCompile(`^Step\s+Id:`),       // "Step Id:"
		*regexp.MustCompile(`uses\s+Open\s+VSX`), // VSX message
		*regexp.MustCompile(`marketplace`),       // marketplace message
		*regexp.MustCompile(`Checked\s+command`), // debug messages
		*regexp.MustCompile(`^回\s`),              // UI icons
		*regexp.MustCompile(`^[〉›>\[\]{}()]+$`),  // Brackets only
	}

	// Patterns that indicate actual response content
	responseIndicators := []string{
		"✅", "❌", "🚀", "📝", "📸", "🔍", "⏱️", "⚠️",
		"回應", "已送出", "執行", "完成", "失敗", "成功",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip very short lines
		if len(line) < 3 {
			continue
		}

		// Check skip patterns
		skip := false
		for _, pattern := range skipPatterns {
			if pattern.MatchString(line) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Keep lines with response indicators
		hasIndicator := false
		for _, indicator := range responseIndicators {
			if strings.Contains(line, indicator) {
				hasIndicator = true
				break
			}
		}

		// Keep lines that look like Chinese sentences or contain indicators
		isChinese := regexp.MustCompile(`[\p{Han}]{3,}`).MatchString(line)

		if hasIndicator || isChinese || len(line) > 20 {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// Join and return
	result := strings.Join(cleanedLines, "\n")

	// If we filtered too much, return more of the original
	if len(result) < 50 && len(text) > 100 {
		log.Println("Cleanup was too aggressive, returning more content")
		return text
	}

	return result
}

// IsAvailable checks if local OCR is available
func (o *LocalOCR) IsAvailable() bool {
	// Check if swift is available
	_, err := exec.LookPath("swift")
	return err == nil
}
