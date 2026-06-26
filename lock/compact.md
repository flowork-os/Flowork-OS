# COMPACT — digest interaksi → GRAPH via agent dedikasi (dream-digester)

> Auto-compact + manual (Compact All / per-agent). Owner: Mr.Dev. 2026-06-26.

## APA
Agent yg interaksi non-deleted > `max_interactions` (default 400) → DIGEST interaksi pending ke
COGNITIVE GRAPH (entitas/relasi) → lalu TRIM interaksi yg udah di-digest (sisain `keep_recent`=60).
Pengalaman GAK ilang (pindah ke graph). Trim cuma nyentuh yg ada di `cognitive_digest_log` (mustahil
buang yg belum ke-graph) = aman, no data-loss.

## JALUR (3 lapis)
- **Auto**: ticker 15 menit (`main.go`) → `AutoCompactAllAgents` (server-side, no client-timeout).
- **Manual per-agent**: `POST /api/agents/compact?id=&force=1`.
- **Manual all**: `POST /api/agents/compact-all`.
- Config GUI: `GET/POST /api/compact/config` {enabled, max_interactions, keep_recent, model}.

## MODEL = AGENT DEDIKASI (owner 2026-06-26) — graph-connected
Digest reasoning compact sekarang DEFAULT lewat **agent `dream-digester`** (CGM EXTRACTOR, model GUI
per-agent, mis. `claude-haiku-4-5`) — konsisten sama enrich (codemap-enricher). Bukan lagi lokal-26B
(lambat). Urutan resolve (`digest_model.go` `buildDigestDepsModel`):
1. `compact_model` KOSONG + `DigestLLMOverride` ke-set → **dream-digester** (model GUI) → hasil ke graph.
2. `compact_model` di-SET owner → direct ke model itu (override).
3. override nil → fallback LOKAL (flowork-brain) → tetep jalan offline (no-subscription).
→ "sesuaikan model compact" = set `router_model` agent dream-digester di GUI, ATAU `compact_model` config.

## VERIFIKASI 2026-06-26
- auto-compact config: enabled, ticker 15mnt, max 400, keep 60. per-agent endpoint app-judge → {ok}.
- compact-all reachable. compact codemap-enricher (1010) via dream-digester(haiku) → `cognitive_digest_log`
  naik (digest → GRAPH jalan). TestKernelFreeze PASS.
- CATATAN: agent volume-SANGAT-tinggi (1000+) manual-compact bisa lewat timeout request (per-batch
  agent invocation) → auto-compact (server-side) yg nyelesain bertahap. Haiku >> lokal-26B (lebih cepat).

## AUTO-SLEEP aman
Compact pakai haiku (cloud) / dream-digester → GAK wake LLM lokal. `llm_idle_sleep.go` UTUH (frozen,
gak disentuh). Routing compact→haiku malah NGURANGIN wake-up LLM lokal.

## FILE
- `internal/agentmgr/digest_model.go` (FROZEN, re-hash 2026-06-26) — buildDigestDepsModel routing.
- `internal/agentmgr/autocompact.go` (FROZEN) — AutoCompactAgent/handlers/config.
- `internal/agentdb/compact.go` (FROZEN) — trim aman by cognitive_digest_log.
- `dream_digester_seed.go` (FROZEN) — agent dream-digester (CGM extractor, model GUI).
