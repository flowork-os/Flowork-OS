//go:build (linux || darwin || windows) && !android

// 📄 Dok: FLowork_os/lock/browser.md
//
// browser_desktop_seam.go — SWITCH-READER BEKU (POLA-B) buat browser_desktop.go (frozen).
// Pisahan dari browser_desktop_ext.go (non-frozen, titik tuning/extension): switch-reader +
// default aman ADA DI SINI biar inti beku self-sufficient (delete-test §6.4: hapus _ext → fungsi
// ini tetep ada, build OK). Switch sejati = ENV (FLOWORK_BROWSER_*), jadi bekuin fungsi pembacanya
// TIDAK ngurangin tuning. Tool browser BARU tetep lewat file sibling `browser_<nama>.go` (init Register).
package builtins

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
)

// browserIdleTimeout — idle sebelum browser di-close otomatis (anti zombie chromium). Default 30
// menit. Override: env FLOWORK_BROWSER_IDLE_MIN (menit). <=0/invalid → default.
func browserIdleTimeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv("FLOWORK_BROWSER_IDLE_MIN")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Minute
		}
	}
	return 30 * time.Minute
}

// browserHeadless — true (default) = chromium headless. env FLOWORK_BROWSER_HEADLESS=0 → headful.
func browserHeadless() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("FLOWORK_BROWSER_HEADLESS"))) {
	case "0", "off", "false", "no":
		return false
	}
	return true
}

// applyExtraBrowserFlags — nambah flag chromium (boolean) dari env FLOWORK_BROWSER_FLAGS (csv).
// Kosong = ga nambah apa-apa (default).
func applyExtraBrowserFlags(l *launcher.Launcher) *launcher.Launcher {
	raw := strings.TrimSpace(os.Getenv("FLOWORK_BROWSER_FLAGS"))
	if raw == "" {
		return l
	}
	for _, f := range strings.Split(raw, ",") {
		if f = strings.TrimSpace(f); f != "" {
			l = l.Set(flags.Flag(f))
		}
	}
	return l
}
