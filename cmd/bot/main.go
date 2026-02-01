package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/applejobs/telegram-remote-controller/config"
	"github.com/applejobs/telegram-remote-controller/internal/bot"
)

func main() {
	log.Println("Starting Telegram Remote Controller...")

	// Load config
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create bot (handler will be set after)
	telegramBot, err := bot.New(cfg.TelegramBotToken, nil)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Set echo handler for testing
	echoHandler := &bot.EchoHandler{Bot: telegramBot}
	telegramBot, _ = bot.New(cfg.TelegramBotToken, echoHandler)
	echoHandler.Bot = telegramBot

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Start bot
	log.Println("Bot is running. Press Ctrl+C to stop.")
	if err := telegramBot.Start(ctx); err != nil {
		log.Printf("Bot stopped: %v", err)
	}

	log.Println("Goodbye!")
}
