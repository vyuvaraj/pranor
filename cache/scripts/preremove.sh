#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-cache 2>/dev/null; then
        echo "Stopping pranor-cache ..."
        systemctl stop pranor-cache || true
        systemctl disable pranor-cache || true
    fi
fi
