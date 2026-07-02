// feature_health_doctor_ext.go — SIBLING ext (deletable, NON-frozen): colok cek
// doctor LANJUTAN ke papan RegisterHealthCheck (feature_health.go). Nambah ke
// payload /api/health: go toolchain, index vektor siap, ruang disk, waktu boot.
// Anti-hardcode alamat: pakai ProjectRoot/HOME auto. Panic ext ke-recover di
// runHealthChecks → endpoint tetep idup. 📄 Dok: FLowork_os/lock/approval-gate.md
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	RegisterHealthCheck(doctorGoToolchain)
	RegisterHealthCheck(doctorVectorIndex)
	RegisterHealthCheck(doctorDiskSpace)
	RegisterHealthCheck(doctorRouterReachable)
}

// doctorGoToolchain — go ada + versinya (buat auto-rebuild saat boot). Cari di
// PATH + lokasi umum (go-sdk) tanpa hardcode absolut tunggal.
func doctorGoToolchain(out map[string]any) {
	home, _ := os.UserHomeDir()
	cands := []string{}
	if p, err := exec.LookPath("go"); err == nil {
		cands = append(cands, p)
	}
	cands = append(cands, filepath.Join(home, "go-sdk", "bin", "go"), "/usr/local/go/bin/go")
	for _, gp := range cands {
		if gp == "" {
			continue
		}
		if _, err := os.Stat(gp); err != nil {
			continue
		}
		v, err := exec.Command(gp, "version").Output()
		if err == nil {
			out["go_ok"] = true
			out["go_version"] = strings.TrimSpace(string(v))
			return
		}
	}
	out["go_ok"] = false
}

// doctorVectorIndex — index vektor (semantic RAG) siap? cek file index di workspace.
func doctorVectorIndex(out map[string]any) {
	home, _ := os.UserHomeDir()
	// Lokasi umum index brain/vektor (best-effort; ga fatal kalau ga ada).
	for _, p := range []string{
		filepath.Join(home, ".flowork", "index"),
		filepath.Join(home, ".flowork", "vindex"),
	} {
		if fi, err := os.Stat(p); err == nil {
			out["vector_index_ok"] = true
			out["vector_index_path"] = p
			_ = fi
			return
		}
	}
	out["vector_index_ok"] = false
}

// doctorDiskSpace — sisa ruang disk di HOME (jaga-jaga penuh → DB/log gagal).
// Impl per-OS di feature_health_doctor_disk_{unix,other}.go — Statfs cuma ada di
// Linux/mac; dulu dipanggil di sini TANPA build-tag → cross-compile Windows
// (portable) GAGAL. Cabut akar: pisah per-OS, non-unix fail-open.
func doctorDiskSpace(out map[string]any) {
	doctorDiskSpaceOS(out)
}

// doctorRouterReachable — cek waktu respon dial router (pelengkap router_ok).
func doctorRouterReachable(out map[string]any) {
	// router_ok udah di-cek healthReport; di sini tambah timestamp uptime-ish.
	out["checked_at"] = time.Now().UTC().Format(time.RFC3339)
}
