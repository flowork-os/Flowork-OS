// codemap_semantic.go — R6 SELF-MAP SEMANTIK (plug-in extension, additive).
// Owner-approved 2026-06-15 (FASE 2 autonomi). Lapisan MAKNA di atas codemap_files
// (struktur deterministik). Tiap file node dapet summary/domain/role hasil analisa LLM
// (prinsip semut: 1 call kecil per file). Tabel TERPISAH — gak sentuh codemap.go /
// codemap_files.go (dua-duanya LOCKED). Schema CREATE IF NOT EXISTS = non-destruktif.

package agentdb

import "time"

// CodemapSemantic — lapisan makna 1 file node (di atas struktur deterministik).
type CodemapSemantic struct {
	Path      string `json:"path"`    // relatif ke codemapRoot (match codemap_files.path)
	Summary   string `json:"summary"` // 1 kalimat: file ini ngapain
	Domain    string `json:"domain"`  // area fungsional: auth, triggers, brain, ui, …
	Role      string `json:"role"`    // peran arsitektur: http-handler, engine, data-store, …
	Model     string `json:"model"`   // model yang ngehasilin (provenance)
	Hash      string `json:"hash"`    // sha256 konten file saat enrich (M1: deteksi file berubah)
	IndexedAt string `json:"indexed_at"`
}

func (s *Store) ensureCodemapSemanticSchema() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS codemap_semantic (
		  path         TEXT PRIMARY KEY,
		  summary      TEXT NOT NULL DEFAULT '',
		  domain       TEXT NOT NULL DEFAULT '',
		  role         TEXT NOT NULL DEFAULT '',
		  model        TEXT NOT NULL DEFAULT '',
		  content_hash TEXT NOT NULL DEFAULT '',
		  indexed_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}
	// M1 migrasi additif: tabel lama (sebelum content_hash) → tambah kolom. Idempotent:
	// kalau kolom udah ada, SQLite error "duplicate column" → diabaikan (bukan kegagalan).
	_, _ = s.db.Exec(`ALTER TABLE codemap_semantic ADD COLUMN content_hash TEXT NOT NULL DEFAULT ''`)
	return nil
}

// UpsertCodemapSemantic — simpan/timpa makna 1 file (idempotent by path).
func (s *Store) UpsertCodemapSemantic(c CodemapSemantic) error {
	if err := s.ensureCodemapSemanticSchema(); err != nil {
		return err
	}
	if c.IndexedAt == "" {
		c.IndexedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := s.db.Exec(`
		INSERT INTO codemap_semantic (path, summary, domain, role, model, content_hash, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
		  summary=excluded.summary, domain=excluded.domain, role=excluded.role,
		  model=excluded.model, content_hash=excluded.content_hash, indexed_at=excluded.indexed_at;
	`, c.Path, c.Summary, c.Domain, c.Role, c.Model, c.Hash, c.IndexedAt)
	return err
}

// CodemapSemanticHashes — map path→content_hash (M1: enrich incremental by-hash, bukan by-path).
// File yang BERUBAH (hash beda) bakal di-enrich ulang; yang sama di-skip.
func (s *Store) CodemapSemanticHashes() (map[string]string, error) {
	if err := s.ensureCodemapSemanticSchema(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT path, content_hash FROM codemap_semantic`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var p, h string
		if err := rows.Scan(&p, &h); err != nil {
			return nil, err
		}
		out[p] = h
	}
	return out, rows.Err()
}

// ListCodemapSemantic — semua baris makna (buat GUI viz + R7 konsumsi).
func (s *Store) ListCodemapSemantic() ([]map[string]any, error) {
	if err := s.ensureCodemapSemanticSchema(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT path, summary, domain, role, model, indexed_at FROM codemap_semantic ORDER BY path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var path, summary, domain, role, model, indexedAt string
		if err := rows.Scan(&path, &summary, &domain, &role, &model, &indexedAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"path": path, "summary": summary, "domain": domain,
			"role": role, "model": model, "indexed_at": indexedAt,
		})
	}
	return out, rows.Err()
}

// CodemapSemanticPaths — set path yang UDAH ke-enrich (buat skip incremental).
func (s *Store) CodemapSemanticPaths() (map[string]bool, error) {
	if err := s.ensureCodemapSemanticSchema(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT path FROM codemap_semantic`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out[p] = true
	}
	return out, rows.Err()
}
