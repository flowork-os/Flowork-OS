// mem_type_registry.go — Canonical MemType enum untuk SELURUH Flowork.
//
// Single source of truth: semua layer (agent, router, GUI, ingest, federation)
// HARUS refer ke file ini untuk daftar tipe memori yang valid.
//
// Dibuat: 2026-06-23 — Phase 2 Memory Typed System.
// Prinsip: Nano modular, plug & play. File baru, zero modifikasi file FROZEN.

package brain

import "strings"

// ──────────────────────────────────────────────────────────────────────────────
// Canonical MemType enum
// ──────────────────────────────────────────────────────────────────────────────

// MemType adalah tipe canonical untuk klasifikasi memori di Flowork.
type MemType string

const (
	// ── Core 4 (dari remember.md skill) ──────────────────────────────────
	MemTypeUser      MemType = "user"      // Data personal tentang owner/pengguna. Amplitude: 999999 (sticky).
	MemTypeFeedback  MemType = "feedback"   // Feedback, koreksi, review dari user. Amplitude: 8000-9000.
	MemTypeProject   MemType = "project"    // Catatan teknis, kode, arsitektur proyek. Amplitude: 5000-7000.
	MemTypeReference MemType = "reference"  // Referensi umum, dokumentasi, pengetahuan. Amplitude: 1000-3000.

	// ── Agent Experience ─────────────────────────────────────────────────
	MemTypeExperience MemType = "experience" // Pengalaman interaksi agent. Promotable via federation.
	MemTypeEureka     MemType = "eureka"     // Insight dari dream cycle. Promotable via federation.

	// ── Verified Knowledge ───────────────────────────────────────────────
	MemTypeFact     MemType = "fact"     // Fakta terverifikasi. Promotable via federation.
	MemTypeAntibody MemType = "antibody" // Anti-hallucination data, koreksi kesalahan.

	// ── System/Structural ────────────────────────────────────────────────
	MemTypeDoctrine MemType = "doctrine" // Aturan konstitusional, sacred rules.
	MemTypeSkill    MemType = "skill"    // Skill yang dipelajari agent.

	// ── Federation Special ───────────────────────────────────────────────
	MemTypeRecoveryInstinct    MemType = "recovery_instinct"    // Instinct recovery yang terfederasi.
	MemTypeCollectiveKnowledge MemType = "collective_knowledge" // Pengetahuan kolektif terfederasi.
)

// AllMemTypes — daftar lengkap semua canonical MemType, urut berdasarkan
// kelompok fungsional.
var AllMemTypes = []MemType{
	MemTypeUser, MemTypeFeedback, MemTypeProject, MemTypeReference,
	MemTypeExperience, MemTypeEureka,
	MemTypeFact, MemTypeAntibody,
	MemTypeDoctrine, MemTypeSkill,
	MemTypeRecoveryInstinct, MemTypeCollectiveKnowledge,
}

// ──────────────────────────────────────────────────────────────────────────────
// Validation
// ──────────────────────────────────────────────────────────────────────────────

// validSet di-build sekali saat init untuk O(1) lookup.
var validSet map[MemType]struct{}

func init() {
	validSet = make(map[MemType]struct{}, len(AllMemTypes))
	for _, mt := range AllMemTypes {
		validSet[mt] = struct{}{}
	}
}

// IsValid mengembalikan true jika s adalah canonical MemType yang dikenal.
func IsValid(s string) bool {
	_, ok := validSet[MemType(s)]
	return ok
}

// Validate mengembalikan MemType canonical jika valid, atau ("", false) jika tidak.
func Validate(s string) (MemType, bool) {
	mt := MemType(s)
	_, ok := validSet[mt]
	if !ok {
		return "", false
	}
	return mt, true
}

// ──────────────────────────────────────────────────────────────────────────────
// Legacy Mapping
// ──────────────────────────────────────────────────────────────────────────────

// legacyMap memetakan tipe lama/non-canonical ke tipe canonical.
// Key: lowercase legacy string → Value: canonical MemType.
var legacyMap = map[string]MemType{
	// GUI legacy types
	"knowledge":        MemTypeReference, // GUI "knowledge" → reference
	"drawer":           MemTypeProject,   // GUI "drawer" → project
	"public_knowledge": MemTypeReference, // GUI "public_knowledge" → reference

	// Router write.go default
	"compounding": MemTypeProject, // Router AddDrawer default → project

	// Alias/typo tolerance
	"ref":        MemTypeReference,
	"refs":       MemTypeReference,
	"usr":        MemTypeUser,
	"exp":        MemTypeExperience,
	"fb":         MemTypeFeedback,
	"proj":       MemTypeProject,
	"anti":       MemTypeAntibody,
	"doc":        MemTypeDoctrine,
	"constitution": MemTypeDoctrine, // constitution → doctrine
}

// MapLegacy memetakan string legacy/non-canonical ke MemType canonical.
// Jika string sudah canonical, dikembalikan apa adanya.
// Jika tidak dikenal sama sekali, fallback ke MemTypeProject.
func MapLegacy(s string) MemType {
	// 1. Sudah canonical?
	if mt, ok := Validate(s); ok {
		return mt
	}

	// 2. Cek legacy map (case-insensitive)
	lower := strings.ToLower(strings.TrimSpace(s))
	if mt, ok := legacyMap[lower]; ok {
		return mt
	}

	// 3. Fallback: project (sesuai router schema default)
	return MemTypeProject
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// Promotable mengembalikan true jika MemType ini boleh dipromosikan via federation.
// Sesuai gate di federation.go: hanya experience, eureka, fact.
func Promotable(mt MemType) bool {
	switch mt {
	case MemTypeExperience, MemTypeEureka, MemTypeFact:
		return true
	default:
		return false
	}
}

// FreshIndexable mengembalikan true jika MemType ini masuk fresh-recall index.
// Sesuai fresh_index.go: hanya recovery_instinct, collective_knowledge.
func FreshIndexable(mt MemType) bool {
	switch mt {
	case MemTypeRecoveryInstinct, MemTypeCollectiveKnowledge:
		return true
	default:
		return false
	}
}

// Sacred mengembalikan true jika MemType ini "keramat" dan tidak boleh di-reclassify.
// Tipe: user (data personal), doctrine (konstitusi), antibody (anti-hallucination).
func Sacred(mt MemType) bool {
	switch mt {
	case MemTypeUser, MemTypeDoctrine, MemTypeAntibody:
		return true
	default:
		return false
	}
}

// GUIOptions mengembalikan daftar MemType yang harus tampil di GUI dropdown,
// dalam urutan yang user-friendly.
func GUIOptions() []MemType {
	return []MemType{
		MemTypeProject,
		MemTypeReference,
		MemTypeFeedback,
		MemTypeUser,
		MemTypeDoctrine,
		MemTypeExperience,
		MemTypeFact,
		MemTypeAntibody,
		MemTypeSkill,
		MemTypeEureka,
		MemTypeRecoveryInstinct,
		MemTypeCollectiveKnowledge,
	}
}

// String implements fmt.Stringer.
func (mt MemType) String() string {
	return string(mt)
}
