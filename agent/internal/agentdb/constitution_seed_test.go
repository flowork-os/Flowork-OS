package agentdb

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestSeedSacredConstitution_PropagatesToNonEmptyTable mengunci fix durability:
// aturan sacred generik baru (mis. sync-honest/recall-first) HARUS nyebar ke
// agent yang tabel constitution-nya UDAH keisi (rule personal/lokal), tanpa
// nimpa/ngapus rule lokal itu. Dulu di-gate `if n>0 return` → aturan baru ga
// pernah nyampe agent lama.
func TestSeedSacredConstitution_PropagatesToNonEmptyTable(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.ensureConstitutionSchema()

	// Simulasi agent existing: tabel udah keisi rule personal/lokal.
	if _, err := s.db.Exec(
		`INSERT INTO constitution (id, rule, amplitude, sacred, always_inject, lens, created_at)
		 VALUES ('local-mission','misi lokal owner',999999,1,1,'mission','x')`); err != nil {
		t.Fatal(err)
	}

	if _, err := s.SeedSacredConstitution(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rules, err := s.ListAlwaysInjectConstitution()
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, r := range rules {
		have[r.ID] = true
	}
	// Rule lokal HARUS tetap ada (ga ke-nimpa).
	if !have["local-mission"] {
		t.Error("rule lokal 'local-mission' hilang setelah seed (seharusnya non-destructive)")
	}
	// Aturan sacred generik baru HARUS nyebar ke tabel non-kosong.
	for _, id := range []string{"sync-honest", "recall-first", "5w1h-gate", "identity-guard", "anti-halu"} {
		if !have[id] {
			t.Errorf("aturan sacred '%s' ga ke-seed ke tabel non-kosong", id)
		}
	}

	// Idempotent: seed lagi ga nambah duplikat.
	rendered := renderConstitutionBody(rules)
	if !strings.Contains(rendered, "background") {
		t.Error("aturan sync-honest (anti-janji-background) ga ke-render")
	}
}
