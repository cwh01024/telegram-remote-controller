# Ticket 004: macOS Automation

## Context
建立 macOS 自動化基礎層，使用 osascript (AppleScript) 執行系統操作。

## Goal
- 封裝 osascript 執行
- 實作常用操作：開啟 App、輸入文字、按鍵
- 錯誤處理

## Files
- `internal/automation/osascript.go`
- `internal/automation/osascript_test.go`

## Key Functions
```go
// 執行 AppleScript
func RunScript(script string) (string, error)

// 開啟應用程式
func OpenApp(appName string) error

// 輸入文字
func TypeText(text string) error

// 按鍵
func PressKey(key string, modifiers ...string) error
```

## Example Scripts
```applescript
-- 開啟 App
tell application "Antigravity" to activate

-- 輸入文字
tell application "System Events"
    keystroke "Hello World"
end tell
```

## Acceptance Criteria
- [x] 能執行 AppleScript
- [x] 能開啟指定 App
- [x] 能模擬鍵盤輸入

## TDD
```go
func TestOpenApp(t *testing.T) {
    err := OpenApp("Finder")
    assert.NoError(t, err)
}
```
