package controller

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileWatcher monitors a directory for new files
type FileWatcher struct {
	watchDir     string
	pollInterval time.Duration
	timeout      time.Duration
	stableDelay  time.Duration // Wait after file appears to ensure it's fully written
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher() *FileWatcher {
	// Create a dedicated response output directory
	watchDir := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/responses"
	os.MkdirAll(watchDir, 0755)
	
	return &FileWatcher{
		watchDir:     watchDir,
		pollInterval: 2 * time.Second,
		timeout:      180 * time.Second, // 3 minutes timeout
		stableDelay:  3 * time.Second,   // Wait 3 seconds after file appears
	}
}

// GetWatchDir returns the directory being watched
func (w *FileWatcher) GetWatchDir() string {
	return w.watchDir
}

// WaitForNewFile waits for a new file to appear in the watch directory
// Returns the path to the new file and its contents
func (w *FileWatcher) WaitForNewFile() (string, string, error) {
	log.Printf("Watching for new files in: %s (timeout: %v)", w.watchDir, w.timeout)
	
	// Get initial file list
	initialFiles := w.getFiles()
	log.Printf("Initial files: %d", len(initialFiles))
	
	startTime := time.Now()
	
	for {
		// Check timeout
		if time.Since(startTime) > w.timeout {
			return "", "", fmt.Errorf("file watcher timed out after %v", w.timeout)
		}
		
		// Get current files
		currentFiles := w.getFiles()
		
		// Find new files
		for path, modTime := range currentFiles {
			if _, exists := initialFiles[path]; !exists {
				// New file found!
				log.Printf("New file detected: %s", path)
				
				// Wait for file to be fully written
				log.Printf("Waiting %v for file to stabilize...", w.stableDelay)
				time.Sleep(w.stableDelay)
				
				// Read file content
				content, err := os.ReadFile(path)
				if err != nil {
					log.Printf("Error reading file: %v", err)
					continue
				}
				
				return path, string(content), nil
			}
			
			// Check if existing file was modified
			if oldModTime, exists := initialFiles[path]; exists {
				if modTime.After(oldModTime) {
					log.Printf("File modified: %s", path)
					time.Sleep(w.stableDelay)
					
					content, err := os.ReadFile(path)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						continue
					}
					
					return path, string(content), nil
				}
			}
		}
		
		time.Sleep(w.pollInterval)
	}
}

// WaitForLatestResponse waits for and returns the most recent response file
func (w *FileWatcher) WaitForLatestResponse(afterTime time.Time) (string, error) {
	log.Printf("Waiting for response file after: %v", afterTime)
	
	startTime := time.Now()
	
	for {
		if time.Since(startTime) > w.timeout {
			return "", fmt.Errorf("timeout waiting for response file")
		}
		
		// Get all response files
		files := w.getFilesWithInfo()
		
		// Find files modified after the start time
		var recentFiles []fileInfo
		for _, f := range files {
			if f.modTime.After(afterTime) {
				recentFiles = append(recentFiles, f)
			}
		}
		
		if len(recentFiles) > 0 {
			// Sort by modification time (newest first)
			sort.Slice(recentFiles, func(i, j int) bool {
				return recentFiles[i].modTime.After(recentFiles[j].modTime)
			})
			
			newest := recentFiles[0]
			log.Printf("Found recent response: %s (modified: %v)", newest.path, newest.modTime)
			
			// Wait for file to be fully written
			time.Sleep(w.stableDelay)
			
			// Read content
			content, err := os.ReadFile(newest.path)
			if err != nil {
				return "", fmt.Errorf("failed to read response file: %w", err)
			}
			
			return string(content), nil
		}
		
		time.Sleep(w.pollInterval)
	}
}

type fileInfo struct {
	path    string
	modTime time.Time
}

// getFiles returns a map of file paths to modification times
func (w *FileWatcher) getFiles() map[string]time.Time {
	files := make(map[string]time.Time)
	
	filepath.WalkDir(w.watchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		// Only watch text/markdown files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".txt" || ext == ".md" || ext == ".json" {
			info, err := d.Info()
			if err == nil {
				files[path] = info.ModTime()
			}
		}
		return nil
	})
	
	return files
}

// getFilesWithInfo returns file info for all watched files
func (w *FileWatcher) getFilesWithInfo() []fileInfo {
	var files []fileInfo
	
	filepath.WalkDir(w.watchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			files = append(files, fileInfo{path: path, modTime: info.ModTime()})
		}
		return nil
	})
	
	return files
}

// FormatResponseForTelegram formats the response content for Telegram
func (w *FileWatcher) FormatResponseForTelegram(content string) string {
	// Clean up the content
	lines := strings.Split(content, "\n")
	var cleaned []string
	
	for _, line := range lines {
		// Remove excessive whitespace
		line = strings.TrimRight(line, " \t")
		cleaned = append(cleaned, line)
	}
	
	// Join and trim
	result := strings.Join(cleaned, "\n")
	result = strings.TrimSpace(result)
	
	// Truncate if too long for Telegram
	if len(result) > 4000 {
		result = result[:4000] + "\n\n...(已截斷)"
	}
	
	return result
}

// CleanupOldFiles removes files older than a certain age
func (w *FileWatcher) CleanupOldFiles(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	
	filepath.WalkDir(w.watchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			os.Remove(path)
			log.Printf("Cleaned up old file: %s", path)
		}
		return nil
	})
}
