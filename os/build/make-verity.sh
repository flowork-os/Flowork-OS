#!/usr/bin/env bash
# make-verity.sh — build a dm-verity hash tree for a rootfs image, and optionally SIGN its
# root hash (chaining with P4a). This is the anti-tamper moat (design Section 7): every block
# of the read-only root is hashed into a Merkle tree; a single changed byte is detected at
# read time. Needs root (veritysetup uses device-mapper).
#   make-verity.sh <rootfs.squashfs> [--sign]
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMG="${1:?usage: make-verity.sh <rootfs.squashfs> [--sign]}"
[ -f "$IMG" ] || { echo "image not found: $IMG" >&2; exit 1; }
command -v veritysetup >/dev/null || { echo "veritysetup (cryptsetup) required" >&2; exit 1; }

# `veritysetup format` only reads the data + writes the hash tree — no device-mapper, so
# it does NOT need root (unlike `verify`, which sets up a dm device).
HASH="$IMG.verity"; RH="$IMG.roothash"
veritysetup format "$IMG" "$HASH" | sed -n 's/^Root hash:[[:space:]]*//p' | tr -d '[:space:]' > "$RH"
echo "[verity] hash tree -> $HASH"
echo "[verity] root hash -> $RH ($(cat "$RH"))"

if [ "${2:-}" = "--sign" ]; then
	# Sign the root hash so the appliance can verify a signed-then-integrity-checked root
	# (full chain of trust: signed root hash [P4a] -> dm-verity-verified squashfs [P3b]).
	"$SELF_DIR/sign-release.sh" "$RH"
fi
