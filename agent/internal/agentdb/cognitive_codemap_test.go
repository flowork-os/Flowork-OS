package agentdb

import (
	"path/filepath"
	"testing"
)

// seed codemap_files + edges lalu LinkCodemapToGraph → code nodes + struktur edge.
func TestLinkCodemapToGraph(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// codemap tables dibuat oleh codemap subsystem; seed manual buat test.
	s.mu.Lock()
	s.db.Exec(`CREATE TABLE IF NOT EXISTS codemap_files (path TEXT PRIMARY KEY, name TEXT, file_type TEXT, line_count INT, layer TEXT, has_tests INT, has_docs INT, health_score INT, recently_touched INT, issues_json TEXT, indexed_at TEXT)`)
	s.db.Exec(`CREATE TABLE IF NOT EXISTS codemap_file_edges (from_path TEXT, to_path TEXT, PRIMARY KEY(from_path,to_path))`)
	s.db.Exec(`INSERT INTO codemap_files (path,name,layer,line_count,health_score) VALUES
		('internal/a.go','a.go','internal',100,90),
		('internal/b.go','b.go','internal',50,80),
		('main.go','main.go','root',200,70)`)
	s.db.Exec(`INSERT INTO codemap_file_edges (from_path,to_path) VALUES ('main.go','internal/a.go'),('internal/a.go','internal/b.go')`)
	s.mu.Unlock()

	n, e, err := s.LinkCodemapToGraph("agent:test")
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	// 3 file + 2 layer (internal, root) = 5 nodes
	if n != 5 {
		t.Fatalf("nodes added = %d, want 5", n)
	}
	// 2 depends_on + 3 part_of = 5 edges
	if e != 5 {
		t.Fatalf("edges added = %d, want 5", e)
	}

	// node code ada + type bener
	node, ok, _ := s.GetNode("agent:test/codemap/main.go")
	if !ok || node.Type != "code" {
		t.Fatalf("code node main.go missing/wrong type: %+v", node)
	}

	// traversal: main.go --depends_on--> internal/a.go (lokasi tubuh)
	out, _, err := s.Neighbors("agent:test/codemap/main.go")
	if err != nil {
		t.Fatal(err)
	}
	foundDep, foundLayer := false, false
	for _, ed := range out {
		if ed.RelationType == "depends_on" && ed.ToID == "agent:test/codemap/internal/a.go" {
			foundDep = true
		}
		if ed.RelationType == "part_of" {
			foundLayer = true
		}
	}
	if !foundDep || !foundLayer {
		t.Fatalf("traversal missing: dep=%v layer=%v", foundDep, foundLayer)
	}

	// idempoten: jalan lagi → 0 node baru
	n2, _, _ := s.LinkCodemapToGraph("agent:test")
	if n2 != 0 {
		t.Fatalf("re-run added %d nodes, want 0 (idempotent)", n2)
	}
}
