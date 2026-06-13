#!/usr/bin/env bash
# update-usb.sh — refresh an already-flashed Flowork OS USB IN PLACE from the current build:
# rewrites the ESP (kernel + initramfs + grub.cfg) and both A/B slots (squashfs + verity
# sidecars), and resets the active pointer to A. Much faster than a full re-flash; the result
# is functionally identical to make-usb-image.sh. Needs root. Does NOT touch the DATA partition.
#   args: DEVICE   (e.g. /dev/sda)
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
DEV="${1:?usage: update-usb.sh /dev/sdX}"
SQ="$OUT/$TAG.rootfs.squashfs"; KERNEL="$OUT/$TAG.vmlinuz"; INITRD="$OUT/$TAG.initramfs-small.gz"
for f in "$SQ" "$KERNEL" "$INITRD"; do [ -f "$f" ] || { echo "missing $f — run build-flowork-os.sh first" >&2; exit 1; }; done

# SAFETY: only ever touch a USB device that carries our partition labels.
[ "$(lsblk -dno TRAN "$DEV" 2>/dev/null)" = usb ] || { echo "$DEV is not a USB device — refusing" >&2; exit 1; }
ESP_PART="$(lsblk -lnpo NAME,LABEL "$DEV" 2>/dev/null | awk '$2=="FLOWORKESP"{print $1}')"
AB_PART="$(lsblk -lnpo NAME,LABEL "$DEV" 2>/dev/null | awk '$2=="FLOWORKAB"{print $1}')"
[ -n "$ESP_PART" ] && [ -n "$AB_PART" ] || { echo "$DEV is not a Flowork OS stick (no FLOWORKESP/FLOWORKAB)" >&2; exit 1; }
SUDO=""; [ "$(id -u)" -ne 0 ] && SUDO="sudo"

E="$(mktemp -d)"; A="$(mktemp -d)"
cleanup(){ $SUDO umount "$E" 2>/dev/null||true; $SUDO umount "$A" 2>/dev/null||true; rmdir "$E" "$A" 2>/dev/null||true; }
trap cleanup EXIT
$SUDO umount "$ESP_PART" "$AB_PART" 2>/dev/null || true

echo "[update] ESP ($ESP_PART): kernel + initramfs + grub.cfg"
$SUDO mount "$ESP_PART" "$E"
$SUDO mkdir -p "$E/boot/grub"
$SUDO cp "$KERNEL" "$E/boot/vmlinuz"
$SUDO cp "$INITRD" "$E/boot/initramfs.gz"
$SUDO tee "$E/boot/grub/grub.cfg" >/dev/null <<'CFG'
set timeout=3
set default=0
menuentry "Flowork OS" {
    linux  /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 net.ifnames=0
    initrd /boot/initramfs.gz
}
menuentry "Flowork OS (safe graphics / nomodeset)" {
    linux  /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 nomodeset net.ifnames=0
    initrd /boot/initramfs.gz
}
CFG

echo "[update] FLOWORKAB ($AB_PART): A/B slots (squashfs + verity)"
$SUDO mount "$AB_PART" "$A"
for s in a b; do
	$SUDO mkdir -p "$A/flowork/slot-$s"
	$SUDO cp "$SQ" "$A/flowork/slot-$s/rootfs.squashfs"
	for e in verity roothash roothash.sig; do
		[ -f "$SQ.$e" ] && $SUDO cp "$SQ.$e" "$A/flowork/slot-$s/rootfs.squashfs.$e"
	done
	$SUDO rm -f "$A/flowork/slot-$s.confirmed" "$A/flowork/slot-$s.tries"
done
printf 'a' | $SUDO tee "$A/flowork/active" >/dev/null

$SUDO sync
echo "[update] DONE — $DEV refreshed from the current build (DATA partition untouched)."
