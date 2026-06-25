// guards.go — GUARD DETERMINISTIK bersama untuk SEMUA semut (agentkit).
//
// ⚠️ FROZEN (chattr +i + hash KERNEL_FREEZE.md, owner-approved 2026-06-25). Guard = yang bikin
// model 26B ga ngamuk; lindungi dari edit-tak-sadar AI. Ubah = SADAR (unfreeze→rebuild semua→
// re-hash→re-freeze). Lihat lock/agentkit.md + ERROR_EDUKASI.md.
//
// AKAR (roadmap AGENTKIT, Rule 5/6): loop tool-calling + guard KE-COPY 6× per agent.
// flail/ghost guard cuma ada di mr-flow; 5 agent lain + template GAK punya → nyalain
// all-tools/#2C buat mereka = FLAIL. File ini = guard yang DI-SHARE (port dari mr-flow
// yang PROVEN): tiap semut yang pakai agentkit otomatis warisan. Fix sekali → semua dapet.
//
// Prinsip: model lokal lemah → HARNESS angkat beban (guard deterministik), jangan ngarep
// model nyadar sendiri. Semua TinyGo-safe (substring match, no regexp, no goroutine).
package agentkit

import (
	"encoding/json"
	"strings"
)

// ============================ FLAIL-GUARD (anti-mantok) ============================
// Port dari agents/mr-flow/flail_guard.go (FROZEN, proven 4/4 unit test).
//
//   ghost-guard = model NARASI niat tanpa manggil tool (janji kosong).
//   flail-guard = model MANGGIL tool SAMA berulang TANPA progress (mantok/muter).
//
// DETEKSI = window-duplikat (BUKAN "tool sama beruntun" — itu false-positive di kerja
// sah kayak baca 10 file beda). Sinyal flail BERSIH: signature `tool|args` SAMA PERSIS
// ≥flailDupMax kali dalam flailWindow call terakhir. AKSI: bukan hard-stop — inject
// koreksi keras bounded (maxFlailNudges); kalau MASIH mantok → eskalasi JUJUR ke owner.

const (
	flailWindow    = 8 // lihat N tool-call terakhir
	flailDupMax    = 3 // signature SAMA muncul ≥N kali dalam window → flailing
	maxFlailNudges = 4 // bounded koreksi (mirror maxGhostNudges)
)

// flailState — di-track per turn (scope tool-loop). Loop yang nyimpen + manggil check().
type flailState struct {
	window []string // ring signature `tool|args` dari call terakhir
	nudges int      // berapa kali udah dikoreksi (bounded)
}

// check — update state dengan call sekarang, balik keputusan:
//   - nudge=true    → loop inject `corrective` sbg pesan user, lanjut (bounded).
//   - escalate=true → udah dikoreksi maxFlailNudges× tapi MASIH mantok → eskalasi jujur.
//   - dua-duanya false → normal.
func (f *flailState) check(toolName, argsJSON string) (corrective string, nudge bool, escalate bool) {
	sig := toolName + "|" + argsJSON
	f.window = append(f.window, sig)
	if len(f.window) > flailWindow {
		f.window = f.window[len(f.window)-flailWindow:]
	}
	dup := 0
	for _, s := range f.window {
		if s == sig {
			dup++
		}
	}
	if dup < flailDupMax {
		return "", false, false
	}
	if f.nudges >= maxFlailNudges {
		return "", false, true // udah dikoreksi cukup, tetep mantok → eskalasi
	}
	f.nudges++
	f.window = f.window[:0] // reset window: kasih kesempatan fresh abis dikoreksi
	return flailCorrective(toolName), true, false
}

// flailCorrective — koreksi deterministik pas mantok (konkret + kasih PILIHAN aksi).
func flailCorrective(toolName string) string {
	return "⚠️ STOP — lo udah manggil tool `" + toolName + "` berkali-kali dengan hasil yang SAMA / ga maju (mantok). Ngulang cara yang sama = buang waktu + owner nunggu sia-sia. PILIH SEKARANG salah satu: (a) pakai tool LAIN yang lebih tepat buat tujuanmu; (b) kalau ga yakin tool mana yang ada, panggil `tool_search` buat nyari; (c) kalau emang lagi NUNGGU sesuatu yang belum siap, panggil `ScheduleWakeup`; (d) kalau infonya udah cukup, kasih hasil-sejauh-ini ke owner. JANGAN ulang `" + toolName + "` dengan args yang sama lagi."
}

// flailEscalation — pesan JUJUR pas model tetep mantok lewat batas koreksi.
func flailEscalation(toolName string) string {
	return "🛑 Gw jujur: mantok di `" + toolName + "` — udah dicoba berkali-kali + udah gw koreksi sendiri tapi belum nemu jalan. Gw stop di sini biar ga muter terus + buang resource. Butuh arahan lo: tujuannya apa persisnya, atau ada tool/cara lain yang lo tau buat ini?"
}

// ============================ GHOST-GUARD (anti-ghosting) ============================
// Port dari agents/mr-flow/main.go (phrase list superset, proven). Model NARASI niat-aksi
// ("tunggu bentar gw cek dulu") TANPA manggil tool → paksa 1 putaran (bounded maxGhostNudges).

