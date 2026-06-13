#!/usr/bin/env bash
# verify-p3b.sh — prove the dm-verity integrity primitive (design Section 7): a hash tree
# verifies an intact rootfs image and REJECTS a tampered one. Needs root (dm-verity).
# Runs on the host; the same primitive guards the read-only root in the appliance once the
# squashfs-on-disk boot path lands (deferred — see BACKLOG).
set -euo pipefail
command -v veritysetup >/dev/null || { echo "veritysetup (cryptsetup) required" >&2; exit 1; }
SUDO=""; [ "$(id -u)" -ne 0 ] && SUDO="sudo"
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT

echo "=== P3b dm-verity integrity verification ==="
dd if=/dev/urandom of="$T/root.img" bs=1M count=8 status=none
RH="$($SUDO veritysetup format "$T/root.img" "$T/root.hash" | sed -n 's/^Root hash:[[:space:]]*//p')"
echo "  root hash: $RH"

pass=0; fail=0
rcof(){ local rc=0; "$@" >/dev/null 2>&1 || rc=$?; echo "$rc"; }
chk_ok(){  if [ "$1" = "0" ]; then echo "  ✅ $2"; pass=$((pass+1)); else echo "  ❌ $2 (rc=$1, want 0)"; fail=$((fail+1)); fi; }
chk_rej(){ if [ "$1" != "0" ]; then echo "  ✅ $2"; pass=$((pass+1)); else echo "  ❌ $2 (accepted; want non-zero)"; fail=$((fail+1)); fi; }

# 1) intact image verifies  (veritysetup verify -> 0)
chk_ok "$(rcof $SUDO veritysetup verify "$T/root.img" "$T/root.hash" "$RH")" "intact rootfs VERIFIED"

# 2) tampered image rejected (veritysetup -> non-zero)
cp "$T/root.img" "$T/tampered"; printf 'X' | dd of="$T/tampered" bs=1 seek=500000 conv=notrunc status=none
chk_rej "$(rcof $SUDO veritysetup verify "$T/tampered" "$T/root.hash" "$RH")" "tampered rootfs REJECTED"

# 3) wrong root hash rejected
WRONG="$(printf '%064d' 0)"
chk_rej "$(rcof $SUDO veritysetup verify "$T/root.img" "$T/root.hash" "$WRONG")" "wrong root hash REJECTED"

echo
if [ "$fail" = 0 ]; then echo "P3b DM-VERITY: ✅ PASS ($pass/3)"; exit 0; fi
echo "P3b DM-VERITY: ❌ FAIL ($pass ok, $fail bad)"; exit 1
