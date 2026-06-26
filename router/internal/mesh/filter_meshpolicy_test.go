package mesh

import (
	"database/sql"
	"testing"
)

// TestMeshSharePolicy buktiin: share OFF → reject (reciprocity); share ON + approve manual →
// flag (quarantine); share ON + approve auto → pass.
func TestMeshSharePolicy(t *testing.T) {
	var run func(*sql.DB, Packet, string) FilterDecision
	for _, f := range extraMeshFilters {
		if f.Name == "share-policy" {
			run = f.Run
		}
	}
	if run == nil {
		t.Fatal("filter share-policy harus terdaftar")
	}
	t.Setenv("FLOWORK_MESH_SHARE", "off")
	if d := run(nil, Packet{}, ""); d.Decision != "reject" {
		t.Fatalf("share OFF harus reject, dapat %q", d.Decision)
	}
	t.Setenv("FLOWORK_MESH_SHARE", "true")
	t.Setenv("FLOWORK_MESH_APPROVE", "manual")
	if d := run(nil, Packet{}, ""); d.Decision != "flag" {
		t.Fatalf("approve manual harus flag (quarantine), dapat %q", d.Decision)
	}
	t.Setenv("FLOWORK_MESH_APPROVE", "auto")
	if d := run(nil, Packet{}, ""); d.Decision != "pass" {
		t.Fatalf("approve auto harus pass, dapat %q", d.Decision)
	}
}
