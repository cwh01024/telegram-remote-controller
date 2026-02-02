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

// Model aliases
var modelAliases = map[string]string{
	"thinking": "Claude Opus 4.5 (Thinking)",
	"opus":     "Claude Opus 4.5 (Thinking)",
	"coding":   "Gemini 3 Pro",
	"gemini":   "Gemini 3 Pro",
	"claude":   "Claude Opus 4.5",
	"sonnet":   "Claude Sonnet 4",
}

// App aliases for common applications
var appAliases = map[string]string{
	"chrome":      "Google Chrome",
	"safari":      "Safari",
	"firefox":     "Firefox",
	"code":        "Visual Studio Code",
	"vscode":      "Visual Studio Code",
	"terminal":    "Terminal",
	"finder":      "Finder",
	"antigravity": "Antigravity",
	"ag":          "Antigravity",
	"slack":       "Slack",
	"discord":     "Discord",
	"notion":      "Notion",
}

// DefaultModel is used when no model is specified
const DefaultModel = "Claude Opus 4.5 (Thinking)"

// Command represents a parsed user command
type Command struct {
	Name    string   // Command name (run, status, screenshot, help)
	Model   string   // Model selection (expanded from alias)
	Args    []string // Additional arguments
	Prompt  string   // The main prompt content (raw, preserved)
	AppName string   // For screenshot: which app to focus first
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
		// Treat non-command messages as run commands with default model
		return &Command{
			Name:   CmdRun,
			Model:  DefaultModel,
			Prompt: input,
		}, nil
	}

	// Find command name (first word)
	spaceIdx := strings.Index(input, " ")
	var cmdPart, rest string
	if spaceIdx == -1 {
		cmdPart = input
		rest = ""
	} else {
		cmdPart = input[:spaceIdx]
		rest = strings.TrimSpace(input[spaceIdx+1:])
	}

	// Extract command name (remove leading /)
	cmdName := strings.ToLower(strings.TrimPrefix(cmdPart, "/"))

	switch cmdName {
	case CmdRun:
		return parseRunCommand(rest)
	case CmdStatus:
		return &Command{Name: CmdStatus}, nil
	case CmdScreenshot:
		return parseScreenshotCommand(rest)
	case CmdHelp:
		return &Command{Name: CmdHelp}, nil
	default:
		return nil, ErrUnknownCommand
	}
}

// parseRunCommand parses a /run command with optional -m flag
// Preserves the entire prompt with spaces
func parseRunCommand(rest string) (*Command, error) {
	cmd := &Command{
		Name:  CmdRun,
		Model: DefaultModel,
	}

	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil, ErrMissingPrompt
	}

	// Check for model flag at the beginning
	if strings.HasPrefix(rest, "-m ") || strings.HasPrefix(rest, "-m\t") {
		// Remove "-m "
		rest = strings.TrimSpace(rest[3:])

		// Find the model name (first word/token)
		spaceIdx := strings.Index(rest, " ")
		if spaceIdx == -1 {
			// Only model name, no prompt
			return nil, ErrMissingPrompt
		}

		modelName := rest[:spaceIdx]
		rest = strings.TrimSpace(rest[spaceIdx+1:])

		// Expand model alias
		cmd.Model = expandModelAlias(modelName)
	}

	// Rest is the entire prompt (preserved with spaces)
	cmd.Prompt = rest

	if cmd.Prompt == "" {
		return nil, ErrMissingPrompt
	}

	return cmd, nil
}

// parseScreenshotCommand parses a /screenshot command with optional app name
func parseScreenshotCommand(rest string) (*Command, error) {
	cmd := &Command{
		Name:    CmdScreenshot,
		AppName: "Antigravity", // Default to Antigravity
	}

	rest = strings.TrimSpace(rest)
	if rest != "" {
		// User specified an app name
		cmd.AppName = expandAppAlias(rest)
	}

	return cmd, nil
}

// expandModelAlias expands a model alias to full name
func expandModelAlias(alias string) string {
	lower := strings.ToLower(alias)
	if full, ok := modelAliases[lower]; ok {
		return full
	}
	// Return as-is if not an alias
	return alias
}

// expandAppAlias expands an app alias to full name
func expandAppAlias(alias string) string {
	lower := strings.ToLower(alias)
	if full, ok := appAliases[lower]; ok {
		return full
	}
	// Return as-is if not an alias (user might have typed the full app name)
	return alias
}

// HelpText returns the help message
func HelpText() string {
	return `🤖 可用指令：

📝 執行 Prompt：
/run <prompt> - 使用預設 model
/run -m <model> <prompt> - 指定 model

🎯 Model 別名：
• thinking / opus → Claude Opus 4.5 (Thinking)
• coding / gemini → Gemini 3 Pro  
• claude → Claude Opus 4.5
• sonnet → Claude Sonnet 4

📸 截圖：
/screenshot - 截取 Antigravity
/screenshot <app> - 截取指定應用程式

📱 App 別名：
• chrome, safari, firefox
• code/vscode, terminal, finder
• ag → Antigravity

🔧 其他：
/status - 檢查系統狀態
/help - 顯示此說明

💡 直接發送文字也會用預設 model 執行！`
}
