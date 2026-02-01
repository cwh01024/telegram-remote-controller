package bot

import (
	"context"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockHandler is a test handler
type MockHandler struct {
	LastMessage *tgbotapi.Message
	HandleErr   error
}

func (m *MockHandler) HandleMessage(ctx context.Context, msg *tgbotapi.Message) error {
	m.LastMessage = msg
	return m.HandleErr
}

func TestNewBotWithInvalidToken(t *testing.T) {
	_, err := New("invalid-token", nil)
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestMockHandler(t *testing.T) {
	handler := &MockHandler{}
	msg := &tgbotapi.Message{
		Text: "test message",
	}

	err := handler.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if handler.LastMessage != msg {
		t.Error("Handler did not receive message")
	}
}

// Note: Full integration tests require a real bot token
// Run with: TELEGRAM_BOT_TOKEN=xxx go test -v -run TestIntegration
func TestIntegration(t *testing.T) {
	t.Skip("Requires real bot token - run manually")
}
