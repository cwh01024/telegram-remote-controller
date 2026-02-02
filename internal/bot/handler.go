package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/applejobs/telegram-remote-controller/internal/auth"
	"github.com/applejobs/telegram-remote-controller/internal/command"
	"github.com/applejobs/telegram-remote-controller/internal/controller"
	"github.com/applejobs/telegram-remote-controller/internal/gemini"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler handles all incoming messages with full functionality
type MainHandler struct {
	Bot     *Bot
	Auth    *auth.Whitelist
	IDE     *controller.IDEController
	Gemini  *gemini.Client
	Capture *controller.ResponseCapture
}

// NewMainHandler creates a new main handler
func NewMainHandler(bot *Bot, allowedUsers []int64) *MainHandler {
	return &MainHandler{
		Bot:     bot,
		Auth:    auth.NewWhitelist(allowedUsers),
		IDE:     controller.NewIDEController(),
		Gemini:  gemini.NewClient(),
		Capture: controller.NewResponseCapture(),
	}
}

// HandleMessage processes incoming messages
func (h *MainHandler) HandleMessage(ctx context.Context, msg *tgbotapi.Message) error {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	// Check authorization
	if !h.Auth.IsAuthorized(userID) {
		log.Printf("Unauthorized access from user %d", userID)
		return h.Bot.SendText(chatID, "⛔ 你沒有使用權限")
	}

	// Parse command
	cmd, err := command.Parse(msg.Text)
	if err != nil {
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ %v", err))
	}

	// Execute command
	switch cmd.Name {
	case command.CmdRun:
		return h.handleRun(chatID, cmd)
	case command.CmdScreenshot:
		return h.handleScreenshot(chatID)
	case command.CmdStatus:
		return h.handleStatus(chatID)
	case command.CmdHelp:
		return h.Bot.SendText(chatID, command.HelpText())
	default:
		return h.Bot.SendText(chatID, "❓ 未知指令，使用 /help 查看說明")
	}
}

// handleRun executes a prompt in Antigravity and captures the response
func (h *MainHandler) handleRun(chatID int64, cmd *command.Command) error {
	h.Bot.SendText(chatID, fmt.Sprintf("🚀 執行中...\nModel: %s\nPrompt: %s",
		orDefault(cmd.Model, "default"), cmd.Prompt))

	// Ensure IDE is ready
	if err := h.IDE.EnsureReady(); err != nil {
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ IDE 未就緒: %v", err))
	}

	// Input the prompt
	if err := h.IDE.InputPrompt(cmd.Prompt); err != nil {
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ 輸入失敗: %v", err))
	}

	// Submit
	if err := h.IDE.Submit(); err != nil {
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ 送出失敗: %v", err))
	}

	h.Bot.SendText(chatID, "✅ 已送出！等待回應中...")

	// Wait for response and capture
	// Wait 10 seconds for initial response
	time.Sleep(10 * time.Second)

	// Take screenshot of the response
	screenshotPath, err := h.IDE.TakeScreenshot()
	if err != nil {
		log.Printf("Failed to capture response: %v", err)
		return h.Bot.SendText(chatID, "✅ 已送出！請使用 /screenshot 查看結果")
	}

	// Send the response screenshot
	h.Bot.SendText(chatID, "📸 回應截圖：")
	if err := h.Bot.SendPhoto(chatID, screenshotPath); err != nil {
		log.Printf("Failed to send response screenshot: %v", err)
		return h.Bot.SendText(chatID, "✅ 已送出！請使用 /screenshot 查看結果")
	}

	return nil
}

// handleScreenshot takes and sends a screenshot
func (h *MainHandler) handleScreenshot(chatID int64) error {
	h.Bot.SendText(chatID, "📸 截圖中...")

	path, err := h.IDE.TakeScreenshot()
	if err != nil {
		log.Printf("Screenshot failed: %v", err)
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ 截圖失敗: %v", err))
	}

	log.Printf("Screenshot saved to: %s", path)

	if err := h.Bot.SendPhoto(chatID, path); err != nil {
		log.Printf("Failed to send photo to Telegram: %v", err)
		return h.Bot.SendText(chatID, fmt.Sprintf("❌ 發送圖片失敗: %v", err))
	}

	return nil
}

// handleStatus returns system status
func (h *MainHandler) handleStatus(chatID int64) error {
	geminiStatus := "❌ 未設定 API Key"
	if h.Gemini.IsAvailable() {
		geminiStatus = "✅ 可用"
	}

	status := fmt.Sprintf(`📊 系統狀態

✅ Bot: 運行中
✅ Auth: 已授權
💻 IDE: Antigravity
🤖 Gemini 摘要: %s

發送 /screenshot 查看螢幕
發送 /run <prompt> 執行並等待回應`, geminiStatus)

	return h.Bot.SendText(chatID, status)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
