#!/usr/bin/env bash
# build-all-agents.sh — build SEMUA agent + template wasm dari source.
#
# KENAPA: agent.wasm GITIGNORED (ga ke-commit). Kalau ga di-build di TIAP pipeline →
# "dev jalan, portable/img GAGAL" (wasm ilang). Script ini = SATU mekanisme konsisten
# dipanggil dev (start.sh) + portable (make-portable.sh) + img (build-flowork-os.sh).
#
# TOOLCHAIN: Go-standar wasip1 (GOOS=wasip1 GOARCH=wasm). BUKAN tinygo — tinygo nempel
# ke Go ≤1.23, gagal di Go 1.25+. Standar wasip1 = multi-OS, selalu jalan, agak gede tapi WORKING.
# Idempotent (rebuild cuma kalau wasm missing / source .go lebih baru). Non-fatal per-agent.
#
# Pakai: ./scripts/build-all-agents.sh [AGENT_ROOT]   (default: parent dari scripts/)
set -u
AGENT_ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
command -v go >/dev/null 2>&1 || { echo "build-all-agents: 'go' ga ada di PATH — skip"; exit 0; }

built=0; failed=0
for dir in "$AGENT_ROOT"/agents/*/ "$AGENT_ROOT"/templates/*/; do
	[ -f "$dir/main.go" ] || continue          # cuma yg buildable (punya main.go)
	w="$dir/agent.wasm"
	if [ -f "$w" ] && [ -z "$(find "$dir" -name '*.go' -newer "$w" 2>/dev/null | head -1)" ]; then
		continue                                 # wasm fresh → skip (idempotent)
	fi
	if ( cd "$dir" && GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o agent.wasm . ); then
		echo "  ✓ agent wasm: $(basename "$dir") ($(stat -c%s "$w" 2>/dev/null) b)"
		built=$((built + 1))
	else
		echo "  ⚠ agent wasm GAGAL: $(basename "$dir") — agent ini ga bakal load sampe ke-build"
		failed=$((failed + 1))
	fi
done
echo "build-all-agents: $built (re)built, $failed gagal"
exit 0
