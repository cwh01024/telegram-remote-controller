package automation

import (
	"strings"
	"testing"
)

func TestRunScript(t *testing.T) {
	// Test simple script that returns a value
	result, err := RunScript(`return "hello"`)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestRunScriptMath(t *testing.T) {
	result, err := RunScript(`return 2 + 2`)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	if result != "4" {
		t.Errorf("Expected '4', got '%s'", result)
	}
}

func TestOpenApp(t *testing.T) {
	// Open Finder (always available on macOS)
	err := OpenApp("Finder")
	if err != nil {
		t.Errorf("OpenApp(Finder) failed: %v", err)
	}
}

func TestIsAppRunning(t *testing.T) {
	// Note: This test may fail due to System Events permission
	// If it times out, ensure Terminal/your IDE has Accessibility permissions
	running, err := IsAppRunning("Finder")
	if err != nil {
		t.Skipf("Skipping due to System Events permission: %v", err)
	}
	if !running {
		t.Error("Finder should be running")
	}
}

func TestClipboard(t *testing.T) {
	testText := "telegram-remote-controller-test"

	// Set clipboard
	err := SetClipboard(testText)
	if err != nil {
		t.Fatalf("SetClipboard failed: %v", err)
	}

	// Get clipboard
	result, err := GetClipboard()
	if err != nil {
		t.Fatalf("GetClipboard failed: %v", err)
	}

	if result != testText {
		t.Errorf("Clipboard content mismatch: expected '%s', got '%s'", testText, result)
	}
}

func TestGetKeyCode(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"return", "36"},
		{"enter", "36"},
		{"tab", "48"},
		{"escape", "53"},
		{"RETURN", "36"}, // Test case insensitivity
	}

	for _, tt := range tests {
		result := getKeyCode(tt.key)
		if result != tt.expected {
			t.Errorf("getKeyCode(%s) = %s, want %s", tt.key, result, tt.expected)
		}
	}
}

func TestRunScriptMultiLine(t *testing.T) {
	lines := []string{
		`set x to 5`,
		`set y to 10`,
		`return x + y`,
	}
	result, err := RunScriptMultiLine(lines)
	if err != nil {
		t.Fatalf("RunScriptMultiLine failed: %v", err)
	}
	if strings.TrimSpace(result) != "15" {
		t.Errorf("Expected '15', got '%s'", result)
	}
}
