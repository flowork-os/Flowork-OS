# 🔋 SADAR-KUOTA — Rate-Limit Awareness Langganan (Router)

> Owner: Aola Sahidin (Mr.Dev) · github.com/flowork-os/Flowork-OS · floworkos.com (white-label)
> Dibikin 2026-06-27. Acuan: cara Claude Code asli nangani limit (source leaked, dipelajari).

## Masalah (kenapa Flowork dulu nabrak limit, Claude Code enggak)
Limit langganan claude.ai = **window 5-jam bergulir** (+ 7-hari), ada tier (Pro < Max < Max-20x).
Anthropic balikin **sisa kuota di HEADER response TIAP call** — Claude Code BACA ini → tau kapan
deket limit → ga nabrak + auto-turun model. Flowork dulu **BUANG** info ini → baru sadar pas kena
**429** → lalu **retry-storm** (6× backoff = ~90 detik) yang malah makin ngabisin window.

## Solusi (3 fase, 1 STATE SHARE — nol duplikat antar-agent)
Semua call lewat `dispatcher.forwardToProvider` (1 choke-point) → kuota dihitung SEKALI di situ,
semua agent baca state yang sama. BUKAN tiap agent ngitung sendiri.

### Header yang dibaca (response Anthropic)
`anthropic-ratelimit-unified-5h-utilization` (0..1) · `-5h-reset` · `-5h-surpassed-threshold` ·
`-7d-utilization` · `-7d-reset` · `-fallback-percentage`.

### Phase 1 — SADAR
- `dispatcher.go` (FROZEN): hook POLA-B `afterAnthropicResponse(resp.Header)` dipanggil tiap
  response (incl 429). Default no-op (self-sufficient). + FIX User-Agent cloak:
  `claude-cli/1.0.0 (flow_router)` → `claude-code/<FLOWORK_CLOAK_VERSION>` (samain Claude Code
  asli biar tier rate-limit langganan bener — UA salah bisa ke-flag tier lain).
- `internal/router/ratelimit_track_ext.go` (NON-frozen sibling): override hook → parse header →
  `RateLimitState` (5h/7d util, surpassed, fallback%) di 1 state share + helper:
  - `RateLimitSnapshot()` — baca state (thread-safe).
  - `SubscriptionNearLimit()` — true kalau 5h ≥ 0.95 ATAU surpassed.
  - `RateLimitHandler` — GET /api/router/ratelimit (JSON).
- `ratelimit_route_ext.go` (package main): wire route via `RegisterExtraRoute`.

### Phase 2 — AUTO-REM (di dispatcher, FROZEN)
Di loop retry: kalau `SubscriptionNearLimit()` true DAN dapet 429 → **JANGAN retry-storm**,
langsung `break` → lompat fallback (rantai: per-agent → default → haiku/lokal). Berhenti
nge-hammer tembok yang ga gerak.

### Phase 3 — TAMPIL DI GUI (statusline ala Claude Code)
- `agent/feature_router_ratelimit_ext.go` (NON-frozen): proxy same-origin GET /api/router/ratelimit
  → router :2402 (hindari CORS; GUI di :1987).
- `agent/web/js/ratelimit_badge.js` (NON-frozen, deletable): badge header **"🔋 5h X%"** polling
  30 dtk, warna ijo<80 / kuning<95 / merah≥95 (atau surpassed). app.js: +1 import+init.

## Switch terkait (fwswitch GUI)
- `FLOWORK_RL_MAX_RETRY` (default 6) — retry 429 sebelum fallback. Set 1-2 = degradasi cepet.
- `FLOWORK_CLOAK_VERSION` (default 2.1.92) — versi di UA `claude-code/<versi>`.

## Verifikasi
`go build ./...` =0 · test parser (parse/no-header/near-limit) PASS (nol LLM) · TestKernelFreeze
PASS · endpoint balik JSON (`seen:false` sampe call Anthropic pertama ngisi).

## Self-sufficiency (Rule #1)
Hook `afterAnthropicResponse` default no-op di frozen → hapus `ratelimit_track_ext.go` +
`ratelimit_route_ext.go` + GUI badge → fitur mati mulus, router/agent utuh. Yang frozen =
mekanik (dispatcher hook + UA + throttle); yang non-frozen = tracker/route/GUI (growth, deletable).
