#!/usr/bin/env bash
# make-data-disk.sh — create a blank DATA disk image for Flowork OS persistence (P1).
# The image is initialized as LUKS by the appliance itself on first boot (flowork-data
# service), so this is just an empty sparse file. No host root/crypto needed.
#   args: [SIZE]   (default 1G)
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$(cd "$SELF_DIR/.." && pwd)/out"
SIZE="${1:-1G}"
mkdir -p "$OUT"
IMG="$OUT/flowork-data.img"
# Remove any existing image first: `truncate` on an existing file keeps its contents,
# which would leave a stale LUKS header (from a prior build, with a different key) and
# make first-boot init think the disk is already encrypted.
rm -f "$IMG"
truncate -s "$SIZE" "$IMG"
echo "[data] blank DATA disk -> $IMG ($SIZE, sparse; LUKS-initialized on first boot)"
