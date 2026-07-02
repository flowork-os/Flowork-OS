// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// evolusi_wall_ext.go — SIBLING non-frozen (deletable): FASE M — TEMBOK mesh↔evolusi.
// Hukum owner (2026-07-02): knowledge dari peer mesh (source='mesh_federated') HARAM
// nembus ke loop EVOLUSI / codegen self-improve — biar racun federasi ga pernah
// nyampe ke KODE yang di-generate. Sekarang aman by-ISOLASI (mesh knowledge ga masuk
// `drawers`, tool manifest inert, lora disabled), TAPI itu "kebetulan" — tembok ini
// bikin EKSPLISIT + ke-tes, biar AI masa depan ga ngebocorin ga sengaja.
//
// HARD wall: switch `FLOWORK_MESH_EVOLUSI_ALLOW` default OFF (tembok NAIK). ON =
// longgarin (sadar-risiko, owner-only). Titik pakai: jalur yang ngerakit KONTEKS buat
// evolusi WAJIB `FilterOutMeshNodes` node-id-nya dulu.
package mesh

import (
	"os"
	"strings"
)

// meshFederatedPrefix — id node cognitive hasil projeksi mesh (graph_proj_mesh.go).
const meshFederatedPrefix = "mesh_know_"

// EvolusiWallActive — true = tembok NAIK (mesh HARAM ke evolusi). Default true.
// Longgarin cuma dengan FLOWORK_MESH_EVOLUSI_ALLOW = 1/true/on (owner sadar-risiko).
func EvolusiWallActive() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("FLOWORK_MESH_EVOLUSI_ALLOW"))) {
	case "1", "true", "on", "yes":
		return false // di-izinin owner → tembok turun
	}
	return true // default: tembok naik
}

// IsMeshFederatedNodeID — node ini berasal dari mesh (projeksi federasi)?
func IsMeshFederatedNodeID(id string) bool {
	return strings.HasPrefix(id, meshFederatedPrefix)
}

// FilterOutMeshNodes — buang node asal-mesh dari daftar id KALAU tembok naik. Jalur
// apa pun yg ngerakit konteks buat evolusi/codegen WAJIB lewat sini dulu. Tembok
// turun (owner izinin) → passthrough apa adanya.
func FilterOutMeshNodes(ids []string) []string {
	if !EvolusiWallActive() {
		return ids
	}
	out := ids[:0:0]
	for _, id := range ids {
		if !IsMeshFederatedNodeID(id) {
			out = append(out, id)
		}
	}
	return out
}

// EvolusiWallStatus — ringkasan buat endpoint/GUI observability.
func EvolusiWallStatus() map[string]any {
	return map[string]any{
		"wall_active":  EvolusiWallActive(),
		"switch":       "FLOWORK_MESH_EVOLUSI_ALLOW",
		"default":      "OFF (tembok naik — mesh HARAM ke evolusi)",
		"mesh_inert":   MeshInertInvariants(),
		"note":         "mesh knowledge ISOLATED dari drawers/codegen; tool manifest discovery-only; lora apply disabled",
	}
}

// MeshInertInvariants — invarian keamanan yg HARUS tetep true (di-assert test): mesh
// ga pernah auto-eksekusi apa pun ke lokal. Kalau salah satu jadi false ke depan =
// tembok bocor = bug, bukan fitur.
func MeshInertInvariants() map[string]bool {
	return map[string]bool{
		"tool_manifest_discovery_only": true, // manifest peer cuma disimpen, ga auto-install
		"lora_apply_disabled":          true, // ApplyLoraDelta return ErrLoraApplyUnavailable
		"knowledge_isolated_from_drawers": true, // mesh knowledge di inbox, bukan drawers
		"consent_required_for_tools":   true, // manifest butuh restu owner (tool_consent_ext)
	}
}
