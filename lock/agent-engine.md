# AI AGENT — ENGINE beku, AGENT editable (owner: agent masih halu)

> Dev: Aola Sahidin. 2026-06-26. Keputusan owner: **freeze ENGINE doang, AGENT-nya JANGAN** —
> "agent belum terbukti work masih halu jadi engine-nya saja". Agent (yang BERPIKIR) belum stabil →
> harus tetap bisa diperbaiki/evolve. Engine (mesin yang MENJALANKAN) udah proven → beku.
> ⚠️ Ini MEMBALIK freeze mr-flow sesi 2026-06-26 lama (keputusan-opus) — agent sekarang DI-UNFREEZE.

## BATAS: ENGINE vs AGENT
**ENGINE (FROZEN)** = mesin generik yang jalanin agent APAPUN:
- `internal/kernel/runtime/{runtime,host,instance}.go` (WASM host) · `internal/kernel/{loader,broker}/*`
- `internal/kernelhost/kernelhost.go` (host-agent bridge)
- `agentkit/{agentkit.go,guards.go}` (SDK agent + guard generik: flail/ghost/defer)
- `internal/agentmgr/agentmgr.go` (manajemen agent) · `internal/tools/{registry,sandbox,sandbox_v3}.go`
- `templates/{agent,group,connector}-template/*` (CETAKAN bikin agent baru — engine, bukan instance)

**AGENT (NON-frozen, EDITABLE)** = definisi tiap agent (yang berpikir, masih halu):
- `agents/mr-flow/{main.go,flail_guard.go,recall_gate.go,working_set.go,recovery_capture.go,manifest.json}`
  + `agents/mr-flow.fwagent/{agent.wasm,manifest.json}` (orchestrator — paling sering di-debug halu)
- `agents/{browse-surfer,browse-reporter,fbspecial,fb-writer,fb-repofinder}/{main.go,go.mod}` (squad)
- deploy `~/.flowork/agents/{codemap-enricher,dream-digester}.fwagent/*` (utility agent)
- **22 entry ini DIKELUARIN dari KERNEL_FREEZE.md + chattr -i** → editable bebas, gak perlu unfreeze.

## KENAPA
Agent = LLM-reasoning (persona/prompt/tool-loop) → masih halu (history-poisoning, flail, ghost).
Bekuin = ga bisa benerin → mandek. Engine = deterministik/proven → bekuin AMAN. Prinsip tetap:
hasil evolusi agent GAK BISA robohin ENGINE (engine beku); agent rusak ≠ engine rusak (sandbox isolasi).

## CARA KERJA (evolusi agent)
- **Benerin agent halu** = edit `agents/<id>/main.go` / `manifest.json` langsung (OPEN, no unfreeze) →
  rebuild wasm. Guard generik (agentkit/guards.go, FROZEN) tetap lindungi dari sisi engine.
- **Agent baru** = copy `templates/agent-template` (FROZEN cetakan) → `agents/<id>/` (editable) → deploy.
- **Model/prompt runtime** = `state.db` / GUI (udah adjustable, by-design).

## SWITCH
`FLOWORK_ORCHESTRATOR` (default mr-flow) — pilih orchestrator (registry.go). Migrasi mr-flow→mr-flow-next
= set ENV, engine gak disentuh. Agent-level = DATA (state.db, manifest), bukan switch.

## VERIFIKASI 2026-06-26
Engine spot-check FROZEN (runtime/kernelhost/agentkit/agentmgr). Agent defs OPEN (mr-flow/main.go,
fbspecial/main.go editable). 22 entry dikeluarin dari manifest. TestKernelFreeze PASS (engine tetap
ke-enforce). Build agent+router OK.
