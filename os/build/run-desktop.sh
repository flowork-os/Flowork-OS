#!/usr/bin/env bash
# run-desktop.sh — boot Flowork OS in a VISIBLE QEMU window you can click into.
# Attaches the local-AI model disk (if built) and user-mode networking so the control
# panel is fully usable. State is ephemeral unless you also attach a DATA disk.
#   env: QEMU_MEM (default 6144), DISPLAY must be set.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
ISO="$OUT/$TAG.squashroot.iso"
MODEL="$OUT/$TAG.model-disk.img"
MEM="${QEMU_MEM:-6144}"
[ -f "$ISO" ] || { echo "no squashroot ISO ($ISO) — run build/build-flowork-os.sh first" >&2; exit 1; }

ACCEL="tcg"; CPU="max"; [ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }
EXTRA=()
[ -f "$MODEL" ] && EXTRA=(-drive "file=$MODEL,if=virtio,format=raw,cache=writeback")

echo "[desktop] booting Flowork OS FULLSCREEN (mem=${MEM}M, accel=$ACCEL)"
echo "[desktop] the kiosk appears after ~30-40s."
echo "[desktop] EXIT back to your desktop:  Ctrl+Alt+F (leave fullscreen) then close the window."
echo "[desktop]   ungrab mouse: Ctrl+Alt+G   |   quit VM: Ctrl+Alt+2 -> type 'quit'"
echo "[desktop] (Flowork OS itself has no 'exit' — it is an appliance; you leave the VM, not the OS.)"
exec qemu-system-x86_64 \
	-machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 4 -m "$MEM" \
	-cdrom "$ISO" -boot d \
	"${EXTRA[@]}" \
	-vga none -device virtio-vga,xres="${FLOWORK_XRES:-1920}",yres="${FLOWORK_YRES:-1080}" \
	-display gtk,zoom-to-fit=on -full-screen \
	-netdev user,id=n0 -device virtio-net-pci,netdev=n0 \
	-usb -device usb-tablet \
	-no-reboot
