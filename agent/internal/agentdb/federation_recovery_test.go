package agentdb

import "testing"

func recovInstinct(id, label, domain, srcKind, status string) CogNode {
	return CogNode{
		ID: id, Label: label, Type: "instinct", WhereDomain: domain,
		SourceKind: srcKind, Status: status, Confidence: 0.85,
	}
}

func TestSelectPromotableRecoveryInstincts(t *testing.T) {
	s := openTestStore(t)

	// Layak: recovery + verified + active.
	if _, err := s.UpsertNode(recovInstinct("agent:test/instinct/recov-not-found",
		"WHEN resource not-found -> retry corrected target", "recovery", "verified", "active")); err != nil {
		t.Fatal(err)
	}
	// TIDAK layak (status shadow).
	_, _ = s.UpsertNode(recovInstinct("agent:test/instinct/recov-timeout",
		"WHEN timeout -> retry backoff", "recovery", "verified", "shadow"))
	// TIDAK layak (bukan domain recovery).
	_, _ = s.UpsertNode(recovInstinct("agent:test/instinct/coding-1",
		"WHEN x -> y", "coding", "verified", "active"))
	// TIDAK layak (bukan verified).
	_, _ = s.UpsertNode(recovInstinct("agent:test/instinct/recov-perm",
		"WHEN permission -> adjust", "recovery", "agent_inferred", "active"))

	got, err := s.SelectPromotableRecoveryInstincts(50)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("eligible=%d want 1 (cuma yg recovery+verified+active)", len(got))
	}
	if got[0].ID != "agent:test/instinct/recov-not-found" {
		t.Fatalf("salah node ke-pilih: %s", got[0].ID)
	}

	// Anti-double: tandai promoted → ga ke-pilih lagi.
	if err := s.MarkPromotedCognitive("node:"+got[0].ID, "remote123", "ok"); err != nil {
		t.Fatalf("mark: %v", err)
	}
	got2, err := s.SelectPromotableRecoveryInstincts(50)
	if err != nil {
		t.Fatal(err)
	}
	if len(got2) != 0 {
		t.Fatalf("setelah promoted harus 0, dapat %d (anti-double bocor)", len(got2))
	}
}

// Gerbang privasi (defense in depth): label dgn path/brand HARUS ke-deteksi ga-bersih.
// (Logika dipakai host PromoteRecoveryInstinctsShared sebelum push.)
func TestRecoveryShareValidityGate(t *testing.T) {
	clean := "WHEN resource not-found -> retry corrected target"
	if StripDeterministic(clean) != clean || ContainsBrand(clean) {
		t.Fatalf("label bersih kok dianggap ga-aman: %q", clean)
	}
	leaky := []string{
		"WHEN file /home/user1/x.go not-found -> retry",  // path
		"WHEN claude api fails -> retry",                 // brand
		"WHEN token ghp_AbCd1234EfGh5678IjKl -> refresh", // token
	}
	for _, l := range leaky {
		if StripDeterministic(l) == l && !ContainsBrand(l) {
			t.Errorf("label ga-aman lolos gerbang: %q", l)
		}
	}
}
