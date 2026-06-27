// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Transport+wiring stabil. Logic di internal/worklog. Dok: lock/worklog.md
//
// feature_worklog.go — PR-0 (roadmap Sistem Saraf Otonom): PAPAN KERJA BERSAMA lintas-agent.
//
// Transport + wiring. Logic INTI ada di `internal/worklog` (dipakai bareng sama builtin tool
// `worklog` buat MANDOR — anti-duplikat, cabut-akar). File ini:
//   1. mount HTTP `GET /api/worklog` (buat GUI owner, di belakang auth single-owner).
//   2. wire hook `builtins.WorklogScanHook` (biar tool worklog di dalam agent bisa scan papan).
//
// Plug-and-play + lock-respecting: file BARU non-frozen (package main) daftar lewat SEAM
// feature-registry. NOL sentuh frozen. Toggle GUI: FLOWORK_WORKLOG. Dok: lock/worklog.md
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/tools/builtins"
	"flowork-gui/internal/worklog"
)

func init() {
	RegisterFeature(Feature{Name: "worklog", Phase: PhaseRoute, Apply: func(d *Deps) {
		d.Mux.HandleFunc("/api/worklog", worklogHandler(d.Host))
		// Wire hook biar builtin tool `worklog` (dipanggil MANDOR dari dalam) bisa scan papan
		// tanpa HTTP/auth. Closure pegang Host.
		builtins.WorklogScanHook = func(activeOnly bool) []worklog.Item {
			if !worklogEnabled() || d.Host == nil {
				return []worklog.Item{}
			}
			return worklog.Collect(d.Host.AgentIDs(), openAgentDB(d.Host),
				worklogOrchestrator(), worklogStaleMin(), activeOnly, time.Now().UTC())
		}
	}})
}

// openAgentDB — closure: agentID → *sql.DB (read agent_runs). Store TIDAK di-close (handle dikelola Host).
func openAgentDB(h *kernelhost.Host) func(string) *sql.DB {
	return func(id string) *sql.DB {
		st, err := h.OpenAgentStore(id)
		if err != nil || st == nil {
			return nil
		}
		return st.DB()
	}
}

// worklogEnabled — default ON (read-only, ga ada biaya prompt/token). OFF lewat GUI/ENV.
func worklogEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_WORKLOG"))
	return v == "" || v == "1" || strings.EqualFold(v, "true")
}

// worklogStaleMin — menit sebelum task running/paused dianggap NYANGKUT (default 60).
func worklogStaleMin() int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("FLOWORK_WORKLOG_STALE_MIN"))); err == nil && n > 0 {
		return n
	}
	return 60
}

// worklogOrchestrator — agent orkestrator (default mr-flow). Selaras switch FLOWORK_ORCHESTRATOR.
func worklogOrchestrator() string {
	if v := strings.TrimSpace(os.Getenv("FLOWORK_ORCHESTRATOR")); v != "" {
		return v
	}
	return "mr-flow"
}

func worklogHandler(h *kernelhost.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !worklogEnabled() {
			_ = json.NewEncoder(w).Encode(map[string]any{"enabled": false, "items": []worklog.Item{}})
			return
		}
		if h == nil {
			_ = json.NewEncoder(w).Encode(map[string]any{"enabled": true, "count": 0, "items": []worklog.Item{}})
			return
		}
		activeOnly := r.URL.Query().Get("all") != "1" // default: cuma yang BELUM kelar
		items := worklog.Collect(h.AgentIDs(), openAgentDB(h),
			worklogOrchestrator(), worklogStaleMin(), activeOnly, time.Now().UTC())
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled":   true,
			"count":     len(items),
			"stale_min": worklogStaleMin(),
			"items":     items,
		})
	}
}
