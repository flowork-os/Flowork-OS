# PARALLEL TOOL CALLS — AKAR fix (was: paksa sequential)

> Owner: Aola Sahidin (Mr.Dev). Gap "Eksekusi Tool paralel" (ROADMAP_FLOWORK_VS_CLAUDE).
> Tujuan: 1 putaran LLM bisa eksekusi BANYAK tool (kayak Claude Code) → jauh lebih cepat +
> ga gampang timeout vs 1 tool/putaran (N tool = N inferensi LLM).

## AKAR yang dicabut (Rule 5 — bukan tambal)
Dulu mr-flow di-PAKSA sequential (`parallel_tool_calls=false` + cuma proses `ToolCalls[0]`).
Itu TAMBAL: gejalanya = router subscription path (Claude via `/messages`) nge-emit 1 pesan
`user` per tool_result → beberapa tool_result beruntun = 2+ pesan same-role beruntun →
Anthropic **HTTP 400 "roles must alternate"**. AKAR sebenarnya: translator OpenAI→Anthropic
ga ngegabung tool_result. Dibetulin di translator, bukan diakalin di agent.

## FILE + PERAN
| File | Peran | Status |
|---|---|---|
| `router/internal/router/tools.go` | `mergeConsecutiveAnthropic()` + `anthropicBlocks()`: gabung pesan same-role beruntun jadi 1 pesan berisi array content-block (banyak tool_result / teks-setelah-tool_result). Dipanggil di `buildAnthropicToolBody`. | **FROZEN** (re-freeze 2026-07-02) |
| `agent/agents/mr-flow/main.go` | Loop tool: eksekusi SEMUA `m.ToolCalls` (dulu `[0]`), rakit 1 assistant msg dgn semua tool_use + append tiap tool_result. `parallel_tool_calls` = switch (default ON). | **FROZEN** (re-freeze 2026-07-02) |
| `agent/internal/fwswitch/registry.go` | Switch GUI `FLOWORK_PARALLEL_TOOLS` (bool, default true). | NON-frozen (seam) |
| `router/internal/router/tools_parallel_test.go` | Bukti: N tool_result → 1 user msg alternate; teks-setelah-tool_result kegabung; 1-tool tetap sama. | test (non-frozen) |

## SWITCH (GUI = kebenaran)
`FLOWORK_PARALLEL_TOOLS` — ON (default) = model boleh minta banyak tool sekaligus. OFF =
balik sequential (kalau ada provider rewel). Call-site default `true` (main.go), konsisten
dgn registry. `mergeConsecutiveAnthropic` SELALU jalan (aman walau sequential — no-op utk 1 tool).

## KENAPA ga perlu switch di router
Merge itu KOREKSI protokol Anthropic (bikin request valid), bukan fitur opsional — selalu benar
buat nyalain. 1 tool_result → tetap 1 pesan (bentuk identik, zero perubahan perilaku lama).

## QC (2026-07-02)
build agent+router OK · vet OK · `go test ./internal/router/` PASS (+3 test merge baru) ·
TestKernelFreeze PASS (hash tools.go+main.go di-update) · gembok aktif (Operation not permitted) ·
tes nyata mr-flow (bahasa manusia) exit 0, ZERO 400/"multiple tool_result"/"roles must alternate".
Delete-test: diff ga nambah dependensi frozen→non-frozen (helper stdlib + env-read + data entry) →
self-sufficiency terjaga by-construction.
