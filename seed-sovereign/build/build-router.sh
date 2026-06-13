#!/usr/bin/env bash
# build-router.sh — build the sovereign Flowork router as a fully static linux/amd64
# binary (the form the appliance ships). Source resolves: env > sibling > GitHub clone.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"; mkdir -p "$OUT"
CACHE="$REPO_DIR/.cache"

GH="https://github.com/flowork-os/Flowork-OS"
SRC="${ROUTER_SRC:-}"
if [ -z "$SRC" ]; then
	for c in "$REPO_DIR/../flowork_Router"; do
		[ -f "$c/go.mod" ] && { SRC="$c"; break; }
	done
fi
if [ -z "$SRC" ]; then
	mkdir -p "$CACHE"
	[ -d "$CACHE/flowork_Router/.git" ] || git clone --depth 1 "$GH" "$CACHE/flowork_Router"
	SRC="$CACHE/flowork_Router"
fi
echo "[router] source: $SRC"

( cd "$SRC" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -tags netgo -ldflags '-s -w' -o "$OUT/flowork-router" . )
file "$OUT/flowork-router" | grep -q 'statically linked' || { echo "not static" >&2; exit 1; }
echo "[router] built $OUT/flowork-router ($(du -h "$OUT/flowork-router" | cut -f1), static)"
