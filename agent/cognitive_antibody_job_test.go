package main

import (
	"testing"

	"flowork-gui/internal/agentdb"
)

func TestRecoveryClassOf(t *testing.T) {
	cases := map[string]string{
		"agent:mr-flow/instinct/recov-not-found": "not-found",
		"agent:worker/instinct/recov-timeout":    "timeout",
		"agent:x/instinct/recov0a1b2c3d":         "", // fallback hash-id (no class) → skip
		"agent:x/concept/foo":                    "", // bukan recovery
		"agent:x/instinct/recov-bad/extra":       "", // ada '/' → skip
	}
	for id, want := range cases {
		if got := agentdb.RecoveryClassOf(id); got != want {
			t.Errorf("RecoveryClassOf(%q)=%q want %q", id, got, want)
		}
	}
}

// F4 inti: kelas yg dipunyai >=N agent DISTINCT → collective; pilih label confidence tertinggi.
func TestComputeCollectiveAntibodies(t *testing.T) {
	perAgent := map[string][]agentdb.ClassLabel{
		"a": {{Class: "not-found", Label: "retry corrected target", Confidence: 0.85}, {Class: "timeout", Label: "backoff", Confidence: 0.85}},
		"b": {{Class: "not-found", Label: "verify new location then retry", Confidence: 0.9}}, // not-found jg di b → konvergen
		"c": {{Class: "permission", Label: "adjust perms", Confidence: 0.85}},
	}
	plans := computeCollectiveAntibodies(perAgent, 2)
	// cuma not-found yg di >=2 agent (a+b). timeout (a only), permission (c only) → bukan.
	if len(plans) != 1 {
		t.Fatalf("harus 1 kelas collective (not-found), dapat %d: %v", len(plans), keys(plans))
	}
	p := plans["not-found"]
	if p == nil || len(p.have) != 2 || !p.have["a"] || !p.have["b"] {
		t.Fatalf("not-found harus dipunyai {a,b}, dapat %+v", p)
	}
	// label = confidence tertinggi (b's 0.9).
	if p.label != "verify new location then retry" {
		t.Errorf("label harus dari conf tertinggi (b), dapat %q", p.label)
	}
}

// minAgents tinggi → ga ada yg lolos (dormant pas konvergensi kurang).
func TestComputeCollectiveAntibodies_Dormant(t *testing.T) {
	perAgent := map[string][]agentdb.ClassLabel{
		"a": {{Class: "not-found", Label: "x", Confidence: 0.85}},
	}
	if plans := computeCollectiveAntibodies(perAgent, 2); len(plans) != 0 {
		t.Fatalf("1 agent harus 0 collective (dormant), dapat %d", len(plans))
	}
}

func keys(m map[string]*antibodyPlan) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}
