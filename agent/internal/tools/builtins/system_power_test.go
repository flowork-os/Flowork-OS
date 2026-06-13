package builtins

import (
	"strings"
	"testing"
)

// TestResolvePowerCmdFor — verifikasi pemetaan (OS, aksi) → argv lintas platform
// TANPA spawn proses (resolvePowerCmdFor murni). Mencakup klaim multi-OS tutorial:
// linux (RasPi/STB), darwin (macOS), windows; Android dibedakan (error edukatif).
func TestResolvePowerCmdFor(t *testing.T) {
	cases := []struct {
		goos, action string
		wantPrefix   string // argv[0] yang diharapkan
		wantErr      bool
	}{
		// Linux (juga Raspberry Pi & STB berbasis Linux)
		{"linux", "shutdown", "systemctl", false},
		{"linux", "reboot", "systemctl", false},
		{"linux", "suspend", "systemctl", false},
		{"linux", "lock", "loginctl", false},
		// macOS
		{"darwin", "shutdown", "osascript", false},
		{"darwin", "reboot", "osascript", false},
		{"darwin", "suspend", "pmset", false},
		{"darwin", "lock", "pmset", false},
		// Windows
		{"windows", "shutdown", "shutdown.exe", false},
		{"windows", "reboot", "shutdown.exe", false},
		{"windows", "suspend", "rundll32.exe", false},
		{"windows", "lock", "rundll32.exe", false},
		// Android dibedakan → error (edukatif)
		{"android", "shutdown", "", true},
		{"android", "reboot", "", true},
		// OS tak dikenal → error edukatif
		{"plan9", "shutdown", "", true},
	}
	for _, c := range cases {
		argv, err := resolvePowerCmdFor(c.goos, c.action)
		if c.wantErr {
			if err == nil {
				t.Errorf("%s/%s: mau error, malah argv=%v", c.goos, c.action, argv)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s/%s: tak terduga error: %v", c.goos, c.action, err)
			continue
		}
		if len(argv) == 0 || argv[0] != c.wantPrefix {
			t.Errorf("%s/%s: argv[0]=%q, mau %q", c.goos, c.action, argvHead(argv), c.wantPrefix)
		}
	}
}

// TestAndroidErrorEducational — error Android harus "memeluk" (tidak menyalahkan agent).
func TestAndroidErrorEducational(t *testing.T) {
	_, err := resolvePowerCmdFor("android", "shutdown")
	if err == nil {
		t.Fatal("android shutdown harus error")
	}
	msg := strings.ToLower(err.Error())
	// harus menjelaskan ALASAN (root) + memberi PETUNJUK, bukan sekadar "not supported"
	if !strings.Contains(msg, "root") {
		t.Errorf("error Android harus jelaskan butuh root; dapat: %q", err.Error())
	}
	if !strings.Contains(msg, "bukan kesalahan") && !strings.Contains(msg, "petunjuk") {
		t.Errorf("error Android harus edukatif (peluk+petunjuk); dapat: %q", err.Error())
	}
}

func argvHead(a []string) string {
	if len(a) == 0 {
		return ""
	}
	return a[0]
}
