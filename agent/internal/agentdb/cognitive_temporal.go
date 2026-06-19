// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval (autonomy grant 2026-06-19).
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-19
// Reason: CGM temporal validity-window (D17/MemPalace) — built + unit-tested. Extend = new file.
//
// cognitive_temporal.go — Temporal validity-window buat edge graph (roadmap D17, ide MemPalace).
//
// Fakta berubah seiring waktu: "dulu Aola suka X, sekarang Y". Daripada hapus edge
// lama (kehilangan sejarah) atau cuma status='obsolete' (kehilangan KAPAN), kita
// kasih tiap edge masa berlaku: valid_from..valid_until. Edge lama gak dibuang —
// di-set valid_until = "true-saat-itu" → bisa di-query timeline.
//
// Migrasi ADDITIVE: cognitive_graph.go (yg bikin tabel) LOCKED → kolom ditambah di
// sini lewat ALTER (idempotent, cek pragma dulu). JANGAN modify file locked.

package agentdb

// ensureTemporalColumns — tambah valid_from/valid_until ke cognitive_edges kalau
// belum ada (idempotent). SQLite ADD COLUMN default WAJIB konstanta → pakai '' (kosong
// = open-ended). Caller WAJIB sudah pegang s.mu + ensureCognitiveGraphSchema().
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

// SupersedeEdge — tandai edge gak berlaku lagi PER tanggal asOf (RFC3339). Set
// valid_until + status='obsolete' (jadi ke-exclude dari recall yg filter active),
// TAPI edge tetap ada buat timeline/audit. Return true kalau ada yg ke-update.
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

// SetEdgeValidFrom — set kapan edge MULAI berlaku (opsional; default '' = sejak dibuat).
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

// EdgesValidAt — edge yang BERLAKU pada waktu asOf (RFC3339), termasuk yg sekarang
// obsolete tapi dulu valid. Kosong ('') di valid_from = sejak awal; '' di valid_until
// = masih berlaku. Buat query "gimana kondisi saat itu" / timeline.
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
