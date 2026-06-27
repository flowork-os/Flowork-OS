// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Detector stabil + teruji. Tuning: switch FLOWORK_DEADAIR(_MIN). Dok: lock/worklog.md
//
// feature_deadair.go — DEAD-AIR DETECTOR (roadmap P0-C). Seam feature-registry.
//
// Owner: *"1 jam ga ada AI kerja = ada yg ga beres (API habis / error)."* Beda docktor (cek PORT
// idup); ini cek KERJAAN GERAK apa nggak. Sinyal: ada tugas AKTIF (agent_runs running/paused) TAPI
// ga ada yang ke-update > ambang → semua beku → anomali (kemungkinan token/quota habis / provider
// down / error beruntun). Idle TANPA tugas = normal (BUKAN anomali).
//
// Host-side poller (pola wakeup_engine), MURAH: cuma baca agent_runs (reuse internal/worklog),
// NOL token/LLM. Alert ke owner via telegram, di-cooldown biar ga spam. Recovery otomatis = nyusul.
// Switch FLOWORK_DEADAIR (default ON) + FLOWORK_DEADAIR_MIN (default 60). Dok: lock/worklog.md.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/worklog"
)

func deadairEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_DEADAIR"))
	return v == "" || v == "1" || strings.EqualFold(v, "true") // default ON (safety net)
}

func deadairMin() int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("FLOWORK_DEADAIR_MIN"))); err == nil && n > 0 {
		return n
	}
	return 60
}

// deadairDecide — PURE (testable): ada tugas aktif TAPI yang paling baru di-update udah lewat
// ambang → anomali (semua beku). Kosong/ga ada timestamp → bukan anomali (fail-safe).
func deadairDecide(items []worklog.Item, now time.Time, thresholdMin int) (anomaly bool, newest time.Time) {
	if len(items) == 0 {
		return false, time.Time{}
	}
	for _, it := range items {
		if t, e := time.Parse(time.RFC3339, it.Updated); e == nil && t.After(newest) {
			newest = t
		}
	}
	if newest.IsZero() {
		return false, newest
	}
	return now.Sub(newest) > time.Duration(thresholdMin)*time.Minute, newest
}

func init() {
	RegisterFeature(Feature{Name: "deadair", Phase: PhaseWire, Apply: func(d *Deps) {
		if !deadairEnabled() || d.Host == nil {
			return
		}
		go deadairLoop(d.Ctx, d.Host)
	}})
}

func deadairLoop(ctx context.Context, host *kernelhost.Host) {
	t := time.NewTicker(15 * time.Minute)
	defer t.Stop()
	var lastAlert time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if !deadairEnabled() {
				continue
			}
			now := time.Now().UTC()
			items := worklog.Collect(host.AgentIDs(), openAgentDB(host),
				worklogOrchestrator(), worklogStaleMin(), true, now) // activeOnly
			anomaly, newest := deadairDecide(items, now, deadairMin())
			if !anomaly {
				continue
			}
			// cooldown: jangan re-alert dalam 1× ambang (anti-spam, hormatin owner istirahat).
			cd := time.Duration(deadairMin()) * time.Minute
			if !lastAlert.IsZero() && now.Sub(lastAlert) < cd {
				continue
			}
			lastAlert = now
			_ = notifyOwnerTelegram(ctx, fmt.Sprintf(
				"⚠️ DEAD-AIR: %d tugas AKTIF tapi diem >%d mnt (update terakhir %s UTC). "+
					"Kemungkinan token/quota habis, provider down, atau error beruntun. "+
					"Cek: check_token + log /tmp/flowork-*.log.",
				len(items), deadairMin(), newest.Format("15:04")))
		}
	}
}
