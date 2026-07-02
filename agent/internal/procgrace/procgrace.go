// procgrace — F-G: matiin proses anak BERTAHAP (graceful), ganti kill-keras.
// 📄 Dok: FLowork_os/lock/procgrace.md
//
// Kill langsung (SIGKILL) = proses anak (app/MCP/adapter) ga sempat flush/tutup
// file/lepas port → data corrupt / port nyangkut. Stop() naik-tahap:
//   1. os.Interrupt (SIGINT) — minta baik-baik
//   2. SIGTERM              — tegas (best-effort; di Windows di-skip aman)
//   3. Kill (SIGKILL)       — paksa (kalau tetep bandel)
// Tiap tahap nunggu FLOWORK_PROC_GRACE_MS (default 2500, cap 30000). Reap via
// Wait internal → ga ninggalin zombie.
//
// MULTI-OS: cuma pakai os.Interrupt + syscall.SIGTERM + Process.Kill (ada di
// semua platform Go; SIGTERM yang ga didukung Windows di-ignore, Kill nutup).
// Core infra (bukan sibling deletable) — dependency-light, unit-tested.
package procgrace

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func graceDur() time.Duration {
	ms := 2500
	if v := strings.TrimSpace(os.Getenv("FLOWORK_PROC_GRACE_MS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 100 {
			ms = n
		}
	}
	if ms > 30000 {
		ms = 30000
	}
	return time.Duration(ms) * time.Millisecond
}

// Stop matiin cmd bertahap lalu reap. Aman dipanggil dgn cmd/Process nil.
// JANGAN panggil cmd.Wait() lagi di caller — Stop udah reap.
func Stop(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()

	grace := graceDur()
	// Tahap 1: SIGINT (minta baik-baik).
	_ = cmd.Process.Signal(os.Interrupt)
	if waitOr(done, grace) {
		return
	}
	// Tahap 2: SIGTERM (tegas; Windows ga dukung → error di-ignore, lanjut kill).
	_ = cmd.Process.Signal(syscall.SIGTERM)
	if waitOr(done, grace) {
		return
	}
	// Tahap 3: SIGKILL (paksa).
	_ = cmd.Process.Kill()
	<-done
}

// waitOr — true kalau proses keburu mati sebelum timeout.
func waitOr(done <-chan struct{}, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-done:
		return true
	case <-t.C:
		return false
	}
}
