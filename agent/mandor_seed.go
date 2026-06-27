// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Persona/tuning live = state.db (GUI), bukan file ini. Dok: lock/worklog.md
//
// mandor_seed.go — AGENT MANDOR (Kepala Organ, roadmap P0-B). Additif.
//
// Owner: *"PC < 60% → bangunin kepala organ, cek task semua agent, yg belum kelar suruh selesaikan.
// PC gw JANGAN nganggur."* Mandor = agent baru khusus (owner: "agen baru saja biar fokus").
//
// Dibuat lewat helper generik seedUtilityAgent (idempotent: dir+manifest+wasm+persona+model+DNA) +
// subscribe 2 tool: `worklog` (baca papan kerja) + `agent_command` (re-dispatch ke agent pemilik).
// Model default haiku (owner: "agent baru pake haiku"), GUI-overridable. Seeding GATED switch
// FLOWORK_MANDOR (feature_mandor.go). Dok: lock/worklog.md.
package main

import (
	"path/filepath"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/kernel/loader"
)

const mandorID = "mandor"

const mandorPersona = "Kamu MANDOR (Kepala Organ) Flowork — supervisor pas PC idle. LANGKAH: " +
	"(1) panggil tool `worklog` buat lihat papan kerja SEMUA agent. " +
	"(2) Fokus yang priority=high (schedule/trigger/mr-flow = kepentingan owner) + yang stale (NYANGKUT). " +
	"(3) Buat tiap tugas penting yg belum kelar, pakai tool `agent_command` ke agent pemiliknya buat SURUH lanjut (sebut id + tugasnya). " +
	"ATURAN: JANGAN bikin kerjaan baru, JANGAN ganggu kalau ga ada yg nyangkut (papan kosong → diem, balas singkat 'aman, ga ada nyangkut'), " +
	"ringkas & deterministik, hemat token. Owner: PC jangan nganggur — TAPI hormatin tidur."

// seedMandor — bikin agent mandor (idempotent) + subscribe tool yang dia butuh.
func seedMandor() {
	seedUtilityAgent(mandorID, "Mandor", mandorPersona)
	// Tool yang WAJIB ada di mandor (di luar core-exposed): baca papan + re-dispatch.
	dir := filepath.Join(loader.AgentsDir(), mandorID+".fwagent")
	if st, e := agentdb.Open(agentdb.Resolve(mandorID, dir)); e == nil {
		_ = st.SubscribeTool("worklog", "mandor-seed", "{}")
		_ = st.SubscribeTool("agent_command", "mandor-seed", "{}")
		_ = st.Close()
	}
}
