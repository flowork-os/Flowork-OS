// 📄 Dok: FLowork_os/lock/CognitiveGraph.md
//
// graph_autosync_seam.go — MEKANISME PAPAN BEKU (POLA-A) buat graph_autosync.go (FROZEN).
// Registry proyeksi sumber→Cognitive-Graph. Pisahan dari graph_autosync_ext.go (non-frozen, titik
// REGISTRASI): mekanisme + default KOSONG aman ADA DI SINI biar inti beku self-sufficient
// (delete-test §6.4: hapus file registrasi → extraGraphProjections kosong → dispatcher no-op, build OK).
//
// Proyeksi BARU: JANGAN edit file ini. Daftar via init(){ RegisterGraphProjection(...) } di file
// SIBLING BARU (mis. graph_proj_xxx.go). Dispatcher frozen manggil runExtraGraphProjections sekali.
package main

import (
	"context"
	"os"
	"strings"

	"flowork-gui/internal/agentdb"
)

// GraphProjection — 1 sumber proyeksi ke Cognitive Graph agent. Run WAJIB idempotent (upsert) +
// fails-open (error → 0,err; JANGAN panik). Switch kosong = selalu jalan; isi ENV key biar bisa
// dimatiin dari GUI (daftarin juga di fwswitch/registry.go).
type GraphProjection struct {
	Name   string
	Switch string
	Run    func(ctx context.Context, store *agentdb.Store, scope string) (int, error)
}

var extraGraphProjections []GraphProjection

// RegisterGraphProjection — titik-extend RESMI. Panggil dari init() file SIBLING BARU.
// Run==nil diabaikan (no-op aman).
func RegisterGraphProjection(p GraphProjection) {
	if p.Run == nil {
		return
	}
	extraGraphProjections = append(extraGraphProjections, p)
}

// graphProjectionSwitchOn — default ON; OFF kalau ENV switch = 0/false/off/no. Switch kosong → ON.
func graphProjectionSwitchOn(key string) bool {
	if strings.TrimSpace(key) == "" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "0", "false", "off", "no":
		return false
	}
	return true
}

// runExtraGraphProjections — dipanggil dispatcher frozen (SyncSourcesToGraph) SEKALI di akhir. Tiap
// proyeksi terdaftar yg switch-nya ON dijalanin. FAILS-OPEN: error 1 proyeksi di-skip, gak ganggu
// core / proyeksi lain. Balikin total node baru/berubah.
func runExtraGraphProjections(ctx context.Context, store *agentdb.Store, scope string) int {
	total := 0
	for _, p := range extraGraphProjections {
		if !graphProjectionSwitchOn(p.Switch) {
			continue
		}
		n, err := p.Run(ctx, store, scope)
		if err == nil {
			total += n
		}
	}
	return total
}
