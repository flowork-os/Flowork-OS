// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// mesh_lockbox_migration.go — SIBLING non-frozen (deletable): migrasi ADDITIVE buat
// FASE M (mesh security lockbox) via seam RegisterMigration (migrate.go frozen, NOL
// disentuh). Kolom `consented` di mesh_tool_manifests = gerbang consent owner
// (default 0 = BELUM di-restui → ga muncul di discovery-consented sampai owner approve).
// Semua ADDITIVE (ALTER ADD COLUMN) → aman auto-update user (HUKUM MUTLAK push).
package store

func init() {
	RegisterMigration(Migration{
		ID:   99010,
		Name: "mesh_tool_manifests_consented",
		SQL:  `ALTER TABLE mesh_tool_manifests ADD COLUMN consented INTEGER NOT NULL DEFAULT 0;`,
	})
}
