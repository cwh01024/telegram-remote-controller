# Ticket 010: Service Daemon

## Context
將 Bot 設為 macOS 開機自啟服務，確保電腦開機後自動運行。

## Goal
- 建立 launchd plist
- 安裝/移除腳本
- 日誌輸出配置

## Files
- `scripts/install.sh`
- `scripts/uninstall.sh`
- `scripts/com.telegram.remote-controller.plist`

## LaunchAgent Plist
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.telegram.remote-controller</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/telegram-remote-controller</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/telegram-remote-controller.log</string>
</dict>
</plist>
```

## Install Script
```bash
#!/bin/bash
cp telegram-remote-controller /usr/local/bin/
cp com.telegram.remote-controller.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.telegram.remote-controller.plist
```

## Acceptance Criteria
- [ ] 服務能隨開機啟動
- [ ] 日誌正確輸出
- [ ] 安裝/移除腳本可用

## TDD
- 手動驗證服務啟動
- 驗證腳本執行正確
