# Ticket 007: Screenshot

## Context
截取螢幕畫面並回傳給 Telegram。讓你遠端查看 IDE 狀態。

## Goal
- 使用 `screencapture` 截圖
- 可選截取整個螢幕或特定視窗
- 回傳圖片到 Telegram

## Files
- `internal/automation/screenshot.go`
- `internal/automation/screenshot_test.go`

## Key Functions
```go
// 截取整個螢幕
func CaptureScreen() (string, error)  // returns file path

// 截取特定視窗
func CaptureWindow(appName string) (string, error)

// 清理舊截圖
func CleanupOldScreenshots() error
```

## Implementation
```bash
# 截取螢幕
screencapture -x /tmp/screenshot.png

# 截取視窗（互動式）
screencapture -w /tmp/screenshot.png
```

## Acceptance Criteria
- [x] 能截圖並儲存
- [x] 能透過 Telegram 發送
- [x] 自動清理舊檔

## TDD
```go
func TestCaptureScreen(t *testing.T) {
    path, err := CaptureScreen()
    assert.NoError(t, err)
    assert.FileExists(t, path)
}
```
