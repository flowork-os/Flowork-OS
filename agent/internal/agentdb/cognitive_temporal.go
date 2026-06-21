// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

func (s *Store) ensureTemporalColumns() {
	rows, err := s.db.Query(`PRAGMA table_info(cognitive_edges)`)
	if err != nil {
		return
	}
	have := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
			have[name] = true
		}
	}
	rows.Close()
	if !have["valid_from"] {
		_, _ = s.db.Exec(`ALTER TABLE cognitive_edges ADD COLUMN valid_from TEXT NOT NULL DEFAULT ''`)
	}
	if !have["valid_until"] {
		_, _ = s.db.Exec(`ALTER TABLE cognitive_edges ADD COLUMN valid_until TEXT NOT NULL DEFAULT ''`)
	}
}

func (s *Store) SupersedeEdge(fromID, toID, relationType, asOf string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	s.ensureTemporalColumns()
	r, err := s.db.Exec(
		`UPDATE cognitive_edges SET valid_until=?, status='obsolete'
		 WHERE from_id=? AND to_id=? AND relation_type=? AND status='active'`,
		asOf, fromID, toID, relationType)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n > 0, nil
}

func (s *Store) SetEdgeValidFrom(fromID, toID, relationType, from string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	s.ensureTemporalColumns()
	_, err := s.db.Exec(
		`UPDATE cognitive_edges SET valid_from=? WHERE from_id=? AND to_id=? AND relation_type=?`,
		from, fromID, toID, relationType)
	return err
}

func (s *Store) EdgesValidAt(asOf string, limit int) ([]GraphEdgeView, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	s.ensureTemporalColumns()
	rows, err := s.db.Query(`
		SELECT from_id, to_id, relation_type, status, strength FROM cognitive_edges
		WHERE (valid_from='' OR valid_from <= ?)
		  AND (valid_until='' OR valid_until > ?)
		ORDER BY strength DESC LIMIT ?`, asOf, asOf, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GraphEdgeView
	for rows.Next() {
		var e GraphEdgeView
		if err := rows.Scan(&e.FromID, &e.ToID, &e.RelationType, &e.Status, &e.Strength); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
