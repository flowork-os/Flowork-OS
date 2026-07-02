// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// provenance_ext.go — SIBLING non-frozen (deletable): FASE M — PROVENANCE per-unit
// knowledge mesh + REVOKE by-origin. Tiap knowledge federasi udah bawa origin_pubkey +
// packet_id di mesh_knowledge_inbox (frozen knowledge.go); di sini gw kasih:
//   (a) query provenance: knowledge dikelompokin per peer asal (trace "ini dari siapa"),
//   (b) REVOKE: kalau 1 peer ternyata jahat → invalidate SEMUA knowledge-nya sekali
//       (status → invalidated) → drop dari graph projection (mesh_federated) sync berikut.
// NOL buka frozen: cuma query/update tabel yang udah ada, fungsi baru di file sibling.
package mesh

import (
	"database/sql"
	"fmt"
	"strings"
)

// PeerProvenance — ringkasan knowledge dari 1 peer asal.
type PeerProvenance struct {
	OriginPubkey string `json:"origin_pubkey"`
	Total        int    `json:"total"`
	Promoted     int    `json:"promoted"`
	Quarantine   int    `json:"quarantine"`
	Dropped      int    `json:"dropped"`
	Invalidated  int    `json:"invalidated"`
}

// ProvenanceByOrigin — kelompokin isi mesh_knowledge_inbox per origin_pubkey dengan
// hitungan per-status. = "peta asal-usul" knowledge federasi (trace + audit).
func ProvenanceByOrigin(db *sql.DB) ([]PeerProvenance, error) {
	rows, err := db.Query(`
		SELECT origin_pubkey,
		       COUNT(*),
		       SUM(CASE WHEN status='promoted'    THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status='quarantine'  THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status='dropped'     THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status='invalidated' THEN 1 ELSE 0 END)
		FROM mesh_knowledge_inbox
		GROUP BY origin_pubkey
		ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PeerProvenance{}
	for rows.Next() {
		var p PeerProvenance
		if err := rows.Scan(&p.OriginPubkey, &p.Total, &p.Promoted,
			&p.Quarantine, &p.Dropped, &p.Invalidated); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// KnowledgeByOrigin — semua entri knowledge dari 1 peer (detail provenance).
func KnowledgeByOrigin(db *sql.DB, originPubkey string, limit int) ([]KnowledgeEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := db.Query(`
		SELECT id, packet_id, origin_pubkey, drawer_content, status, arrived_at
		FROM mesh_knowledge_inbox WHERE origin_pubkey = ?
		ORDER BY id DESC LIMIT ?`, originPubkey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []KnowledgeEntry{}
	for rows.Next() {
		var e KnowledgeEntry
		if err := rows.Scan(&e.ID, &e.PacketID, &e.OriginPubkey,
			&e.DrawerContent, &e.Status, &e.ArrivedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// RevokeByOrigin — REVOKE semua knowledge dari 1 peer: status → invalidated (kecuali
// yang udah dropped). Idempotent. Balik jumlah baris ke-revoke. Efek: sync graph
// berikut (graph_proj_mesh baca status='promoted') otomatis buang node peer ini.
// Opsional +penalti karma biar peer jahat turun reputasi (dipanggil handler).
func RevokeByOrigin(db *sql.DB, originPubkey string) (int64, error) {
	if strings.TrimSpace(originPubkey) == "" {
		return 0, fmt.Errorf("origin_pubkey required")
	}
	res, err := db.Exec(
		`UPDATE mesh_knowledge_inbox SET status = ?
		 WHERE origin_pubkey = ? AND status != ?`,
		StatusInvalidated, originPubkey, StatusDropped)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
