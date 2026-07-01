# PROMPT CACHING (Claude) — AKAR fix gap "Prompt Caching"

> Owner: Aola Sahidin (Mr.Dev). Gap #1 ROADMAP (dampak Sangat Tinggi: TPM/latensi/biaya).
> Dulu: jalur Claude mr-flow (`buildAnthropicToolBody`) kirim system+tool-schema+history MENTAH
> tiap turn (0 cache) → prefix statik gede dibayar & di-prefill ulang tiap putaran.

## AKAR + solusi
Prompt caching Anthropic udah **GA**: cukup sisipin `cache_control:{type:ephemeral}` di body
(`anthropic-version: 2023-06-01`) — TANPA header beta, TANPA nyentuh peniruan auth langganan
(low-risk). Breakpoint dipasang di 3 bagian STATIS lintas-turn:
1. **system** (persona/konstitusi/doktrin) — bagian paling gede & sama tiap turn.
2. **tool-schema terakhir** — cache seluruh array tool (stabil).
3. **block terakhir pesan terakhir** — cache prefix history inkremental.

Efek: turn ke-2+ baca prefix dari cache (~90% lebih murah + prefill lebih cepat → latensi turun,
anti-timeout). Ini melengkapi fix parallel-tools (lock/parallel-tools.md).

## FILE
| File | Peran | Status |
|---|---|---|
| `router/internal/router/tools.go` | `promptCacheEnabled()` + `markLastBlockCache()`; `buildAnthropicToolBody` sisip cache_control (system/tools/last-msg); parse `cache_read/cache_creation` di respons (stream + non-stream) → `prompt_tokens_details` + log observability. | **FROZEN** (re-freeze 2026-07-02) |
| `agent/internal/fwswitch/registry.go` | Switch GUI `FLOWORK_PROMPT_CACHE` (bool, default true). | NON-frozen (seam) |
| `router/internal/router/tools_cache_test.go` | Bukti breakpoint kepasang (ON) & absen (OFF). | test |

## SWITCH (kill-switch cepat)
`FLOWORK_PROMPT_CACHE` — ON (default). Kalau ada provider nolak cache_control → set `off` di GUI
Setting (lintas-proses, live ≤3 dtk, ga perlu restart) → balik kirim mentah. Local model path
TIDAK kesentuh (cuma jalur Anthropic).

## QC (2026-07-02) — TERVALIDASI LIVE
build router OK · vet OK · `go test ./internal/router/` PASS (+2 test cache, +3 test merge) ·
TestKernelFreeze PASS (hash tools.go update) · gembok aktif · local-model path sehat.
**LIVE (Claude Pro/Max Subscription ON):** log router bukti caching jalan —
`anthropic cache: read=0 create=10903` (turn awal tulis cache) → `read=7987 create=3815`
(turn lanjut BACA 7987 token dari cache, cuma fresh_input=6 di-prefill). ZERO 400/"roles must
alternate". Model respons = `claude-haiku-4-5` (bukan fallback local).
