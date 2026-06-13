#!/usr/bin/env bash
# flowork-creator-launch.sh — one-click launcher for the Flowork OS Creator (.desktop Exec).
# Build (needs docker) + flash (dd) require root, so the server runs via pkexec (graphical
# password prompt). The browser is opened as the REAL user, and the user's ~/.flowork is passed
# as the Full-profile data default (so it's correct even though the server runs as root).
set -e
SELF="$(cd "$(dirname "$0")" && pwd)"   # .../os/builder
OSDIR="$(cd "$SELF/.." && pwd)"          # .../os
BIN="$SELF/flowork-creator"
ADDR="127.0.0.1:8799"
USERHOME="$HOME"

# Build the binary on first run (needs Go). Prebuilt by install-desktop.sh so this is a fallback.
if [ ! -x "$BIN" ]; then
	( cd "$SELF" && GOWORK=off go build -o "$BIN" . ) || {
		command -v zenity >/dev/null && zenity --error --text="Could not build flowork-creator (is Go installed?)"
		exit 1
	}
fi

# Open the browser as the user once the (elevated) server is listening.
(
	for _ in $(seq 1 60); do
		curl -fsS -o /dev/null "http://$ADDR/api/info" 2>/dev/null && break
		sleep 0.3
	done
	( command -v xdg-open >/dev/null && xdg-open "http://$ADDR" ) 2>/dev/null || true
) &

# Run the server elevated. pkexec pops a graphical auth dialog; keep DISPLAY for any GUI bits.
exec pkexec env DISPLAY="${DISPLAY:-}" XAUTHORITY="${XAUTHORITY:-}" \
	"$BIN" -no-browser -addr "$ADDR" -os "$OSDIR" -data-default "$USERHOME/.flowork"
