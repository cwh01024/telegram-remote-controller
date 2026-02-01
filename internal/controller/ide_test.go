package controller

import (
	"os"
	"testing"
)

func TestNewIDEController(t *testing.T) {
	ctrl := NewIDEController()
	if ctrl == nil {
		t.Fatal("NewIDEController returned nil")
	}
	if ctrl.appName != AntigravityAppName {
		t.Errorf("Expected app name %s, got %s", AntigravityAppName, ctrl.appName)
	}
}

func TestInputPromptShort(t *testing.T) {
	// Skip in CI or when IDE is not available
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	ctrl := NewIDEController()
	// This would actually open the IDE and type, so skip in automated tests
	t.Skip("Requires manual testing with IDE open")

	_ = ctrl.InputPrompt("test")
}

func TestScreenshotCapture(t *testing.T) {
	s := NewScreenshot()
	path, err := s.CaptureScreen()
	if err != nil {
		t.Fatalf("CaptureScreen failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Screenshot file was not created")
	}

	// Cleanup
	os.Remove(path)
}

func TestNewScreenshot(t *testing.T) {
	s := NewScreenshot()
	if s == nil {
		t.Fatal("NewScreenshot returned nil")
	}
	if s.outputDir != "/tmp" {
		t.Errorf("Expected output dir /tmp, got %s", s.outputDir)
	}
}
