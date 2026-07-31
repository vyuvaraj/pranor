#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-hub 2>/dev/null; then
        echo "Stopping pranor-hub ..."
        systemctl stop pranor-hub || true
        systemctl disable pranor-hub || true
    fi
fi
