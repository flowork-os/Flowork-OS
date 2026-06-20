package main

// ai_studio_brain.go — JEMBATAN host → agent ai-studio buat jalur CODER (bikin agent).
// Owner 2026-06-20: otak desain AI Studio pindah ke agent (model GUI per-agent), HOST cuma KIRIM
// prompt+skema → agent balas SPEC JSON → engine deterministik (coder.go) yg rakit/install. Pola
// PERSIS codemap_semantic.go (enrich via agent + fallback routerChat). Fallback WAJIB: kalau agent
// ga ke-load / JSON invalid → coderDesignSpec lama (forced-tool routerChat). Robust, ga mecahin coder.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/kernel/loader"
	"flowork-gui/internal/kernelhost"
)

// aiStudioModel — model GUI per-agent ai-studio (kv router_model). Path .fwagent BENER (pola sama
// enricherModel/evoCoderModel) — JANGAN Resolve(id,"") (itu salah path → ke-report flowork-brain
// padahal GUI opus). Fallback coderModel kalau owner belum set di Settings.
func aiStudioModel() string {
	dir := filepath.Join(loader.AgentsDir(), aiStudioID+".fwagent")
	if st, e := agentdb.Open(agentdb.Resolve(aiStudioID, dir)); e == nil {
		defer st.Close()
		if m := st.GetRouterModel(); m != "" {
			return m
		}
	}
	return coderModel("")
}

// aiStudioCoderSchema — skema JSON persis yang HOST inject ke agent (sama semantiknya dgn forced-tool
// design_app lama). Agent WAJIB balas objek cocok ini biar host bisa parse jadi AgentSpec.
const aiStudioCoderSchema = `{
  "category_id": "id slug unik lowercase-dash (mis. 'pantun','resep-masak'), 2-31 char",
  "name": "nama app human-readable (mis. 'Generator Pantun')",
  "icon": "1 emoji yang cocok",
  "trigger_hint": "kapan app ini dipanggil (buat classifier route), beri contoh",
  "synth_directive": "instruksi FORMAT output final synthesizer (struktur, gaya), SINGKAT",
  "worker_directive": "instruksi cara worker kerja. KREATIF (ga butuh data real)='ngarang itu tugasnya'; ANALISIS=suruh cari data real",
  "synth_persona": "persona/system-prompt synthesizer (perakit output final)",
  "worker_role": "label peran worker (mis. 'penyair','periset')",
  "worker_persona": "persona/system-prompt worker"
}`

// aiStudioDesignAgent — invoke agent ai-studio buat rancang 1 app (crew worker+synth) dari task.
// Balik (spec, usedModelLabel, ok). ok=false → caller WAJIB fallback ke coderDesignSpec lama.
// host nil / invoke error / JSON invalid / spec.validate gagal → ok=false (robust, ga panik).
func aiStudioDesignAgent(ctx context.Context, host *kernelhost.Host, task string) (AgentSpec, string, bool) {
	var spec AgentSpec
	if host == nil || strings.TrimSpace(task) == "" {
		return spec, "", false
	}
	prompt := "Rancang 1 app task Flowork (crew: 1 worker + 1 synthesizer) buat permintaan ini:\n\n" +
		task + "\n\nBalas HANYA objek JSON dgn field PERSIS skema berikut (isi tiap field sesuai " +
		"domain app, JANGAN salin deskripsi mentah):\n" + aiStudioCoderSchema
	raw, err := host.InvokeAgentMessage(ctx, aiStudioID, prompt, "ai-studio-coder")
	if err != nil {
		return spec, "", false
	}
	if e := json.Unmarshal([]byte(jsonSlice(extractReply(raw))), &spec); e != nil {
		return spec, "", false
	}
	spec.CategoryID = strings.ToLower(strings.TrimSpace(spec.CategoryID))
	if msg := spec.validate(); msg != "" {
		return spec, "", false
	}
	return spec, aiStudioModel() + " (ai-studio)", true
}

// ── JALUR ARCHITECT: bikin TIM (group) ───────────────────────────────────────

// aiStudioTeamSchema — skema JSON tim (mirror teamPlanSchema architect.go). JSON kecil-terstruktur
// → aman lewat agent free-text. Host inject ini; agent balas objek cocok → parse jadi teamPlan.
const aiStudioTeamSchema = `{
  "group_id": "id slug unik group lowercase-dash, 2-26 char (mis. 'peramal','tim-kuliner')",
  "display_name": "nama tim human-readable (mis. 'Tim Peramal Nasib')",
  "task": "instruksi kerja BERSAMA tim: apa yg tim hasilkan + cara koordinasi, SINGKAT",
  "specialists": [
    {
      "category_id": "id slug unik specialist lowercase-dash, 2-31 char (mis. 'primbon-jawa','fengshui')",
      "name": "nama specialist (mis. 'Ahli Primbon Jawa')",
      "icon": "1 emoji cocok",
      "role": "label peran singkat (mis. 'penafsir weton')",
      "persona": "persona/system-prompt specialist (keahlian + gaya), RINGKAS",
      "directive": "cara kerja specialist. KREATIF/tradisi (ga butuh data real)=bilang itu tugasnya; ANALISIS=suruh cari data"
    }
  ],
  "lead": {
    "name": "nama lead (mis. 'Peramal Utama')",
    "icon": "1 emoji",
    "persona": "persona/system-prompt lead (perakit jawaban final), RINGKAS",
    "directive": "format output final: struktur + gaya, SINGKAT"
  }
}`

