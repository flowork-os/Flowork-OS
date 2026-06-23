// reclassify_rules.go — Classification rules engine untuk reclassify daemon.
//
// Defines heuristic rules untuk auto-classify drawers berdasarkan content analysis.
// Digunakan oleh reclassify.go (goroutine background) dan brain_classify.go (agent pre-write).
//
// Prinsip: Cabut akar — rules berbasis konten, bukan tambal default.
// Dibuat: 2026-06-23 — Phase 2 Memory Typed System.

package brain

import "strings"

// ──────────────────────────────────────────────────────────────────────────────
// ClassifyRule — satu aturan klasifikasi
// ──────────────────────────────────────────────────────────────────────────────

// ClassifyRule mendefinisikan satu aturan heuristik untuk menentukan MemType.
// Rules diproses berurutan dari prioritas tertinggi ke terendah.
// Rule pertama yang Match() == true menentukan hasilnya.
type ClassifyRule struct {
	Name     string                                   // nama rule (untuk logging/debug)
	Priority int                                      // semakin tinggi = semakin prioritas
	Target   MemType                                  // tipe canonical target
	Match    func(content, wing, room string) bool     // fungsi pencocokan
}

// ──────────────────────────────────────────────────────────────────────────────
// Default Rules — urut berdasarkan prioritas (tinggi → rendah)
// ──────────────────────────────────────────────────────────────────────────────

// defaultRules adalah daftar aturan klasifikasi built-in.
// Urutan: Sacred types dulu (user, doctrine, antibody), lalu spesifik, lalu umum.
var defaultRules = []ClassifyRule{
	// ── Priority 100: Sacred / User ──────────────────────────────────────
	{
		Name:     "user_personal_info",
		Priority: 100,
		Target:   MemTypeUser,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			// Deteksi info personal: nama owner, keluarga, preferensi pribadi
			personalKeywords := []string{
				"aola", "mr.dev", "mr dev", "mrdev",
				"istri", "anak", "keluarga", "family",
				"ulang tahun", "birthday", "alamat rumah",
				"nomor hp", "phone number", "email pribadi",
				"hobi", "hobby", "favorit", "favorite",
				"nama saya", "my name", "tentang saya", "about me",
			}
			for _, kw := range personalKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			// Wing/room khusus personal
			lw := strings.ToLower(wing)
			if lw == "personal" || lw == "user" || lw == "owner" {
				return true
			}
			return false
		},
	},

	// ── Priority 95: Doctrine / Constitution ─────────────────────────────
	{
		Name:     "doctrine_constitution",
		Priority: 95,
		Target:   MemTypeDoctrine,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			doctrineKeywords := []string{
				"constitution", "konstitusi", "sacred rule",
				"aturan emas", "golden rule", "prinsip dasar",
				"fundamental principle", "doktrin", "doctrine",
				"haram", "wajib", "mutlak",
			}
			for _, kw := range doctrineKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			if lw == "constitution" || lw == "doctrine" || lw == "sacred" {
				return true
			}
			return false
		},
	},

	// ── Priority 90: Antibody ────────────────────────────────────────────
	{
		Name:     "antibody_correction",
		Priority: 90,
		Target:   MemTypeAntibody,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			abKeywords := []string{
				"antibody", "anti-hallucination", "hallucination",
				"koreksi fatal", "fatal correction", "jangan halusinasi",
				"fakta salah", "wrong fact", "misinformation",
				"immune", "imun",
			}
			for _, kw := range abKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			if lw == "antibody" || lw == "immune" || lw == "correction" {
				return true
			}
			return false
		},
	},

	// ── Priority 80: Feedback ────────────────────────────────────────────
	{
		Name:     "feedback_review",
		Priority: 80,
		Target:   MemTypeFeedback,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			fbKeywords := []string{
				"feedback:", "review:", "koreksi:", "correction:",
				"saran:", "suggestion:", "perbaiki:", "fix:",
				"tolong ubah", "please change", "seharusnya",
				"should be", "lebih baik kalau", "better if",
				"user bilang", "user said", "dev bilang",
			}
			for _, kw := range fbKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			if lw == "feedback" || lw == "review" {
				return true
			}
			return false
		},
	},

	// ── Priority 70: Fact ────────────────────────────────────────────────
	{
		Name:     "verified_fact",
		Priority: 70,
		Target:   MemTypeFact,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			factKeywords := []string{
				"fakta:", "fact:", "verified:", "terverifikasi:",
				"confirmed:", "dikonfirmasi:", "truth:", "kebenaran:",
			}
			for _, kw := range factKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			return false
		},
	},

	// ── Priority 60: Skill ───────────────────────────────────────────────
	{
		Name:     "learned_skill",
		Priority: 60,
		Target:   MemTypeSkill,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			skillKeywords := []string{
				"skill:", "cara:", "how to:", "langkah:",
				"step by step:", "tutorial:", "procedure:",
				"prosedur:", "teknik:", "technique:",
			}
			for _, kw := range skillKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			if lw == "skill" || lw == "skills" || lw == "technique" {
				return true
			}
			return false
		},
	},

	// ── Priority 50: Reference ───────────────────────────────────────────
	{
		Name:     "reference_knowledge",
		Priority: 50,
		Target:   MemTypeReference,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			refKeywords := []string{
				"referensi:", "reference:", "dokumentasi:", "documentation:",
				"panduan:", "guide:", "manual:", "spesifikasi:", "specification:",
				"api:", "endpoint:", "schema:", "format:",
			}
			for _, kw := range refKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			if lw == "reference" || lw == "documentation" || lw == "docs" || lw == "knowledge" {
				return true
			}
			return false
		},
	},

	// ── Priority 40: Eureka (dream-generated) ────────────────────────────
	{
		Name:     "eureka_insight",
		Priority: 40,
		Target:   MemTypeEureka,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			if strings.Contains(lower, "eureka:") || strings.Contains(lower, "insight:") {
				return true
			}
			lw := strings.ToLower(wing)
			if lw == "eureka" || lw == "dream" || lw == "insight" {
				return true
			}
			lr := strings.ToLower(room)
			if lr == "eureka" || lr == "dream_cycle" || lr == "cognitive_graph" {
				return true
			}
			return false
		},
	},

	// ── Priority 30: Recovery Instinct ───────────────────────────────────
	{
		Name:     "recovery_instinct",
		Priority: 30,
		Target:   MemTypeRecoveryInstinct,
		Match: func(content, wing, room string) bool {
			lw := strings.ToLower(wing)
			lr := strings.ToLower(room)
			if lw == "recovery" || lw == "instinct" || lw == "recovery_instinct" {
				return true
			}
			if lr == "recovery" || lr == "instinct" {
				return true
			}
			lower := strings.ToLower(content)
			return strings.Contains(lower, "recovery instinct:") || strings.Contains(lower, "instinct pattern:")
		},
	},

	// ── Priority 25: Collective Knowledge ────────────────────────────────
	{
		Name:     "collective_knowledge",
		Priority: 25,
		Target:   MemTypeCollectiveKnowledge,
		Match: func(content, wing, room string) bool {
			lw := strings.ToLower(wing)
			lr := strings.ToLower(room)
			if lw == "collective" || lw == "collective_knowledge" {
				return true
			}
			if lr == "collective" || lr == "collective_knowledge" {
				return true
			}
			lower := strings.ToLower(content)
			return strings.Contains(lower, "collective knowledge:") || strings.Contains(lower, "pengetahuan kolektif:")
		},
	},

	// ── Priority 20: Project (broad catch) ───────────────────────────────
	{
		Name:     "project_technical",
		Priority: 20,
		Target:   MemTypeProject,
		Match: func(content, wing, room string) bool {
			lower := strings.ToLower(content)
			projKeywords := []string{
				"bug:", "error:", "fix:", "implement:", "todo:",
				"feature:", "fitur:", "deploy:", "build:",
				"commit:", "merge:", "branch:", "release:",
				"config:", "konfigurasi:", "architecture:",
				"refactor:", "optimize:", "optimasi:",
			}
			for _, kw := range projKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
			lw := strings.ToLower(wing)
			projWings := []string{
				"project", "dev", "development", "engineering",
				"backend", "frontend", "infra", "infrastructure",
				"training_data", "compounding",
			}
			for _, pw := range projWings {
				if lw == pw {
					return true
				}
			}
			return false
		},
	},
}

