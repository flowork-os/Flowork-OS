// cognitive_deadletter.go — #4: DEAD-LETTER task → CGM. Task background yang GAGAL permanen
// (agent_runs.state='error') = dead-letter. Projeksi ke graph (type 'dead_letter', nyambung hub
// brain-root) → agent SADAR kegagalannya, bisa `graph_recall` + belajar (mirip edu_error).
// Non-frozen extension; pakai UpsertNode/UpsertEdge (frozen). MIRROR-only (gak ngubah agent_runs).
package agentdb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SyncDeadLettersToGraph — projeksi agent_runs state='error' → node 'dead_letter' + edge member_of
// → brain-root. Idempotent. limit = ambil terbaru (default 100 kalau <=0). scope kosong→agent:local.
func (s *Store) SyncDeadLettersToGraph(scope string, limit int) (int, error) {
	if scope == "" {
		scope = "agent:local"
	}
	if limit <= 0 {
		limit = 100
	}
	hubID := scope + "/concept/brain-root"
	_, _ = s.UpsertNode(CogNode{ID: hubID, Label: "Brain Root", Type: "concept", WhereDomain: "system",
		SourceKind: "verified", SourceRef: "deadletter", Confidence: 1.0, Status: "active"})

	s.mu.Lock()
	rows, err := s.db.Query(`SELECT id, COALESCE(label,''), COALESCE(checkpoint,'') FROM agent_runs
		WHERE state='error' ORDER BY updated DESC LIMIT ?`, limit)
	if err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("query dead-letter: %w", err)
	}
	type dl struct{ id, label, checkpoint string }
	var items []dl
	for rows.Next() {
		var d dl
		if rows.Scan(&d.id, &d.label, &d.checkpoint) == nil && d.id != "" {
			items = append(items, d)
		}
	}
	rows.Close()
	s.mu.Unlock()

	n := 0
	for _, d := range items {
		// ambil "output" (pesan error) dari checkpoint JSON; fallback raw.
		why := d.checkpoint
		var meta map[string]any
		if json.Unmarshal([]byte(d.checkpoint), &meta) == nil {
			if o, ok := meta["output"].(string); ok && strings.TrimSpace(o) != "" {
				why = o
			}
		}
		if len(why) > 400 {
			why = why[:400]
		}
		label := strings.TrimSpace(d.label)
		if label == "" {
			label = d.id
		}
		nodeID := scope + "/dead_letter/" + d.id
		if _, e := s.UpsertNode(CogNode{
			ID: nodeID, Label: "DeadLetter: " + label, Type: "dead_letter", Why: why,
			WhereDomain: "task", Properties: `{"kind":"failed_task"}`,
			SourceKind: "verified", SourceRef: d.id, Confidence: 1.0, Status: "active",
		}); e != nil {
			continue
		}
		if s.UpsertEdge(CogEdge{FromID: nodeID, ToID: hubID, RelationType: "member_of",
			Strength: 0.5, Confidence: 1.0, SourceKind: "verified", SourceRef: "deadletter", Status: "active"}) == nil {
			n++
		}
	}
	return n, nil
}
