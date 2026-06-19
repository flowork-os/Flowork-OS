package agentdb

import "testing"

func TestTemporal_SupersedeAndTimeline(t *testing.T) {
	s := openTestStore(t)
	_, _ = s.UpsertNode(CogNode{ID: "a/p/aola", Label: "Aola", Type: "person", Status: "active"})
	_, _ = s.UpsertNode(CogNode{ID: "a/c/x", Label: "X", Type: "concept", Status: "active"})
	_, _ = s.UpsertNode(CogNode{ID: "a/c/y", Label: "Y", Type: "concept", Status: "active"})

	// Aola prefers X (active, open-ended)
	if err := s.UpsertEdge(CogEdge{FromID: "a/p/aola", ToID: "a/c/x", RelationType: "prefers", Status: "active"}); err != nil {
		t.Fatal(err)
	}

	const cut = "2026-06-19T12:00:00Z"
	// fact changed: stop being valid at `cut`
	ok, err := s.SupersedeEdge("a/p/aola", "a/c/x", "prefers", cut)
	if err != nil || !ok {
		t.Fatalf("supersede: ok=%v err=%v", ok, err)
	}
	// now Aola prefers Y (new active edge)
	_ = s.UpsertEdge(CogEdge{FromID: "a/p/aola", ToID: "a/c/y", RelationType: "prefers", Status: "active"})

	// the old edge is obsolete → excluded from current recall (status filter)
	out, _, _ := s.Neighbors("a/p/aola")
	for _, e := range out {
		if e.ToID == "a/c/x" {
			t.Fatal("superseded edge should not appear in active Neighbors")
		}
	}

	// timeline: BEFORE cut, Aola->X was valid
	before, err := s.EdgesValidAt("2026-06-19T06:00:00Z", 100)
	if err != nil {
		t.Fatal(err)
	}
	if !hasEdge(before, "a/c/x") {
		t.Fatalf("before cut, X should be valid: %+v", before)
	}
	// timeline: AFTER cut, Aola->X no longer valid; Y is
	after, _ := s.EdgesValidAt("2026-06-19T18:00:00Z", 100)
	if hasEdge(after, "a/c/x") {
		t.Fatal("after cut, X should NOT be valid")
	}
	if !hasEdge(after, "a/c/y") {
		t.Fatal("after cut, Y should be valid (open-ended)")
	}
}

func hasEdge(es []GraphEdgeView, toID string) bool {
	for _, e := range es {
		if e.ToID == toID {
			return true
		}
	}
	return false
}
