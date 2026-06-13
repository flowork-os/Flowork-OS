#!/usr/bin/env bash
# make-iso-squash.sh — build the SQUASHFS-ROOT bootable ISO (product/USB shape):
# GRUB loads the kernel + the small stage-1 initramfs; the initramfs mounts
# flowork/rootfs.squashfs (on this same media) as an overlay root.
#   args: OUT_DIR TAG
set -euo pipefail
OUT="$1"; TAG="$2"
command -v grub-mkrescue >/dev/null || { echo "grub-mkrescue missing; skip squashroot ISO"; exit 0; }
command -v xorriso       >/dev/null || { echo "xorriso missing; skip squashroot ISO";       exit 0; }

for f in "$TAG.vmlinuz" "$TAG.initramfs-small.gz" "$TAG.rootfs.squashfs"; do
	[ -f "$OUT/$f" ] || { echo "missing $f; cannot build squashroot ISO" >&2; exit 1; }
done

ISODIR="$(dirname "$OUT")/.work/iso-squash"
rm -rf "$ISODIR"; mkdir -p "$ISODIR/boot/grub" "$ISODIR/flowork"
cp "$OUT/$TAG.vmlinuz"              "$ISODIR/boot/vmlinuz"
cp "$OUT/$TAG.initramfs-small.gz"  "$ISODIR/boot/initramfs.gz"
cp "$OUT/$TAG.rootfs.squashfs"     "$ISODIR/flowork/rootfs.squashfs"
# dm-verity integrity sidecars (if built): /init verifies the squashfs against these.
for ext in verity roothash roothash.sig; do
	[ -f "$OUT/$TAG.rootfs.squashfs.$ext" ] && \
		cp "$OUT/$TAG.rootfs.squashfs.$ext" "$ISODIR/flowork/rootfs.squashfs.$ext"
done

cat > "$ISODIR/boot/grub/grub.cfg" <<CFG
set timeout=1
set default=0
menuentry "Flowork OS (squashfs root)" {
    linux  /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 net.ifnames=0 ${FLOWORK_EXTRA_CMDLINE:-}
    initrd /boot/initramfs.gz
}
CFG

grub-mkrescue -o "$OUT/$TAG.squashroot.iso" "$ISODIR" >/dev/null 2>&1
echo "squashroot ISO -> $TAG.squashroot.iso ($(du -h "$OUT/$TAG.squashroot.iso" | cut -f1))"
