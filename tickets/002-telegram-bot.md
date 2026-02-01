# Ticket 002: Telegram Bot

## Context
實作 Telegram Bot 基礎功能，連接 Telegram API 收發訊息。這是用戶遠端控制的介面入口。

## Goal
- 建立 Bot client 連接 Telegram
- 接收訊息 (long polling)
- 發送文字/圖片回覆
- 優雅處理關閉

## Files
- `internal/bot/client.go`
- `internal/bot/client_test.go`
- `config/config.go`

## Key Interface
```go
type Bot interface {
    Start(ctx context.Context) error
    Stop() error
    SendText(chatID int64, text string) error
    SendPhoto(chatID int64, photoPath string) error
}
```

## Env Vars
- `TELEGRAM_BOT_TOKEN`

## Acceptance Criteria
- [x] Bot 連接成功
- [x] 能接收並 log 訊息
- [x] 能發送回覆

## TDD
```go
func TestBotSendMessage(t *testing.T) {
    // Mock API, verify message sent
}
```
