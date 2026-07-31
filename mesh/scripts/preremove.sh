#!/bin/sh
set -e
# Stop the service if running as a systemd unit
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pranor-mesh 2>/dev/null; then
        echo "Stopping pranor-mesh ..."
        systemctl stop pranor-mesh || true
        systemctl disable pranor-mesh || true
    fi
fi
