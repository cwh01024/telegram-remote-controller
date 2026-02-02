package bot

import (
	"context"
	"fmt"
	"log"

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
	Monitor *controller.ResponseMonitor
}

// NewMainHandler creates a new main handler
func NewMainHandler(bot *Bot, allowedUsers []int64) *MainHandler {
	return &MainHandler{
		Bot:     bot,
		Auth:    auth.NewWhitelist(allowedUsers),
		IDE:     controller.NewIDEController(),
		Gemini:  gemini.NewClient(),
		Monitor: controller.NewResponseMonitor(),
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
		return h.handleScreenshot(chatID, cmd.AppName)
	case command.CmdStatus:
		return h.handleStatus(chatID)
	case command.CmdHelp:
		return h.Bot.SendText(chatID, command.HelpText())
	default:
		return h.Bot.SendText(chatID, "❓ 未知指令，使用 /help 查看說明")
	}
}

// handleRun executes a prompt in Antigravity and waits for response
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

	h.Bot.SendText(chatID, "✅ 已送出！等待回應完成中...（監測螢幕變化）")

	// Cleanup old screenshots
	h.Monitor.CleanupOldScreenshots()

	// Wait for screen to stabilize (response complete)
	screenshotPath, err := h.Monitor.WaitForStableScreen()
	if err != nil {
		log.Printf("Response monitoring failed: %v", err)
		return h.Bot.SendText(chatID, "⏱️ 監測超時。使用 /screenshot 查看結果。")
	}

	// Send the response screenshot
	h.Bot.SendText(chatID, "📸 回應完成：")
	if err := h.Bot.SendPhoto(chatID, screenshotPath); err != nil {
		log.Printf("Failed to send response screenshot: %v", err)
		return h.Bot.SendText(chatID, "❌ 發送截圖失敗，請使用 /screenshot")
	}

	return nil
}

// handleScreenshot takes and sends a screenshot of the specified app
func (h *MainHandler) handleScreenshot(chatID int64, appName string) error {
	h.Bot.SendText(chatID, fmt.Sprintf("📸 截圖 %s 中...", appName))

	// Focus the specified app first
	log.Printf("Focusing app: %s", appName)
	if err := h.IDE.FocusApp(appName); err != nil {
		log.Printf("Warning: failed to focus %s: %v", appName, err)
	}

	path, err := h.IDE.TakeScreenshotRaw()
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
📸 回應偵測: 螢幕輪詢（6秒穩定）
🤖 Gemini 摘要: %s

📝 /run 會監測螢幕等待回應完成
📸 /screenshot <app> 截取指定應用程式`, geminiStatus)

	return h.Bot.SendText(chatID, status)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
