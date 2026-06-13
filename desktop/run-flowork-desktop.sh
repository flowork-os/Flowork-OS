#!/usr/bin/env bash
# run-flowork-desktop.sh — the Flowork DESKTOP app: run the agent + router pair locally and
# open the control panel in your browser. One command, clean shutdown on Ctrl-C.
#
#   env: FLOWORK_HOME (default ~/.flowork), AGENT_ADDR (127.0.0.1:1987),
#        ROUTER_ADDR (127.0.0.1:2402), NO_BROWSER=1 to skip opening a browser.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SELF_DIR/.." && pwd)"
HOMEDIR="${FLOWORK_HOME:-$HOME/.flowork}"
AGENT_ADDR="${AGENT_ADDR:-127.0.0.1:1987}"
ROUTER_ADDR="${ROUTER_ADDR:-127.0.0.1:2402}"
BIN="$REPO/desktop/bin"; mkdir -p "$BIN" "$HOMEDIR"

build() { # $1 module dir, $2 out name
	[ -x "$BIN/$2" ] && [ -z "${FLOWORK_REBUILD:-}" ] && return 0
	echo "[desktop] building $2 …"
	( cd "$REPO/$1" && CGO_ENABLED=0 go build -ldflags '-s -w' -o "$BIN/$2" . )
}
build agent  flowork-agent
build router  flowork-router

PIDS=()
stop() {
	echo; echo "[desktop] stopping…"
	for p in "${PIDS[@]:-}"; do [ -n "$p" ] && kill "$p" 2>/dev/null || true; done
	wait 2>/dev/null || true
	echo "[desktop] stopped. (Your data is safe in $HOMEDIR)"
}
trap stop INT TERM EXIT

echo "[desktop] router → $ROUTER_ADDR"
HOME="$HOMEDIR" FLOWORK_HOME="$HOMEDIR" FLOWORK_ROUTER_ADDR="$ROUTER_ADDR" "$BIN/flowork-router" &
PIDS+=($!)
sleep 1
echo "[desktop] agent  → $AGENT_ADDR"
HOME="$HOMEDIR" FLOWORK_HOME="$HOMEDIR" "$BIN/flowork-agent" -addr "$AGENT_ADDR" &
PIDS+=($!)

# Wait for the panel, then open a browser.
URL="http://$AGENT_ADDR/"
for _ in $(seq 1 40); do
	curl -fsS -o /dev/null "$URL" 2>/dev/null && break
	# a 3xx (login redirect) also means the agent is up
	curl -sS -o /dev/null "$URL" 2>/dev/null && break
	sleep 0.5
done
echo "[desktop] panel ready: $URL   (router drawer reachable from the panel header)"
if [ -z "${NO_BROWSER:-}" ]; then
	( command -v xdg-open >/dev/null && xdg-open "$URL" ) 2>/dev/null \
		|| ( command -v open >/dev/null && open "$URL" ) 2>/dev/null \
		|| echo "[desktop] open $URL in your browser"
fi
echo "[desktop] running. Press Ctrl+C to stop both."
wait
