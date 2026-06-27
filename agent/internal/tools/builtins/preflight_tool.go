// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Tool stabil. Tuning: switch FLOWORK_URGENCY/FLOWORK_BUSY_PCT. Dok: lock/worklog.md
//
// preflight_tool.go — builtin `preflight` (roadmap: Shadow-P3 pre-flight + Cost-of-thought meter).
//
// Owner anti-halu (Rule 1): JANGAN ngarang ramalan ("nanti overheat/dapet sekian"). Tool ini kasih
// DATA NYATA buat keputusan sebelum task berat: beban PC sekarang (load %) + mode urgensi (hemat/
// normal/deadly) + saran konkret. BUKAN simulasi fantasi — angka real dari /proc/loadavg.
// Cost-of-thought: mode urgensi (FLOWORK_URGENCY) di-surface biar agent self-moderate (hemat=ringkas,
// deadly=buka penuh). Mapping ke model-tier otomatis = follow-up router. Dok: lock/worklog.md.
package builtins

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"

	"flowork-gui/internal/tools"
)

func init() { tools.Register(&preflightTool{}) }

type preflightTool struct{}

func (preflightTool) Name() string       { return "preflight" }
func (preflightTool) Capability() string { return "state:read" }
func (preflightTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Cek DATA NYATA sebelum kerjaan berat: beban CPU sekarang (load %), PC berat apa nggak, + mode urgensi (hemat/normal/deadly). Pakai buat putusin effort/lanjut — BUKAN nebak/ngarang. Return {load_pct, busy, urgency, advice}.",
		Params:      nil,
		Returns:     "{load_pct, busy, urgency, advice}",
	}
}

func (preflightTool) Run(ctx context.Context, _ map[string]any) (tools.Result, error) {
	pct, ok := preflightLoadPct()
	busyPct := preflightEnvFloat("FLOWORK_BUSY_PCT", 90)
	urgency := preflightUrgency()
	busy := ok && pct > busyPct
	return tools.Result{Output: map[string]any{
		"load_pct": pct,
		"load_ok":  ok,
		"busy":     busy,
		"urgency":  urgency,
		"advice":   preflightAdvice(pct, ok, busy, urgency),
	}}, nil
}

// preflightUrgency — mode urgensi (cost-of-thought). hemat | normal | deadly. Default normal.
func preflightUrgency() string {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("FLOWORK_URGENCY")))
	switch v {
	case "hemat", "deadly":
		return v
	default:
		return "normal"
	}
}

// preflightAdvice — saran konkret dari data (PURE, testable).
func preflightAdvice(pct float64, ok, busy bool, urgency string) string {
	if urgency == "deadly" {
		return "URGENSI DEADLY: buka kapasitas penuh, jalan terus walau berat."
	}
	if !ok {
		return "beban PC ga ke-baca (OS non-linux?) — lanjut normal, pantau manual."
	}
	if busy {
		if urgency == "hemat" {
			return "PC BERAT + mode hemat: tunda task berat / pecah kecil, jangan barengan."
		}
		return "PC BERAT: pertimbangin jeda/standby task berat, atau tawari owner."
	}
	if urgency == "hemat" {
		return "PC santai tapi mode HEMAT: kerja ringkas, jangan over-think."
	}
	return "PC santai: aman lanjut normal."
}

func preflightLoadPct() (float64, bool) {
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

func preflightEnvFloat(key string, def float64) float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(key)), 64); err == nil && v > 0 {
		return v
	}
	return def
}
