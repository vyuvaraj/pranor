#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-flow 2>/dev/null; then
        echo "Stopping pranor-flow ..."
        systemctl stop pranor-flow || true
        systemctl disable pranor-flow || true
    fi
fi
