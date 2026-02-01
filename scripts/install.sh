#!/bin/bash
# Install script for Telegram Remote Controller

set -e

BINARY_NAME="telegram-remote-controller"
INSTALL_DIR="/usr/local/bin"
PLIST_NAME="com.telegram.remote-controller.plist"
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"

echo "Building $BINARY_NAME..."
cd "$(dirname "$0")/.."
go build -o "$BINARY_NAME" ./cmd/bot

echo "Installing binary to $INSTALL_DIR..."
sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Installing LaunchAgent..."
mkdir -p "$LAUNCH_AGENTS_DIR"
cp "scripts/$PLIST_NAME" "$LAUNCH_AGENTS_DIR/"

echo "Loading LaunchAgent..."
launchctl load "$LAUNCH_AGENTS_DIR/$PLIST_NAME" 2>/dev/null || true

echo ""
echo "✅ Installation complete!"
echo ""
echo "Before starting, set your bot token:"
echo "  export TELEGRAM_BOT_TOKEN='your-token-here'"
echo ""
echo "To start the service:"
echo "  launchctl start com.telegram.remote-controller"
echo ""
echo "To check status:"
echo "  launchctl list | grep telegram"
echo ""
