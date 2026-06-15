// === LOCKED FILE (soft) === Status: STABLE — owner-approved 2026-06-16 (autonomous sprint A1).
// LOCKED ≠ FREEZE (boleh diedit dgn izin owner). Tested: evolve_pillars_test.go (8 case classify + kanonik).
//
// evolve_pillars.go — A1 GOVERNANCE EVOLUSI (additive, plug-in). Owner-approved 2026-06-16.
//
// 5 PILAR TUJUAN (diputus owner+Opus 2026-06-16, lihat roadmap_asi.md "A1-DESIGN"). SETIAP
// proposal evolusi WAJIB nyentuh ≥1 pilar — kalau nggak = "ngelantur" (di luar tujuan organisme).
// Ini gerbang PERTAMA & PALING MURAH (pre-debat): sebelum dewan adversarial (Pembela/Penantang/
// Hakim) yang butuh model ≥4.7, saring dulu yang jelas-jelas ngelantur.
//
// Klasifikasi di sini DETERMINISTIK (keyword, generous) = fallback yang gak butuh LLM + jalan di
// model lokal. Pas dewan penuh aktif (cloud), proposer boleh DECLARE pilar sendiri; classifier ini
// tetap jadi validator/penyaring murah. ADDITIVE: nol ubah perilaku lama (kolom baru + fungsi murni).

package agentdb

import "strings"

// EvolvePillar — satu pilar tujuan + kata kunci pengenalnya.
type EvolvePillar struct {
	ID    string // id kanonik (dipakai di DB + gate)
	Emoji string
	Label string
	kw    []string // keyword (lowercase) penanda proposal nyentuh pilar ini
}

// EvolvePillars — 5 pilar kanonik. URUTAN = hierarki desain: ekonomi+keamanan = dua kaki yg nahan
// hidup, warga+kecerdasan = pengali, mandiri = tujuan akhir. Keamanan = LANTAI KERAS (lihat gate dewan).
var EvolvePillars = []EvolvePillar{
	{ID: "ekonomi", Emoji: "💰", Label: "Ekonomi (cari duit / biayai diri)", kw: []string{
		"ekonomi", "duit", "uang", "income", "pemasukan", "bounty", "hackerone", "bugcrowd",
		"affiliate", "afiliasi", "konten", "monetiz", "monetisasi", "revenue", "bayar", "finance",
		"finansial", "wallet", "dompet", "jual", "profit", "cuan", "trading", "royalti", "donasi", "sponsor",
	}},
	{ID: "keamanan", Emoji: "🛡️", Label: "Keamanan Flowork (lantai keras)", kw: []string{
		"keamanan", "aman", "security", "secure", "hardening", "harden", "sandbox", "audit", "immune",
		"imun", "antibody", "antibodi", "vuln", "vulnerab", "exploit", "defense", "defensif", "proteksi",
		"protect", "enkripsi", "encrypt", "auth", "isolasi", "isolat", "attack", "serangan", "malware",
		"rce", "injection", "injeksi", "ssrf", "xss", "spoof", "leak", "bocor", "privacy", "privasi",
	}},
	{ID: "warga", Emoji: "🤝", Label: "Manfaat + mudahin warga/agent lain", kw: []string{
		"warga", "user", "pengguna", "mudah", "mempermudah", "permudah", "bantu", "membantu", "helpful",
		"usability", "ux", "onboarding", "dokumentasi", "docs", "komunitas", "community", "share",
		"berbagi", "gampang", "simpel", "sederhana", "accessible", "aksesibilitas", "ramah", "intuitif",
	}},
	{ID: "kecerdasan", Emoji: "🧠", Label: "Kecerdasan + evolusi", kw: []string{
		"cerdas", "pintar", "kecerdasan", "evolusi", "evolve", "learning", "belajar", "reasoning",
		"nalar", "skill", "knowledge", "pengetahuan", "brain", "otak", "capability", "kemampuan",
		"intelligence", "akurasi", "accuracy", "kualitas jawaban", "konteks", "memori", "memory", "rag",
	}},
	{ID: "mandiri", Emoji: "♾️", Label: "Hidup tanpa owner (mandiri/self-heal)", kw: []string{
		"mandiri", "self-heal", "selfheal", "self heal", "rollback", "autonomous", "otonom", "survive",
		"bertahan", "recovery", "recover", "watchdog", "resilience", "resilient", "uptime", "swadaya",
		"self-suffic", "failover", "heal", "auto-restart", "auto restart", "boot", "tetap hidup", "tahan",
	}},
}

// ClassifyPillars — pilar mana saja yang DISENTUH teks (rationale+goal+kind digabung caller).
// Balik id pilar dalam urutan kanonik (EvolvePillars). Generous: cocok kalau salah satu keyword muncul.
func ClassifyPillars(text string) []string {
	s := strings.ToLower(text)
	out := []string{}
	for _, p := range EvolvePillars {
		for _, k := range p.kw {
			if strings.Contains(s, k) {
				out = append(out, p.ID)
				break
			}
		}
	}
	return out
}

// PillarFit — TRUE kalau teks nyentuh ≥1 pilar (lolos gerbang murah). FALSE = "ngelantur".
func PillarFit(text string) bool {
	return len(ClassifyPillars(text)) > 0
}

// PillarsLabel — gabung id pilar jadi label pendek buat ditaruh di proposal/GUI (mis. "💰ekonomi 🛡️keamanan").
func PillarsLabel(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	byID := map[string]EvolvePillar{}
	for _, p := range EvolvePillars {
		byID[p.ID] = p
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			parts = append(parts, p.Emoji+p.ID)
		} else {
			parts = append(parts, id)
		}
	}
	return strings.Join(parts, " ")
}
