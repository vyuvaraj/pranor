#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-notify 2>/dev/null; then
        echo "Stopping pranor-notify ..."
        systemctl stop pranor-notify || true
        systemctl disable pranor-notify || true
    fi
fi
