// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// tool_consent_ext.go — SIBLING non-frozen (deletable): FASE M — CONSENT GATE tool
// manifest mesh. Dulu manifest peer disimpen otomatis (inert/discovery-only) TAPI
// nongol di discovery tanpa restu owner. Sekarang: kolom `consented` (default 0,
// migrasi additive di store) = gerbang. Manifest asing = PENDING sampai owner approve
// di GUI. Discovery-consented cuma nampilin yang di-restui. NOL buka frozen: kolom via
// RegisterMigration, query/update di file sibling ini.
package mesh

import (
	"database/sql"
	"fmt"
	"strings"
)

// ToolManifestRow — 1 manifest tool dari peer (buat antrian consent + discovery).
type ToolManifestRow struct {
	ToolName     string `json:"tool_name"`
	OriginPubkey string `json:"origin_pubkey"`
	ManifestJSON string `json:"manifest_json"`
	Signature    string `json:"signature"`
	ArrivedAt    string `json:"arrived_at"`
	Consented    bool   `json:"consented"`
}

// PendingToolManifests — manifest yang BELUM di-restui owner (consented=0). Antrian
// consent buat GUI.
func PendingToolManifests(db *sql.DB, limit int) ([]ToolManifestRow, error) {
	return listToolManifests(db, "COALESCE(consented,0) = 0", limit)
}

// ConsentedToolManifests — manifest yang UDAH di-restui owner (discovery aman).
func ConsentedToolManifests(db *sql.DB, limit int) ([]ToolManifestRow, error) {
	return listToolManifests(db, "COALESCE(consented,0) = 1", limit)
}

func listToolManifests(db *sql.DB, where string, limit int) ([]ToolManifestRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query(`
		SELECT tool_name, origin_pubkey, manifest_json, signature, arrived_at, COALESCE(consented,0)
		FROM mesh_tool_manifests WHERE `+where+`
		ORDER BY arrived_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ToolManifestRow{}
	for rows.Next() {
		var r ToolManifestRow
		var c int
		if err := rows.Scan(&r.ToolName, &r.OriginPubkey, &r.ManifestJSON,
			&r.Signature, &r.ArrivedAt, &c); err != nil {
			return nil, err
		}
		r.Consented = c == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetToolConsent — owner approve/tolak 1 manifest (by tool_name + origin_pubkey).
// approve=true → consented=1 (boleh muncul di discovery). false → consented=0
// (balik pending). Balik error kalau manifest ga ketemu.
func SetToolConsent(db *sql.DB, toolName, originPubkey string, approve bool) error {
	if strings.TrimSpace(toolName) == "" || strings.TrimSpace(originPubkey) == "" {
		return fmt.Errorf("tool_name + origin_pubkey required")
	}
	v := 0
	if approve {
		v = 1
	}
	res, err := db.Exec(
		`UPDATE mesh_tool_manifests SET consented = ? WHERE tool_name = ? AND origin_pubkey = ?`,
		v, toolName, originPubkey)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("manifest %q dari %q ga ketemu", toolName, originPubkey)
	}
	return nil
}
