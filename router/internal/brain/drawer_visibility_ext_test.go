package brain

import (
	"database/sql"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

func visTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// drawers minimal (subset kolom relevan + visibility ditambah lazy).
	_, err = db.Exec(`CREATE TABLE drawers (
		id TEXT PRIMARY KEY, content TEXT NOT NULL DEFAULT '', deleted_at TIMESTAMP)`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func TestDrawerVisibilityDefaultPrivate(t *testing.T) {
	drawerVisOnce = sync.Once{} // fresh lazy-migration per test (in-memory db baru)
	db := visTestDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO drawers(id,content) VALUES('d1','rahasia')`)

	// default = private (fail-closed), ga shareable.
	if v := DrawerVisibility(db, "d1"); v != VisibilityPrivate {
		t.Fatalf("default harus private, dapet %q", v)
	}
	if IsShareable(db, "d1") {
		t.Fatal("private ga boleh shareable")
	}
	// owner tandai public → shareable.
	if err := SetDrawerVisibility(db, "d1", "public"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if !IsShareable(db, "d1") {
		t.Fatal("public harus shareable")
	}
	// balikin ke private.
	if err := SetDrawerVisibility(db, "d1", "private"); err != nil {
		t.Fatalf("set2: %v", err)
	}
	if IsShareable(db, "d1") {
		t.Fatal("balik private → ga shareable")
	}
	// drawer ga ada → error + fail-closed private.
	if err := SetDrawerVisibility(db, "ghost", "public"); err == nil {
		t.Fatal("set drawer ga ada harus error")
	}
	if DrawerVisibility(db, "ghost") != VisibilityPrivate {
		t.Fatal("drawer ga ada harus dianggap private")
	}
}

func TestVisibilityNormalizeFailClosed(t *testing.T) {
	// nilai aneh → private (fail-closed).
	for _, v := range []string{"", "PUBLICKY", "shared", "yes", "1"} {
		if normalizeVisibility(v) != VisibilityPrivate {
			t.Fatalf("%q harus jadi private (fail-closed)", v)
		}
	}
	if normalizeVisibility("PUBLIC") != VisibilityPublic {
		t.Fatal("PUBLIC (case-insensitive) harus public")
	}
}
