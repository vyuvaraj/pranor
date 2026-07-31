#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-trace 2>/dev/null; then
        echo "Stopping pranor-trace ..."
        systemctl stop pranor-trace || true
        systemctl disable pranor-trace || true
    fi
fi
