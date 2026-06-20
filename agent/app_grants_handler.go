// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-20
// Reason: Endpoint izin-app per-agent + applyAppCaps. Tested end-to-end via auth
//   (login owner → GET list → POST toggle → cap broker live). Owner-only gate
//   konsisten /api/agents/config & /api/apps.

package main

// app_grants_handler.go — "GUI = KEBENARAN" fitur izin-app per-agent (owner
// 2026-06-20: "buat filtur baru, dia diijinin pake app apa saja, list app ngak
// boleh pake json — app ada → muncul, dihapus → ilang dari list").
//
// Endpoint /api/agents/apps?id=<agent>:
//   GET  → list app TERINSTALL (sumber: apps.Manager = folder ~/.flowork/apps,
//          BUKAN json) tiap-tiap dengan flag `permitted` per-agent. App di-
//          uninstall → ilang dari Manager.List() → ilang dari list ini otomatis.
//   POST → {app_id, allow} toggle izin. allow=true grant cap app:<id> ke broker
//          LIVE (langsung bisa tanpa restart), false cabut. GUI = kebenaran.

import (
	"encoding/json"
	"net/http"
	"strings"

	"flowork-gui/internal/agentdb"
	fwapps "flowork-gui/internal/apps"
	"flowork-gui/internal/kernelhost"
)

func appGrantsHandler(mgr *fwapps.Manager, host *kernelhost.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if !appPathIDRe.MatchString(id) {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid id"})
			return
		}
		store, err := host.OpenAgentStore(id)
		if err != nil {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		defer store.Close()

		switch r.Method {
		case http.MethodGet:
			granted := map[string]bool{}
			if g, gerr := store.ListAppGrants(); gerr == nil {
				for _, a := range g {
					granted[a] = true
				}
			}
			// List DIGERAKIN apps.Manager (realita folder), bukan tabel izin →
			// app dihapus otomatis ga muncul. Row izin sisa utk app mati = inert.
			out := []map[string]any{}
			for _, a := range mgr.List() {
				out = append(out, map[string]any{
					"id": a.ID, "name": a.Name, "permitted": granted[a.ID],
				})
			}
			tfWriteJSON(w, 0, map[string]any{"apps": out, "count": len(out)})

		case http.MethodPost:
			var body struct {
				AppID string `json:"app_id"`
				Allow bool   `json:"allow"`
			}
			if derr := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); derr != nil {
				tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid body"})
				return
			}
			appID := strings.TrimSpace(body.AppID)
			if !appPathIDRe.MatchString(appID) {
				tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid app_id"})
				return
			}
			var serr error
			if body.Allow {
				serr = store.GrantApp(appID)
			} else {
				serr = store.RevokeApp(appID)
			}
			if serr != nil {
				tfWriteJSON(w, http.StatusInternalServerError, map[string]any{"error": serr.Error()})
				return
			}
			applyAppCaps(host, id, store) // sinkron broker LIVE — langsung berlaku
			tfWriteJSON(w, 0, map[string]any{"ok": true, "app_id": appID, "permitted": body.Allow})

		default:
			tfWriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET/POST only"})
		}
	}
}

// applyAppCaps — set capability app:* di broker = PERSIS isi app_grants agent
// (live). Cap non-app dipertahankan apa adanya. Dipanggil saat toggle (POST) +
// saat boot. Ini yang bikin "uncentang app → cap dicabut → agent ga bisa pake".
func applyAppCaps(host *kernelhost.Host, agentID string, store *agentdb.Store) {
	if host == nil || host.Broker == nil || store == nil {
		return
	}
	grants, err := store.ListAppGrants()
	if err != nil {
		return
	}
	final := []string{}
	for _, c := range host.Broker.Approved(agentID) {
		if strings.HasPrefix(c, "app:") {
			continue // buang semua app cap lama, rebuild dari grants
		}
		final = append(final, c)
	}
	for _, a := range grants {
		final = append(final, "app:"+a)
	}
	host.Broker.Approve(agentID, final)
}
