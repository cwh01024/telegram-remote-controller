package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// EchoHandler is a simple handler that echoes messages back
type EchoHandler struct {
	Bot *Bot
}

// HandleMessage echoes the received message
func (h *EchoHandler) HandleMessage(ctx context.Context, msg *tgbotapi.Message) error {
	reply := "收到: " + msg.Text
	log.Printf("Echoing to %d: %s", msg.Chat.ID, reply)
	return h.Bot.SendText(msg.Chat.ID, reply)
}
