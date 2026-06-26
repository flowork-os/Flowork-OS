// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// Gate share/approve mesh (tier-2 jantung) → dok lock/mesh-sharing.md  ⚠️ FROZEN.
// Switch GUI fwswitch FLOWORK_MESH_SHARE/APPROVE. Cara: flowork-secrets/CARAFREEZE.MD.
// Pola freeze: lock/frozen-core.md

package mesh

import "database/sql"

func init() {
	RegisterMeshFilter(MeshFilter{
		Name: "share-policy",
		Run: func(db *sql.DB, pkt Packet, drawerContent string) FilterDecision {
			if !meshShareEnabled() {
				return FilterDecision{Layer: "L0-share-policy", Decision: "reject", Reason: "mesh share OFF — reciprocity: tak nerima pengetahuan"}
			}
			if !meshApproveAuto() {
				return FilterDecision{Layer: "L0-share-policy", Decision: "flag", Reason: "approve manual — masuk antrian (quarantine), nunggu owner approve"}
			}
			return FilterDecision{Layer: "L0-share-policy", Decision: "pass"}
		},
	})
}
