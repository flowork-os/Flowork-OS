#!/usr/bin/env bash
# verify-ollama.sh — prove P2 (sovereign local AI): boot the appliance with a model disk and
# NO network, and confirm Ollama runs a local Qwen inference. Offline = the QEMU VM has no NIC,
# so a returned completion can only come from the on-device model.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO_DIR/out"
TAG="flowork-os-$(cat "$REPO_DIR/VERSION" 2>/dev/null || echo 0.0.0)"
ISO="$OUT/$TAG.squashroot.iso"; MODELIMG="$OUT/$TAG.model-disk.img"
MODEL="${FLOWORK_MODEL:-qwen2.5:0.5b}"
SERIAL="$OUT/serial-ollama.log"
[ -f "$ISO" ] || { echo "no squashroot ISO — build first" >&2; exit 1; }
ACCEL="tcg"; CPU="max"; [ -e /dev/kvm ] && [ -w /dev/kvm ] && { ACCEL="kvm"; CPU="host"; }

echo "=== P2 local-AI (Ollama) verification — OFFLINE ==="
[ -f "$MODELIMG" ] || bash "$SELF_DIR/make-model-disk.sh" "$OUT" "$TAG" "$MODEL" >/dev/null
# rebuild the ISO with the local-AI self-test enabled
FLOWORK_EXTRA_CMDLINE="flowork.aitest=1" bash "$SELF_DIR/make-iso-squash.sh" "$OUT" "$TAG" >/dev/null

pkill -x qemu-system-x86_64 2>/dev/null || true; sleep 1
rm -f "$SERIAL"
echo "[ollama] booting (NO network; model disk attached)"
# NOTE: deliberately NO -netdev — the VM has no path to the internet.
timeout 260 qemu-system-x86_64 -machine q35 -accel "$ACCEL" -cpu "$CPU" -smp 4 -m 4096 \
	-cdrom "$ISO" -boot d \
	-drive file="$MODELIMG",if=virtio,format=raw,cache=writeback \
	-display none -serial "file:$SERIAL" -no-reboot >/dev/null 2>&1 &
qp=$!
for i in $(seq 1 110); do
	grep -qE 'OLLAMA_AI self-test done' "$SERIAL" 2>/dev/null && break
	kill -0 "$qp" 2>/dev/null || break
	sleep 3
done
sleep 1; kill "$qp" 2>/dev/null || true; wait "$qp" 2>/dev/null || true

# restore the clean ISO (no self-test flag)
bash "$SELF_DIR/make-iso-squash.sh" "$OUT" "$TAG" >/dev/null

echo "--- local-AI serial ---"; grep -E 'OLLAMA|ROUTER_AI' "$SERIAL" | sed 's/^/  /' | head -14
ollama_ok=0; router_ok=0
grep -q 'OLLAMA_AI RESULT: OK' "$SERIAL" && ollama_ok=1
grep -q 'ROUTER_AI RESULT: OK' "$SERIAL" && router_ok=1
echo "RESULT: ollama_offline=$ollama_ok  router_to_ollama(P2b)=$router_ok"
if [ "$ollama_ok" = 1 ] && [ "$router_ok" = 1 ]; then
	echo "P2 + P2b LOCAL-AI: ✅ PASS (offline Qwen, and the router routes to it)"; exit 0
fi
[ "$ollama_ok" = 1 ] && { echo "P2 LOCAL-AI: ✅ (ollama) — P2b router routing not green; see $SERIAL"; exit 1; }
echo "P2 LOCAL-AI: ❌ FAIL — see $SERIAL"; exit 1
