// fw-app-adapter — binary CLI-Adapter Core generik (ROADMAP_REPO_TO_APP F1).
// 📄 Dok arsitektur lengkap: FLowork_os/lock/apps-adopt.md (frozen — JANGAN edit tanpa unfreeze sadar).
// Dipanggil engine app sbg core_entry (runtime:process). cwd = folder app (proc.go set cmd.Dir),
// tempat adapter.json + kode repo berada. Baca stdio baris-JSON, jembatani ke command repo.
//
// Build sekali (ikut pipeline portable multi-OS), white-label. Lihat internal/apps/cliadapter.
package main

import (
	"fmt"
	"os"

	"flowork-gui/internal/apps/cliadapter"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fw-app-adapter: getwd:", err)
		os.Exit(1)
	}
	if err := cliadapter.Run(os.Stdin, os.Stdout, cwd); err != nil {
		fmt.Fprintln(os.Stderr, "fw-app-adapter:", err)
		os.Exit(1)
	}
}
