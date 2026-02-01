#!/bin/bash
# Uninstall script for Telegram Remote Controller

set -e

BINARY_NAME="telegram-remote-controller"
INSTALL_DIR="/usr/local/bin"
PLIST_NAME="com.telegram.remote-controller.plist"
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"

echo "Stopping service..."
launchctl stop com.telegram.remote-controller 2>/dev/null || true
launchctl unload "$LAUNCH_AGENTS_DIR/$PLIST_NAME" 2>/dev/null || true

echo "Removing LaunchAgent..."
rm -f "$LAUNCH_AGENTS_DIR/$PLIST_NAME"

echo "Removing binary..."
sudo rm -f "$INSTALL_DIR/$BINARY_NAME"

echo ""
echo "✅ Uninstall complete!"
