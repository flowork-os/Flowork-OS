#!/usr/bin/env bash
# make-iso.sh — wrap the kernel + ephemeral initramfs into a GRUB bootable ISO.
# For P0 the ISO boots the SAME full-rootfs initramfs as the direct-kernel path
# (robust: no boot-time squashfs mount / module juggling). The squashfs artifact is
# produced separately as the P1 persistence base.
#   args: OUT_DIR TAG ROOTFS_DIR KVER
set -euo pipefail
OUT="$1"; TAG="$2"; ROOTFS="$3"; KVER="${4:-}"

command -v grub-mkrescue >/dev/null || { echo "grub-mkrescue missing; skip ISO"; exit 0; }
command -v xorriso       >/dev/null || { echo "xorriso missing; skip ISO";       exit 0; }

ISODIR="$(dirname "$OUT")/.work/iso"
rm -rf "$ISODIR"; mkdir -p "$ISODIR/boot/grub"
cp "$OUT/$TAG.vmlinuz"       "$ISODIR/boot/vmlinuz"
cp "$OUT/$TAG.initramfs.gz"  "$ISODIR/boot/initramfs.gz"

cat > "$ISODIR/boot/grub/grub.cfg" <<'CFG'
set timeout=1
set default=0
menuentry "Flowork OS (ephemeral)" {
    linux  /boot/vmlinuz rdinit=/init console=tty0 console=ttyS0,115200 quiet loglevel=4
    initrd /boot/initramfs.gz
}
CFG

grub-mkrescue -o "$OUT/$TAG.iso" "$ISODIR" >/dev/null 2>&1
echo "ISO -> $TAG.iso ($(du -h "$OUT/$TAG.iso" | cut -f1))"
