# Ticket 001: Project Setup

## Context
建立 Telegram Remote Controller 專案。這個程式會透過 Telegram 遠端控制這台 Mac 上的 Antigravity IDE。

## Goal
- 初始化 Go module
- 建立專案目錄結構
- 配置 GitHub repository
- 設定 Makefile

## Directory Structure
```
telegram-remote-controller/
├── cmd/bot/main.go
├── internal/
│   ├── bot/
│   ├── auth/
│   ├── command/
│   ├── automation/
│   └── controller/
├── scripts/
├── go.mod
├── .gitignore
├── README.md
└── Makefile
```

## Acceptance Criteria
- [x] `go build ./...` 成功
- [x] GitHub repo 建立並推送
- [x] README 說明專案用途

## TDD
```go
// main_test.go - 驗證編譯通過
func TestMain(t *testing.T) {
    // placeholder
}
```
