package agentdb

import (
	"path/filepath"
	"testing"
)

// Co-reference: alias identitas → canonical, deterministik (fix fragmentasi identitas).
// Fixture GENERIC (no real identity) — privasi: data owner asli gak masuk repo.
func TestCoreferenceIdentityResolve(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	scope := "agent:t"
	canon := scope + "/person/owner"
	if _, err := s.UpsertNode(CogNode{ID: canon, Label: "Owner Test", Type: "person", Status: "active", SourceKind: "user_said", Confidence: 1.0}); err != nil {
		t.Fatalf("seed canonical: %v", err)
	}
	for _, a := range []string{"User", "Nick", "saya", "gw"} {
		if err := s.RegisterIdentityAlias(scope, a, canon); err != nil {
			t.Fatalf("register %s: %v", a, err)
		}
	}

	// alias → canonical (termasuk normalize case + whitespace)
	for _, in := range []string{"User", "user", "  NICK ", "saya", "gw"} {
		if id, ok := s.resolveCanonicalIdentity(scope, in, "person"); !ok || id != canon {
			t.Fatalf("resolve %q = (%q,%v), want (%q,true)", in, id, ok, canon)
		}
	}
	// non-alias person → JANGAN merge (tetep node sendiri)
	if id, ok := s.resolveCanonicalIdentity(scope, "Other Person", "person"); ok {
		t.Fatalf("non-alias resolved to %q, want false", id)
	}
	// type bukan person → skip (jangan tarik concept "User" ke owner)
	if _, ok := s.resolveCanonicalIdentity(scope, "User", "concept"); ok {
		t.Fatalf("concept resolved, want false (person-only)")
	}
	// scope beda → gak nyambung (isolasi antar-agent)
	if _, ok := s.resolveCanonicalIdentity("agent:other", "User", "person"); ok {
		t.Fatalf("cross-scope resolved, want false")
	}
	// canonical mati (obsolete) → jangan resolve ke node mati
	s.mu.Lock()
	_, _ = s.db.Exec(`UPDATE cognitive_nodes SET status='obsolete' WHERE id=?`, canon)
	s.mu.Unlock()
	if _, ok := s.resolveCanonicalIdentity(scope, "User", "person"); ok {
		t.Fatalf("resolved to obsolete canonical, want false")
	}
}
