#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-tunnel 2>/dev/null; then
        echo "Stopping pranor-tunnel ..."
        systemctl stop pranor-tunnel || true
        systemctl disable pranor-tunnel || true
    fi
fi
