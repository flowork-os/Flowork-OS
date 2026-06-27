// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Nambah dukungan OS lain: sibling non-frozen `init(){ idleLoadReader = ... }`
//    (POLA B var-default). Tuning ambang = DATA (rule config). Pola lengkap: lock/worklog.md
//
// type_idle.go — TRIGGER TIPE BARU (seam Register): "PC idle / beban rendah".
//
// Bagian roadmap Sistem Saraf Otonom P0-B: pas PC sepi (load < ambang) → fire event →
// (lewat rule data) bangunin MANDOR buat rekonsiliasi papan kerja. Owner: "PC < 60% → bangunin
// kepala organ". Anti-spam: cooldown (ga fire tiap poll selama idle). Fail-safe: ga bisa baca
// beban (OS non-linux / error) → TIDAK fire (diem aman, ga ngaco).
//
// Lock-respecting: file BARU, daftar via `Register` (seam triggers.go FROZEN tetap utuh).
// Mekanisme doang — "idle → ngapain / agent mana" itu DATA (rule trigger via GUI/API). Dok: lock/trigger-schedule.md.
package triggers

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func init() { Register(&idleType{}) }

type idleType struct{}

func (t *idleType) ID() string            { return "idle" }
func (t *idleType) Name() string          { return "PC Idle (beban CPU rendah)" }
func (t *idleType) Mode() string          { return "poll" }
func (t *idleType) PayloadKeys() []string { return []string{"load_pct", "threshold"} }
func (t *idleType) ConfigSchema() []Field {
	return []Field{
		{Key: "threshold", Label: "Ambang idle (load %)", Type: "text", Default: "60", Required: true,
			Help: "load CPU di BAWAH % ini = idle (default 60). 1.00 loadavg per-core = 100%."},
		{Key: "cooldown_min", Label: "Cooldown (menit)", Type: "text", Default: "30", Required: false,
			Help: "jeda minimal antar-fire selama idle, biar ga spam (default 30)."},
	}
}
func (t *idleType) OnWebhook(_ map[string]string, _ []byte) ([]Event, error) { return nil, nil }

// idleLoadReader — SEAM (POLA B var-default): default baca load Linux. OS lain (windows/darwin)
// override via sibling NON-frozen `init(){ idleLoadReader = myReader }` TANPA buka file beku ini.
var idleLoadReader = readLoadPercent

func (t *idleType) Check(cfg map[string]string, state string) ([]Event, string, error) {
	pct, ok := idleLoadReader()
	if !ok {
		return nil, state, nil // ga bisa baca beban → diem (fail-safe, ga fire)
	}
	threshold := parseF(cfg["threshold"], 60)
	cooldown := time.Duration(int(parseF(cfg["cooldown_min"], 30))) * time.Minute
	now := time.Now()
	if !idleShouldFire(pct, threshold, parseTime(state), now, cooldown) {
		// pertahanin state lama kalau idle-tapi-cooldown; reset ke "" kalau ga idle (biar fire instan pas idle lagi).
		if pct >= threshold {
			return nil, "", nil
		}
		return nil, state, nil
	}
	ev := Event{Key: now.Format(time.RFC3339), Payload: map[string]string{
		"load_pct":  strconv.FormatFloat(pct, 'f', 1, 64),
		"threshold": strconv.FormatFloat(threshold, 'f', 0, 64),
	}}
	return []Event{ev}, now.Format(time.RFC3339), nil
}

// idleShouldFire — true kalau idle (load < threshold) DAN cooldown sejak fire terakhir lewat.
// Pure (testable): ga nyentuh OS/jam.
func idleShouldFire(loadPct, threshold float64, lastFire, now time.Time, cooldown time.Duration) bool {
	if loadPct >= threshold {
		return false
	}
	if lastFire.IsZero() {
		return true // baru masuk idle → fire
	}
	return now.Sub(lastFire) >= cooldown
}

// readLoadPercent — load1 / jumlah-core * 100 (%). Linux: /proc/loadavg. OS lain: (0,false) → fail-safe.
func readLoadPercent() (float64, bool) {
	if runtime.GOOS != "linux" {
		return 0, false // TODO multi-OS (windows/darwin) nanti via seam; skarang fail-safe no-fire
	}
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return 0, false
	}
	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}
	cores := runtime.NumCPU()
	if cores < 1 {
		cores = 1
	}
	return load1 / float64(cores) * 100, true
}

func parseF(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil && v > 0 {
		return v
	}
	return def
}

func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(s)); err == nil {
		return t
	}
	return time.Time{}
}
