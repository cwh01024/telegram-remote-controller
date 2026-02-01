# Ticket 009: Error Handling

## Context
統一的錯誤處理和重試機制。確保遠端操作穩定可靠。

## Goal
- 統一錯誤類型定義
- 重試機制
- 錯誤回報到 Telegram

## Files
- `internal/errors/errors.go`
- `internal/errors/retry.go`
- `internal/errors/errors_test.go`

## Error Types
```go
type ErrorCode int

const (
    ErrBotConnection ErrorCode = iota
    ErrAuthFailed
    ErrCommandParse
    ErrAutomation
    ErrIDENotReady
    ErrTimeout
)

type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
}
```

## Retry Helper
```go
func Retry(attempts int, delay time.Duration, fn func() error) error
```

## Acceptance Criteria
- [ ] 錯誤能正確分類
- [ ] 重試機制運作正常
- [ ] 錯誤訊息發送到 Telegram

## TDD
```go
func TestRetry(t *testing.T) {
    count := 0
    err := Retry(3, time.Millisecond, func() error {
        count++
        if count < 3 { return errors.New("fail") }
        return nil
    })
    assert.NoError(t, err)
    assert.Equal(t, 3, count)
}
```
