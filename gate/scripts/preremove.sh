#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-gate 2>/dev/null; then
        echo "Stopping pranor-gate ..."
        systemctl stop pranor-gate || true
        systemctl disable pranor-gate || true
    fi
fi
