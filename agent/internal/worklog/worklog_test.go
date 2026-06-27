package worklog

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newRunDB(t *testing.T, rows [][3]string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE agent_runs (id TEXT PRIMARY KEY, label TEXT, state TEXT, checkpoint TEXT, updated TEXT)`); err != nil {
		t.Fatalf("schema: %v", err)
	}
	for _, r := range rows { // {id, state, updated}
		if _, err := db.Exec(`INSERT INTO agent_runs(id,label,state,updated) VALUES(?,?,?,?)`, r[0], "", r[1], r[2]); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	return db
}

func TestCollect(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Minute).Format(time.RFC3339)
	old := now.Add(-120 * time.Minute).Format(time.RFC3339)
	older := now.Add(-300 * time.Minute).Format(time.RFC3339)

	dbs := map[string]*sql.DB{
		"mr-flow": newRunDB(t, [][3]string{{"r-mf", "running", fresh}}), // orkestrator → high
		"ant-a": newRunDB(t, [][3]string{
			{"r-run-old", "running", old},       // normal, NYANGKUT
			{"r-paused-fresh", "paused", fresh},  // normal, fresh
			{"r-done", "done", older},            // dibuang activeOnly
		}),
		"ant-b": newRunDB(t, [][3]string{{"r-stopped", "stopped", older}}),
		"ant-empty": func() *sql.DB { db, _ := sql.Open("sqlite", ":memory:"); db.SetMaxOpenConns(1); return db }(),
	}
	ids := []string{"mr-flow", "ant-a", "ant-b", "ant-empty"}
	open := func(id string) *sql.DB { return dbs[id] }

	got := Collect(ids, open, "mr-flow", 60, true, now)
	if len(got) != 3 {
		t.Fatalf("activeOnly: mau 3, dapet %d: %+v", len(got), got)
	}
	if got[0].Agent != "mr-flow" || got[0].Priority != "high" {
		t.Errorf("item[0] mau mr-flow/high, dapet %s/%s", got[0].Agent, got[0].Priority)
	}
	if got[1].ID != "r-run-old" || !got[1].Stale {
		t.Errorf("item[1] mau r-run-old & stale, dapet %s stale=%v", got[1].ID, got[1].Stale)
	}
	if got[2].ID != "r-paused-fresh" || got[2].Stale {
		t.Errorf("item[2] mau r-paused-fresh & ga stale, dapet %s stale=%v", got[2].ID, got[2].Stale)
	}
	if all := Collect(ids, open, "mr-flow", 60, false, now); len(all) != 5 {
		t.Fatalf("all: mau 5, dapet %d", len(all))
	}
}

func TestPriorityOf(t *testing.T) {
	if PriorityOf("ant-x", "Laporan [schedule] jam 7", "mr-flow") != "high" {
		t.Error("label [schedule] mau high")
	}
	if PriorityOf("ant-x", "riset biasa", "mr-flow") != "normal" {
		t.Error("label biasa mau normal")
	}
	if PriorityOf("mr-flow", "apa aja", "mr-flow") != "high" {
		t.Error("orkestrator mau high")
	}
}
