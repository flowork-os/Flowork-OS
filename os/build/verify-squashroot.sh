#!/usr/bin/env bash
# verify-squashroot.sh — boot the SQUASHFS-ROOT ISO in QEMU (from CD/USB, via GRUB) and
# assert: the overlay root comes up, and the agent + router still serve. This is the
# product/USB-shaped boot (real-fs root), as opposed to the in-RAM PoC initramfs.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
ISO="$OUT/$TAG.squashroot.iso"
SERIAL="$OUT/serial-squash.log"; QMP="$OUT/qmp-squash.sock"; SHOT="$OUT/screen-squash.png"
MEM="${QEMU_MEM:-4096}"; BOOT_TIMEOUT="${BOOT_TIMEOUT:-180}"
[ -f "$ISO" ] || { echo "no squashroot ISO ($ISO) — run build-flowork-os.sh first" >&2; exit 1; }
rm -f "$SERIAL" "$QMP" "$SHOT"

ACCEL="tcg"; CPU="max"
[ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }
echo "[squashroot] booting $ISO (mem=${MEM}M accel=$ACCEL) from CD"

qemu-system-x86_64 \
	-machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 2 -m "$MEM" \
	-cdrom "$ISO" -boot d \
	-device virtio-vga -display none \
	-netdev user,id=n0 -device virtio-net-pci,netdev=n0 \
	-serial "file:$SERIAL" -qmp "unix:$QMP,server,nowait" -no-reboot &
QPID=$!
trap 'kill $QPID 2>/dev/null || true' EXIT

agent=0 router=0 overlay=0 t=0
while [ "$t" -lt "$BOOT_TIMEOUT" ]; do
	kill -0 "$QPID" 2>/dev/null || { echo "[squashroot] qemu exited early"; break; }
	if [ -f "$SERIAL" ]; then
		grep -q 'overlay root ready' "$SERIAL" && overlay=1
		grep -q AGENT_UP  "$SERIAL" && agent=1
		grep -q ROUTER_UP "$SERIAL" && router=1
	fi
	[ "$agent" = 1 ] && [ "$router" = 1 ] && break
	sleep 3; t=$((t + 3))
done

echo "[squashroot] overlay_root=$overlay agent=$agent router=$router (after ${t}s)"
sleep 6
python3 "$SELF_DIR/qmp.py" "$QMP" screendump "{\"filename\":\"$SHOT\",\"format\":\"png\"}" >/dev/null 2>&1 || true
[ -f "$SHOT" ] && echo "[squashroot] screenshot -> $SHOT"
echo "----- FLOWORK serial lines -----"; grep -E 'FLOWORK|AGENT_UP|ROUTER_UP|overlay' "$SERIAL" 2>/dev/null | head -20
kill "$QPID" 2>/dev/null || true

if [ "$overlay" = 1 ] && [ "$agent" = 1 ] && [ "$router" = 1 ]; then
	echo "SQUASHROOT: ✅ PASS (overlay root + agent + router)"; exit 0
fi
echo "SQUASHROOT: ❌ not green — see $SERIAL"; exit 2
