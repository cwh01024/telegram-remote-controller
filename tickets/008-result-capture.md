# Ticket 008: Result Capture

## Context
擷取 Antigravity 的回應結果並回報到 Telegram。這是最複雜的部分，需要偵測執行完成。

## Goal
- 偵測 IDE 回應完成
- 擷取回應文字
- 回報到 Telegram

## Files
- `internal/controller/capture.go`
- `internal/controller/capture_test.go`

## Strategies
1. **輪詢截圖 + OCR** - 複雜但通用
2. **剪貼簿** - 手動複製後讀取
3. **檔案監控** - 如果有 log 檔案

## Proposed Approach
使用簡化流程：
1. 等待固定時間 或 偵測 UI 變化
2. 截圖回傳讓你自己看
3. 可選：使用 `pbpaste` 讀取剪貼簿

```go
type ResultCapture interface {
    // 等待執行完成（timeout 秒）
    WaitForCompletion(timeout time.Duration) error
    
    // 截圖當前狀態
    CaptureCurrentState() (string, error)
}
```

## Acceptance Criteria
- [x] 能等待指定時間
- [x] 能截圖當前狀態
- [x] 回報結果到 Telegram

## TDD
```go
func TestWaitAndCapture(t *testing.T) {
    cap := NewResultCapture()
    err := cap.WaitForCompletion(5 * time.Second)
    assert.NoError(t, err)
}
```
