// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without explicit owner (Mr.Dev) approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-03
//
// ⚠️⚠️ PERINGATAN UNTUK AI MANAPUN (TERMASUK CLAUDE PASCA-COMPACT) ⚠️⚠️
//   JANGAN ubah / refactor / "rapihin" file ini tanpa Mr.Dev minta EKSPLISIT.
//   Ini jantung anti-halu deterministik (antibody injection). Kalau lo lupa
//   konteks gara-gara context window ke-compact: STOP, tanya Mr.Dev dulu.
//   Logika sengaja simpel + fails-open. "Perbaikan" sepihak = ngerusak.
//
// mistakeenrich.go — ANTIBODY INJECTION (Phase 2 yang di-defer di
// internal/brain/mistakes.go baris 28-31: "inject MAX 3 antibody relevant").
//
// TUJUAN:
//   Nyambungin pipa yang putus: mistakes journal (knowledge SALAH yang
//   terbukti) selama ini cuma DISIMPEN, ga pernah nyampe ke model. File ini
//   narik mistakes relevan, rank by KARMA (hit_count) × relevansi (keyword
//   overlap), inject MAX 3 sebagai system-message "antibodi" SEBELUM LLM
//   dipanggil. Hasil: model lemah sekalipun (qwen-7b) ga ngulang halu yang
//   udah kecatat — DETERMINISTIK, no GPU, makin sering kecatat makin kuat.
//
// PRINSIP (Mr.Dev): "deterministik = kuat, LLM lemah = rapuh". Jangan ngarep
// model manggil mistake_recall sendiri — PAKSA lewat injeksi di gateway.
//
// Fails open: brain mati / DB ga ada / kosong / error → skip, request tetap
// jalan normal. Ga pernah bikin request gagal.

package router

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/flowork-os/flowork_Router/internal/brain"
	"github.com/flowork-os/flowork_Router/internal/store"
)

const (
	// antibodyMaxInject — hard cap jumlah antibodi yang disuntik (anti
	// over-prompt; sesuai catatan penulis brain/mistakes.go "MAX 3").
	antibodyMaxInject = 3
	// antibodyUniversalKarma — mistake dgn hit_count >= ini dianggap antibodi
	// UNIVERSAL (selalu kandidat walau ga ada keyword overlap), karena udah
	// kebukti berkali-kali. Di bawah ini wajib relevan (overlap > 0).
	antibodyUniversalKarma = 10
	// antibodyMistakeTier — tier mistakes yang dipakai (promosi global dari agent).
	antibodyMistakeTier = "global"
	// antibodyHalfLifeDays — decay: tiap N hari TANPA di-reinforce, bobot recency
	// jadi separuh. Antibodi yg berhenti relevan pelan-pelan pudar dari injeksi.
	antibodyHalfLifeDays = 30.0
	// antibodyRecencyFloor — recency ga pernah 0 (antibodi fondasi karma tinggi
	// tetep kandidat walau lama), tapi kalah sama yg fresh+relevan.
	antibodyRecencyFloor = 0.1
)

// maybeInjectAntibodies — dipanggil dari dispatcher SETELAH maybeEnrichBrain.
// Mutate req.Messages in-place (nambah 1 system message antibodi). Best-effort.
func maybeInjectAntibodies(ctx context.Context, req *OpenAIRequest, settings *store.Settings) {
	if settings == nil || !settings.Brain.Enabled {
		return
	}
	if settings.Brain.DBPath != "" {
		brain.SetDBPath(settings.Brain.DBPath)
	}
	if !brain.Available() {
		return
	}
	query := lastUserText(req.Messages)
	if query == "" {
		return
	}
	ab := relevantAntibodies(ctx, query, antibodyMaxInject)
	if len(ab) == 0 {
		return
	}
	sys := buildAntibodySystem(ab)
	if sys == "" {
		return
	}
	// "augment" mode: sisip setelah blok system caller, jangan dominasi persona.
	req.Messages = injectSystem(req.Messages, sys, "augment")
	log.Printf("flow_router antibody: injected %d antibody (query=%.48q)", len(ab), query)
}

// relevantAntibodies — perkawinan KARMA × RELEVANSI:
//
//	score = karma(hit_count) × (1 + 2·overlap)
//
// Recency (decay) implicit: ListMistakes ngurut updated_at DESC, sort stable
// jaga urutan itu buat tie-break (yang baru di-reinforce menang). Filter
// anti-noise: skip mistake yang ga nyambung KECUALI karma-nya universal.
func relevantAntibodies(ctx context.Context, query string, max int) []brain.Mistake {
	all, err := brain.ListMistakes(ctx, antibodyMistakeTier, "", 200)
	if err != nil || len(all) == 0 {
		return nil
	}
	return rankAntibodies(all, query, max, time.Now())
}

