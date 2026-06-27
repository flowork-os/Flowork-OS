// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Tool stabil. Hook di-wire feature_journal.go. Dok: lock/worklog.md
//
// journal_tool.go — builtin `journal`: agent self-refleksi "apa yang udah gue/koloni pelajari"
// (pelajaran kegagalan, eureka, antibody, skill, insting) lintas-agent. Logic di internal/journal;
// akses Host di-wire dari package main lewat JournalScanHook. nil = balik kosong. Dok: lock/worklog.md.
package builtins

import (
	"context"

	"flowork-gui/internal/journal"
	"flowork-gui/internal/tools"
)

// JournalScanHook — di-set package main (akses Host). nil = belum di-wire → kosong.
var JournalScanHook func(lessonsPerAgent int) []journal.Summary

func init() { tools.Register(&journalTool{}) }

type journalTool struct{}

func (journalTool) Name() string       { return "journal" }
func (journalTool) Capability() string { return "state:read" }
func (journalTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Jurnal pengalaman KOLONI: ringkas apa yang tiap agent UDAH PELAJARI — pelajaran dari kegagalan (mistakes), eureka, antibody anti-halu, skill self-authored, insting. Buat refleksi/anti-ulang kesalahan. Return {totals, items:[{agent,mistakes,eureka,antibody,skills,instincts,lessons}]}.",
		Params:      nil,
		Returns:     "{totals, items:[...]}",
	}
}

func (journalTool) Run(ctx context.Context, _ map[string]any) (tools.Result, error) {
	var items []journal.Summary
	if JournalScanHook != nil {
		items = JournalScanHook(3)
	}
	return tools.Result{Output: map[string]any{
		"totals": journal.Totals(items),
		"items":  items,
	}}, nil
}
