package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/applejobs/telegram-remote-controller/internal/auth"
	"github.com/applejobs/telegram-remote-controller/internal/command"
	"github.com/applejobs/telegram-remote-controller/internal/controller"
	"github.com/applejobs/telegram-remote-controller/internal/ocr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler handles all incoming messages with full functionality
type MainHandler struct {
	Bot     *Bot
	Auth    *auth.Whitelist
	IDE     *controller.IDEController
	OCR     *ocr.LocalOCR
	Monitor *controller.ResponseMonitor
}

// NewMainHandler creates a new main handler
func NewMainHandler(bot *Bot, allowedUsers []int64) *MainHandler {
	return &MainHandler{
		Bot:     bot,
		Auth:    auth.NewWhitelist(allowedUsers),
		IDE:     controller.NewIDEController(),
		OCR:     ocr.NewLocalOCR(),
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

// handleRun executes a prompt in Antigravity and extracts the response
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

	h.Bot.SendText(chatID, "✅ 已送出！等待回應中...\n（每 5 秒監測，穩定 10 秒後提取回應）")

	// Cleanup old screenshots
	h.Monitor.CleanupOldScreenshots()

	// Wait for screen to stabilize (response complete)
	screenshotPath, err := h.Monitor.WaitForStableScreen()
	if err != nil {
		log.Printf("Response monitoring failed: %v", err)
		return h.Bot.SendText(chatID, "⏱️ 監測超時。使用 /screenshot 查看結果。")
	}

	// Use local OCR to extract text
	h.Bot.SendText(chatID, "🔍 正在讀取回應內容（本地 OCR）...")

	responseText, err := h.OCR.ExtractText(screenshotPath)
	if err != nil {
		log.Printf("Local OCR failed: %v", err)
		// Fallback to sending screenshot
		h.Bot.SendText(chatID, "⚠️ 文字提取失敗，發送截圖：")
		return h.Bot.SendPhoto(chatID, screenshotPath)
	}

	// Send the extracted response text
	if len(responseText) > 4000 {
		// Telegram has 4096 char limit
		return h.Bot.SendText(chatID, fmt.Sprintf("📝 回應：\n\n%s...\n\n_（已截斷，完整回應 %d 字）_", responseText[:4000], len(responseText)))
	}

	return h.Bot.SendText(chatID, fmt.Sprintf("📝 回應：\n\n%s", responseText))
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
	ocrStatus := "❌ 不可用"
	if h.OCR.IsAvailable() {
		ocrStatus = "✅ macOS Vision OCR"
	}

	status := fmt.Sprintf(`📊 系統狀態

✅ Bot: 運行中
✅ Auth: 已授權
💻 IDE: Antigravity
📸 回應偵測: 每 5 秒監測，穩定 10 秒
🔍 文字提取: %s

📝 /run 會監測螢幕並用本地 OCR 提取文字回應
📸 /screenshot <app> 截取指定應用程式`, ocrStatus)

	return h.Bot.SendText(chatID, status)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
