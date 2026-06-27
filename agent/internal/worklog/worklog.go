// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Logic stabil + teruji. Nambah fitur: file/seam baru. Dok: lock/worklog.md
//
// Package worklog — INTI papan kerja bersama lintas-agent (PR-0 Sistem Saraf Otonom).
//
// Logic agregasi DIPISAH dari transport (HTTP feature_worklog.go) + dari konsumen (builtin tool
// `worklog` buat MANDOR) → SATU sumber kebenaran, anti-duplikat (cabut-akar). Scan `agent_runs`
// tiap agent → 1 papan: siapa ngerjain apa, mana NYANGKUT, mana PRIORITAS owner.
//
// Sumber data = `agent_runs` (internal/tools/builtins/agent_run.go): id,label,state,checkpoint,updated.
// Murni (ga nyentuh OS/jam langsung — `now` di-pass) → gampang di-test. Dok: lock/worklog.md.
package worklog

import (
	"database/sql"
	"sort"
	"strings"
	"time"
)

type Item struct {
	Agent    string `json:"agent"`
	ID       string `json:"id"`
	Label    string `json:"label"`
	State    string `json:"state"`
	Updated  string `json:"updated"`
	Stale    bool   `json:"stale"`
	Priority string `json:"priority"` // "high" (schedule/trigger/mr-flow = kepentingan owner) | "normal"
}

// OriginMarkers — penanda asal di label buat deteksi tugas dari schedule/trigger/wakeup.
// (mr-flow ke-deteksi by agent-id, lebih reliable; ini buat sisanya.)
var OriginMarkers = []string{"[schedule]", "[trigger]", "[wakeup]", "[cron]"}

// IsStale — task running/paused yang `updated`-nya lewat cutoff = NYANGKUT.
func IsStale(state, updated string, cutoff time.Time) bool {
	if state != "running" && state != "paused" {
		return false
	}
	if updated == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, updated)
	return err == nil && t.Before(cutoff)
}

// PriorityOf — "high" kalau punya kepentingan owner: orkestrator (mr-flow) ATAU label bertanda asal.
func PriorityOf(agent, label, orchestrator string) string {
	if agent == orchestrator {
		return "high"
	}
	low := strings.ToLower(label)
	for _, m := range OriginMarkers {
		if strings.Contains(low, m) {
			return "high"
		}
	}
	return "normal"
}

// Collect — scan agent_runs tiap agent → papan terpadu. openDB balikin *sql.DB per-agent (nil = skip).
// activeOnly = buang done/stopped. now = jam acuan (stale + dependable test).
func Collect(ids []string, openDB func(agentID string) *sql.DB, orchestrator string, staleMin int, activeOnly bool, now time.Time) []Item {
	items := []Item{}
	cutoff := now.Add(-time.Duration(staleMin) * time.Minute)
	for _, id := range ids {
		db := openDB(id)
		if db == nil {
			continue
		}
		// Agent yang belum pernah pakai agent_run → tabel ga ada → Query error → skip (fail-safe).
		rows, err := db.Query(`SELECT id, COALESCE(label,''), state, COALESCE(updated,'') FROM agent_runs`)
		if err != nil {
			continue
		}
		for rows.Next() {
			it := Item{Agent: id}
			if err := rows.Scan(&it.ID, &it.Label, &it.State, &it.Updated); err != nil {
				continue
			}
			if activeOnly && (it.State == "done" || it.State == "stopped") {
				continue
			}
			it.Stale = IsStale(it.State, it.Updated, cutoff)
			it.Priority = PriorityOf(it.Agent, it.Label, orchestrator)
			items = append(items, it)
		}
		rows.Close()
	}
	// Urut: PRIORITAS owner (high) dulu → NYANGKUT → paling lama ga di-update.
	sort.Slice(items, func(i, j int) bool {
		hi, hj := items[i].Priority == "high", items[j].Priority == "high"
		if hi != hj {
			return hi
		}
		if items[i].Stale != items[j].Stale {
			return items[i].Stale
		}
		return items[i].Updated < items[j].Updated
	})
	return items
}
