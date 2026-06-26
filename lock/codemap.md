# CODEMAP — peta kode self-aware (GUI Codemap)

> Dev: Aola Sahidin. 2026-06-26. Agent sadar struktur kode-dirinya: file + import + layer + makna
> (enrich) → di-viz + masuk Cognitive Graph. Cara kerja mesin: `os/`.

## JALUR
```
codemap.js (GUI)
  → /api/codemap/{status,graph,semantic,docs,zombies,reindex,enrich}
  → walker.go (jalan pohon repo, skip noise) → goparser.go (parse Go: func/import/struct)
  → agentdb/codemap.go + codemap_files.go (simpan node+edge file/import)
  → enrich: agentmgr/codemap_semantic.go + agentdb/codemap_semantic.go (LLM summary/domain/role,
    by-hash incremental) → cognitive_codemap_semantic.go (tempel makna ke CGM)
  → codemap_tools.go + codemap_files_tool.go (tool buat agent baca peta kode)
```

## FROZEN — engine/jalur (10 file, chattr +i + KERNEL_FREEZE + TestKernelFreeze)
`codemap/{walker,goparser}.go`, `agentdb/{codemap,codemap_files,codemap_semantic,cognitive_codemap_
semantic}.go`, `agentmgr/{codemap,codemap_semantic}.go`, `tools/builtins/{codemap_tools,codemap_files_
tool}.go`. Strip + header white-label.

## SEAM — extend TANPA buka frozen
- **Bahasa baru** (parser Python/JS/dst) = file sibling BARU `codemap/<lang>parser.go` (goparser.go =
  Go-only; walker dispatch by-ext). Zero edit frozen.
- **Enrich** = agent `codemap-enricher` (NON-frozen, editable — lihat lock/agent-engine.md), prompt/
  model GUI-adjustable.

## SWITCH
- `FLOWORK_CODEMAP_AUTOENRICH` (default ON) + `_MIN` (interval menit) — auto-enrich file berubah
  (by-hash, murah pas stabil). `FLOWORK_CGM_CODEMAP` (projeksi ke CGM). Di registry.go.
- Parse sendiri = deterministik (bukan LLM) → gak perlu switch.

## NON-FROZEN (seam)
`web/tabs/codemap.js` (GUI), parser bahasa BARU (file baru), agent `codemap-enricher` (definisi
agent = editable, owner: agent jangan frozen).

## VERIFIKASI 2026-06-26
QC live (login :1987): `/api/codemap/status` → 420 node / 433 edge · `/graph` → 420/433 ·
`/semantic` → 160 row enrich. Engine 10 file FROZEN (TestKernelFreeze PASS). Build OK.
