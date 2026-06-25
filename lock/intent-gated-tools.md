# INTENT-GATED TOOLS (#9) — prune tool-schema ke yg relevan (hemat token)

> Owner: Aola Sahidin (Mr.Dev) · github.com/flowork-os/Flowork-OS · 2026-06-26.

## AKAR (Rule 5) — token
Diagnosis: ~15.8k token/turn, **biang #1 = tool-schema statis ~55%** prompt. Tiap request kirim
SEMUA skema tool (24+) walau cuma butuh 2-3 → boros (token = syarat hidup koloni).

## SOLUSI — `maybeFilterTools` (router, semantic gating)
Router PRUNE `req.Tools` ke yg RELEVAN sebelum kirim ke model lokal:
1. Embed query user terakhir (bge-m3 lokal) + embed deskripsi tiap tool.
2. Skor cosine query↔tool → keep **top-K** (default 12) yg ≥ **threshold** (default 0.30).
3. **ALWAYS-KEEP (escape-hatch, DI LUAR top-K):** `tool_search`, `tool_lookup` (#2C — kalau tool
   kepruned model masih bisa NEMU/AMBIL balik → pruning AMAN, anti salah-gate), `StructuredOutput`,
   `ScheduleWakeup`, + tool yg UDAH dipanggil di history turn ini.
4. **Fail-open total**: switch off / no tools / no query / embedder mati / error embed → req UTUH.

File: `router/internal/router/dynamic_tools.go` (NON-frozen). Di-wire di `dispatcher.go` +
`dispatcher_stream.go` setelah `maybeInjectInstinct` (sejajar hook injeksi lain).

## SWITCH (GUI Switch Fitur / ENV — [[flowork-settings-gui]] fwswitch)
| Key | Default | Guna |
|---|---|---|
| `FLOWORK_DYNAMIC_TOOLS` | **off** | master switch (legacy alias `FLOW_ROUTER_DYNAMIC_TOOLS=1` masih jalan) |
| `FLOWORK_DYNAMIC_TOOLS_TOPK` | 12 | max tool relevan (di luar escape-hatch). Kecil=hemat tapi rawan starve |
| `FLOWORK_DYNAMIC_TOOLS_MINSCORE` | 0.30 | cosine minimum tool dianggap relevan |

Semua ke-manage GUI (prefix `FLOWORK_` → fwswitch apply ke router lintas-proses). Default OFF →
NOL perubahan perilaku sampai owner nyalain.

## VERIFIKASI
- Live: switch ON via file GUI → request 15 tool, topK=5 → log `dynamic_tools: filtered tools
  from 15 down to 5`. Token OFF=5422 → ON=4941 (test kecil; tool-schema asli jauh lebih gede →
  hemat di produksi jauh lebih besar, skala ke ~55% biang).
- Rule-9: mr-flow bahasa-manusia dgn filter ON (topK 12) → tetep eksekusi tool, no starve/ghost.

## CATATAN
- Komplementer dgn #2C defer-tools (agent-side) — ini router-side, jalan buat SEMUA caller termasuk external.
- Butuh embedder lokal (bge-m3) ke-register di router; kalau ga ada → fail-open (req utuh).
