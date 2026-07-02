// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// drawer_visibility_ext.go — SIBLING non-frozen (deletable): FASE B — visibility
// publik-vs-privat per-drawer. Kolom `visibility` di `drawers` (default 'private'):
// knowledge lokal DEFAULT ga boleh bocor ke mesh; cuma yang di-tandai 'public'
// eksplisit yang shareable. Guard `IsShareable` = titik tunggal yang WAJIB dipanggil
// jalur share-ke-mesh manapun (sekarang + masa depan). init.go frozen NOL disentuh:
// kolom ditambah via ALTER idempotent lazy (sync.Once, abaikan "duplicate column").
package brain

import (
	"database/sql"
	"strings"
	"sync"
)

const (
	VisibilityPrivate = "private" // default: ga pernah ke-share ke mesh
	VisibilityPublic  = "public"  // owner tandai eksplisit → boleh di-share
)

var drawerVisOnce sync.Once

// ensureVisibilityColumn — ALTER TABLE drawers ADD COLUMN visibility (idempotent).
// SQLite ga punya ADD COLUMN IF NOT EXISTS → error "duplicate column" pas udah ada
// itu NORMAL, diabaikan. Dipanggil lazy sebelum tiap operasi visibility.
func ensureVisibilityColumn(db *sql.DB) {
	drawerVisOnce.Do(func() {
		_, _ = db.Exec(`ALTER TABLE drawers ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private'`)
	})
}

// normalizeVisibility — cuma 'public' yang di-anggap publik; selain itu → private
// (fail-closed: default aman, ga bocor).
func normalizeVisibility(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), VisibilityPublic) {
		return VisibilityPublic
	}
	return VisibilityPrivate
}

// DrawerVisibility — balik 'public'/'private' buat 1 drawer. Drawer ga ada / kolom
// kosong → 'private' (fail-closed).
func DrawerVisibility(db *sql.DB, id string) string {
	ensureVisibilityColumn(db)
	var v sql.NullString
	err := db.QueryRow(`SELECT visibility FROM drawers WHERE id = ?`, id).Scan(&v)
	if err != nil || !v.Valid {
		return VisibilityPrivate
	}
	return normalizeVisibility(v.String)
}

// SetDrawerVisibility — set visibility 1 drawer (owner action). Balik error kalau
// drawer ga ada.
func SetDrawerVisibility(db *sql.DB, id, visibility string) error {
	ensureVisibilityColumn(db)
	res, err := db.Exec(`UPDATE drawers SET visibility = ? WHERE id = ?`,
		normalizeVisibility(visibility), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// IsShareable — GUARD tunggal: boleh ga drawer ini di-share ke mesh? HANYA 'public'.
// Jalur share-ke-mesh manapun WAJIB lewat sini (fail-closed: default private = tolak).
func IsShareable(db *sql.DB, id string) bool {
	return DrawerVisibility(db, id) == VisibilityPublic
}

// CountVisibility — {public,private} buat observability GUI.
func CountVisibility(db *sql.DB) (public, private int) {
	ensureVisibilityColumn(db)
	rows, err := db.Query(`SELECT COALESCE(visibility,'private'), COUNT(*) FROM drawers WHERE deleted_at IS NULL GROUP BY 1`)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		var n int
		if rows.Scan(&v, &n) == nil {
			if normalizeVisibility(v) == VisibilityPublic {
				public += n
			} else {
				private += n
			}
		}
	}
	return public, private
}
