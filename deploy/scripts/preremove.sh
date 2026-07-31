#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-deploy 2>/dev/null; then
        echo "Stopping pranor-deploy ..."
        systemctl stop pranor-deploy || true
        systemctl disable pranor-deploy || true
    fi
fi
