package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	// Parse allowed users from env
	allowedUsers := parseAllowedUsers()
	log.Printf("Allowed users: %v", allowedUsers)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create bot (handler will be set after)
	telegramBot, err := bot.New(cfg.TelegramBotToken, nil)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create main handler with auth
	handler := bot.NewMainHandler(telegramBot, allowedUsers)

	// Recreate bot with handler
	telegramBot, _ = bot.New(cfg.TelegramBotToken, handler)
	handler.Bot = telegramBot

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

// parseAllowedUsers parses ALLOWED_USER_ID from environment
func parseAllowedUsers() []int64 {
	env := os.Getenv("ALLOWED_USER_ID")
	if env == "" {
		// Default to allowing all (for testing)
		log.Println("Warning: ALLOWED_USER_ID not set, allowing all users")
		return nil
	}

	var users []int64
	for _, s := range strings.Split(env, ",") {
		s = strings.TrimSpace(s)
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			users = append(users, id)
		}
	}
	return users
}
