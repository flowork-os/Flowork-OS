// === LOCKED FILE ===
// Status: STABLE — brain/intelligence subsystem, audited working 2026-06-07. Corpus wing/room enumeration over the 5M-drawer brain (powers /api/brain/wing).
// Do not edit without owner approval.

// wing_enum.go — enumerate drawer per WING (corpus kategori: exploitdb, hackerone,
// red_team, ...). READ-ONLY + paginated (offset) buat nyisir corpus gede
// (exploitdb 44.955) jadi sumber topik distilasi scanner. File baru — views.go LOCKED.

package brain

import "context"

// ListByWing — drawer dari satu wing, paginated. Read-only. limit<=0 → 100, cap 500.
// roomLike (opsional, mis. "%webapps%") → filter room SQL biar fetch cuma kategori
// relevan (anti boros nyisir local/dos). Urut importance (resumable lewat offset).
func ListByWing(ctx context.Context, wing, roomLike string, limit, offset, maxContentLen int) ([]Snippet, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	db, err := Open()
	if err != nil {
		return nil, err
	}
	query := `SELECT id, wing, room, content FROM drawers WHERE deleted_at IS NULL AND wing = ? `
	args := []any{wing}
	if roomLike != "" {
		query += `AND room LIKE ? `
		args = append(args, roomLike)
	}
	query += `ORDER BY importance DESC, filed_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Snippet
	for rows.Next() {
		var s Snippet
		if err := rows.Scan(&s.DrawerID, &s.Wing, &s.Room, &s.Content); err != nil {
			continue
		}
		if maxContentLen > 0 {
			s.Content = truncateRunes(s.Content, maxContentLen)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
