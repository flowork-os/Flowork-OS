// cognitive_orphan.go — Bagian 4: backfill node ORPHAN → hub "brain-root" (member_of).
// Node projeksi sumber (skill/constitution/edu/knowledge/code) di-`put` sbg NODE tanpa edge →
// orphan (ngambang di viz). Link ke hub biar graph nyambung penuh (pola lock/CognitiveGraph.md FIX-2).
// Non-frozen extension; pakai UpsertNode/UpsertEdge (frozen, divalidasi ValidRelations).
package agentdb

import "fmt"

// BackfillOrphansToHub — pastikan hub brain-root ada, lalu link semua node orphan → hub via
// member_of. Idempotent. Balikin jumlah edge baru. scope kosong → "agent:local".
func (s *Store) BackfillOrphansToHub(scope string) (int, error) {
	if scope == "" {
		scope = "agent:local"
	}
	hubID := scope + "/concept/brain-root"
	if _, err := s.UpsertNode(CogNode{
		ID: hubID, Label: "Brain Root", Type: "concept", WhereDomain: "system",
		SourceKind: "verified", SourceRef: "orphan-backfill", Confidence: 1.0, Status: "active",
	}); err != nil {
		return 0, fmt.Errorf("hub upsert: %w", err)
	}

	// kumpulin id orphan (no edge), kecuali hub sendiri. Query dulu, baru Upsert (hindari lock nested).
	s.mu.Lock()
	rows, err := s.db.Query(`
		SELECT id FROM cognitive_nodes n
		WHERE n.id <> ?
		  AND NOT EXISTS (SELECT 1 FROM cognitive_edges e WHERE e.from_id = n.id OR e.to_id = n.id)`, hubID)
	if err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("query orphan: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil && id != "" {
			ids = append(ids, id)
		}
	}
	rows.Close()
	s.mu.Unlock()

	n := 0
	for _, id := range ids {
		if e := s.UpsertEdge(CogEdge{
			FromID: id, ToID: hubID, RelationType: "member_of", Strength: 0.5,
			Confidence: 1.0, SourceKind: "verified", SourceRef: "orphan-backfill", Status: "active",
		}); e == nil {
			n++
		}
	}
	return n, nil
}
