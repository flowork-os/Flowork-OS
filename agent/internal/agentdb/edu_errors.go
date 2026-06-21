// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"database/sql"
	"fmt"
	"time"
)

type EduError struct {
	Code        string `json:"code"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	Remediation string `json:"remediation"`
	SyncedAt    string `json:"synced_at"`
}

func (s *Store) UpsertEduError(e EduError) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e.Code == "" || e.Category == "" || e.Title == "" || e.Explanation == "" {
		return fmt.Errorf("code + category + title + explanation required")
	}

	const (
		maxText  = 4 * 1024
		maxTitle = 256
	)
	if len(e.Explanation) > maxText {
		e.Explanation = e.Explanation[:maxText] + "…"
	}
	if len(e.Remediation) > maxText {
		e.Remediation = e.Remediation[:maxText] + "…"
	}
	if len(e.Title) > maxTitle {
		e.Title = e.Title[:maxTitle] + "…"
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT INTO educational_errors_cache(code, category, title, explanation, remediation, synced_at)
		 VALUES(?, ?, ?, ?, ?, ?)
		 ON CONFLICT(code) DO UPDATE SET
		     category    = excluded.category,
		     title       = excluded.title,
		     explanation = excluded.explanation,
		     remediation = excluded.remediation,
		     synced_at   = excluded.synced_at,
		     deleted_at  = NULL`,
		e.Code, e.Category, e.Title, e.Explanation, e.Remediation, ts,
	)
	if err != nil {
		return fmt.Errorf("upsert edu error: %w", err)
	}
	return nil
}

func (s *Store) LookupEduError(code string) (EduError, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if code == "" {
		return EduError{}, fmt.Errorf("code required")
	}

	var e EduError
	err := s.db.QueryRow(
		`SELECT code, category, title, explanation, remediation, synced_at
		 FROM educational_errors_cache WHERE code = ? AND deleted_at IS NULL`,
		code,
	).Scan(&e.Code, &e.Category, &e.Title, &e.Explanation, &e.Remediation, &e.SyncedAt)
	if err == sql.ErrNoRows {
		return EduError{Code: code}, nil
	}
	if err != nil {
		return EduError{}, fmt.Errorf("lookup edu error: %w", err)
	}
	return e, nil
}

func (s *Store) ListEduErrors(category string, limit int) ([]EduError, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 || limit > 500 {
		limit = 50
	}

	query := `SELECT code, category, title, explanation, remediation, synced_at
	          FROM educational_errors_cache WHERE deleted_at IS NULL`
	args := []any{}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	query += ` ORDER BY synced_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query edu errors: %w", err)
	}
	defer rows.Close()

	var out []EduError
	for rows.Next() {
		var e EduError
		if err := rows.Scan(&e.Code, &e.Category, &e.Title, &e.Explanation,
			&e.Remediation, &e.SyncedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) CountEduErrors() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int64
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM educational_errors_cache WHERE deleted_at IS NULL`,
	).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
