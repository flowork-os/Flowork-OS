// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import "strings"

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

	var st string
	if e := s.db.QueryRow(`SELECT status FROM cognitive_nodes WHERE id=?`, canonicalID).Scan(&st); e != nil || st != "active" {
		return "", false
	}
	return canonicalID, true
}