// ──────────────────────────────────────────────────────────────────────────────
// Classification Engine
// ──────────────────────────────────────────────────────────────────────────────

// ClassifyContent menganalisis konten drawer dan mengembalikan MemType canonical
// yang paling cocok berdasarkan rules engine.
//
// Parameter currentType digunakan sebagai referensi — jika currentType sudah
// canonical DAN Sacred(), maka TIDAK akan di-reclassify (protection).
//
// Jika tidak ada rule yang cocok, fallback ke currentType (jika valid) atau MemTypeProject.
func ClassifyContent(content, wing, room, currentType string) MemType {
	// Protection: jangan reclassify tipe Sacred
	if mt, ok := Validate(currentType); ok && Sacred(mt) {
		return mt
	}

	// Jalankan rules dari prioritas tertinggi
	// (defaultRules sudah urut berdasarkan prioritas di kode)
	for _, rule := range defaultRules {
		if rule.Match(content, wing, room) {
			return rule.Target
		}
	}

	// Tidak ada rule yang match — cek apakah currentType sudah canonical
	if mt, ok := Validate(currentType); ok {
		return mt
	}

	// Last resort: map legacy atau fallback project
	return MapLegacy(currentType)
}

// ClassifyContentStrict sama seperti ClassifyContent, tapi HANYA reclassify
// jika confidence tinggi (minimal 1 rule match). Jika tidak ada match,
// kembalikan currentType apa adanya (tanpa fallback).
func ClassifyContentStrict(content, wing, room, currentType string) (MemType, bool) {
	// Protection: Sacred types tidak boleh di-reclassify
	if mt, ok := Validate(currentType); ok && Sacred(mt) {
		return mt, false // false = tidak berubah
	}

	for _, rule := range defaultRules {
		if rule.Match(content, wing, room) {
			current := MemType(currentType)
			changed := current != rule.Target
			return rule.Target, changed
		}
	}

	// Tidak ada match — kembalikan apa adanya
	return MemType(currentType), false
}
