#!/usr/bin/env bash
# verify-p4a.sh — prove the update signature-verification core (anti-MITM, design Section 8):
# a valid signature is ACCEPTED, a tampered artifact is REJECTED, and a wrong key is REJECTED.
# Runs entirely on the host (the crypto is identical in the appliance) — fast, no QEMU.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
UPDATE="$REPO_DIR/rootfs-overlay/usr/local/bin/flowork-update"
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT

echo "=== P4a update signature verification ==="
"$SELF_DIR/make-update-keys.sh" >/dev/null 2>&1 || true
PUB="$REPO_DIR/config/update-pub.pem"
KEY="$REPO_DIR/config/update.key"

head -c 2000000 /dev/urandom > "$T/rootfs.squashfs"
"$SELF_DIR/sign-release.sh" "$T/rootfs.squashfs" "$KEY" >/dev/null

pass=0; fail=0
# run a command that may fail without tripping `set -e`; return its exit code
rcof(){ local rc=0; "$@" >/dev/null 2>&1 || rc=$?; echo "$rc"; }
chk(){ if [ "$1" = "$2" ]; then echo "  ✅ $3"; pass=$((pass+1)); else echo "  ❌ $3 (got rc=$1 want $2)"; fail=$((fail+1)); fi; }

# 1) valid signature accepted
chk "$(rcof sh "$UPDATE" verify "$T/rootfs.squashfs" "$T/rootfs.squashfs.sig" "$PUB")" 0 "valid signature ACCEPTED"

# 2) tampered artifact rejected
cp "$T/rootfs.squashfs" "$T/tampered"; printf 'X' | dd of="$T/tampered" bs=1 seek=10 conv=notrunc 2>/dev/null
chk "$(rcof sh "$UPDATE" verify "$T/tampered" "$T/rootfs.squashfs.sig" "$PUB")" 1 "tampered artifact REJECTED"

# 3) wrong key rejected
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$T/other.key" 2>/dev/null
openssl pkey -in "$T/other.key" -pubout -out "$T/other.pub" 2>/dev/null
chk "$(rcof sh "$UPDATE" verify "$T/rootfs.squashfs" "$T/rootfs.squashfs.sig" "$T/other.pub")" 1 "wrong public key REJECTED"

echo
if [ "$fail" = 0 ]; then echo "P4a UPDATE-VERIFY: ✅ PASS ($pass/3)"; exit 0; fi
echo "P4a UPDATE-VERIFY: ❌ FAIL ($pass ok, $fail bad)"; exit 1