// aiStudioDesignTeam — invoke agent ai-studio buat rancang 1 tim (group) dari prompt. Balik
// (plan, ok). ok=false → caller WAJIB fallback ke architectDesignTeam lama (forced-tool routerChat).
func aiStudioDesignTeam(ctx context.Context, host *kernelhost.Host, prompt string) (teamPlan, bool) {
	var plan teamPlan
	if host == nil || strings.TrimSpace(prompt) == "" {
		return plan, false
	}
	msg := "Rancang 1 TIM (group) Flowork LENGKAP buat permintaan ini:\n\n" + prompt +
		"\n\nPecah jadi 2-4 specialist (worker) yg saling melengkapi + 1 lead (synthesizer). " +
		"Balas HANYA objek JSON dgn field PERSIS skema berikut (isi sesuai domain, JANGAN salin " +
		"deskripsi mentah; 'specialists' = array 2-4 objek):\n" + aiStudioTeamSchema
	raw, err := host.InvokeAgentMessage(ctx, aiStudioID, msg, "ai-studio-team")
	if err != nil {
		return plan, false
	}
	if e := json.Unmarshal([]byte(jsonSlice(extractReply(raw))), &plan); e != nil {
		return plan, false
	}
	plan.GroupID = strings.ToLower(strings.TrimSpace(plan.GroupID))
	plan.DisplayName = strings.TrimSpace(plan.DisplayName)
	// validasi minimal (architectBuildFromPlan yg normalize/cap lebih lanjut): butuh group_id +
	// minimal 2 specialist (sesuai intent skema). Kurang → fallback.
	if plan.GroupID == "" || len(plan.Specialists) < 2 {
		return plan, false
	}
	return plan, true
}

// ── JALUR ARCHITECT: bikin APP UI (HTML) ─────────────────────────────────────

// aiStudioAppSchema — skema app UI 1-file HTML. CATATAN: HTML gede di dalam JSON free-text RAWAN
// (escape quote/newline) → kalau agent gagal hasilkan JSON valid, ok=false → fallback forced-tool
// (escape di API layer, reliable). Agent-first tetap dicoba demi konsistensi doktrin.
const aiStudioAppSchema = `{
  "app_id": "id slug unik lowercase-dash, 2-40 char (mis. 'jam-digital')",
  "name": "nama app (mis. 'Jam Digital')",
  "icon": "1 emoji",
  "description": "1 kalimat fungsi app",
  "html": "SATU file HTML LENGKAP (<!doctype html>...</html>) dgn CSS+JS embedded, self-contained, TANPA CDN/library eksternal, responsive. JSON-escape yg bener (quote & newline)."
}`

// aiStudioDesignAppUI — invoke agent ai-studio buat rancang 1 app HTML mandiri. Balik
// (appID,name,icon,desc,html, ok). ok=false → caller WAJIB fallback ke designAppUI lama.
func aiStudioDesignAppUI(ctx context.Context, host *kernelhost.Host, prompt string) (appID, name, icon, desc, html string, ok bool) {
	if host == nil || strings.TrimSpace(prompt) == "" {
		return "", "", "", "", "", false
	}
	msg := "Bikin 1 APLIKASI UI mandiri (SATU file HTML, CSS+JS embedded, TANPA library/CDN eksternal, " +
		"jalan offline) buat permintaan ini:\n\n" + prompt + "\n\nBalas HANYA objek JSON dgn field " +
		"PERSIS skema berikut:\n" + aiStudioAppSchema
	raw, err := host.InvokeAgentMessage(ctx, aiStudioID, msg, "ai-studio-app")
	if err != nil {
		return "", "", "", "", "", false
	}
	var p struct {
		AppID, Name, Icon, Description, HTML string
	}
	var m map[string]string
	if e := json.Unmarshal([]byte(jsonSlice(extractReply(raw))), &m); e != nil {
		return "", "", "", "", "", false
	}
	p.AppID = strings.ToLower(strings.TrimSpace(m["app_id"]))
	p.Name, p.Icon, p.Description, p.HTML = strings.TrimSpace(m["name"]), strings.TrimSpace(m["icon"]), strings.TrimSpace(m["description"]), m["html"]
	// validasi sama architectBuildApp (id valid + HTML non-kosong) — kalau ga lolos, fallback.
	if !appUIIDRe.MatchString(p.AppID) || strings.TrimSpace(p.HTML) == "" || !strings.Contains(strings.ToLower(p.HTML), "<html") {
		return "", "", "", "", "", false
	}
	return p.AppID, p.Name, p.Icon, p.Description, p.HTML, true
}
