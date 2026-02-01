package command

import (
	"errors"
	"strings"
)

// Command types
const (
	CmdRun        = "run"
	CmdStatus     = "status"
	CmdScreenshot = "screenshot"
	CmdHelp       = "help"
)

// Command represents a parsed user command
type Command struct {
	Name   string   // Command name (run, status, screenshot, help)
	Model  string   // Optional model selection
	Args   []string // Additional arguments
	Prompt string   // The main prompt content
}

// Errors
var (
	ErrEmptyInput     = errors.New("empty input")
	ErrUnknownCommand = errors.New("unknown command")
	ErrMissingPrompt  = errors.New("missing prompt")
)

// Parse parses a user message into a Command
func Parse(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, ErrEmptyInput
	}

	// Check if it's a command (starts with /)
	if !strings.HasPrefix(input, "/") {
		// Treat non-command messages as run commands
		return &Command{
			Name:   CmdRun,
			Prompt: input,
		}, nil
	}

	// Split into parts
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, ErrEmptyInput
	}

	// Extract command name (remove leading /)
	cmdName := strings.ToLower(strings.TrimPrefix(parts[0], "/"))

	switch cmdName {
	case CmdRun:
		return parseRunCommand(parts[1:])
	case CmdStatus:
		return &Command{Name: CmdStatus}, nil
	case CmdScreenshot:
		return &Command{Name: CmdScreenshot}, nil
	case CmdHelp:
		return &Command{Name: CmdHelp}, nil
	default:
		return nil, ErrUnknownCommand
	}
}

// parseRunCommand parses a /run command with optional -m flag
func parseRunCommand(args []string) (*Command, error) {
	cmd := &Command{
		Name: CmdRun,
	}

	// Check for model flag
	i := 0
	for i < len(args) {
		if args[i] == "-m" && i+1 < len(args) {
			cmd.Model = args[i+1]
			i += 2
			continue
		}
		break
	}

	// Rest is the prompt
	if i < len(args) {
		cmd.Prompt = strings.Join(args[i:], " ")
	}

	if cmd.Prompt == "" {
		return nil, ErrMissingPrompt
	}

	return cmd, nil
}

// HelpText returns the help message
func HelpText() string {
	return `可用指令：

/run <prompt> - 執行 prompt
/run -m <model> <prompt> - 使用指定 model 執行
/status - 檢查系統狀態
/screenshot - 截取螢幕畫面
/help - 顯示此說明

支援的 models: gemini, claude, o3

直接發送文字也會被當作 /run 處理。`
}
