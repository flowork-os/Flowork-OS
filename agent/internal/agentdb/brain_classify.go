// brain_classify.go — Pre-write classifier utility untuk agent-side brain.
//
// Menyediakan fungsi SuggestMemType() yang bisa dipanggil oleh kode baru
// SEBELUM memanggil brain_add / AddBrainDrawer.
//
// CATATAN: File brain_local.go (brain_add tool) dan brain_drawers.go (AddBrainDrawer)
// adalah FROZEN. Classifier ini TIDAK memodifikasi file tersebut.
// Caller baru (tool baru, wrapper, atau middleware) bisa memanggil SuggestMemType()
// dan kemudian meneruskan hasilnya sebagai parameter mem_type ke brain_add.
//
// Prinsip: Switch / jalan pintas — bukan bongkar file FROZEN.
// Dibuat: 2026-06-23 — Phase 2 Memory Typed System.

package agentdb

import "strings"

func init() {
	MemTypeClassifierHook = func(content, wing, room, currentType string) string {
		return SuggestMemType(content, wing, room, currentType)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Heuristic Classifier
// ──────────────────────────────────────────────────────────────────────────────

// SuggestMemType menganalisis konten dan konteks (wing/room) untuk menyarankan
// mem_type canonical yang paling cocok.
//
// Ini adalah versi agent-side dari reclassify_rules.go di router.
// Bedanya: fungsi ini dipanggil SEBELUM write (preventive),
// sedangkan reclassify daemon jalan SETELAH write (corrective).
//
// Jika suggestedType sudah diberikan dan valid, fungsi ini menghormatinya
// (tidak override pilihan eksplisit dari agent/user).
func SuggestMemType(content, wing, room, suggestedType string) string {
	// Jika caller sudah kasih tipe eksplisit yang valid → hormati
	if suggestedType != "" && IsCanonicalMemType(suggestedType) {
		return suggestedType
	}

	// Jika suggestedType ada tapi legacy → normalize dulu
	if suggestedType != "" {
		normalized := NormalizeMemType(suggestedType)
		if IsCanonicalMemType(normalized) && normalized != memTypeExperience {
			// Hanya terima normalisasi jika hasilnya bukan default fallback
			return normalized
		}
	}

	// Auto-classify berdasarkan konten
	return classifyByContent(content, wing, room)
}

// classifyByContent — heuristic classifier agent-side.
// Mirip dengan reclassify_rules.go tapi lebih ringan (tanpa struct ClassifyRule).
func classifyByContent(content, wing, room string) string {
	lower := strings.ToLower(content)
	lw := strings.ToLower(wing)
	lr := strings.ToLower(room)

	// ── User/Personal (prioritas tertinggi) ──────────────────────────────
	if containsAny(lower, "aola", "mr.dev", "mr dev", "mrdev",
		"istri", "anak", "keluarga", "family",
		"ulang tahun", "birthday", "alamat rumah",
		"nama saya", "my name", "tentang saya", "about me",
		"nomor hp", "phone number", "email pribadi") {
		return memTypeUser
	}
	if lw == "personal" || lw == "user" || lw == "owner" {
		return memTypeUser
	}

	// ── Doctrine ─────────────────────────────────────────────────────────
	if containsAny(lower, "constitution", "konstitusi", "sacred rule",
		"aturan emas", "golden rule", "doktrin", "doctrine",
		"haram", "wajib", "mutlak") {
		return memTypeDoctrine
	}
	if lw == "constitution" || lw == "doctrine" || lw == "sacred" {
		return memTypeDoctrine
	}

	// ── Antibody ─────────────────────────────────────────────────────────
	if containsAny(lower, "antibody", "anti-hallucination", "hallucination",
		"koreksi fatal", "fakta salah", "wrong fact", "immune", "imun") {
		return memTypeAntibody
	}
	if lw == "antibody" || lw == "immune" {
		return memTypeAntibody
	}

	// ── Feedback ─────────────────────────────────────────────────────────
	if containsAny(lower, "feedback:", "review:", "koreksi:", "correction:",
		"saran:", "suggestion:", "perbaiki:", "tolong ubah",
		"seharusnya", "should be", "lebih baik kalau") {
		return memTypeFeedback
	}
	if lw == "feedback" || lw == "review" {
		return memTypeFeedback
	}

	// ── Fact ─────────────────────────────────────────────────────────────
	if containsAny(lower, "fakta:", "fact:", "verified:", "terverifikasi:",
		"confirmed:", "dikonfirmasi:") {
		return memTypeFact
	}

	// ── Skill ────────────────────────────────────────────────────────────
	if containsAny(lower, "skill:", "cara:", "how to:", "langkah:",
		"step by step:", "tutorial:", "procedure:", "prosedur:") {
		return memTypeSkill
	}
	if lw == "skill" || lw == "skills" {
		return memTypeSkill
	}

	// ── Reference ────────────────────────────────────────────────────────
	if containsAny(lower, "referensi:", "reference:", "dokumentasi:", "documentation:",
		"panduan:", "guide:", "manual:", "spesifikasi:", "api:", "schema:") {
		return memTypeReference
	}
	if lw == "reference" || lw == "documentation" || lw == "docs" || lw == "knowledge" {
		return memTypeReference
	}

	// ── Eureka ───────────────────────────────────────────────────────────
	if containsAny(lower, "eureka:", "insight:") {
		return memTypeEureka
	}
	if lw == "eureka" || lw == "dream" || lr == "eureka" || lr == "dream_cycle" {
		return memTypeEureka
	}

	// ── Recovery / Collective (federation special) ───────────────────────
	if lw == "recovery" || lw == "instinct" || lw == "recovery_instinct" {
		return memTypeRecoveryInstinct
	}
	if lw == "collective" || lw == "collective_knowledge" {
		return memTypeCollectiveKnowledge
	}

	// ── Project (broad catch) ────────────────────────────────────────────
	if containsAny(lower, "bug:", "error:", "fix:", "implement:", "todo:",
		"feature:", "fitur:", "deploy:", "build:", "commit:", "merge:",
		"refactor:", "optimize:", "config:", "architecture:") {
		return memTypeProject
	}
	if lw == "project" || lw == "dev" || lw == "development" ||
		lw == "backend" || lw == "frontend" || lw == "infra" {
		return memTypeProject
	}

	// ── Default: experience (agent context) ──────────────────────────────
	return memTypeExperience
}

// containsAny mengembalikan true jika s mengandung salah satu dari keywords.
func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
