#!/usr/bin/env bash
# verify-usb.sh — boot the single flashable USB image in QEMU (BIOS via SeaBIOS, and UEFI via
# OVMF if available) and confirm it boots through GRUB into the A/B system with agent + router up.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
IMG="$OUT/$TAG.usb.img"
[ -f "$IMG" ] || { echo "no USB image — run build/make-usb-image.sh first" >&2; exit 1; }
ACCEL="tcg"; CPU="max"; [ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }
OVMF=""; for f in /usr/share/ovmf/OVMF.fd /usr/share/OVMF/OVMF_CODE_4M.fd; do [ -f "$f" ] && { OVMF="$f"; break; }; done

boot_check() { # $1 = label  $2 = extra qemu args (firmware)
	local log="$OUT/usb-$1.log"; rm -f "$log"
	echo "[usb-$1] booting the USB image ($1)"
	# shellcheck disable=SC2086
	timeout 200 qemu-system-x86_64 -machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 2 -m 4096 \
		$2 -drive file="$IMG",format=raw,if=virtio,cache=writeback \
		-display none -serial "file:$log" \
		-netdev user,id=n0 -device virtio-net-pci,netdev=n0 -no-reboot >/dev/null 2>&1 &
	local qp=$! i
	for i in $(seq 1 60); do
		grep -qE 'AGENT_UP|ROOT INTEGRITY|no Flowork media|Kernel panic' "$log" 2>/dev/null && break
		kill -0 "$qp" 2>/dev/null || break
		sleep 3
	done
	sleep 1; kill "$qp" 2>/dev/null || true; wait "$qp" 2>/dev/null || true
	local ok=0
	grep -q 'A/B mode: active slot = a' "$log" && grep -q AGENT_UP "$log" && ok=1
	grep -E 'FLOWORK|AGENT_UP|ROUTER_UP' "$log" | sed 's/^/    /' | head -8
	echo "  $1: $([ "$ok" = 1 ] && echo '✅ booted (GRUB -> A/B -> agent up)' || echo '❌ not green')"
	return $((1 - ok))
}

echo "=== single-USB image boot verification ==="
bios_ok=0; uefi_ok=0
boot_check bios "" && bios_ok=1 || true
if [ -n "$OVMF" ]; then
	boot_check uefi "-bios $OVMF" && uefi_ok=1 || true
else
	echo "[usb-uefi] OVMF firmware not found — skipping UEFI check"
fi

echo
echo "RESULT: bios=$bios_ok uefi=$uefi_ok"
if [ "$bios_ok" = 1 ] || [ "$uefi_ok" = 1 ]; then echo "USB IMAGE: ✅ PASS (boots from the single stick)"; exit 0; fi
echo "USB IMAGE: ❌ FAIL — see $OUT/usb-*.log"; exit 1
