// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval (autonomy grant 2026-06-19/21).
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-21
// Reason: CGM co-reference / canonical-identity resolution (fix fragmentasi identitas owner, bug
//   ke-temu D21 benchmark) — built + unit-tested (build/vet/test green). GENERIC: zero owner
//   identity di file ini (data alias di-seed lokal, BUKAN di repo publik — privasi D8).
//   Extend = new file, jangan modify ini.
//
package agentdb

import "strings"

// cognitive_coref.go — CO-REFERENCE / canonical-identity resolution (roadmap §4.4 follow-on).
//
// Masalah (bug ke-temu D21 recall benchmark): ResolveByEmbedding (cognitive_resolve.go)
// cuma nge-merge node dgn LABEL-synonym (embedding mirip: "mobil"/"car"). TAPI co-reference —
// entitas SAMA disebut NAMA BEDA (mis. "User"/"I"/"saya"/nama-panggilan vs nama-lengkap) —
// embedding-nya BEDA → gak ke-merge → identitas owner PECAH jadi banyak node → graph_recall
// seed di 1 node gak nyampe fakta yg nempel di node lain (recall turun, fakta penting ke-miss).
//
// Fix: tabel alias EKSPLISIT (per-scope) alias→canonical_id. Sebelum bikin node person baru,
// cek alias dulu (DETERMINISTIK, bukan embedding) → resolve ke canonical → anti-fragmentasi.
// GENERIC + portable: mekanisme di sini netral; DATA alias owner di-seed per-agent (mis. lewat
// twin-setup / script identity-merge), BUKAN hardcode di kode bersama.

func normalizeAlias(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func (s *Store) ensureIdentityAliasSchema() {
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS cognitive_identity_alias (
		scope        TEXT NOT NULL DEFAULT '',
		alias        TEXT NOT NULL,
		canonical_id TEXT NOT NULL,
		created_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (scope, alias)
	)`)
}

// RegisterIdentityAlias daftarin 1 alias→canonical (idempotent, alias di-normalize).
// Dipakai twin-setup / identity-merge buat ngajarin "nama-nama lain" si owner.
func (s *Store) RegisterIdentityAlias(scope, alias, canonicalID string) error {
	alias = normalizeAlias(alias)
	canonicalID = strings.TrimSpace(canonicalID)
	if alias == "" || canonicalID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureIdentityAliasSchema()
	_, err := s.db.Exec(
		`INSERT INTO cognitive_identity_alias (scope, alias, canonical_id) VALUES (?,?,?)
		 ON CONFLICT(scope, alias) DO UPDATE SET canonical_id=excluded.canonical_id`,
		scope, alias, canonicalID)
	return err
}

// resolveCanonicalIdentity: kalau `label` (HANYA type person) = alias identitas terdaftar
// utk `scope` DAN canonical node-nya masih ada+active → return canonical id. Dipanggil di
// resolveNodeID SEBELUM ResolveByEmbedding (deterministik > embedding utk identitas).
func (s *Store) resolveCanonicalIdentity(scope, label, typ string) (string, bool) {
	if typ != "person" {
		return "", false
	}
	alias := normalizeAlias(label)
	if alias == "" {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureIdentityAliasSchema()
	var canonicalID string
	if err := s.db.QueryRow(
		`SELECT canonical_id FROM cognitive_identity_alias WHERE scope=? AND alias=?`,
		scope, alias).Scan(&canonicalID); err != nil || strings.TrimSpace(canonicalID) == "" {
		return "", false
	}
	// canonical WAJIB masih ada + active (jangan resolve ke node mati/kehapus)
	var st string
	if e := s.db.QueryRow(`SELECT status FROM cognitive_nodes WHERE id=?`, canonicalID).Scan(&st); e != nil || st != "active" {
		return "", false
	}
	return canonicalID, true
}
