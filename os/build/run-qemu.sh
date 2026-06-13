#!/usr/bin/env bash
# run-qemu.sh — boot the Flowork OS ephemeral artifact in QEMU and verify P0.
# Headless: serial -> log (health), QMP -> framebuffer screenshot (kiosk).
# No writable disk is attached (success criterion S7: zero host-disk writes).
set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
VERSION="$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.1.0-p0)"
TAG="flowork-os-$VERSION"
MEM="${QEMU_MEM:-6144}"
BOOT_TIMEOUT="${BOOT_TIMEOUT:-180}"

KERNEL="$OUT/$TAG.vmlinuz"
INITRD="$OUT/$TAG.initramfs.gz"
SERIAL="$OUT/serial.log"
QMP="$OUT/qmp.sock"
SHOT="$OUT/screen.png"
[ -f "$KERNEL" ] && [ -f "$INITRD" ] || { echo "missing artifacts; run build-flowork-os.sh first" >&2; exit 1; }
rm -f "$SERIAL" "$QMP" "$SHOT"

ACCEL="tcg"; CPU="max"
if [ -e /dev/kvm ] && [ -r /dev/kvm ] && [ -w /dev/kvm ]; then ACCEL="kvm"; CPU="host"; fi

# Attach the persistent DATA disk if it exists (P1). Without it, the boot is fully
# ephemeral (P0 — no writable disk attached, success criterion S7 still holds).
DATA_IMG="$OUT/flowork-data.img"
DATA_ARGS=()
DISKMODE="ephemeral (no writable disk)"
if [ -f "$DATA_IMG" ]; then
	DATA_ARGS=(-drive "file=$DATA_IMG,if=virtio,format=raw,cache=writeback")
	DISKMODE="persistent DATA: $DATA_IMG"
fi
echo "[run] booting $TAG  (mem=${MEM}M accel=$ACCEL cpu=$CPU, $DISKMODE)"

# console order matters: the LAST console= becomes /dev/console, which is where the
# health probe writes AGENT_UP/ROUTER_UP. Keep ttyS0 last so it lands in serial.log.
qemu-system-x86_64 \
	-machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 2 -m "$MEM" \
	-kernel "$KERNEL" -initrd "$INITRD" \
	-append "rdinit=/init console=tty0 console=ttyS0,115200 loglevel=4 net.ifnames=0 ${EXTRA_APPEND:-}" \
	-device virtio-vga -display none \
	-netdev user,id=n0 -device virtio-net-pci,netdev=n0 \
	"${DATA_ARGS[@]}" \
	-serial "file:$SERIAL" \
	-qmp "unix:$QMP,server,nowait" \
	-no-reboot &
QPID=$!
trap 'kill $QPID 2>/dev/null || true' EXIT

# --- wait for boot health ----------------------------------------------------
agent=0 router=0 t=0
while [ "$t" -lt "$BOOT_TIMEOUT" ]; do
	kill -0 "$QPID" 2>/dev/null || { echo "[run] qemu exited early"; break; }
	if [ -f "$SERIAL" ]; then
		grep -q AGENT_UP  "$SERIAL" && agent=1
		grep -q ROUTER_UP "$SERIAL" && router=1
	fi
	[ "$agent" = 1 ] && [ "$router" = 1 ] && break
	sleep 2; t=$((t + 2))
done

echo "[run] health: agent=$agent router=$router (after ${t}s)"

# --- capture the kiosk framebuffer (chromium software-render is slow to paint) -
# Skippable: persistence runs (verify-p1) don't need the kiosk pixels, only the serial.
if [ -z "${SKIP_SCREENSHOT:-}" ]; then
	KIOSK_SETTLE="${KIOSK_SETTLE:-30}"
	echo "[run] letting the kiosk paint (${KIOSK_SETTLE}s + a late shot)..."
	sleep "$KIOSK_SETTLE"
	python3 "$SELF_DIR/qmp.py" "$QMP" screendump "{\"filename\":\"$SHOT\",\"format\":\"png\"}" || true
	[ -f "$SHOT" ] && echo "[run] screenshot -> $SHOT ($(du -h "$SHOT" | cut -f1))"
	sleep 18
	python3 "$SELF_DIR/qmp.py" "$QMP" screendump "{\"filename\":\"$OUT/screen-late.png\",\"format\":\"png\"}" || true
	[ -f "$OUT/screen-late.png" ] && echo "[run] late screenshot -> $OUT/screen-late.png ($(du -h "$OUT/screen-late.png" | cut -f1))"
fi

echo "----- tail of serial.log -----"
tail -n 40 "$SERIAL" 2>/dev/null || true
echo "------------------------------"

kill "$QPID" 2>/dev/null || true
if [ "$agent" = 1 ] && [ "$router" = 1 ]; then
	echo "[run] P0 core PASS (agent + router up). Inspect $SHOT for kiosk (S6)."
	exit 0
fi
echo "[run] P0 core NOT yet green — see serial.log"
exit 2
