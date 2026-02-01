package command

import (
	"testing"
)

func TestParseRunCommand(t *testing.T) {
	cmd, err := Parse("/run hello world")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Name != CmdRun {
		t.Errorf("Expected name 'run', got '%s'", cmd.Name)
	}
	if cmd.Prompt != "hello world" {
		t.Errorf("Expected prompt 'hello world', got '%s'", cmd.Prompt)
	}
}

func TestParseRunWithModel(t *testing.T) {
	cmd, err := Parse("/run -m claude fix this bug")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Model != "claude" {
		t.Errorf("Expected model 'claude', got '%s'", cmd.Model)
	}
	if cmd.Prompt != "fix this bug" {
		t.Errorf("Expected prompt 'fix this bug', got '%s'", cmd.Prompt)
	}
}

func TestParseStatus(t *testing.T) {
	cmd, err := Parse("/status")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Name != CmdStatus {
		t.Errorf("Expected name 'status', got '%s'", cmd.Name)
	}
}

func TestParseScreenshot(t *testing.T) {
	cmd, err := Parse("/screenshot")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Name != CmdScreenshot {
		t.Errorf("Expected name 'screenshot', got '%s'", cmd.Name)
	}
}

func TestParseHelp(t *testing.T) {
	cmd, err := Parse("/help")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Name != CmdHelp {
		t.Errorf("Expected name 'help', got '%s'", cmd.Name)
	}
}

func TestParseNonCommand(t *testing.T) {
	// Non-command messages should be treated as run
	cmd, err := Parse("just a regular message")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if cmd.Name != CmdRun {
		t.Errorf("Expected name 'run', got '%s'", cmd.Name)
	}
	if cmd.Prompt != "just a regular message" {
		t.Errorf("Expected prompt 'just a regular message', got '%s'", cmd.Prompt)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err != ErrEmptyInput {
		t.Errorf("Expected ErrEmptyInput, got %v", err)
	}
}

func TestParseUnknown(t *testing.T) {
	_, err := Parse("/unknown")
	if err != ErrUnknownCommand {
		t.Errorf("Expected ErrUnknownCommand, got %v", err)
	}
}

func TestParseRunMissingPrompt(t *testing.T) {
	_, err := Parse("/run")
	if err != ErrMissingPrompt {
		t.Errorf("Expected ErrMissingPrompt, got %v", err)
	}
}

func TestParseRunWithModelMissingPrompt(t *testing.T) {
	_, err := Parse("/run -m claude")
	if err != ErrMissingPrompt {
		t.Errorf("Expected ErrMissingPrompt, got %v", err)
	}
}

func TestCaseInsensitive(t *testing.T) {
	commands := []string{"/RUN test", "/Run test", "/run test"}
	for _, input := range commands {
		cmd, err := Parse(input)
		if err != nil {
			t.Errorf("Parse(%s) failed: %v", input, err)
			continue
		}
		if cmd.Name != CmdRun {
			t.Errorf("Parse(%s): expected 'run', got '%s'", input, cmd.Name)
		}
	}
}

func TestHelpText(t *testing.T) {
	help := HelpText()
	if help == "" {
		t.Error("HelpText() returned empty string")
	}
}
