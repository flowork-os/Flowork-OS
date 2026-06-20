package agentdb

import (
	"path/filepath"
	"testing"
)

// TestListSelfPromptSlots_LatestPerSlot mengunci fix bug SQL render self-prompt.
//
// Bug lama (antipattern): `WHERE version IN (SELECT MAX(version) ... GROUP BY slot)`
// — subquery balikin HIMPUNAN max-version lintas-slot (mis. {4,1}); lalu
// `version IN (4,1)` salah match baris versi-rendah dari slot LAIN yang
// kebetulan punya angka versi sama (00_constitution v1 ke-ambil gara-gara
// 01_twin max-nya 1). Akibatnya render bisa milih versi LAMA → misi/aturan
// sacred hilang dari prompt.
//
// Fix: korelasi per-slot `WHERE version = (SELECT MAX(version) WHERE slot=sp.slot)`.
// Test ini memastikan tiap slot mengembalikan TEPAT versi tertingginya walau
// angka versi tumpang-tindih antar-slot.
func TestListSelfPromptSlots_LatestPerSlot(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// 00_constitution: v1..v4 (versi-rendah v1 jadi jebakan bug).
	for _, body := range []string{"const-v1", "const-v2", "const-v3", "const-v4"} {
		if _, err := s.SetSelfPrompt("00_constitution", body, "", 0); err != nil {
			t.Fatalf("set constitution: %v", err)
		}
	}
	// 01_twin: cuma v1 — max-nya (1) yang dulu meracuni himpunan IN.
	if _, err := s.SetSelfPrompt("01_twin", "twin-v1", "", 0); err != nil {
		t.Fatalf("set twin: %v", err)
	}

	slots, err := s.ListSelfPromptSlots()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	got := map[string]SelfPrompt{}
	for _, sp := range slots {
		if prev, dup := got[sp.Slot]; dup {
			t.Fatalf("slot %q muncul >1x (v%d & v%d) — list harus 1 baris/slot",
				sp.Slot, prev.Version, sp.Version)
		}
		got[sp.Slot] = sp
	}

	if len(got) != 2 {
		t.Fatalf("harap 2 slot unik, dapat %d", len(got))
	}
	if c := got["00_constitution"]; c.Version != 4 || c.Body != "const-v4" {
		t.Errorf("00_constitution: harap v4/const-v4, dapat v%d/%q (bug: ke-ambil versi lama)", c.Version, c.Body)
	}
	if tw := got["01_twin"]; tw.Version != 1 || tw.Body != "twin-v1" {
		t.Errorf("01_twin: harap v1/twin-v1, dapat v%d/%q", tw.Version, tw.Body)
	}
}
