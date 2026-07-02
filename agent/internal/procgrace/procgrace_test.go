package procgrace

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestStop_NilSafe(t *testing.T) {
	Stop(nil)                 // proses nil
	Stop(&exec.Cmd{})         // Process nil
}

func TestStop_TerminatesSleep(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Skipf("sleep ga ada: %v", err)
	}
	t.Setenv("FLOWORK_PROC_GRACE_MS", "300")
	start := time.Now()
	Stop(cmd)
	if el := time.Since(start); el > 5*time.Second {
		t.Fatalf("Stop kelamaan: %v (harusnya ke-terminate cepat)", el)
	}
	// Proses harus udah mati & di-reap: signal ke process yg udah selesai → error.
	if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
		t.Fatal("proses harusnya udah mati setelah Stop (signal 0 ga error)")
	}
}

func TestGraceDur_SwitchAndCap(t *testing.T) {
	t.Setenv("FLOWORK_PROC_GRACE_MS", "500")
	if graceDur() != 500*time.Millisecond {
		t.Errorf("switch ga kebaca: %v", graceDur())
	}
	t.Setenv("FLOWORK_PROC_GRACE_MS", "999999")
	if graceDur() != 30*time.Second {
		t.Errorf("cap 30s ga jalan: %v", graceDur())
	}
	t.Setenv("FLOWORK_PROC_GRACE_MS", "5") // < min → default
	if graceDur() != 2500*time.Millisecond {
		t.Errorf("di bawah min harus default: %v", graceDur())
	}
}
