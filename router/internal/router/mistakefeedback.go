// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without explicit owner (Mr.Dev) approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-03
//
// ⚠️⚠️ PERINGATAN UNTUK AI MANAPUN (TERMASUK CLAUDE PASCA-COMPACT) ⚠️⚠️
//   JANGAN ubah/lemahin tanpa Mr.Dev minta EKSPLISIT. Ini loop self-learning
//   anti-halu (pasangan mistakeenrich.go). Dijaga wiring_invariant_auditor.
//
// mistakefeedback.go — FEEDBACK LOOP: tutup lingkaran self-learning.
//
// IDE: antibody (mistakeenrich.go) NYUNTIK koreksi ke prompt. File ini BACA
// hasil model: kalau model masih ngeluarin task_run dgn category NON-kanonik
// (= halu lolos), upsert mistake → KARMA antibody NAIK (hit_count +) → next
// time antibody di-rank lebih tinggi + makin sering keinject. Immune response:
// makin sering ketemu "patogen" (halu), antibody makin kuat. No GPU, otomatis.
//
// Best-effort + async: ga nambah latency, ga pernah bikin request gagal.

package router

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/flowork-os/flowork_Router/internal/brain"
	"github.com/flowork-os/flowork_Router/internal/store"
)

// canonicalTaskCategories — daftar kategori task VALID (HARUS sinkron dgn
// Flowork_Agent task_categories + antibodi seed). Di luar ini = halu.
var canonicalTaskCategories = map[string]struct{}{
	"saham": {}, "crypto": {}, "music": {}, "promo": {},
}

// Antibodi yang di-reinforce. HARUS match seed (UNIQUE category+title) supaya
// SubmitMistake nambah karma ke row yang SAMA, bukan bikin baru.
const (
	antibodyFbCategory = "logic"
	antibodyFbTitle    = "task_run wajib kategori kanonik"
	antibodyFbContent  = "Saat user minta analisa, 'category' di task_run WAJIB dari daftar valid: saham, crypto, music, promo. JANGAN ngarang 'analysis'/'stock'/'security_stock'. 'subject' = entitas murni tanpa suffix '.JK'."
	antibodyFbHit      = 3 // = minHitCount; tiap halu kedeteksi nambah +3 karma
)

// maybeReinforceAntibody — dipanggil SETELAH completion (async) di dispatcher.
// Deteksi halu kategori di response → naikin karma antibody. Best-effort.
func maybeReinforceAntibody(ctx context.Context, resp *OpenAIResponse, settings *store.Settings) {
	if settings == nil || !settings.Brain.Enabled || resp == nil {
		return
	}
	if !brain.Available() {
		return
	}
	bad := detectNonCanonicalTaskRun(resp)
	if bad == "" {
		return // bersih (atau bukan task_run) — ga ada yang di-reinforce
	}
	if _, _, err := brain.SubmitMistake(ctx, antibodyFbCategory, antibodyFbTitle,
		antibodyFbContent, "router-feedback", antibodyFbHit); err != nil {
		log.Printf("flow_router antibody-feedback: submit gagal: %v", err)
		return
	}
	log.Printf("flow_router antibody-feedback: halu kategori %q kedeteksi → karma antibody +%d", bad, antibodyFbHit)
}

// detectNonCanonicalTaskRun — PURE (unit-testable): balik category JELEK pertama
// dari tool_call task_run di response, atau "" kalau bersih / bukan task_run.
func detectNonCanonicalTaskRun(resp *OpenAIResponse) string {
	if resp == nil {
		return ""
	}
	for _, ch := range resp.Choices {
		if len(ch.Message.ToolCalls) == 0 {
			continue
		}
		var calls []openAIToolCall
		if json.Unmarshal(ch.Message.ToolCalls, &calls) != nil {
			continue
		}
		for _, c := range calls {
			if c.Function.Name != "task_run" || c.Function.Arguments == "" {
				continue
			}
			var args struct {
				Category string `json:"category"`
			}
			if json.Unmarshal([]byte(c.Function.Arguments), &args) != nil {
				continue
			}
			cat := strings.ToLower(strings.TrimSpace(args.Category))
			if cat == "" {
				continue
			}
			if _, ok := canonicalTaskCategories[cat]; !ok {
				return cat // halu: kategori di luar daftar kanonik
			}
		}
	}
	return ""
}
