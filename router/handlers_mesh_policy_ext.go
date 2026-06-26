// handlers_mesh_policy_ext.go — GET /api/mesh/policy → state share/approve buat GUI. Via seam
// routes_ext.go (RegisterExtraRoute). NON-frozen, deletable. Atur nilainya di tab Switch Fitur.
//
// Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
package main

import (
	"net/http"

	"github.com/flowork-os/flowork_Router/internal/mesh"
)

func init() {
	RegisterExtraRoute(func(mux *http.ServeMux) {
		mux.HandleFunc("/api/mesh/policy", meshPolicyStateHandler)
	})
}

func meshPolicyStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	approve := "manual"
	if mesh.ApproveAuto() {
		approve = "auto"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"share":   mesh.ShareEnabled(),
		"approve": approve,
	})
}
