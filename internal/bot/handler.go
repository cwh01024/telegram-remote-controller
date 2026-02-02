package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/applejobs/telegram-remote-controller/internal/auth"
	"github.com/applejobs/telegram-remote-controller/internal/command"
	"github.com/applejobs/telegram-remote-controller/internal/controller"
	"github.com/applejobs/telegram-remote-controller/internal/notes"
	"github.com/applejobs/telegram-remote-controller/internal/web"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainHandler handles all incoming messages with full functionality
type MainHandler struct {
	Bot       *Bot
	Auth      *auth.Whitelist
	IDE       *controller.IDEController
	Watcher   *controller.FileWatcher
	NoteStore *notes.Store
	WebServer *web.Server

	// Response watching state
	watchingMutex sync.Mutex
	watchChatID   int64
}

// NewMainHandler creates a new main handler
func NewMainHandler(bot *Bot, allowedUsers []int64) *MainHandler {
	// Initialize components
	noteStore := notes.NewStore()
	webServer := web.NewServer(noteStore, 8080)

	h := &MainHandler{
		Bot:       bot,
		Auth:      auth.NewWhitelist(allowedUsers),
		IDE:       controller.NewIDEController(),
		Watcher:   controller.NewFileWatcher(),
		NoteStore: noteStore,
		WebServer: webServer,
	}

	// Set default watch chat ID to first allowed user
	if len(allowedUsers) > 0 {
		h.watchChatID = allowedUsers[0]
		log.Printf("Default chat ID set to: %d", h.watchChatID)
	}

	// Start background file watcher
	go h.backgroundWatcher()

	// Start Web UI
	go func() {
		if err := h.WebServer.Start(); err != nil {
			log.Printf("Web server failed: %v", err)
		}
	}()

	return h
}

// backgroundWatcher continuously monitors for new response files
func (h *MainHandler) backgroundWatcher() {
	responseDir := h.Watcher.GetWatchDir()
	log.Printf("Background watcher started, monitoring: %s", responseDir)

	// Get initial file state
	initialFiles := h.getFileStates(responseDir)

	for {
		time.Sleep(2 * time.Second)

		// Get current file state
		currentFiles := h.getFileStates(responseDir)

		// Check for new or modified files
		for path, modTime := range currentFiles {
			oldModTime, exists := initialFiles[path]

			// New file or modified file
			if !exists || modTime.After(oldModTime) {
				log.Printf("Detected file change: %s", path)

				// Wait for file to be fully written
				time.Sleep(2 * time.Second)

				// Read content
				content, err := os.ReadFile(path)
				if err != nil {
					log.Printf("Failed to read file: %v", err)
					continue
				}

				// Get the watching chat ID
				h.watchingMutex.Lock()
				chatID := h.watchChatID
				h.watchingMutex.Unlock()

				if chatID != 0 {
					// Format and send
					formatted := h.Watcher.FormatResponseForTelegram(string(content))
					h.Bot.SendText(chatID, fmt.Sprintf("📝 回應：\n\n%s", formatted))
					log.Printf("Sent response to chat %d (%d chars)", chatID, len(formatted))
				} else {
					log.Println("No active chat to send response to")
				}

				// Update file state
				initialFiles[path] = modTime
			}
		}

		// Update initial files with any new files
		for path, modTime := range currentFiles {
			initialFiles[path] = modTime
		}
	}
}

// getFileStates returns modification times for all files in a directory
func (h *MainHandler) getFileStates(dir string) map[string]time.Time {
	states := make(map[string]time.Time)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Only watch text/markdown files
		ext := filepath.Ext(path)
		if ext == ".txt" || ext == ".md" {
			states[path] = info.ModTime()
		}
		return nil
	})

	return states
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

	// Update watching chat ID (so responses go to the right chat)
	h.watchingMutex.Lock()
	h.watchChatID = chatID
	h.watchingMutex.Unlock()

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
	case command.CmdNotes:
		return h.handleNotes(chatID, cmd)
	case command.CmdStatus:
		return h.handleStatus(chatID)
	case command.CmdHelp:
		return h.Bot.SendText(chatID, command.HelpText())
	default:
		return h.Bot.SendText(chatID, "❓ 未知指令，使用 /help 查看說明")
	}
}

// handleRun shows the prompt and waits for response file
func (h *MainHandler) handleRun(chatID int64, cmd *command.Command) error {
	responseDir := h.Watcher.GetWatchDir()

	// Clean up old files
	h.Watcher.CleanupOldFiles(1 * time.Hour)

	// Show the prompt to user and explain the process
	return h.Bot.SendText(chatID, fmt.Sprintf(`🚀 收到 Prompt:
%s

📋 請在 Antigravity 執行此 Prompt

✅ 回應完成後，將回應保存到:
%s/response.md

Bot 會自動偵測並發送回應給你。`, cmd.Prompt, responseDir))
}

// handleNotes adds a note or shows the web UI link
func (h *MainHandler) handleNotes(chatID int64, cmd *command.Command) error {
	if cmd.Prompt == "" {
		// No content, show Web UI info
		count := h.NoteStore.Count()
		return h.Bot.SendText(chatID, fmt.Sprintf(`💡 Ideas / Notes

📝 目前筆記： %d 則
🌐 Web UI: http://localhost:8080

使用方式：
/notes <你的想法> - 新增一則筆記`, count))
	}

	// Add note
	note := h.NoteStore.Add(cmd.Prompt)
	return h.Bot.SendText(chatID, fmt.Sprintf("✅ Idea 已保存！\nID: %s\n\n可在 Web UI 查看。", note.ID))
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

	// Notes count
	notesCount := h.NoteStore.Count()

	h.watchingMutex.Lock()
	watchingChat := h.watchChatID
	h.watchingMutex.Unlock()

	status := fmt.Sprintf(`📊 系統狀態

✅ Bot: 運行中
✅ 背景監聽: 已啟動
🌐 Web UI: http://localhost:8080
📁 回應目錄: %s
   狀態: %s
   檔案數: %d
💡 筆記數: %d
💬 當前 Chat ID: %d

📝 /run <問題> - 執行 prompt
💡 /notes <想法> - 記錄 idea`, responseDir, dirExists, fileCount, notesCount, watchingChat)

	return h.Bot.SendText(chatID, status)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
