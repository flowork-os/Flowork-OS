#!/bin/bash
# Flowork — macOS double-click STOP launcher (agent + router).
cd "$(dirname "$0")" || exit 1
exec ./stop.sh --pause
