//go:build !linux && !darwin

// feature_health_doctor_disk_other.go — stub doctorDiskSpace buat OS tanpa
// syscall.Statfs (Windows dkk). FAIL-OPEN: ga bisa ngukur ≠ disk penuh — jangan
// nge-degraded-in health cuma karena metrik ga tersedia. (Windows pakai
// GetDiskFreeSpaceEx via golang.org/x/sys kalau nanti dibutuhin — sekarang NO
// dependency baru, aturan push: rebuild user offline jangan gagal.)
// 📄 Dok: FLowork_os/lock/approval-gate.md
package main

func doctorDiskSpaceOS(out map[string]any) {
	out["disk_ok"] = true
	out["disk_free_gb"] = -1 // -1 = unknown (metrik ga tersedia di OS ini)
}