const ghostNudgeMsg = "⚠️ Lo barusan bilang mau ngelakuin sesuatu (cek/list/scan/cari/tunggu) TAPI GA manggil tool apa-apa. Itu artinya owner nunggu jawaban yang ga bakal datang (ghosting) — DILARANG. LAKUIN SEKARANG di balasan ini: panggil tool yang lo maksud (mis. file_list, glob, grep, file_read). KALAU emang harus nunggu sesuatu yang belum siap / kerja lama, panggil ScheduleWakeup(delaySeconds, reason, prompt) biar lo kebangun otomatis & lanjut sendiri. JANGAN jawab teks doang lagi."

// ghostPhrases — sinyal NIAT-AKSI khas ghosting. TinyGo-safe: substring match (no regexp).
var ghostPhrases = []string{
	"tunggu bentar", "tunggu sebentar", "tunggu ya",
	"bentar gw", "bentar ya", "sebentar gw", "sebentar ya",
	"gw cek dulu", "cek dulu ya", "gw lihat dulu", "gw liat dulu",
	"gw list dulu", "gw scan dulu", "gw cari dulu", "gw periksa dulu",
	"gw kerjain dulu", "gw proses dulu", "lagi gw cek", "lagi gw proses",
	"hasilnya nyusul", "nyusul ya", "stay tuned",
	"nanti gw kabarin", "nanti gw lapor", "gw kabarin nanti",
	// POLA KELANJUTAN/LOOP (model narasi "lanjut ke huruf b" lalu BERHENTI, ga chain).
	"lanjut ke huruf", "mulai ke huruf", "lanjut ke karakter", "lanjut ke prefix",
	"lanjut ke pencarian", "pencarian berikutnya", "alfabet berikutnya", "scan berikutnya",
	"lanjut ke alfabet", "lanjut scan", "lanjutin scan", "mulai scanning", "mulai scan",
	"proses looping dimulai", "proses scanning dimulai", "iterasi berikutnya", "lanjut ke iterasi",
	"berikutnya...", "berikutnya…", "lanjut ke tahap berikutnya",
}

// looksLikeGhostPromise — true kalau teks (tanpa tool-call) nyinyalin niat-aksi ga ditepatin.
func looksLikeGhostPromise(s string) bool {
	low := strings.ToLower(s)
	for _, p := range ghostPhrases {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}

// ====================== RECOVERY-CAPTURE (D32 INC-2 self-learning) ======================
// Port dari agents/mr-flow/recovery_capture.go (FROZEN). Pas tool ERROR lalu tool SAMA
// SUKSES dalam loop turn = agent recover → mistake_log recovery-instinct. Best-effort
// (tool ga ada/gagal = graceful, ga crash). TinyGo-safe.

// recoveryCaptureSkip — tool meta/recall/notify yg error→sukses-nya BUKAN "recovery" bermakna.
var recoveryCaptureSkip = map[string]bool{
	"mistake_log": true, "graph_recall": true, "brain_search": true,
	"brain_search_shared": true, "instinct_recall": true, "tool_search": true,
	"telegram_send": true, "ScheduleWakeup": true, "now": true,
	"memory_get": true, "interaction_recall": true, "tool_lookup": true,
}

// toolErrClass — hasil tool error host ({"error":...}) → KELAS error ringkas (BEBAS
// path/identifier owner, privasi). Sukses ({"ok":true}) / non-JSON → "".
func toolErrClass(result string) string {
	t := strings.TrimSpace(result)
	if !strings.HasPrefix(t, "{") {
		return ""
	}
	var m map[string]any
	if json.Unmarshal([]byte(t), &m) != nil {
		return ""
	}
	if ok, _ := m["ok"].(bool); ok {
		return ""
	}
	es, _ := m["error"].(string)
	es = strings.ToLower(strings.TrimSpace(es))
	if es == "" {
		return ""
	}
	switch {
	case strings.Contains(es, "not found"), strings.Contains(es, "no such"), strings.Contains(es, "tidak ada"), strings.Contains(es, "404"):
		return "not-found"
	case strings.Contains(es, "permission"), strings.Contains(es, "denied"), strings.Contains(es, "ditolak"):
		return "permission"
	case strings.Contains(es, "timeout"), strings.Contains(es, "timed out"), strings.Contains(es, "deadline"):
		return "timeout"
	case strings.Contains(es, "already exists"), strings.Contains(es, "sudah ada"):
		return "already-exists"
	case strings.Contains(es, "invalid"), strings.Contains(es, "parse"), strings.Contains(es, "syntax"), strings.Contains(es, "unmarshal"), strings.Contains(es, "required"):
		return "invalid-input"
	case strings.Contains(es, "dispatch"), strings.Contains(es, "tool http"), strings.Contains(es, "capability"), strings.Contains(es, "denied by cap"):
		return "blocked"
	default:
		return "error"
	}
}

// captureRecovery — seen[tool] = kelas error tool yg BELUM ke-recover dalam loop turn.
// Pas tool SAMA sukses setelah error → mistake_log. Side-effect only.
func captureRecovery(tool, result string, seen map[string]string) {
	if recoveryCaptureSkip[tool] {
		return
	}
	if cls := toolErrClass(result); cls != "" {
		seen[tool] = cls
		return
	}
	cls, had := seen[tool]
	if !had {
		return
	}
	delete(seen, tool)
	_ = runTool("mistake_log", map[string]any{
		"category":       "workflow",
		"title":          "recovery: " + tool + "/" + cls,
		"content":        "WHEN " + tool + " gagal (" + cls + ") -> RECOVERED: agent berhasil di percobaan berikutnya (perbaiki argumen / ganti pendekatan).",
		"context_origin": "auto-recovery",
	})
}
