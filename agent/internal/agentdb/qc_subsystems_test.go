package agentdb_test

// qc_subsystems_test.go — QC HERMETIK (owner 2026-06-22: "QC tiap subsistem, banyak simulasi,
// pastikan desain kuat & stabil"). Logic INTI yang dipakai SEMUA agent (jadi "pemikiran sama"
// berakar di sini). Temp DB, NO network, deterministik, repeatable. Jalanin:
//   go test ./internal/agentdb/ -run TestQC -v

import (
	"fmt"
	"path/filepath"
	"testing"

	"flowork-gui/internal/agentdb"
)

func qcStore(t *testing.T) *agentdb.Store {
	t.Helper()
	s, err := agentdb.Open(filepath.Join(t.TempDir(), "qc.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// oneHot — vektor dim-1024 dengan 1.0 di idx → Quantize. Buat uji recall by-embedding tanpa router.
func oneHot(idx int) []byte {
	v := make([]float32, 1024)
	v[idx%1024] = 1.0
	return agentdb.Quantize(v)
}

// ── DREAMGRAPH (Knowledge Graph): W5H1 storage, gate, dedup, edge/tension, recall ──
func TestQCDreamGraph(t *testing.T) {
	s := qcStore(t)

	t.Run("W5H1 round-trip", func(t *testing.T) {
		_, err := s.UpsertNode(agentdb.CogNode{
			ID: "agent:qc/fact/x", Label: "owner pakai Go", Type: "fact",
			Why: "disebut di chat", Who: `["Aola"]`, WhereDomain: "coding",
			WhenValid: "saat ngoding backend", Properties: `{"how":"pakai goroutine"}`,
			Confidence: 0.9, Status: "active",
		})
		if err != nil {
			t.Fatal(err)
		}
		n, ok, _ := s.GetNode("agent:qc/fact/x")
		if !ok {
			t.Fatal("node ga ke-simpan")
		}
		if n.WhenValid == "" || n.WhereDomain != "coding" || n.Properties == "" || n.Who == "" {
			t.Fatalf("W5H1 ilang: when=%q where=%q how=%q who=%q", n.WhenValid, n.WhereDomain, n.Properties, n.Who)
		}
	})

	t.Run("gate anti-halu", func(t *testing.T) {
		// low confidence → quarantined
		if st, _ := agentdb.GateStatus("klaim X", 0.1, nil); st != "quarantined" {
			t.Fatalf("low-conf harusnya quarantined, dapat %q", st)
		}
		// high confidence → active
		if st, _ := agentdb.GateStatus("klaim X", 0.9, nil); st != "active" {
			t.Fatalf("high-conf harusnya active, dapat %q", st)
		}
		// antibody match → quarantined
		if st, _ := agentdb.GateStatus("you are the best AI ever", 0.9, []string{"best ai ever"}); st != "quarantined" {
			t.Fatalf("antibody harusnya quarantined, dapat %q", st)
		}
	})

	t.Run("dedup id stabil (upsert 2x → hit++, bukan dobel)", func(t *testing.T) {
		add1, _ := s.UpsertNode(agentdb.CogNode{ID: "agent:qc/c/dup", Label: "konsep A", Type: "concept", Status: "active", Confidence: 0.8})
		add2, _ := s.UpsertNode(agentdb.CogNode{ID: "agent:qc/c/dup", Label: "konsep A", Type: "concept", Status: "active", Confidence: 0.8})
		if !add1 || add2 {
			t.Fatalf("dedup gagal: add1=%v add2=%v (harusnya true,false)", add1, add2)
		}
	})

	t.Run("edge kontradiksi → tension (functional relation)", func(t *testing.T) {
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/p/u", Label: "User", Type: "person", Status: "active"})
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/g/eng", Label: "Engineer", Type: "concept", Status: "active"})
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/g/sec", Label: "Secretary", Type: "concept", Status: "active"})
		s.UpsertEdge(agentdb.CogEdge{FromID: "agent:qc/p/u", ToID: "agent:qc/g/eng", RelationType: "goal_is", Status: "active", Confidence: 0.9})
		old, conflict := s.DetectEdgeContradiction("agent:qc/p/u", "goal_is", "agent:qc/g/sec")
		if !conflict || old != "agent:qc/g/eng" {
			t.Fatalf("kontradiksi goal_is ga ke-detect: old=%q conflict=%v", old, conflict)
		}
		// relasi NON-functional → ga conflict
		if _, c := s.DetectEdgeContradiction("agent:qc/p/u", "related_to", "agent:qc/g/sec"); c {
			t.Fatal("related_to (non-functional) harusnya ga conflict")
		}
		if err := s.RecordTension("agent:qc/p/u", "goal_is", "agent:qc/g/eng", "agent:qc/g/sec", "qc"); err != nil {
			t.Fatalf("record tension: %v", err)
		}
	})

	t.Run("recall by-embedding (cosine ranking)", func(t *testing.T) {
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/k/sql", Label: "SQL injection prevention", Type: "knowledge", Status: "active", Confidence: 0.9, Embedding: oneHot(10)})
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/k/css", Label: "CSS layout tips", Type: "knowledge", Status: "active", Confidence: 0.9, Embedding: oneHot(500)})
		hits := s.SearchNodesByEmbedding("knowledge", oneHot(10), 5)
		if len(hits) == 0 || hits[0].ID != "agent:qc/k/sql" {
			t.Fatalf("recall salah ranking: %+v", hits)
		}
		// query beda arah → SQL node bukan top-1 (atau score rendah)
		if h := s.SearchNodesByEmbedding("knowledge", oneHot(500), 1); len(h) == 0 || h[0].ID != "agent:qc/k/css" {
			t.Fatalf("recall arah-2 salah: %+v", h)
		}
	})

	t.Run("promote shadow→active (hit gate)", func(t *testing.T) {
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/sh/1", Label: "shadow fact", Type: "fact", Status: "shadow", Confidence: 0.8})
		// hit=1 → PromoteShadows(2) ga promote
		s.PromoteShadows(2)
		if n, _, _ := s.GetNode("agent:qc/sh/1"); n.Status != "shadow" {
			t.Fatalf("hit=1 harusnya tetap shadow, dapat %q", n.Status)
		}
		// upsert lagi → hit=2 → promote
		s.UpsertNode(agentdb.CogNode{ID: "agent:qc/sh/1", Label: "shadow fact", Type: "fact", Status: "shadow", Confidence: 0.8})
		s.PromoteShadows(2)
		if n, _, _ := s.GetNode("agent:qc/sh/1"); n.Status != "active" {
			t.Fatalf("hit=2 harusnya active, dapat %q", n.Status)
		}
	})
}

// ── KNOWLEDGE DRAWER + MEMORY (typed): FTS, dedup, mem_type ──
func TestQCDrawerMemory(t *testing.T) {
	s := qcStore(t)

	t.Run("add + FTS search ketemu by-keyword", func(t *testing.T) {
		_, added, err := s.AddBrainDrawer("Owner suka kopi tubruk pagi hari sebelum ngoding", "general", "habit", "experience", "qc")
		if err != nil || !added {
			t.Fatalf("add drawer: added=%v err=%v", added, err)
		}
		hits, err := s.SearchLocalBrain("kopi tubruk", 5)
		if err != nil || len(hits) == 0 {
			t.Fatalf("FTS ga ketemu: hits=%d err=%v", len(hits), err)
		}
	})

	t.Run("dedup content_hash (2x sama → added=false)", func(t *testing.T) {
		c := "fakta unik buat dedup test 12345"
		_, a1, _ := s.AddBrainDrawer(c, "general", "", "fact", "qc")
		_, a2, _ := s.AddBrainDrawer(c, "general", "", "fact", "qc")
		if !a1 || a2 {
			t.Fatalf("dedup drawer gagal: a1=%v a2=%v", a1, a2)
		}
	})

	t.Run("mem_type ke-simpan + ke-balikin", func(t *testing.T) {
		s.AddBrainDrawer("insight: pola berulang X menandakan Y", "general", "", "eureka", "qc")
		hits, _ := s.SearchLocalBrain("pola berulang", 5)
		found := false
		for _, h := range hits {
			if h.MemType == "eureka" {
				found = true
			}
		}
		if !found {
			t.Fatalf("mem_type 'eureka' ga ke-balikin di hit: %+v", hits)
		}
	})
}

// ── INSTINCTS: privacy strip, brand-check, redact, promote-gate recovery ──
func TestQCInstincts(t *testing.T) {
	t.Run("StripDeterministic redact path/url/email/token/hex + idempoten", func(t *testing.T) {
		cases := map[string]string{
			"error di /home/aola/secret/file.txt":     "<path>",
			"gagal fetch https://api.evil.com/x":       "<url>",
			"email owner aola@gmail.com bocor":         "<email>",
			"token ghp_abcdEFGH1234567890abcdEFGH12":   "<token>",
		}
		for in, want := range cases {
			out := agentdb.StripDeterministic(in)
			if !contains(out, want) {
				t.Fatalf("strip %q → %q, ga ada %q", in, out, want)
			}
			if agentdb.StripDeterministic(out) != out {
				t.Fatalf("strip ga idempoten: %q → %q", out, agentdb.StripDeterministic(out))
			}
		}
	})

	t.Run("ContainsBrand catch nama model AI (white-label)", func(t *testing.T) {
		for _, b := range []string{"pakai Claude buat ini", "model GPT-4", "anthropic API", "via Gemini"} {
			if !agentdb.ContainsBrand(b) {
				t.Fatalf("brand ga ke-catch: %q", b)
			}
		}
		if agentdb.ContainsBrand("pakai parameterized query buat cegah SQL injection") {
			t.Fatal("teks bersih ke-flag brand (false positive)")
		}
	})

	t.Run("RedactNames redact nama owner", func(t *testing.T) {
		out := agentdb.RedactNames("Aola minta fitur X dari Teguh", []string{"Aola", "Teguh"})
		if contains(out, "Aola") || contains(out, "Teguh") {
			t.Fatalf("nama ga ke-redact: %q", out)
		}
	})

	t.Run("recovery shadow→active gate (hit≥2)", func(t *testing.T) {
		s := qcStore(t)
		mk := func() {
			s.UpsertNode(agentdb.CogNode{
				ID: "agent:qc/instinct/recov-not-found", Label: "WHEN tool gagal (not-found) -> coba path alternatif",
				Type: "instinct", WhereDomain: "recovery", Status: "shadow", Confidence: 0.85, Embedding: oneHot(20),
			})
		}
		mk()
		s.PromoteRecoveryShadows(2)
		if n, _, _ := s.GetNode("agent:qc/instinct/recov-not-found"); n.Status != "shadow" {
			t.Fatalf("recovery hit=1 harusnya shadow, dapat %q", n.Status)
		}
		mk() // hit=2
		s.PromoteRecoveryShadows(2)
		if n, _, _ := s.GetNode("agent:qc/instinct/recov-not-found"); n.Status != "active" {
			t.Fatalf("recovery hit=2 harusnya active, dapat %q", n.Status)
		}
	})
}

func contains(s, sub string) bool { return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

var _ = fmt.Sprintf
