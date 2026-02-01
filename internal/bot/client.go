package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler handles incoming messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg *tgbotapi.Message) error
}

// Bot represents the Telegram bot client
type Bot struct {
	api     *tgbotapi.BotAPI
	handler MessageHandler
}

// New creates a new Bot instance
func New(token string, handler MessageHandler) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:     api,
		handler: handler,
	}, nil
}

// Start begins polling for updates
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopping...")
			b.api.StopReceivingUpdates()
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if b.handler != nil {
				if err := b.handler.HandleMessage(ctx, update.Message); err != nil {
					log.Printf("Error handling message: %v", err)
				}
			}
		}
	}
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

// SendText sends a text message to a chat
func (b *Bot) SendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}

// SendPhoto sends a photo to a chat
func (b *Bot) SendPhoto(chatID int64, photoPath string) error {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(photoPath))
	_, err := b.api.Send(photo)
	return err
}

// SendMarkdown sends a markdown-formatted message
func (b *Bot) SendMarkdown(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	return err
}
