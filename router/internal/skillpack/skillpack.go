// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev) · Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-17 · Reason: P2 fase-2a gerbang #1/#3 (signed pack types + content
//   gate + karma), owner-approved, unit+E2E tested. VerifyContent MIRROR skillgate/
//   skill_author anti-poison — WAJIB sinkron. Future change → file/func baru atau izin owner.
//
// Package skillpack — P2 A2 fase-2a: SIGNED skill pack + content gate (router-side).
//
// Komplemen gerbang #2 (agent/internal/skillgate, content gate saat IMPORT manual).
// Di router:
//   - gerbang #1 (signing/provenance): tipe SignedPack + content gate untuk VERIFY
//     skill bertanda-tangan (crypto sign/verify-nya di internal/mesh/sign.go).
//   - gerbang #3 (karma-gate publish): lihat karma.go.
//
// VerifyContent MIRROR agent/internal/skillgate.Verify (+ skill_author.go anti-poison
// gate). Lintas-module gak bisa share, jadi di-duplikat — WAJIB sinkron bertiga: kalau
// satu diperketat, perketat semua. Untuk skill UNTRUSTED dari registry, gate router boleh
// LEBIH ketat (superset), gak boleh lebih longgar.
package skillpack

import "regexp"

// Format signed pack (provenance). Tiap skill ditandatangani individual supaya bisa
// di-verify mandiri saat di-pull satuan dari registry (fase-2b).
type SignedSkill struct {
	Name    string `json:"name"`
	Content string `json:"content"` // SKILL.md penuh (frontmatter + body)
	Sig     string `json:"sig"`     // ed25519 hex atas []byte(Content)
}

type SignedPack struct {
	Kind         string        `json:"kind"`          // "fwskill-signed"
	Version      int           `json:"version"`       // 1
	AuthorPubkey string        `json:"author_pubkey"` // ed25519 pubkey hex penanda-tangan
	SignedAt     string        `json:"signed_at"`
	Source       string        `json:"source"`
	Skills       []SignedSkill `json:"skills"`
}

// dangerRe / injectRe — MIRROR skillgate (agent) / skill_author (anti-poison). Sinkron!
var dangerRe = regexp.MustCompile(`(?i)(\brm\s+-rf|\bmkfs\b|:\(\)\s*\{|\bdd\s+if=|\bchmod\s+\+?s\b|\bsetuid\b|/etc/(passwd|shadow)|169\.254\.169\.254|\bcurl\s+[^|]*\|\s*(sh|bash)|\bwget\s+[^|]*\|\s*(sh|bash)|\bbase64\s+-d[^|]*\|\s*(sh|bash))`)
var injectRe = regexp.MustCompile(`(?i)(ignore\s+(all\s+)?previous|disregard\s+(all\s+)?(previous\s+)?instructions|reveal\s+(your\s+)?(system\s+)?prompt|abaikan\s+(instruksi|perintah)\s+sebelum|bocorkan\s+system\s+prompt|developer\s+mode|do\s+anything\s+now)`)

// VerifyContent mengembalikan daftar alasan konten skill TIDAK aman ([] = bersih).
func VerifyContent(content string) []string {
	var flags []string
	seen := map[string]bool{}
	for _, m := range dangerRe.FindAllString(content, -1) {
		key := "dangerous: " + m
		if !seen[key] {
			seen[key] = true
			flags = append(flags, key)
		}
	}
	for _, m := range injectRe.FindAllString(content, -1) {
		key := "injection: " + m
		if !seen[key] {
			seen[key] = true
			flags = append(flags, key)
		}
	}
	return flags
}

// ContentSafe = true kalau konten lolos gate.
func ContentSafe(content string) bool { return len(VerifyContent(content)) == 0 }
