// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Tool stabil + teruji (mandor tool_call kebukti). Dok: lock/worklog.md
//
// worklog_tool.go — builtin tool `worklog`: MANDOR baca papan kerja bersama dari DALAM
// agent (tanpa HTTP/auth). Logic agregasi di `internal/worklog`; akses lintas-agent (Host.AgentIDs +
// OpenAgentStore) di-wire dari package main lewat hook `WorklogScanHook` (di feature_worklog.go,
// PhaseRoute). Hook nil (belum di-wire / fitur OFF) → balik kosong (fail-safe). Dok: lock/worklog.md
package builtins

import (
	"context"

	"flowork-gui/internal/tools"
	"flowork-gui/internal/worklog"
)

// WorklogScanHook — di-set package main (punya akses Host). nil = belum di-wire → tool balik kosong.
var WorklogScanHook func(activeOnly bool) []worklog.Item

func init() { tools.Register(&worklogTool{}) }

type worklogTool struct{}

func (worklogTool) Name() string       { return "worklog" }
func (worklogTool) Capability() string { return "state:read" }
func (worklogTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Papan kerja BERSAMA lintas-agent: siapa ngerjain apa, mana NYANGKUT (stale), mana PRIORITAS (high=schedule/trigger/mr-flow). Buat MANDOR rekonsiliasi tugas pas PC idle. Default cuma tugas BELUM kelar; all=true ikutin done/stopped.",
		Params: []tools.Param{
			{Name: "all", Type: tools.ParamBool, Description: "true = ikutin tugas done/stopped (default false: cuma yang belum kelar)"},
		},
		Returns: "{count, items:[{agent,id,label,state,updated,stale,priority}]}",
	}
}

func (worklogTool) Run(ctx context.Context, args map[string]any) (tools.Result, error) {
	all, _ := args["all"].(bool)
	var items []worklog.Item
	if WorklogScanHook != nil {
		items = WorklogScanHook(!all) // activeOnly = !all
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"agent": it.Agent, "id": it.ID, "label": it.Label, "state": it.State,
			"updated": it.Updated, "stale": it.Stale, "priority": it.Priority,
		})
	}
	return tools.Result{Output: map[string]any{"count": len(out), "items": out}}, nil
}
