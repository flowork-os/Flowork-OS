// graph_proj_mesh.go — proyeksi PENGETAHUAN MESH (federasi) → Cognitive Graph (DreamGraph).
// Owner: Aola Sahidin · github.com/flowork-os/Flowork-OS · floworkos.com
//
// AKAR (audit 2026-06-26): pengetahuan dari peer mesh yg lolos filter 9-lapis cuma di-mark
// status='promoted' di mesh_knowledge_inbox — BERHENTI di situ, GAK nyambung ke Knowledge Graph
// (DreamGraph projeksi knowledge dari `drawers`, bukan dari inbox). Jadi ilmu federasi diterima
// tapi gak ke-graph / gak ke-recall. Ditutup di sini via seam RegisterGraphProjection (zero buka
// frozen — graph_extras.go TIDAK disentuh).
//
// Hasil: tiap pengetahuan mesh 'promoted' jadi node type 'knowledge' source 'mesh_federated' +
// edge member_of→hub 'flowork' → nyambung penuh ke Cognitive Graph + ke-recall lintas-subsistem.
// Idempotent (cleanup source dulu), mirror-only (gak hapus inbox). Switch FLOWORK_DREAMGRAPH_MESH.
package brain

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/flowork-os/flowork_Router/internal/store"
)

func init() {
	RegisterGraphProjection(GraphProjection{
		Name:   "mesh-knowledge",
		Switch: "FLOWORK_DREAMGRAPH_MESH",
		Run:    syncMeshKnowledgeToGraph,
	})
}

// syncMeshKnowledgeToGraph — mirror mesh_knowledge_inbox(status='promoted') → cognitive graph.
// Jalan dalam tx SyncGraphExtended (dipanggil runExtraGraphProjectionsTx). Balikin jumlah node.
func syncMeshKnowledgeToGraph(ctx context.Context, tx *sql.Tx) (int, error) {
	if _, err := tx.ExecContext(ctx, "DELETE FROM cognitive_edges WHERE from_id LIKE 'mesh_know_%'"); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM cognitive_nodes WHERE source = 'mesh_federated'"); err != nil {
		return 0, err
	}
	// mesh_knowledge_inbox ada di STORE DB (data.sqlite), beda dari cognitive_nodes (brain DB).
	// store.Open() = singleton (shared conn) → baca promoted dari sana, tulis node ke brain tx.
	sdb, serr := store.Open()
	if serr != nil {
		return 0, serr
	}
	rows, err := sdb.QueryContext(ctx,
		"SELECT packet_id, origin_pubkey, drawer_content FROM mesh_knowledge_inbox WHERE status = 'promoted'")
	if err != nil {
		return 0, err
	}
	type mk struct{ pid, origin, content string }
	var items []mk
	for rows.Next() {
		var x mk
		if rows.Scan(&x.pid, &x.origin, &x.content) == nil && x.pid != "" {
			items = append(items, x)
		}
	}
	rows.Close()
	n := 0
	for _, it := range items {
		nodeID := "mesh_know_" + it.pid
		origin := it.origin
		if len(origin) > 12 {
			origin = origin[:12]
		}
		props := fmt.Sprintf(`{"kind":"mesh_knowledge","origin":"%s"}`, origin)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO cognitive_nodes (id, label, type, properties, source)
			VALUES (?, ?, 'knowledge', ?, 'mesh_federated')
			ON CONFLICT(id) DO UPDATE SET label=excluded.label, type=excluded.type,
			  properties=excluded.properties, source=excluded.source, last_accessed=CURRENT_TIMESTAMP`,
			nodeID, "Mesh: "+clip(it.content, 60), props); err != nil {
			return n, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO cognitive_edges (from_id, to_id, relation_type, strength)
			VALUES (?, 'flowork', 'member_of', 1.0)
			ON CONFLICT(from_id, to_id, relation_type) DO UPDATE SET strength=1.0`, nodeID); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
