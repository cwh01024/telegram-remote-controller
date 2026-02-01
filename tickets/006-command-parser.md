# Ticket 006: Command Parser

## Context
解析來自 Telegram 的指令和參數。將用戶訊息轉換為結構化的 Command。

## Goal
- 解析 `/command` 格式
- 支援選項 `-m model`
- 提取 prompt 內容

## Files
- `internal/command/parser.go`
- `internal/command/parser_test.go`

## Supported Commands
```
/run <prompt>           # 執行 prompt
/run -m claude <prompt> # 指定 model
/status                 # 檢查狀態
/screenshot             # 截圖
/help                   # 說明
```

## Data Structure
```go
type Command struct {
    Name    string            // "run", "status", etc
    Model   string            // 可選 model
    Args    []string          // 其他參數
    Prompt  string            // 主要 prompt 內容
}

func Parse(input string) (*Command, error)
```

## Acceptance Criteria
- [ ] 正確解析 /run
- [ ] 正確解析 -m 參數
- [ ] 處理無效指令

## TDD
```go
func TestParseRun(t *testing.T) {
    cmd, _ := Parse("/run hello world")
    assert.Equal(t, "run", cmd.Name)
    assert.Equal(t, "hello world", cmd.Prompt)
}

func TestParseWithModel(t *testing.T) {
    cmd, _ := Parse("/run -m claude fix bug")
    assert.Equal(t, "claude", cmd.Model)
    assert.Equal(t, "fix bug", cmd.Prompt)
}
```
