package agentdb

import "testing"

func TestSelfHeal_DecayPruneOrphan(t *testing.T) {
	s := openTestStore(t)

	// identity node (protected) — must survive even when orphaned
	_, _ = s.UpsertNode(CogNode{ID: "a/person/aola", Label: "Aola", Type: "person", Status: "active"})
	// concept nodes + a weak edge that will be pruned
	_, _ = s.UpsertNode(CogNode{ID: "a/c/b", Label: "B", Type: "concept", Status: "active"})
	_, _ = s.UpsertNode(CogNode{ID: "a/c/orphan", Label: "Orphan", Type: "concept", Status: "active"})
	// weak edge a->b (strength below floor after we set it low)
	_ = s.UpsertEdge(CogEdge{FromID: "a/person/aola", ToID: "a/c/b", RelationType: "related_to", Strength: 0.05, Status: "active"})

	st, err := s.SelfHeal(SelfHealOpts{MinStrength: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if st.EdgesPruned < 1 {
		t.Fatalf("expected weak edge pruned, stats=%+v", st)
	}

	// B was only connected by the pruned edge → orphan concept → removed.
	// orphan concept → removed. Aola (person, protected) → stays even though orphaned.
	if _, ok, _ := s.GetNode("a/person/aola"); !ok {
		t.Fatal("protected identity node (person) must NOT be removed when orphaned")
	}
	if _, ok, _ := s.GetNode("a/c/orphan"); ok {
		t.Fatal("orphan concept node should be removed")
	}
	if _, ok, _ := s.GetNode("a/c/b"); ok {
		t.Fatal("concept node orphaned after edge prune should be removed")
	}
}

func TestSelfHeal_Decay(t *testing.T) {
	s := openTestStore(t)
	_, _ = s.UpsertNode(CogNode{ID: "a/c/x", Label: "X", Type: "concept", Status: "active"})
	_, _ = s.UpsertNode(CogNode{ID: "a/c/y", Label: "Y", Type: "concept", Status: "active"})
	_ = s.UpsertEdge(CogEdge{FromID: "a/c/x", ToID: "a/c/y", RelationType: "related_to", Strength: 1.0, Status: "active"})

	if _, err := s.SelfHeal(SelfHealOpts{DecayFactor: 0.9, MinStrength: 0.05}); err != nil {
		t.Fatal(err)
	}
	out, _, _ := s.Neighbors("a/c/x")
	if len(out) != 1 || out[0].Strength > 0.95 || out[0].Strength < 0.85 {
		t.Fatalf("strength after 0.9 decay = %v, want ~0.9", out)
	}
}
