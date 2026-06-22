// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md
//
// recovery_share_job.go — D32-INC4: share recovery-instinct generik (hasil INC-3) ke
// shared-brain router → IMUNITAS KOLEKTIF (agent lain recall via brain_search_shared yg
// udah ada). Ticker non-beku, resilient (router mati → skip, retry tick berikut).
//
// Cabut-akar (audit INC-4): "consensus 9-lapis" cuma 6/9 nyata + antibody-kolektif router
// belum ada → INC-4 REUSE yg ADA (federation drawer + gate privasi), BUKAN consensus palsu.
// Consensus N-of-M penuh = roadmap F (blocked multi-peer mesh). Recovery-instinct sini udah
// privacy-safe by-construction (INC-3); tetap di-DOUBLE-CHECK deterministik sebelum keluar.

package main

import (
	"context"
	"log"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/routerclient"
)

// PromoteRecoveryInstinctsShared — tiap agent: recovery-instinct AKTIF+verified yg belum
// di-share → (privacy double-check) → PromoteDrawer ke shared-brain (mem_type=
// recovery_instinct) → MarkPromotedCognitive (anti-double). Return jumlah ke-share.
func PromoteRecoveryInstinctsShared(ctx context.Context, host *kernelhost.Host) int {
	rc := routerclient.New("")
	shared := 0
	for _, id := range host.AgentIDs() {
		store, err := host.OpenAgentStore(id)
		if err != nil {
			continue
		}
		eligible, lerr := store.SelectPromotableRecoveryInstincts(50)
		if lerr != nil || len(eligible) == 0 {
			store.Close()
			continue
		}
		for _, n := range eligible {
			label := n.Label
			// PRIVACY double-check (defense in depth): label HARUS bersih. Kalau strip ngubah
			// (ada path/url/token/hex) ATAU ada brand → JANGAN keluar; tandai biar ga dicoba ulang.
			if agentdb.StripDeterministic(label) != label || agentdb.ContainsBrand(label) {
				_ = store.MarkPromotedCognitive("node:"+n.ID, "", "blocked-privacy")
				continue
			}
			pctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			resp, perr := rc.PromoteDrawer(pctx, routerclient.PromoteDrawerReq{
				Content: label, Wing: "recovery", Room: "recovery_instinct", MemType: "recovery_instinct",
			})
			cancel()
			if perr != nil {
				// Router mati/error → STOP agent ini tick ini (retry tick berikut), JANGAN mark ok.
				break
			}
			_ = store.MarkPromotedCognitive("node:"+n.ID, resp.ID, "ok")
			shared++
			log.Printf("[recovery-share] %s instinct %s → shared-brain (imunitas kolektif)", id, n.ID)
		}
		store.Close()
	}
	return shared
}
