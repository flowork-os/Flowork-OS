#!/usr/bin/env bash
# verify-verity.sh — prove dm-verity AT BOOT (design Section 7, enforced):
#  (1) the legitimate squashroot ISO boots with integrity active + services up;
#  (2) a tampered squashfs (one byte flipped, original hash tree kept) is DETECTED by the
#      kernel and the system REFUSES to boot (no AGENT_UP).
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
ISO="$OUT/$TAG.squashroot.iso"
[ -f "$ISO" ] || { echo "no squashroot ISO ($ISO) — run build-flowork-os.sh first" >&2; exit 1; }
[ -f "$OUT/$TAG.rootfs.squashfs.verity" ] || { echo "no verity sidecars — rebuild with veritysetup present" >&2; exit 1; }

ACCEL="tcg"; CPU="max"; [ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }

bootwait() { # $1 iso  $2 log  $3 timeout  ; returns when a terminal marker appears
	local iso="$1" log="$2" to="${3:-120}"; rm -f "$log"
	timeout "$to" qemu-system-x86_64 -machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 2 -m 4096 \
		-cdrom "$iso" -boot d -display none \
		-netdev user,id=n0 -device virtio-net-pci,netdev=n0 \
		-serial "file:$log" -no-reboot >/dev/null 2>&1 &
	local qp=$! i
	for i in $(seq 1 60); do
		grep -qE 'AGENT_UP|INTEGRITY CHECK FAILED|verity: .*corrupt|VERIFICATION FAILED' "$log" 2>/dev/null && break
		sleep 2
	done
	sleep 2; kill "$qp" 2>/dev/null || true; wait "$qp" 2>/dev/null || true
}

echo "=== dm-verity at-boot verification ==="
pkill -x qemu-system-x86_64 2>/dev/null || true; sleep 1

echo "--- case 1: legitimate ISO (expect integrity active + AGENT_UP) ---"
bootwait "$ISO" "$OUT/verity-legit.log" 140
grep -E 'integrity device active|AGENT_UP|ROUTER_UP' "$OUT/verity-legit.log" | sed 's/^/  /' || true
legit_ok=0
grep -q 'integrity device active' "$OUT/verity-legit.log" && grep -q AGENT_UP "$OUT/verity-legit.log" && legit_ok=1

echo "--- case 2: tampered ISO (expect detection + NO boot) ---"
TD="$REPO_DIR/.work/iso-tamper"; rm -rf "$TD"; mkdir -p "$TD/boot/grub" "$TD/flowork"
cp "$OUT/$TAG.vmlinuz" "$TD/boot/vmlinuz"
cp "$OUT/$TAG.initramfs-small.gz" "$TD/boot/initramfs.gz"
cp "$OUT/$TAG.rootfs.squashfs" "$TD/flowork/rootfs.squashfs"
printf '\xFF' | dd of="$TD/flowork/rootfs.squashfs" bs=1 seek=50 conv=notrunc status=none
cp "$OUT/$TAG.rootfs.squashfs.verity"   "$TD/flowork/rootfs.squashfs.verity"
cp "$OUT/$TAG.rootfs.squashfs.roothash" "$TD/flowork/rootfs.squashfs.roothash"
cat > "$TD/boot/grub/grub.cfg" <<'CFG'
set timeout=1
set default=0
menuentry "Flowork OS (tamper test)" { linux /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 ; initrd /boot/initramfs.gz }
CFG
grub-mkrescue -o "$OUT/$TAG.tamper.iso" "$TD" >/dev/null 2>&1
bootwait "$OUT/$TAG.tamper.iso" "$OUT/verity-tamper.log" 100
grep -iE 'verity: .*corrupt|INTEGRITY CHECK FAILED|AGENT_UP' "$OUT/verity-tamper.log" | sed 's/^/  /' || true
tamper_ok=0
if ! grep -q AGENT_UP "$OUT/verity-tamper.log" && \
   grep -qiE 'verity: .*corrupt|INTEGRITY CHECK FAILED' "$OUT/verity-tamper.log"; then tamper_ok=1; fi
rm -f "$OUT/$TAG.tamper.iso"

echo
echo "RESULT: legit_boots_verified=$legit_ok  tamper_rejected=$tamper_ok"
if [ "$legit_ok" = 1 ] && [ "$tamper_ok" = 1 ]; then echo "DM-VERITY BOOT: ✅ PASS"; exit 0; fi
echo "DM-VERITY BOOT: ❌ FAIL — see $OUT/verity-legit.log / verity-tamper.log"; exit 1
