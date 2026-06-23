// mem_type_compat.go — Compatibility layer untuk agent-side brain operations.
//
// Mirror dari canonical MemType yang didefinisikan di router/internal/brain/mem_type_registry.go.
// Karena agent dan router adalah Go module terpisah, file ini menyediakan
// fungsi validasi dan mapping yang sama tanpa cross-module dependency.
//
// PENTING: Jika menambah tipe baru di router/mem_type_registry.go,
// tambahkan juga di sini agar tetap sinkron.
//
// Prinsip: Nano modular, plug & play. File baru, zero modifikasi file FROZEN.
// Dibuat: 2026-06-23 — Phase 2 Memory Typed System.

package agentdb

import "strings"

// ──────────────────────────────────────────────────────────────────────────────
// Canonical MemType Constants (mirrored from router)
// ──────────────────────────────────────────────────────────────────────────────

const (
	memTypeUser              = "user"
	memTypeFeedback          = "feedback"
	memTypeProject           = "project"
	memTypeReference         = "reference"
	memTypeExperience        = "experience"
	memTypeEureka            = "eureka"
	memTypeFact              = "fact"
	memTypeAntibody          = "antibody"
	memTypeDoctrine          = "doctrine"
	memTypeSkill             = "skill"
	memTypeRecoveryInstinct  = "recovery_instinct"
	memTypeCollectiveKnowledge = "collective_knowledge"
)

// canonicalTypes — set untuk O(1) lookup validasi.
var canonicalTypes = map[string]struct{}{
	memTypeUser:                {},
	memTypeFeedback:            {},
	memTypeProject:             {},
	memTypeReference:           {},
	memTypeExperience:          {},
	memTypeEureka:              {},
	memTypeFact:                {},
	memTypeAntibody:            {},
	memTypeDoctrine:            {},
	memTypeSkill:               {},
	memTypeRecoveryInstinct:    {},
	memTypeCollectiveKnowledge: {},
}

// agentLegacyMap memetakan tipe legacy agent-side ke canonical.
var agentLegacyMap = map[string]string{
	// GUI legacy
	"knowledge":        memTypeReference,
	"drawer":           memTypeProject,
	"public_knowledge": memTypeReference,
	// Router legacy
	"compounding":      memTypeProject,
	// Aliases
	"ref":              memTypeReference,
	"refs":             memTypeReference,
	"usr":              memTypeUser,
	"exp":              memTypeExperience,
	"fb":               memTypeFeedback,
	"proj":             memTypeProject,
	"anti":             memTypeAntibody,
	"doc":              memTypeDoctrine,
	"constitution":     memTypeDoctrine,
}

// ──────────────────────────────────────────────────────────────────────────────
// Public API
// ──────────────────────────────────────────────────────────────────────────────

// IsCanonicalMemType mengembalikan true jika s adalah tipe canonical yang valid.
func IsCanonicalMemType(s string) bool {
	_, ok := canonicalTypes[s]
	return ok
}

// NormalizeMemType menerima string mem_type mentah dan mengembalikan
// string canonical yang valid. Flow:
//  1. Jika sudah canonical → return apa adanya.
//  2. Jika ada di legacy map → return mapping.
//  3. Fallback → "experience" (default agent-side, sesuai brain_drawers.go).
func NormalizeMemType(raw string) string {
	raw = strings.TrimSpace(raw)

	// Kosong → default agent
	if raw == "" {
		return memTypeExperience
	}

	// Sudah canonical?
	if _, ok := canonicalTypes[raw]; ok {
		return raw
	}

	// Cek legacy map (case-insensitive)
	lower := strings.ToLower(raw)
	if mapped, ok := agentLegacyMap[lower]; ok {
		return mapped
	}

	// Tidak dikenal → default experience (agent context)
	return memTypeExperience
}

// NormalizeMemTypeForRouter sama seperti NormalizeMemType tapi fallback
// ke "project" (default router context, bukan agent context).
func NormalizeMemTypeForRouter(raw string) string {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return memTypeProject
	}

	if _, ok := canonicalTypes[raw]; ok {
		return raw
	}

	lower := strings.ToLower(raw)
	if mapped, ok := agentLegacyMap[lower]; ok {
		return mapped
	}

	return memTypeProject
}
