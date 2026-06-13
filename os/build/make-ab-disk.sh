#!/usr/bin/env bash
# make-ab-disk.sh — build a writable A/B disk image (ext4) from a built squashfs + verity
# sidecars: two slots (b initially = a) + an `active` pointer = a. Populated with
# `mkfs.ext4 -d` so NO root is needed. This is the writable USB shape that A/B update needs.
#   args: OUT_DIR TAG [SIZE]
set -euo pipefail
OUT="$1"; TAG="$2"; SIZE="${3:-2G}"
SQ="$OUT/$TAG.rootfs.squashfs"
[ -f "$SQ" ] || { echo "no squashfs: $SQ" >&2; exit 1; }
command -v mkfs.ext4 >/dev/null || { echo "mkfs.ext4 (e2fsprogs) required" >&2; exit 1; }

STAGE="$(dirname "$OUT")/.work/ab-stage"; rm -rf "$STAGE"
mkdir -p "$STAGE/flowork/slot-a" "$STAGE/flowork/slot-b"
for s in a b; do
	cp "$SQ" "$STAGE/flowork/slot-$s/rootfs.squashfs"
	for ext in verity roothash roothash.sig; do
		[ -f "$SQ.$ext" ] && cp "$SQ.$ext" "$STAGE/flowork/slot-$s/rootfs.squashfs.$ext"
	done
done
printf 'a' > "$STAGE/flowork/active"

IMG="$OUT/$TAG.ab-disk.img"
rm -f "$IMG"; truncate -s "$SIZE" "$IMG"
mkfs.ext4 -F -q -L FLOWORKAB -d "$STAGE" "$IMG"
echo "A/B disk -> $IMG ($SIZE; slots a+b, active=a)"
