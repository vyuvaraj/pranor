#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-auth 2>/dev/null; then
        echo "Stopping pranor-auth ..."
        systemctl stop pranor-auth || true
        systemctl disable pranor-auth || true
    fi
fi
