package controller

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ResponseMonitor monitors for response completion by comparing screenshots
type ResponseMonitor struct {
	screenshotDir string
	pollInterval  time.Duration
	stableCount   int // Number of identical screenshots to consider stable
	timeout       time.Duration
}

// NewResponseMonitor creates a new response monitor
func NewResponseMonitor() *ResponseMonitor {
	dir := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/screenshots"
	os.MkdirAll(dir, 0755)
	return &ResponseMonitor{
		screenshotDir: dir,
		pollInterval:  5 * time.Second,   // Check every 5 seconds
		stableCount:   2,                 // Need 2 identical screenshots (10 seconds stable)
		timeout:       120 * time.Second, // Wait up to 2 minutes
	}
}

// WaitForStableScreen waits for the screen to stop changing
// Returns the path to the final screenshot
func (m *ResponseMonitor) WaitForStableScreen() (string, error) {
	log.Printf("Monitoring screen for stable state (timeout: %v)...", m.timeout)

	startTime := time.Now()
	var lastHash []byte
	stableCounter := 0
	var lastScreenshot string

	for {
		// Check timeout
		if time.Since(startTime) > m.timeout {
			if lastScreenshot != "" {
				log.Println("Timeout reached, returning last screenshot")
				return lastScreenshot, nil
			}
			return "", fmt.Errorf("response monitoring timed out after %v", m.timeout)
		}

		// Capture screenshot
		screenshotPath, err := m.captureScreen()
		if err != nil {
			log.Printf("Warning: failed to capture screen: %v", err)
			time.Sleep(m.pollInterval)
			continue
		}

		// Calculate hash
		currentHash, err := m.fileHash(screenshotPath)
		if err != nil {
			log.Printf("Warning: failed to hash screenshot: %v", err)
			time.Sleep(m.pollInterval)
			continue
		}

		// Compare with last hash
		if bytes.Equal(currentHash, lastHash) {
			stableCounter++
			log.Printf("Screen stable (%d/%d)...", stableCounter, m.stableCount)

			if stableCounter >= m.stableCount {
				log.Printf("Screen stable! Response complete.")
				return screenshotPath, nil
			}
		} else {
			// Screen changed, reset counter
			if stableCounter > 0 {
				log.Println("Screen changed, resetting stability counter")
			}
			stableCounter = 0
			lastHash = currentHash
			lastScreenshot = screenshotPath
		}

		time.Sleep(m.pollInterval)
	}
}

// captureScreen takes a quick screenshot for comparison
func (m *ResponseMonitor) captureScreen() (string, error) {
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("monitor_%d.png", timestamp)
	path := filepath.Join(m.screenshotDir, filename)

	cmd := exec.Command("screencapture", "-x", "-C", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture failed: %w", err)
	}

	return path, nil
}

// fileHash calculates MD5 hash of a file
func (m *ResponseMonitor) fileHash(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum(data)
	return hash[:], nil
}

// CleanupOldScreenshots removes old monitoring screenshots
func (m *ResponseMonitor) CleanupOldScreenshots() {
	entries, err := os.ReadDir(m.screenshotDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-10 * time.Minute)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Remove files older than 10 minutes starting with "monitor_"
		if info.ModTime().Before(cutoff) && len(entry.Name()) > 8 && entry.Name()[:8] == "monitor_" {
			os.Remove(filepath.Join(m.screenshotDir, entry.Name()))
		}
	}
}
