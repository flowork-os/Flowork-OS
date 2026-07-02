package mesh

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// lockboxTestDB — mesh_knowledge_inbox + mesh_tool_manifests minimal (mirror schema store).
func lockboxTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE mesh_knowledge_inbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			packet_id TEXT NOT NULL UNIQUE, origin_pubkey TEXT NOT NULL,
			drawer_content TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'shadow',
			arrived_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE mesh_tool_manifests (
			tool_name TEXT NOT NULL, origin_pubkey TEXT NOT NULL,
			manifest_json TEXT NOT NULL DEFAULT '{}', signature TEXT NOT NULL DEFAULT '',
			arrived_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			consented INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (tool_name, origin_pubkey));`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func TestRevokeByOrigin(t *testing.T) {
	db := lockboxTestDB(t)
	defer db.Close()
	seed := func(pid, origin, status string) {
		db.Exec(`INSERT INTO mesh_knowledge_inbox(packet_id,origin_pubkey,drawer_content,status) VALUES(?,?,?,?)`,
			pid, origin, "isi", status)
	}
	seed("p1", "peerBAD", "promoted")
	seed("p2", "peerBAD", "quarantine")
	seed("p3", "peerBAD", "dropped") // dropped ga di-revoke
	seed("p4", "peerGOOD", "promoted")

	n, err := RevokeByOrigin(db, "peerBAD")
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if n != 2 { // p1+p2 (p3 dropped di-skip)
		t.Fatalf("mau 2 ke-revoke, dapet %d", n)
	}
	// peerGOOD ga ke-sentuh
	var goodStatus string
	db.QueryRow(`SELECT status FROM mesh_knowledge_inbox WHERE packet_id='p4'`).Scan(&goodStatus)
	if goodStatus != "promoted" {
		t.Fatalf("peerGOOD ga boleh ke-revoke, status=%s", goodStatus)
	}
	// provenance ngerangkum bener
	prov, err := ProvenanceByOrigin(db)
	if err != nil {
		t.Fatalf("prov: %v", err)
	}
	var bad *PeerProvenance
	for i := range prov {
		if prov[i].OriginPubkey == "peerBAD" {
			bad = &prov[i]
		}
	}
	if bad == nil || bad.Invalidated != 2 || bad.Dropped != 1 {
		t.Fatalf("provenance peerBAD salah: %+v", bad)
	}
}

func TestToolConsentGate(t *testing.T) {
	db := lockboxTestDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO mesh_tool_manifests(tool_name,origin_pubkey,manifest_json) VALUES('scanner','peerX','{}')`)

	// default consented=0 → muncul di pending, KOSONG di consented.
	pend, _ := PendingToolManifests(db, 10)
	if len(pend) != 1 {
		t.Fatalf("mau 1 pending, dapet %d", len(pend))
	}
	cons, _ := ConsentedToolManifests(db, 10)
	if len(cons) != 0 {
		t.Fatalf("mau 0 consented sebelum approve, dapet %d", len(cons))
	}
	// owner approve → pindah ke consented.
	if err := SetToolConsent(db, "scanner", "peerX", true); err != nil {
		t.Fatalf("consent: %v", err)
	}
	pend, _ = PendingToolManifests(db, 10)
	cons, _ = ConsentedToolManifests(db, 10)
	if len(pend) != 0 || len(cons) != 1 {
		t.Fatalf("abis approve: pending=%d consented=%d (mau 0/1)", len(pend), len(cons))
	}
	// manifest ga ada → error.
	if err := SetToolConsent(db, "ghost", "peerX", true); err == nil {
		t.Fatal("consent manifest ga ada harus error")
	}
}

func TestEvolusiWall(t *testing.T) {
	os.Unsetenv("FLOWORK_MESH_EVOLUSI_ALLOW")
	if !EvolusiWallActive() {
		t.Fatal("tembok default HARUS aktif (mesh haram ke evolusi)")
	}
	ids := []string{"local_1", "mesh_know_abc", "local_2", "mesh_know_def"}
	got := FilterOutMeshNodes(ids)
	if len(got) != 2 || got[0] != "local_1" || got[1] != "local_2" {
		t.Fatalf("tembok naik harus buang node mesh, dapet %v", got)
	}
	// owner longgarin → passthrough.
	os.Setenv("FLOWORK_MESH_EVOLUSI_ALLOW", "1")
	defer os.Unsetenv("FLOWORK_MESH_EVOLUSI_ALLOW")
	if EvolusiWallActive() {
		t.Fatal("switch ON harus nurunin tembok")
	}
	if got := FilterOutMeshNodes(ids); len(got) != 4 {
		t.Fatalf("tembok turun harus passthrough semua, dapet %v", got)
	}
}
