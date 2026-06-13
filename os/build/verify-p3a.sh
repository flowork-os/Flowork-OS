#!/usr/bin/env bash
# verify-p3a.sh — boot with the app-sandbox self-test enabled and assert that the
# Python and Go sample apps run confined and cannot read the owner's state.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$(cd "$SELF_DIR/.." && pwd)/out"

echo "=== P3a app-sandbox verification ==="
rm -f "$OUT/flowork-data.img"   # ephemeral boot is fine for the sandbox self-test

SKIP_SCREENSHOT=1 BOOT_TIMEOUT="${BOOT_TIMEOUT:-110}" EXTRA_APPEND="flowork.selftest=1" \
	bash "$SELF_DIR/run-qemu.sh" >/dev/null 2>&1 || true

echo "--- sandbox self-test output (serial) ---"
grep -E 'SANDBOX' "$OUT/serial.log" | sed 's/^/  /' || true

if grep -q 'SANDBOX RESULT: OK' "$OUT/serial.log"; then
	echo "P3a SANDBOX: ✅ PASS"
	exit 0
fi
echo "P3a SANDBOX: ❌ FAIL — inspect $OUT/serial.log"
exit 1
