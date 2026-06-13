#!/bin/bash
# ============================================================================
# Flowork — macOS double-click launcher.
# Double-click in Finder → Terminal opens and boots the full stack
# (router + agent). The agent starts schedules & triggers automatically.
# First run builds the binaries (needs Go 1.25+). Panel: http://127.0.0.1:1987
# ============================================================================
cd "$(dirname "$0")" || exit 1
exec ./start.sh --pause
