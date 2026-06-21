// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"fmt"
	"time"
)

func (s *Store) SetMistakePromoted(id int64, promotedToID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return fmt.Errorf("mistake id required")
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	promotedStr := fmt.Sprintf("%d", promotedToID)

	_, err := s.db.Exec(
		`UPDATE mistakes_local SET
		    tier           = 'promoted',
		    promoted_at    = ?,
		    promoted_to_id = ?
		 WHERE id = ? AND tier != 'promoted' AND deleted_at IS NULL`,
		ts, promotedStr, id,
	)
	if err != nil {
		return fmt.Errorf("set mistake promoted: %w", err)
	}
	return nil
}

func (s *Store) ListMistakesEligibleForPromote(minHitCount int64, limit int) ([]Mistake, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if minHitCount < 1 {
		minHitCount = 3
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.db.Query(
		`SELECT id, category, title, content, context_origin, tier, hit_count,
		        last_hit_at, created_at,
		        COALESCE(promoted_at, ''), COALESCE(promoted_to_id, '')
		 FROM mistakes_local
		 WHERE deleted_at IS NULL
		   AND tier = 'raw'
		   AND hit_count >= ?
		   AND (promoted_to_id IS NULL OR promoted_to_id = '')
		 ORDER BY hit_count DESC, last_hit_at DESC
		 LIMIT ?`,
		minHitCount, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query eligible: %w", err)
	}
	defer rows.Close()

	var out []Mistake
	for rows.Next() {
		var m Mistake
		if err := rows.Scan(&m.ID, &m.Category, &m.Title, &m.Content,
			&m.ContextOrigin, &m.Tier, &m.HitCount,
			&m.LastHitAt, &m.CreatedAt, &m.PromotedAt, &m.PromotedToID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
