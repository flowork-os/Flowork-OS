// fw-http-adapter — binary HTTP-Adapter Core (ROADMAP_REPO_TO_APP F5, kontrak HTTP).
// core_entry app buat repo SERVER (streamlit/fastapi/express/dll): spawn server → tunggu port →
// jembatani op ke HTTP. cwd = folder app (proc.go set cmd.Dir). Lihat internal/apps/httpadapter.
package main

import (
	"fmt"
	"os"

	"flowork-gui/internal/apps/httpadapter"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fw-http-adapter: getwd:", err)
		os.Exit(1)
	}
	if err := httpadapter.Run(os.Stdin, os.Stdout, cwd); err != nil {
		fmt.Fprintln(os.Stderr, "fw-http-adapter:", err)
		os.Exit(1)
	}
}
