#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-vault 2>/dev/null; then
        echo "Stopping pranor-vault ..."
        systemctl stop pranor-vault || true
        systemctl disable pranor-vault || true
    fi
fi
