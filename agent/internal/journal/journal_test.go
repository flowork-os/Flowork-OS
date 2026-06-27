package journal

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func mkAgentDB(t *testing.T, mistakes int, eureka int) *sql.DB {
	t.Helper()
	db, _ := sql.Open("sqlite", ":memory:")
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE mistakes_local (id INTEGER PRIMARY KEY, title TEXT, hit_count INT, last_hit_at TEXT, deleted_at TEXT)`)
	db.Exec(`CREATE TABLE brain_drawers (id INTEGER PRIMARY KEY, wing TEXT, deleted_at TEXT)`)
	db.Exec(`CREATE TABLE brain_antibody (id INTEGER PRIMARY KEY)`)
	db.Exec(`CREATE TABLE skills (id INTEGER PRIMARY KEY, archived INT)`)
	db.Exec(`CREATE TABLE cognitive_nodes (id INTEGER PRIMARY KEY, type TEXT)`)
	for i := 0; i < mistakes; i++ {
		db.Exec(`INSERT INTO mistakes_local(title,hit_count) VALUES(?,?)`, "pelajaran-"+string(rune('A'+i)), i+1)
	}
	for i := 0; i < eureka; i++ {
		db.Exec(`INSERT INTO brain_drawers(wing) VALUES('eureka')`)
	}
	return db
}

func TestJournalCollect(t *testing.T) {
	dbs := map[string]*sql.DB{
		"rich":  mkAgentDB(t, 3, 2), // total 5
		"thin":  mkAgentDB(t, 1, 0), // total 1
		"empty": mkAgentDB(t, 0, 0), // di-skip
	}
	ids := []string{"thin", "rich", "empty"}
	open := func(id string) *sql.DB { return dbs[id] }

	got := Collect(ids, open, 3)
	if len(got) != 2 {
		t.Fatalf("mau 2 agent (empty di-skip), dapet %d: %+v", len(got), got)
	}
	if got[0].Agent != "rich" { // urut by total desc
		t.Errorf("item[0] mau rich (terkaya), dapet %s", got[0].Agent)
	}
	if got[0].Mistakes != 3 || got[0].Eureka != 2 {
		t.Errorf("rich counts salah: %+v", got[0])
	}
	if len(got[0].Lessons) != 3 || got[0].Lessons[0].Hits < got[0].Lessons[1].Hits {
		t.Errorf("lessons mestinya 3 + urut hit desc: %+v", got[0].Lessons)
	}
	tot := Totals(got)
	if tot["mistakes"] != 4 || tot["eureka"] != 2 || tot["agents"] != 2 {
		t.Errorf("totals salah: %+v", tot)
	}
}
