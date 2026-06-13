#!/usr/bin/env bash
# verify-ab.sh — prove A/B auto-update (design Section 8): boots the active slot, a pointer
# switch boots the other slot, a healthy boot is confirmed, and a bad/unconfirmed slot
# auto-rolls-back. Boots -kernel/-initrd + the writable A/B disk in QEMU; edits the disk
# between boots via a loop mount (needs sudo for writes).
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
KERNEL="$OUT/$TAG.vmlinuz"; INITRD="$OUT/$TAG.initramfs-small.gz"; ABIMG="$OUT/$TAG.ab-disk.img"
[ -f "$KERNEL" ] && [ -f "$INITRD" ] || { echo "build first (kernel/initramfs missing)" >&2; exit 1; }
ACCEL="tcg"; CPU="max"; [ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }
MNT="$OUT/.abmnt"; mkdir -p "$MNT"
pass=0; fail=0
ok(){ echo "  ✅ $1"; pass=$((pass+1)); }
no(){ echo "  ❌ $1"; fail=$((fail+1)); }

freshdisk(){ bash "$SELF_DIR/make-ab-disk.sh" "$OUT" "$TAG" 2G >/dev/null; }
mountab(){ sudo mount -o loop "$ABIMG" "$MNT"; }
umountab(){ sudo umount "$MNT" 2>/dev/null || true; }
setactive(){ printf '%s' "$1" | sudo tee "$MNT/flowork/active" >/dev/null; }

qemu_boot(){ # $1 log  $2 timeout
	rm -f "$1"
	timeout "${2:-140}" qemu-system-x86_64 -machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 2 -m 4096 \
		-kernel "$KERNEL" -initrd "$INITRD" \
		-append "rdinit=/init console=tty0 console=ttyS0,115200 loglevel=4 net.ifnames=0" \
		-drive file="$ABIMG",if=virtio,format=raw,cache=writeback \
		-display none -netdev user,id=n0 -device virtio-net-pci,netdev=n0 \
		-serial "file:$1" -no-reboot >/dev/null 2>&1 &
	local qp=$! i
	for i in $(seq 1 70); do
		grep -qE 'FLOWORK_AB slot=. confirmed|rolling back|refusing to boot' "$1" 2>/dev/null && break
		sleep 2
	done
	sleep 2; kill "$qp" 2>/dev/null || true; wait "$qp" 2>/dev/null || true
}

echo "=== A/B auto-update verification ==="
pkill -x qemu-system-x86_64 2>/dev/null || true; sleep 1
freshdisk

echo "--- AB.1/AB.3: boot active slot a, confirm on health ---"
qemu_boot "$OUT/ab1.log" 140
grep -q 'active slot = a' "$OUT/ab1.log" && grep -q AGENT_UP "$OUT/ab1.log" && ok "AB.1 booted slot a + agent up" || no "AB.1 boot slot a"
mountab; [ -f "$MNT/flowork/slot-a.confirmed" ] && ok "AB.3 healthy boot confirmed slot a" || no "AB.3 confirm"; umountab

echo "--- AB.2: switch pointer to b, boots slot b ---"
mountab; setactive b; sudo rm -f "$MNT/flowork/slot-b.confirmed" "$MNT/flowork/slot-b.tries"; umountab
qemu_boot "$OUT/ab2.log" 140
grep -q 'active slot = b' "$OUT/ab2.log" && grep -q AGENT_UP "$OUT/ab2.log" && ok "AB.2 switch booted slot b + agent up" || no "AB.2 boot slot b"

echo "--- AB.4: corrupt slot b (verity) -> auto-rollback to a ---"
mountab; setactive b; sudo rm -f "$MNT/flowork/slot-b.confirmed"; printf '\xFF' | sudo dd of="$MNT/flowork/slot-b/rootfs.squashfs" bs=1 seek=50 conv=notrunc status=none; umountab
qemu_boot "$OUT/ab4.log" 100
grep -qiE 'rolling back b -> a' "$OUT/ab4.log" && ok "AB.4 corrupt slot detected -> rollback" || no "AB.4 rollback msg"
mountab; [ "$(cat "$MNT/flowork/active")" = a ] && ok "AB.4 active pointer reverted to a" || no "AB.4 pointer revert"; umountab
qemu_boot "$OUT/ab4b.log" 140
grep -q 'active slot = a' "$OUT/ab4b.log" && grep -q AGENT_UP "$OUT/ab4b.log" && ok "AB.4 recovered: slot a boots after rollback" || no "AB.4 recovery boot"

echo "--- AB.5: unconfirmed slot b (tries>3) -> auto-rollback to a ---"
freshdisk
mountab; setactive b; sudo rm -f "$MNT/flowork/slot-b.confirmed"; echo 3 | sudo tee "$MNT/flowork/slot-b.tries" >/dev/null; umountab
qemu_boot "$OUT/ab5.log" 90
grep -qiE 'rolling back b -> a' "$OUT/ab5.log" && ok "AB.5 unconfirmed slot rolled back after tries>3" || no "AB.5 tries rollback"
mountab; [ "$(cat "$MNT/flowork/active")" = a ] && ok "AB.5 active pointer reverted to a" || no "AB.5 pointer revert"; umountab

echo
echo "RESULT: pass=$pass fail=$fail"
[ "$fail" = 0 ] && { echo "A/B AUTO-UPDATE: ✅ PASS"; exit 0; } || { echo "A/B AUTO-UPDATE: ❌ FAIL"; exit 1; }
