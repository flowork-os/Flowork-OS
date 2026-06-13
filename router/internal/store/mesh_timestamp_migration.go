// === LOCKED FILE === STABLE — DO NOT MODIFY without owner approval (Aola Sahidin / Mr.Dev).
// Locked 2026-06-13 after the mesh fix was verified end-to-end (node→node→promoted).
package store

// 2026-06-13 (owner-approved, autonomous mesh task): additive migration. mesh_packets never
// stored the packet's signed TimestampNS, so a packet reloaded for gossip forwarding had ts=0 —
// its ed25519 signature then failed verification on the receiver and NO forwarded packet was ever
// accepted. Persisting timestamp_ns lets the signed bytes survive the store→forward round-trip.
// New file (does not modify the locked schema migration); idempotent ALTER.
func init() {
	RegisterMigration(Migration{
		ID:   99002, // after the mesh table (101) and the existing late migration (99001)
		Name: "mesh_packets_timestamp_ns",
		SQL:  `ALTER TABLE mesh_packets ADD COLUMN timestamp_ns INTEGER NOT NULL DEFAULT 0;`,
	})
}
