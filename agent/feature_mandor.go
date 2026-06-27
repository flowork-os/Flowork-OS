// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Wiring stabil. Rule live = trigger_rules DB (GUI). Dok: lock/worklog.md
//
// feature_mandor.go — wiring AGENT MANDOR (roadmap P0-B). Seam feature-registry.
//
// Switch FLOWORK_MANDOR (default OFF): ON → (1) seed agent mandor + tool-nya, (2) seed RULE trigger
// `idle→mandor` (DATA): pas PC idle (load<ambang + cooldown) → invoke mandor buat rekonsiliasi papan.
// Default OFF biar ga bikin agent dormant + ke-reap sebelum dipakai. Dok: lock/worklog.md.
//
// REM (anti "bakar terus"): idle (load rendah = owner ga hammer) + cooldown 30mnt + persona mandor
// "papan kosong → diem". Deteksi owner-aktif eksplisit + token-budget gate = refinement nyusul.
package main

import (
	"os"
	"strings"

	"flowork-gui/internal/floworkdb"
)

const mandorRuleID = "idle-mandor"

// mandorEnabled — switch FLOWORK_MANDOR. Default OFF (nyalain kalau wiring operasional lengkap).
func mandorEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_MANDOR"))
	return v == "1" || strings.EqualFold(v, "true")
}

func init() {
	RegisterFeature(Feature{Name: "mandor", Phase: PhaseSeed, Apply: func(d *Deps) {
		if !mandorEnabled() {
			return
		}
		seedMandor()
		seedMandorRule(d.FDB)
	}})
}

// seedMandorRule — bikin rule trigger idle→mandor (idempotent: skip kalau ID udah ada → hormatin
// edit/disable owner di GUI). nil FDB = skip aman.
func seedMandorRule(fdb *floworkdb.Store) {
	if fdb == nil {
		return
	}
	if existing, err := fdb.ListTriggers(); err == nil {
		for _, t := range existing {
			if t.ID == mandorRuleID {
				return // udah ada → jangan timpa (owner mungkin udah atur)
			}
		}
	}
	_ = fdb.UpsertTrigger(floworkdb.Trigger{
		ID:         mandorRuleID,
		Name:       "Mandor — supervisor idle",
		TypeID:     "idle",
		Config:     `{"threshold":"60","cooldown_min":"30"}`,
		Target:     mandorID,
		TargetKind: "agent",
		Prompt: "PC lagi idle (load {{load_pct}}%). Jalanin tugas MANDOR kamu: panggil tool `worklog`, " +
			"fokus yang priority=high + stale, lalu `agent_command` ke agent pemiliknya buat lanjut. " +
			"Kalau papan kosong / ga ada yang nyangkut → diem, balas singkat 'aman'.",
		Deliver: "",
		Enabled: true,
	})
}
