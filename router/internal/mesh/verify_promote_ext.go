// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// verify_promote_ext.go — SIBLING non-frozen (deletable): FASE M — VERIFY-ON-PROMOTE.
// Akar celah: (a) auto-pipeline nge-promote tanpa ngecek integritas NODE LOKAL (kalau
// node kita udah di-tamper, promote = nyerep knowledge ke otak yg mungkin udah dibajak);
// (b) approve MANUAL (handlers_mesh_approve_ext.go frozen) cuma flip status TANPA
// re-verify signature paket asli. Fix:
//   1. Filter `verify-integrity` (RegisterMeshFilter, jalan sebelum L9 di auto-pipeline):
//      node lokal ke-tamper (tier-2 CoreClean=false) → REJECT (jangan promote apa pun).
//   2. Helper VerifyPacketForPromote: re-verify TANDA-TANGAN ed25519 paket asli dari
//      mesh_packets + integritas node → dipakai endpoint approve-verified (hardened).
// NOL buka frozen: filter lewat registry, helper fungsi baru.
package mesh

import (
	"database/sql"
	"fmt"
)

func init() {
	RegisterMeshFilter(MeshFilter{
		Name:   "verify-integrity",
		Switch: "FLOWORK_MESH_VERIFY_PROMOTE", // default ON (kosong→on lewat meshFilterSwitchOn)
		Run: func(db *sql.DB, pkt Packet, drawerContent string) FilterDecision {
			// Node lokal ke-tamper → JANGAN promote knowledge apa pun (otak mungkin dibajak).
			if !CoreClean() {
				return FilterDecision{
					Layer:    "verify-integrity",
					Decision: "reject",
					Reason:   "node tampered (tier-2 integrity gate FAIL) — promote diblok",
				}
			}
			return FilterDecision{Layer: "verify-integrity", Decision: "pass"}
		},
	})
}

// VerifyPacketForPromote — re-verify SEBELUM promote manual (hardened approve):
//  1. integritas node lokal (tier-2) harus clean,
//  2. paket asli (mesh_packets) masih ada & TANDA-TANGAN ed25519-nya valid.
//
// Balik nil kalau aman di-promote; error jelas kalau ada yang gagal. Paket ga ketemu
// di mesh_packets (mis. cuma masuk lewat jalur non-transport) → error eksplisit biar
// owner sadar (fail-closed, bukan diam-diam lolos).
func VerifyPacketForPromote(db *sql.DB, packetID string) error {
	if !CoreClean() {
		return fmt.Errorf("node tampered (tier-2 integrity gate FAIL) — promote ditolak")
	}
	var p Packet
	err := db.QueryRow(
		`SELECT packet_id, origin_pubkey, packet_type, payload_json, signature, ttl, hop_count, timestamp_ns
		 FROM mesh_packets WHERE packet_id = ?`, packetID).
		Scan(&p.PacketID, &p.OriginPubkey, &p.PacketType, &p.PayloadJSON,
			&p.Signature, &p.TTL, &p.HopCount, &p.TimestampNS)
	if err == sql.ErrNoRows {
		return fmt.Errorf("paket asli %q ga ketemu di mesh_packets — ga bisa re-verify tanda tangan", packetID)
	}
	if err != nil {
		return err
	}
	if verr := p.Verify(); verr != nil {
		return fmt.Errorf("tanda tangan paket %q INVALID saat re-verify: %w", packetID, verr)
	}
	return nil
}
