// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. CLI owner stabil. Tuning lewat flag (-section/-amp/-db). Pola: lock/worklog.md
//
// brain-addconst — CLI owner buat NAMBAH 1 aturan konstitusi sacred ke brain LIVE.
//
// KENAPA ADA: editionGate (HTTP /api/brain/constitution) nge-lock konstitusi di edisi FREE
// (anti-rebrand). Tapi OWNER (lokal, sudah single-owner di mesinnya) tetep harus bisa nambah
// DOKTRINNYA SENDIRI ke brain yang lagi jalan — tanpa flip "corporate" (boundary owner). CLI ini
// jalur edition-INDEPENDENT (owner-local = inheren owner-authed): manggil brain.AddConstitution
// (jalur INTERNAL yg sama dipake seed boot), bukan lewat HTTP gate. Cabut-akar: pisahin "owner
// nambah doktrin" dari "rebrand identitas" — gate rebrand TETAP utuh. Dok: lock/worklog.md / brain.md.
//
// Idempotent: skip kalau section udah ada (hidup). Append-only (hormatin doktrin brain).
// Pakai: go run ./cmd/brain-addconst -section AOLA-014_X -amp 999999 -content-file /path/x.txt
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/flowork-os/flowork_Router/internal/brain"
)

func main() {
	section := flag.String("section", "", "section konstitusi (mis. AOLA-014_TASK_DISIPLIN)")
	amp := flag.Float64("amp", 999999, "amplitude (sacred = 999999)")
	source := flag.String("source", "flow_router_admin", "source")
	contentFile := flag.String("content-file", "", "file teks isi konstitusi (UTF-8)")
	dbPath := flag.String("db", "", "path DB brain (mis. dari log router 'Brain: ...'). Kosong = auto (sidecar.BrainDB).")
	flag.Parse()

	if *section == "" || *contentFile == "" {
		fmt.Fprintln(os.Stderr, "ERR: -section dan -content-file wajib")
		os.Exit(2)
	}
	// Standalone CLI ga punya env/cwd router → sidecar.BrainDB() bisa meleset. Kasih -db eksplisit
	// (path dari log router "Brain: ..."). Override path tunggal sebelum OpenRW.
	if *dbPath != "" {
		brain.SetDBPath(*dbPath)
	}
	raw, err := os.ReadFile(*contentFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR baca content-file:", err)
		os.Exit(1)
	}
	content := string(raw)

	db, err := brain.OpenRW()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR OpenRW:", err)
		os.Exit(1)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM constitution WHERE section=? AND deleted_at IS NULL`, *section).Scan(&n); err != nil {
		fmt.Fprintln(os.Stderr, "ERR cek existing:", err)
		os.Exit(1)
	}
	if n > 0 {
		fmt.Printf("SKIP: section %q udah ada (%d) — append-only, ga dobel\n", *section, n)
		return
	}
	id, err := brain.AddConstitution(context.Background(), *section, content, *amp, *source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR AddConstitution:", err)
		os.Exit(1)
	}
	fmt.Printf("OK injected section=%q id=%d amp=%.0f len=%d\n", *section, id, *amp, len(content))
}
