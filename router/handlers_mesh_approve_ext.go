// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// Approve manual pengetahuan mesh → dok lock/mesh-sharing.md  ⚠️ FROZEN.
// Switch GUI fwswitch FLOWORK_MESH_SHARE/APPROVE. Cara: flowork-secrets/CARAFREEZE.MD.
// Pola freeze: lock/frozen-core.md

package main

import (
	"encoding/json"
	"net/http"

	"github.com/flowork-os/flowork_Router/internal/mesh"
	"github.com/flowork-os/flowork_Router/internal/store"
)

func init() {
	RegisterExtraRoute(func(mux *http.ServeMux) {
		mux.HandleFunc("/api/mesh/knowledge/pending", meshKnowledgePendingHandler)
		mux.HandleFunc("/api/mesh/knowledge/approve", meshKnowledgeDecisionHandler(mesh.StatusPromoted))
		mux.HandleFunc("/api/mesh/knowledge/reject", meshKnowledgeDecisionHandler(mesh.StatusDropped))
	})
}

func meshKnowledgePendingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	d, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items, err := mesh.ListKnowledge(d, mesh.StatusQuarantine, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "count": len(items)})
}

func meshKnowledgeDecisionHandler(newStatus string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			PacketID string `json:"packet_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.PacketID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "packet_id wajib"})
			return
		}
		d, err := store.Open()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if err := mesh.PromoteKnowledge(d, body.PacketID, newStatus); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "packet_id": body.PacketID, "status": newStatus})
	}
}
