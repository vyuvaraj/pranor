#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-pool 2>/dev/null; then
        echo "Stopping pranor-pool ..."
        systemctl stop pranor-pool || true
        systemctl disable pranor-pool || true
    fi
fi
