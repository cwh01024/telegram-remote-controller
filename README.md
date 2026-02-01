# Telegram Remote Controller

透過 Telegram Bot 遠端控制 Mac 上的 Antigravity IDE。

## 功能

- 🤖 Telegram Bot 介面
- 🔐 用戶白名單驗證
- ⌨️ macOS 自動化控制 IDE
- 📸 截圖回報
- 🚀 開機自啟服務

## 安裝

```bash
go build -o telegram-remote-controller ./cmd/bot
./scripts/install.sh
```

## 使用指令

```
/run <prompt>           # 執行 prompt
/run -m claude <prompt> # 指定 model
/status                 # 檢查狀態
/screenshot             # 截圖
/help                   # 說明
```

## 配置

設定環境變數：
```bash
export TELEGRAM_BOT_TOKEN="your-bot-token"
```

## 開發

```bash
go test ./... -v
go build ./...
```

## License

MIT
