// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Stabil + teruji. Tuning: switch FLOWORK_BUSY_ALERT(_PCT). Dok: lock/worklog.md
//
// feature_busy.go — REFLEX BEBAN-TINGGI (roadmap P0-A). Seam feature-registry.
//
// Owner: *"PC over-ideal 90% — JANGAN langsung suruh matiin, TAWARI + kasih kesadaran. Bahaya putus
// coding."* → pas load tinggi, KASIH KESADARAN + TAWARI (owner yang putusin), JANGAN auto-matiin.
//
// INSIGHT KUNCI: pas PC udah berat, JANGAN invoke LLM (nambah beban!). Jadi reflex ini HOST-SIDE
// MURNI: baca load → notify owner Telegram (no token, no LLM, no nambah load). Owner-active detection
// eksplisit (input/editor) = refinement nyusul. Switch FLOWORK_BUSY_ALERT (default ON) + _PCT (90).
// Dok: lock/worklog.md.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func busyAlertEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_BUSY_ALERT"))
	return v == "" || v == "1" || strings.EqualFold(v, "true") // default ON
}

func busyPct() float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("FLOWORK_BUSY_PCT")), 64); err == nil && v > 0 {
		return v
	}
	return 90
}

// busyShouldAlert — PURE (testable): load > ambang DAN cooldown sejak alert terakhir lewat.
func busyShouldAlert(loadPct float64, ok bool, threshold float64, lastAlert, now time.Time, cooldown time.Duration) bool {
	if !ok || loadPct <= threshold {
		return false
	}
	if lastAlert.IsZero() {
		return true
	}
	return now.Sub(lastAlert) >= cooldown
}

// readHostLoadPercent — load1/core*100 (Linux /proc/loadavg). OS lain → (0,false) fail-safe.
// (Sengaja kecil + lokal di package main; type_idle.go punya pembaca sendiri di sisi triggers.)
func readHostLoadPercent() (float64, bool) {
	if runtime.GOOS != "linux" {
		return 0, false
	}
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, false
	}
	f := strings.Fields(string(raw))
	if len(f) == 0 {
		return 0, false
	}
	l1, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0, false
	}
	c := runtime.NumCPU()
	if c < 1 {
		c = 1
	}
	return l1 / float64(c) * 100, true
}

func init() {
	RegisterFeature(Feature{Name: "busy-reflex", Phase: PhaseWire, Apply: func(d *Deps) {
		if !busyAlertEnabled() {
			return
		}
		go busyLoop(d.Ctx)
	}})
}

func busyLoop(ctx context.Context) {
	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()
	var lastAlert time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if !busyAlertEnabled() {
				continue
			}
			pct, ok := readHostLoadPercent()
			now := time.Now().UTC()
			if !busyShouldAlert(pct, ok, busyPct(), lastAlert, now, 30*time.Minute) {
				continue
			}
			lastAlert = now
			_ = notifyOwnerTelegram(ctx, fmt.Sprintf(
				"⚠️ PC lagi BERAT (load %.0f%%, di atas %.0f%%). Mau gue jeda task berat / standby dulu? "+
					"Gue GA bakal matiin sendiri — tinggal bilang, bro. (Kalau lo lagi coding, lanjut aja, ini cuma kabar.)",
				pct, busyPct()))
		}
	}
}
