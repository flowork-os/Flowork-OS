// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// 📄 Dok: FLowork_os/lock/mesh-lockbox.md
//
// handlers_mesh_lockbox_ext.go — SIBLING non-frozen (deletable): endpoint FASE M/B
// (mesh security lockbox) via seam RegisterExtraRoute (routes_ext.go frozen, NOL
// disentuh). Semua loopback-only (owner/local action). Path BARU (ga bentrok frozen):
//
//	FASE B visibility (drawers publik-vs-privat):
//	  GET  /api/brain/drawer/visibility?id=      → {id, visibility}
//	  POST /api/brain/drawer/visibility          → {id, visibility}  (set)
//	  GET  /api/brain/visibility/stats           → {public, private}
//	FASE M provenance + revoke:
//	  GET  /api/mesh/provenance                   → per-origin summary
//	  GET  /api/mesh/provenance/peer?pubkey=      → detail knowledge 1 peer
//	  POST /api/mesh/provenance/revoke            → {origin_pubkey} invalidate semua
//	FASE M verify-on-promote (approve hardened):
//	  POST /api/mesh/knowledge/approve-verified   → {packet_id} re-verify → promote
//	FASE M tembok evolusi:
//	  GET  /api/mesh/evolusi-wall                  → status tembok
//	FASE M consent gate tool manifest:
//	  GET  /api/mesh/tools/pending                 → antrian belum di-restui
//	  GET  /api/mesh/tools/consented               → udah di-restui (discovery aman)
//	  POST /api/mesh/tools/consent                 → {tool_name, origin_pubkey, approve}
package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flowork-os/flowork_Router/internal/brain"
	"github.com/flowork-os/flowork_Router/internal/mesh"
	"github.com/flowork-os/flowork_Router/internal/store"
)

func init() {
	RegisterExtraRoute(func(mux *http.ServeMux) {
		mux.HandleFunc("/api/brain/drawer/visibility", lockboxDrawerVisibilityHandler)
		mux.HandleFunc("/api/brain/visibility/stats", lockboxVisibilityStatsHandler)
		mux.HandleFunc("/api/mesh/provenance", lockboxProvenanceHandler)
		mux.HandleFunc("/api/mesh/provenance/peer", lockboxProvenancePeerHandler)
		mux.HandleFunc("/api/mesh/provenance/revoke", lockboxRevokeHandler)
		mux.HandleFunc("/api/mesh/knowledge/approve-verified", lockboxApproveVerifiedHandler)
		mux.HandleFunc("/api/mesh/evolusi-wall", lockboxEvolusiWallHandler)
		mux.HandleFunc("/api/mesh/tools/pending", lockboxToolsPendingHandler)
		mux.HandleFunc("/api/mesh/tools/consented", lockboxToolsConsentedHandler)
		mux.HandleFunc("/api/mesh/tools/consent", lockboxToolConsentHandler)
	})
}

// ── FASE B: visibility drawers ───────────────────────────────────────────────

func lockboxDrawerVisibilityHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	db, err := brain.OpenRW()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	switch r.Method {
	case http.MethodGet:
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id wajib"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "visibility": brain.DrawerVisibility(db, id)})
	case http.MethodPost:
		var body struct {
			ID         string `json:"id"`
			Visibility string `json:"visibility"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if strings.TrimSpace(body.ID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id wajib"})
			return
		}
		if err := brain.SetDrawerVisibility(db, body.ID, body.Visibility); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": body.ID, "visibility": brain.DrawerVisibility(db, body.ID)})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET atau POST"})
	}
}

func lockboxVisibilityStatsHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	db, err := brain.OpenRW()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	pub, priv := brain.CountVisibility(db)
	writeJSON(w, http.StatusOK, map[string]any{"public": pub, "private": priv})
}

// ── FASE M: provenance + revoke ──────────────────────────────────────────────

func lockboxProvenanceHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items, err := mesh.ProvenanceByOrigin(db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "count": len(items)})
}

func lockboxProvenancePeerHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	pubkey := strings.TrimSpace(r.URL.Query().Get("pubkey"))
	if pubkey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "pubkey wajib"})
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items, err := mesh.KnowledgeByOrigin(db, pubkey, 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "count": len(items)})
}

func lockboxRevokeHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	var body struct {
		OriginPubkey string `json:"origin_pubkey"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	n, err := mesh.RevokeByOrigin(db, body.OriginPubkey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	// Peer jahat → penalti karma biar reputasi turun (best-effort).
	_ = mesh.AdjustKarma(db, body.OriginPubkey, -0.2, "revoked-by-owner")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "revoked": n, "origin_pubkey": body.OriginPubkey})
}

// ── FASE M: verify-on-promote (approve hardened) ─────────────────────────────

func lockboxApproveVerifiedHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	var body struct {
		PacketID string `json:"packet_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if strings.TrimSpace(body.PacketID) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "packet_id wajib"})
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	// RE-VERIFY dulu (integritas node + tanda tangan paket) SEBELUM promote.
	if verr := mesh.VerifyPacketForPromote(db, body.PacketID); verr != nil {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "verify-on-promote GAGAL: " + verr.Error()})
		return
	}
	if err := mesh.PromoteKnowledge(db, body.PacketID, mesh.StatusPromoted); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "packet_id": body.PacketID, "status": mesh.StatusPromoted, "verified": true})
}

// ── FASE M: tembok evolusi ───────────────────────────────────────────────────

func lockboxEvolusiWallHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, mesh.EvolusiWallStatus())
}

// ── FASE M: consent gate tool manifest ───────────────────────────────────────

func lockboxToolsPendingHandler(w http.ResponseWriter, r *http.Request) {
	lockboxToolsList(w, r, false)
}

func lockboxToolsConsentedHandler(w http.ResponseWriter, r *http.Request) {
	lockboxToolsList(w, r, true)
}

func lockboxToolsList(w http.ResponseWriter, r *http.Request, consented bool) {
	if !loopbackOnly(w, r) {
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	var items []mesh.ToolManifestRow
	if consented {
		items, err = mesh.ConsentedToolManifests(db, 100)
	} else {
		items, err = mesh.PendingToolManifests(db, 100)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "count": len(items)})
}

func lockboxToolConsentHandler(w http.ResponseWriter, r *http.Request) {
	if !loopbackOnly(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	var body struct {
		ToolName     string `json:"tool_name"`
		OriginPubkey string `json:"origin_pubkey"`
		Approve      bool   `json:"approve"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if err := mesh.SetToolConsent(db, body.ToolName, body.OriginPubkey, body.Approve); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "tool_name": body.ToolName, "consented": body.Approve})
}
