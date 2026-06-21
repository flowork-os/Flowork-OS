// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"fmt"
	"strings"
)

func (s *Store) SearchMistakes(query string, limit int) ([]Mistake, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	var toks []string
	for _, f := range strings.Fields(strings.ToLower(query)) {
		f = strings.Trim(f, ".,:;!?()[]{}\"'`")
		if len(f) >= 3 {
			toks = append(toks, f)
		}
	}
	if len(toks) == 0 {
		return []Mistake{}, nil
	}

	var conds []string
	var args []any
	for _, t := range toks {
		conds = append(conds, "(lower(title) LIKE ? OR lower(content) LIKE ?)")
		like := "%" + t + "%"
		args = append(args, like, like)
	}
	q := `SELECT id, category, title, content, context_origin, tier, hit_count,
	             last_hit_at, created_at
	        FROM mistakes_local
	       WHERE deleted_at IS NULL AND (` + strings.Join(conds, " OR ") + `)
	       ORDER BY hit_count DESC, last_hit_at DESC
	       LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("search mistakes: %w", err)
	}
	defer rows.Close()

	out := []Mistake{}
	for rows.Next() {
		var m Mistake
		if err := rows.Scan(&m.ID, &m.Category, &m.Title, &m.Content, &m.ContextOrigin,
			&m.Tier, &m.HitCount, &m.LastHitAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
