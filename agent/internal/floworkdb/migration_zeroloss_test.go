package floworkdb

import (
	"path/filepath"
	"testing"
)

// TestMigrationZeroLoss is the owner's guarantee, mechanised: take a database that
// was created by an OLDER Flowork release (a task_categories table WITHOUT the
// later synth_directive / worker_directive columns) holding real user data, then
// run the CURRENT schema-ensure path and prove:
//
//	1. every old user row survives with its values intact (no data loss),
//	2. the new columns are added non-destructively with their defaults,
//	3. running the upgrade again is idempotent (no error, no change).
//
// This is the CODE-vs-DATA contract: upgrades only ADD; they never drop or rewrite
// what the user already had.
func TestMigrationZeroLoss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowork.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// --- simulate an OLD-release DB: task_categories minus the columns that were
	// added by later additive migrations. CREATE TABLE IF NOT EXISTS in the current
	// EnsureTaskSchema() will then be a no-op for this table, exercising the ALTER path.
	if _, err := st.db.Exec(`CREATE TABLE task_categories (
		id           TEXT PRIMARY KEY,
		name         TEXT NOT NULL DEFAULT '',
		icon         TEXT NOT NULL DEFAULT '',
		trigger_hint TEXT NOT NULL DEFAULT '',
		synthesizer  TEXT NOT NULL DEFAULT '',
		enabled      INTEGER NOT NULL DEFAULT 1,
		created_at   TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		t.Fatalf("create old-schema table: %v", err)
	}
	// Real user data the upgrade must never lose.
	if _, err := st.db.Exec(
		`INSERT INTO task_categories (id, name, icon, enabled) VALUES (?,?,?,?)`,
		"cat-finance", "Investment Desk", "📈", 1); err != nil {
		t.Fatalf("seed user row: %v", err)
	}

	// Guard: the old columns really are absent before the upgrade.
	if st.columnExists("task_categories", "synth_directive") {
		t.Fatal("precondition failed: synth_directive already present in old schema")
	}

	// --- run the REAL upgrade path ---
	if err := st.EnsureTaskSchema(); err != nil {
		t.Fatalf("EnsureTaskSchema (upgrade): %v", err)
	}

	// 1. user row survived with values intact.
	var name, icon string
	var enabled int
	if err := st.db.QueryRow(
		`SELECT name, icon, enabled FROM task_categories WHERE id=?`, "cat-finance",
	).Scan(&name, &icon, &enabled); err != nil {
		t.Fatalf("user row lost after upgrade: %v", err)
	}
	if name != "Investment Desk" || icon != "📈" || enabled != 1 {
		t.Fatalf("user data corrupted: name=%q icon=%q enabled=%d", name, icon, enabled)
	}

	// 2. new columns added non-destructively, defaulting to ''.
	for _, col := range []string{"synth_directive", "worker_directive"} {
		if !st.columnExists("task_categories", col) {
			t.Fatalf("new column %q not added by upgrade", col)
		}
	}
	var synth string
	if err := st.db.QueryRow(
		`SELECT synth_directive FROM task_categories WHERE id=?`, "cat-finance",
	).Scan(&synth); err != nil {
		t.Fatalf("read new column: %v", err)
	}
	if synth != "" {
		t.Fatalf("new column default not '': got %q", synth)
	}

	// 3. idempotent: a second upgrade is a no-op and keeps the row.
	if err := st.EnsureTaskSchema(); err != nil {
		t.Fatalf("EnsureTaskSchema (second run): %v", err)
	}
	var count int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM task_categories`).Scan(&count); err != nil {
		t.Fatalf("count after second upgrade: %v", err)
	}
	if count != 1 {
		t.Fatalf("row count changed after idempotent upgrade: got %d, want 1", count)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}
