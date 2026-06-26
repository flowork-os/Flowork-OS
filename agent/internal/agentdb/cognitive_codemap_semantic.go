// cognitive_codemap_semantic.go — #2: sambungin hasil ENRICH (codemap_semantic) ke code-node CGM.
// LinkCodemapToGraph cuma projeksi STRUKTUR (line/health/layer). Ini nempelin MAKNA enrich
// (summary/domain/role) ke node `<scope>/codemap/<path>` → `Why`=summary (graph_recall nyurfacing
// "file ini ngapain") + domain/role di properties. Non-frozen extension; pakai UpsertNode (frozen).
package agentdb

import "fmt"

// AttachCodemapSemanticToGraph — buat tiap file yg UDAH ke-enrich, re-upsert code-node-nya dgn
// STRUKTUR (dari codemap_files, biar gak ke-timpa kosong) + MAKNA (summary/domain/role). Idempotent.
// Balikin jumlah node yg ke-enrich-graph. scope kosong → "agent:local".
func (s *Store) AttachCodemapSemanticToGraph(scope string) (int, error) {
	if scope == "" {
		scope = "agent:local"
	}
	s.mu.Lock()
	rows, err := s.db.Query(`
		SELECT cs.path, cs.summary, cs.domain, cs.role,
		       COALESCE(cf.line_count,0), COALESCE(cf.health_score,0), COALESCE(cf.layer,'')
		FROM codemap_semantic cs
		LEFT JOIN codemap_files cf ON cf.path = cs.path
		WHERE TRIM(cs.summary) <> ''`)
	if err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("query semantic: %w", err)
	}
	type row struct {
		path, summary, domain, role, layer string
		lineCount, health                  int
	}
	var items []row
	for rows.Next() {
		var r row
		if rows.Scan(&r.path, &r.summary, &r.domain, &r.role, &r.lineCount, &r.health, &r.layer) == nil && r.path != "" {
			items = append(items, r)
		}
	}
	rows.Close()
	s.mu.Unlock()

	n := 0
	for _, r := range items {
		dom := r.domain
		if dom == "" {
			dom = "codebase"
		}
		props := fmt.Sprintf(`{"line_count":%d,"health":%d,"layer":%q,"domain":%q,"role":%q,"enriched":true}`,
			r.lineCount, r.health, r.layer, r.domain, r.role)
		// Why = summary → graph_recall nyurfacing makna file. Type tetap "code" (konsisten LinkCodemap).
		if _, e := s.UpsertNode(CogNode{
			ID: scope + "/codemap/" + r.path, Label: r.path, Type: "code",
			Why: r.summary, WhereDomain: dom, Properties: props,
			SourceKind: "verified", SourceRef: r.path, Confidence: 1.0, Status: "active",
		}); e == nil {
			n++
		}
	}
	return n, nil
}
