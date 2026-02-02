package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/applejobs/telegram-remote-controller/internal/auth"
	"github.com/applejobs/telegram-remote-controller/internal/command"
	"github.com/applejobs/telegram-remote-controller/internal/controller"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler handles all incoming messages with full functionality
type MainHandler struct {
	Bot     *Bot
	Auth    *auth.Whitelist
	IDE     *controller.IDEController
	Watcher *controller.FileWatcher
}

// NewMainHandler creates a new main handler
func NewMainHandler(bot *Bot, allowedUsers []int64) *MainHandler {
	return &MainHandler{
		Bot:     bot,
		Auth:    auth.NewWhitelist(allowedUsers),
		IDE:     controller.NewIDEController(),
		Watcher: controller.NewFileWatcher(),
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

// handleRun executes a prompt in Antigravity and watches for response file
func (h *MainHandler) handleRun(chatID int64, cmd *command.Command) error {
	h.Bot.SendText(chatID, fmt.Sprintf("🚀 執行中...\nModel: %s\nPrompt: %s",
		orDefault(cmd.Model, "default"), cmd.Prompt))

	// Record time before submission
	startTime := time.Now()

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

	responseDir := h.Watcher.GetWatchDir()
	h.Bot.SendText(chatID, fmt.Sprintf("✅ 已送出！\n\n監聽回應目錄: %s\n\n將回應寫入上述目錄的 .txt 或 .md 檔案即可收到通知。", responseDir))

	// Clean up old response files
	h.Watcher.CleanupOldFiles(1 * time.Hour)

	// Wait for response file
	content, err := h.Watcher.WaitForLatestResponse(startTime)
	if err != nil {
		log.Printf("File watcher timed out: %v", err)
		return h.Bot.SendText(chatID, "⏱️ 等待回應檔案超時（3分鐘）。\n\n請將回應寫入: "+responseDir)
	}

	// Format and send response
	formatted := h.Watcher.FormatResponseForTelegram(content)
	return h.Bot.SendText(chatID, fmt.Sprintf("📝 回應：\n\n%s", formatted))
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
	responseDir := h.Watcher.GetWatchDir()

	// Check if response directory exists
	dirExists := "✅ 存在"
	if _, err := os.Stat(responseDir); os.IsNotExist(err) {
		dirExists = "❌ 不存在"
	}

	// Count response files
	files, _ := filepath.Glob(filepath.Join(responseDir, "*"))
	fileCount := len(files)

	status := fmt.Sprintf(`📊 系統狀態

✅ Bot: 運行中
✅ Auth: 已授權
💻 IDE: Antigravity
📁 回應目錄: %s
   狀態: %s
   檔案數: %d

📝 使用方式:
1. 發送 /run <問題>
2. Antigravity 回應後，將內容保存到回應目錄
3. Bot 自動偵測並發送給你

📸 /screenshot <app> 截取指定應用程式`, responseDir, dirExists, fileCount)

	return h.Bot.SendText(chatID, status)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
