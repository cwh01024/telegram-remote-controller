package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	TelegramBotToken string
	AllowedUsers     []int64
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		AllowedUsers:     []int64{}, // Will be populated from config file
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.TelegramBotToken == "" {
		return ErrMissingToken
	}
	return nil
}

// ErrMissingToken is returned when TELEGRAM_BOT_TOKEN is not set
var ErrMissingToken = configError("TELEGRAM_BOT_TOKEN environment variable is not set")

type configError string

func (e configError) Error() string {
	return string(e)
}
