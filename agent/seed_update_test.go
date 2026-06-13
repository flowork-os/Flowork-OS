package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"flowork-gui/internal/floworkdb"
)

// TestSeedNewByIdAdditive is the P1.5 auto-update guarantee for bundled content:
// shipping a NEW schedule rule in a later release installs it on an already-populated
// machine, WITHOUT overwriting a rule the user edited and WITHOUT resurrecting one the
// user deleted. Backed by the "seeded_schedule_ids" kv ledger in seedSocialDefaults.
func TestSeedNewByIdAdditive(t *testing.T) {
	dir := t.TempDir()
	// seedSocialDefaults reads "seed/social.seed.json" relative to the cwd.
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	writeSeed := func(rules ...map[string]string) {
		var sched []map[string]string
		for _, r := range rules {
			sched = append(sched, r)
		}
		doc := map[string]any{"schedule": sched, "group_config": map[string]string{}, "config_groups": []string{}}
		b, _ := json.Marshal(doc)
		if err := os.MkdirAll(filepath.Join(dir, "seed"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "seed", "social.seed.json"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ruleA := map[string]string{"id": "rule-a", "name": "Rule A", "cron": "0 9 * * *", "target": "grp", "prompt": "/a", "kind": "group"}
	ruleB := map[string]string{"id": "rule-b", "name": "Rule B", "cron": "0 10 * * *", "target": "grp", "prompt": "/b", "kind": "group"}
	ruleC := map[string]string{"id": "rule-c", "name": "Rule C", "cron": "0 11 * * *", "target": "grp", "prompt": "/c", "kind": "group"}

	fdb, err := floworkdb.Open(filepath.Join(dir, "flowork.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer fdb.Close()
	if err := fdb.EnsureTriggerSchema(); err != nil { // boot does this before seeding
		t.Fatalf("ensure trigger schema: %v", err)
	}

	ids := func() map[string]string { // id -> name
		ts, err := fdb.ListTriggers()
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		m := map[string]string{}
		for _, x := range ts {
			m[x.ID] = x.Name
		}
		return m
	}

	// 1) FRESH install: ship A + B → both installed.
	writeSeed(ruleA, ruleB)
	seedSocialDefaults(fdb, dir)
	got := ids()
	if len(got) != 2 || got["rule-a"] == "" || got["rule-b"] == "" {
		t.Fatalf("fresh: expected {A,B}, got %v", got)
	}

	// 2a) user EDITS rule A (renames it).
	a := floworkdb.Trigger{ID: "rule-a", Name: "Rule A (user edited)", TypeID: "time",
		Config: `{"cron":"0 9 * * *"}`, Target: "grp", TargetKind: "group", Prompt: "/a", Deliver: "telegram", Enabled: true}
	if err := fdb.UpsertTrigger(a); err != nil {
		t.Fatalf("user edit: %v", err)
	}
	// 2b) a later RELEASE ships A + B + C → only C is new.
	writeSeed(ruleA, ruleB, ruleC)
	seedSocialDefaults(fdb, dir)
	got = ids()
	if len(got) != 3 || got["rule-c"] == "" {
		t.Fatalf("update: expected C added (3 rules), got %v", got)
	}
	if got["rule-a"] != "Rule A (user edited)" {
		t.Fatalf("update CLOBBERED the user's edit: rule-a name = %q", got["rule-a"])
	}

	// 3) user DELETES rule B, then the same release boots again → B must NOT come back.
	if err := fdb.DeleteTrigger("rule-b"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	seedSocialDefaults(fdb, dir)
	got = ids()
	if _, back := got["rule-b"]; back {
		t.Fatalf("deleted rule-b was resurrected by re-seed: %v", got)
	}
	if len(got) != 2 {
		t.Fatalf("after delete+reseed: expected {A,C}, got %v", got)
	}

	// ledger records every id ever seeded (A,B,C) regardless of current presence.
	raw, _ := fdb.GetKV("seeded_schedule_ids")
	var ledger []string
	_ = json.Unmarshal([]byte(raw), &ledger)
	if len(ledger) != 3 {
		t.Fatalf("ledger should remember 3 ids (a,b,c), got %v", ledger)
	}
}
