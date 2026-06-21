// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import "fmt"

type codemapFileRow struct {
	path, name, layer string
	lineCount, health int
}

func (s *Store) LinkCodemapToGraph(scope string) (int, int, error) {
	if scope == "" {
		scope = "agent:local"
	}

	s.mu.Lock()
	s.ensureCognitiveGraphSchema()
	var files []codemapFileRow
	rows, err := s.db.Query(`SELECT path, name, layer, line_count, health_score FROM codemap_files`)
	if err != nil {
		s.mu.Unlock()
		return 0, 0, fmt.Errorf("read codemap_files: %w", err)
	}
	for rows.Next() {
		var f codemapFileRow
		if rows.Scan(&f.path, &f.name, &f.layer, &f.lineCount, &f.health) == nil {
			files = append(files, f)
		}
	}
	rows.Close()

	type fe struct{ from, to string }
	var fedges []fe
	er2, err := s.db.Query(`SELECT from_path, to_path FROM codemap_file_edges`)
	if err == nil {
		for er2.Next() {
			var e fe
			if er2.Scan(&e.from, &e.to) == nil {
				fedges = append(fedges, e)
			}
		}
		er2.Close()
	}
	s.mu.Unlock()

	if len(files) == 0 {
		return 0, 0, fmt.Errorf("codemap kosong — index dulu (codemap belum jalan?)")
	}

	codeID := func(path string) string { return scope + "/codemap/" + path }
	layerID := func(layer string) string { return scope + "/code_layer/" + slug(layer) }

	nodesAdded, edgesAdded := 0, 0
	layersSeen := map[string]bool{}

	for _, f := range files {
		label := f.path
		props := fmt.Sprintf(`{"line_count":%d,"health":%d,"layer":%q}`, f.lineCount, f.health, f.layer)
		added, _ := s.UpsertNode(CogNode{
			ID: codeID(f.path), Label: label, Type: "code",
			WhereDomain: "codebase", Properties: props,
			SourceKind: "verified", SourceRef: f.path, Confidence: 1.0, Status: "active",
		})
		if added {
			nodesAdded++
		}
		if f.layer != "" {
			lid := layerID(f.layer)
			if !layersSeen[f.layer] {
				layersSeen[f.layer] = true
				if a, _ := s.UpsertNode(CogNode{
					ID: lid, Label: f.layer, Type: "code_layer", WhereDomain: "codebase",
					SourceKind: "verified", SourceRef: "codemap:layer", Confidence: 1.0, Status: "active",
				}); a {
					nodesAdded++
				}
			}
			if s.UpsertEdge(CogEdge{
				FromID: codeID(f.path), ToID: lid, RelationType: "part_of",
				Confidence: 1.0, SourceKind: "verified", SourceRef: "codemap", Status: "active",
			}) == nil {
				edgesAdded++
			}
		}
	}

	for _, e := range fedges {
		if e.from == "" || e.to == "" || e.from == e.to {
			continue
		}
		if s.UpsertEdge(CogEdge{
			FromID: codeID(e.from), ToID: codeID(e.to), RelationType: "depends_on",
			Confidence: 1.0, SourceKind: "verified", SourceRef: "codemap", Status: "active",
		}) == nil {
			edgesAdded++
		}
	}
	return nodesAdded, edgesAdded, nil
}
