# Ticket 005: IDE Controller

## Context
控制 Antigravity IDE 的專用模組。依賴 Ticket 004 的 automation 層。

## Goal
- 開啟/聚焦 Antigravity
- 輸入 prompt 到 IDE
- 觸發執行（送出訊息）
- 等待回應

## Files
- `internal/controller/ide.go`
- `internal/controller/ide_test.go`

## Key Interface
```go
type IDEController interface {
    // 確保 IDE 開啟並聚焦
    EnsureReady() error
    
    // 輸入 prompt
    InputPrompt(prompt string) error
    
    // 送出（按 Enter 或快捷鍵）
    Submit() error
    
    // 選擇 model（如果支援）
    SelectModel(model string) error
}
```

## Implementation Notes
- Antigravity 是 Electron app
- 輸入區可能需要先 focus
- 送出可能是 Cmd+Enter 或 Enter

## Acceptance Criteria
- [ ] 能開啟 Antigravity
- [ ] 能輸入文字到輸入框
- [ ] 能觸發送出

## TDD
```go
func TestInputPrompt(t *testing.T) {
    ctrl := NewIDEController()
    err := ctrl.InputPrompt("test prompt")
    assert.NoError(t, err)
}
```
