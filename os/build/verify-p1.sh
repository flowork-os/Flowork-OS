#!/usr/bin/env bash
# verify-p1.sh — prove persistence: boot twice on the SAME encrypted DATA disk and
# assert the boot-counter (which lives inside /root/.flowork) reaches 2 on the second
# boot. That can only happen if the LUKS volume was unlocked, mounted, and survived.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$(cd "$SELF_DIR/.." && pwd)/out"

echo "=== P1 persistence verification ==="
"$SELF_DIR/make-data-disk.sh" 1G          # fresh blank DATA disk

run_boot() { # $1 = label
	SKIP_SCREENSHOT=1 KIOSK_SETTLE=1 BOOT_TIMEOUT="${BOOT_TIMEOUT:-120}" \
		bash "$SELF_DIR/run-qemu.sh" >/dev/null 2>&1 || true
	cp "$OUT/serial.log" "$OUT/serial-$1.log"
	grep -E 'FLOWORK_PERSIST|initializing LUKS|persistent /root/.flowork' "$OUT/serial-$1.log" | tail -4
}

echo "--- boot #1 (fresh disk: expect LUKS init, boot_count=1) ---"
B1="$(run_boot boot1)"; echo "$B1"
echo "--- boot #2 (same disk: expect boot_count=2, flowork_db=yes) ---"
B2="$(run_boot boot2)"; echo "$B2"

b1="$(grep -oE 'boot_count=[0-9]+' "$OUT/serial-boot1.log" | tail -1 || true)"
b2="$(grep -oE 'boot_count=[0-9]+' "$OUT/serial-boot2.log" | tail -1 || true)"
db2="$(grep -oE 'flowork_db=(yes|no)' "$OUT/serial-boot2.log" | tail -1 || true)"
luks="$(grep -c 'initializing LUKS' "$OUT/serial-boot1.log" || true)"
mounted="$(grep -c 'persistent /root/.flowork mounted (encrypted)' "$OUT/serial-boot2.log" || true)"

echo
echo "RESULT: boot1[$b1] boot2[$b2 $db2] luks_init_boot1=$luks encrypted_mount_boot2=$mounted"
if [ "$b1" = "boot_count=1" ] && [ "$b2" = "boot_count=2" ] && [ "$mounted" -ge 1 ]; then
	echo "P1 PERSIST: ✅ PASS"
	exit 0
fi
echo "P1 PERSIST: ❌ FAIL — inspect $OUT/serial-boot1.log and $OUT/serial-boot2.log"
exit 1
