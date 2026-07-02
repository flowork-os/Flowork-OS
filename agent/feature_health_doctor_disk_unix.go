//go:build linux || darwin

// feature_health_doctor_disk_unix.go — impl doctorDiskSpace buat Linux/mac
// (syscall.Statfs). Pasangan: feature_health_doctor_disk_other.go (stub non-unix).
// 📄 Dok: FLowork_os/lock/approval-gate.md
package main

import (
	"os"
	"syscall"
)

func doctorDiskSpaceOS(out map[string]any) {
	home, _ := os.UserHomeDir()
	var st syscall.Statfs_t
	if err := syscall.Statfs(home, &st); err != nil {
		out["disk_ok"] = false
		return
	}
	freeGB := float64(st.Bavail) * float64(st.Bsize) / (1 << 30)
	out["disk_free_gb"] = int(freeGB)
	out["disk_ok"] = freeGB > 1.0 // < 1GB = warning
	if freeGB <= 1.0 {
		out["status"] = "degraded"
	}
}