// rankAntibodies — PURE (no DB, `now` di-inject biar testable): perkawinan
// KARMA × RELEVANSI × DECAY.
//
//	score = karma(hit_count) × (1 + 2·overlap) × recency(updated_at)
//
// recency = decay eksponensial dari updated_at (di-refresh tiap reinforce).
// Antibodi yg berhenti relevan → recency turun → pudar dari top-N otomatis.
func rankAntibodies(all []brain.Mistake, query string, max int, now time.Time) []brain.Mistake {
	qTokens := tokenSet(query)
	type scored struct {
		m     brain.Mistake
		score float64
	}
	ranked := make([]scored, 0, len(all))
	for _, m := range all {
		overlap := overlapCount(qTokens, tokenSet(m.Title+" "+m.Content+" "+m.Category))
		if overlap == 0 && m.HitCount < antibodyUniversalKarma {
			continue // ga relevan + ga universal → skip (anti-noise)
		}
		karma := m.HitCount
		if karma < 1 {
			karma = 1
		}
		rec := recencyFactor(m.UpdatedAt, now)
		ranked = append(ranked, scored{m, float64(karma) * float64(1+2*overlap) * rec})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	out := make([]brain.Mistake, 0, max)
	for i := 0; i < len(ranked) && i < max; i++ {
		out = append(out, ranked[i].m)
	}
	return out
}

// recencyFactor — decay eksponensial (half-life antibodyHalfLifeDays), floor
// antibodyRecencyFloor. Fresh = 1.0; makin lama updated_at = makin kecil.
// Format updated_at ga keparse → anggap tua (floor), jangan crash.
func recencyFactor(updatedAt string, now time.Time) float64 {
	t, ok := parseMistakeTime(updatedAt)
	if !ok {
		return antibodyRecencyFloor
	}
	days := now.Sub(t).Hours() / 24
	if days < 0 {
		days = 0
	}
	f := math.Pow(0.5, days/antibodyHalfLifeDays)
	if f < antibodyRecencyFloor {
		f = antibodyRecencyFloor
	}
	if f > 1 {
		f = 1
	}
	return f
}

// parseMistakeTime — coba format SubmitMistake (RFC3339) + fallback sqlite datetime.
func parseMistakeTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// buildAntibodySystem — render antibodi jadi system message. "kuat N" = hit_count
// = berapa kali kesalahan ini kecatat (karma).
func buildAntibodySystem(ms []brain.Mistake) string {
	if len(ms) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Antibodi — kesalahan TERBUKTI, JANGAN diulang\n")
	b.WriteString("Pelajaran dari kesalahan lampau (\"kuat\" = berapa kali kecatat; makin tinggi makin wajib dipatuhi):\n\n")
	for i, m := range ms {
		title := strings.TrimSpace(m.Title)
		content := strings.TrimSpace(m.Content)
		fmt.Fprintf(&b, "%d. [%s · kuat %d] %s", i+1, m.Category, m.HitCount, title)
		if content != "" {
			fmt.Fprintf(&b, " — %s", content)
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

// tokenSet — lowercase, pecah non-alfanumerik, buang token pendek/stopword.
func tokenSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	}) {
		if len(f) < 3 {
			continue
		}
		if _, stop := antibodyStopwords[f]; stop {
			continue
		}
		out[f] = struct{}{}
	}
	return out
}

// overlapCount — jumlah token sama antara dua set.
func overlapCount(a, b map[string]struct{}) int {
	// iterate yang lebih kecil.
	if len(b) < len(a) {
		a, b = b, a
	}
	n := 0
	for t := range a {
		if _, ok := b[t]; ok {
			n++
		}
	}
	return n
}

// antibodyStopwords — kata umum ID/EN yang ga ngebantu relevansi.
var antibodyStopwords = map[string]struct{}{
	"yang": {}, "untuk": {}, "dari": {}, "dengan": {}, "dan": {}, "atau": {},
	"the": {}, "and": {}, "for": {}, "with": {}, "this": {}, "that": {},
	"buat": {}, "kalau": {}, "saja": {}, "lagi": {}, "sudah": {}, "akan": {},
}
