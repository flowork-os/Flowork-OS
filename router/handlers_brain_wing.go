// handlers_brain_wing.go — enumerate drawer per WING (corpus → sumber topik
// distilasi scanner). Read-only, paginated. File baru (views handler LOCKED).

package main

import (
	"net/http"

	"github.com/flowork-os/flowork_Router/internal/brain"
	"github.com/flowork-os/flowork_Router/internal/store"
)

// brainWingHandler — GET /api/brain/wing?wing=exploitdb&limit=N&offset=M
// Balik drawer 1 wing (paginated) buat nyisir corpus jadi topik. Read-only.
func brainWingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	d, _ := store.Open()
	s, _ := store.LoadSettings(d)
	applyBrainPath(s)
	if !brain.Available() {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "drawers": []any{}})
		return
	}
	wing := r.URL.Query().Get("wing")
	if wing == "" {
		http.Error(w, "wing required", http.StatusBadRequest)
		return
	}
	limit := atoiDefault(r.URL.Query().Get("limit"), 100)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	roomLike := r.URL.Query().Get("room_like")
	drawers, err := brain.ListByWing(r.Context(), wing, roomLike, limit, offset, 1200)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": true, "wing": wing, "count": len(drawers), "offset": offset, "drawers": drawers,
	})
}
