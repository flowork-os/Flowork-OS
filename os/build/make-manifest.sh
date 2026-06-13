#!/usr/bin/env bash
# make-manifest.sh — emit the release manifest.json that drives auto-update. Devices poll this tiny
# file from the public release, compare versions, and pull only what changed (per component), with
# sha256 + compatibility gates. Run after the artifacts exist in OUT.
#
#   usage: make-manifest.sh <VERSION> <RELEASED_ISO> [OUT_DIR]
#   stdout: manifest.json
set -euo pipefail
VERSION="${1:?usage: make-manifest.sh <VERSION> <RELEASED_ISO> [OUT_DIR]}"
RELEASED="${2:?need RELEASED timestamp (ISO8601) — pass from CI, scripts cannot read the clock}"
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${3:-$SELF_DIR/../out}"
TAG="flowork-os-$VERSION"

sha()  { if [ -f "$1" ]; then sha256sum "$1" | cut -d' ' -f1; else echo ""; fi; }
size() { if [ -f "$1" ]; then stat -c%s "$1" 2>/dev/null || echo 0; else echo 0; fi; }

# Print a JSON object for an asset (filename under OUT), or null when absent. $2 = optional extra
# trailing JSON (e.g. compression note), already comma-prefixed.
emit_obj() {
	local name="$1" extra="${2:-}" p="$OUT/$1"
	if [ -f "$p" ]; then
		printf '{ "asset": "%s", "sha256": "%s", "size": %s%s }' "$name" "$(sha "$p")" "$(size "$p")" "$extra"
	else
		printf 'null'
	fi
}

AGENT="$(emit_obj flowork-agent)"
ROUTER="$(emit_obj flowork-router)"
OSIMG="$(emit_obj "$TAG-public.usb.img.zst" ', "compressed": "zstd"')"
PORTABLE="$(emit_obj flowork-portable.zip)"

cat <<JSON
{
  "schema": 1,
  "version": "$VERSION",
  "released": "$RELEASED",
  "channel": "stable",
  "min_compatible": { "agent": "$VERSION", "router": "$VERSION" },
  "components": {
    "agent_linux_amd64": $AGENT,
    "router_linux_amd64": $ROUTER
  },
  "os_image": $OSIMG,
  "portable": $PORTABLE
}
JSON
