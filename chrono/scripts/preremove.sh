#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-chrono 2>/dev/null; then
        echo "Stopping pranor-chrono ..."
        systemctl stop pranor-chrono || true
        systemctl disable pranor-chrono || true
    fi
fi
