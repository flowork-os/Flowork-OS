// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Transport+wiring stabil. Logic di internal/journal. Dok: lock/worklog.md
//
// feature_journal.go — JURNAL PENGALAMAN (roadmap P2 SURFACE). Seam feature-registry.
//   1. HTTP GET /api/journal (buat panel GUI owner, di belakang auth single-owner).
//   2. wire hook builtins.JournalScanHook (biar tool `journal` bisa scan lintas-agent).
// Switch FLOWORK_JOURNAL (default ON, read-only no-token). Dok: lock/worklog.md.
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"flowork-gui/internal/journal"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/tools/builtins"
)

func journalEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_JOURNAL"))
	return v == "" || v == "1" || strings.EqualFold(v, "true")
}

func init() {
	RegisterFeature(Feature{Name: "journal", Phase: PhaseRoute, Apply: func(d *Deps) {
		d.Mux.HandleFunc("/api/journal", journalHandler(d.Host))
		builtins.JournalScanHook = func(lessons int) []journal.Summary {
			if !journalEnabled() || d.Host == nil {
				return []journal.Summary{}
			}
			return journal.Collect(d.Host.AgentIDs(), journalOpenDB(d.Host), lessons)
		}
	}})
}

func journalOpenDB(h *kernelhost.Host) func(string) *sql.DB {
	return func(id string) *sql.DB {
		st, err := h.OpenAgentStore(id)
		if err != nil || st == nil {
			return nil
		}
		return st.DB()
	}
}

func journalHandler(h *kernelhost.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !journalEnabled() || h == nil {
			_ = json.NewEncoder(w).Encode(map[string]any{"enabled": journalEnabled(), "items": []journal.Summary{}})
			return
		}
		items := journal.Collect(h.AgentIDs(), journalOpenDB(h), 5)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"totals":  journal.Totals(items),
			"items":   items,
		})
	}
}
