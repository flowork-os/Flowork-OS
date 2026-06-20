// === LOCKED FILE (soft) === Status: STABLE — owner-approved 2026-06-20 (enrich jadi agent, model GUI).
// LOCKED ≠ FREEZE (boleh diedit dgn izin owner). Verified: enrich via agent pakai opus GUI.
package main

// codemap_enricher_seed.go — AGENT enrich codemap (owner 2026-06-20: "enrich jadi AGENT, semua
// agent kumpul di menu Agent, model JANGAN hardcode = cacat arsitektur, owner ganti model kapan
// aja lewat GUI"). Dulu enrich = LLM hardcode di host (codemapSemanticSummarizer→routerChat).
// Sekarang otaknya pindah ke agent `codemap-enricher`: persona cartographer di DB, model dari
// Settings per-agent (kv router_model). Host cuma KIRIM isi file → agent balas JSON. Idempotent.

import (
	"os"
	"path/filepath"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/agentmgr"
	"flowork-gui/internal/kernel/loader"
)

const enricherID = "codemap-enricher"

// enricherPersona — SAMA semantiknya dgn sys-prompt summarizer lama, tapi sekarang di DB persona
// agent (bisa di-tweak owner di GUI). Output WAJIB JSON {summary,domain,role} biar host bisa parse.
const enricherPersona = "Kamu CODEBASE CARTOGRAPHER Flowork. Dikasih SATU file source, balas HANYA " +
	"objek JSON ringkas (tanpa markdown/fence/prosa): " +
	`{"summary":"satu kalimat: file ini ngapain","domain":"area fungsional (mis. auth/triggers/brain/ui/codemap/finance/router/orchestrator)","role":"peran arsitektur (mis. http-handler/engine/data-store/config/parser/wasm-agent/test)"}. ` +
	"JSON only."

// seedCodemapEnricher — bikin agent codemap-enricher (idempoten). Model TIDAK di-hardcode
// (owner set di Settings GUI). Dipanggil boot SETELAH template wasm ke-build.
func seedCodemapEnricher() {
	agentsDir := loader.AgentsDir()
	dir := filepath.Join(agentsDir, enricherID+".fwagent")
	if _, e := os.Stat(dir); e != nil {
		tplWasm, err := os.ReadFile(filepath.Join("templates", "agent-template", "agent.wasm"))
		if err != nil || len(tplWasm) == 0 {
			return // template belum ke-build → skip (start.sh build dulu)
		}
		if mk := os.MkdirAll(filepath.Join(dir, "workspace"), 0o755); mk != nil {
			return
		}
		_ = os.WriteFile(filepath.Join(dir, "manifest.json"), evoMemberManifest(enricherID, "Codemap Enricher"), 0o644)
		_ = os.WriteFile(filepath.Join(dir, "agent.wasm"), tplWasm, 0o644)
		_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("agent.wasm\nworkspace/*.db\nworkspace/*.db-*\n"), 0o644)
	}
	// persona + tools STANDAR minimal (pure summarizer: ga butuh fs/codemap, isi file dikirim host).
	if st, e := agentdb.Open(agentdb.Resolve(enricherID, dir)); e == nil {
		_ = st.SetPrompt(enricherPersona)
		for _, t := range []string{"brain_search_shared", "ScheduleWakeup"} {
			_ = st.SubscribeTool(t, "seed:codemap-enricher", "{}")
		}
		_ = st.Close()
	}
	agentmgr.ProvisionAgentDNA(enricherID) // warga penuh (konstitusi sacred + DNA)
}
