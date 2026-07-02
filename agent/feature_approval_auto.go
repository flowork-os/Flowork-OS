// feature_approval_auto.go — AUTO-APPROVE (owner 2026-07-02: "buat automatis biar
// ga perlu approve"). NON-frozen sibling (deletable). Poller cepet: pas
// FLOWORK_APPROVAL_AUTO ON, approve SEMUA antrian approval pending lintas-agent
// otomatis → agent ga ke-block, ga ada notif numpuk. Approved berlaku 1 jam per
// tool+args (mekanisme frozen), jadi combo yg sama abis itu lolos langsung.
//
// KENAPA perlu (walau mode bypass): gate sensitive-args (state.db) + sensitive-tool
// di sandbox_v3 (FROZEN) tetap bikin pending TANPA peduli FLOWORK_APPROVAL_MODE.
// Auto-approver ini yg beneran bikin "automatis" buat owner. Matiin:
// FLOWORK_APPROVAL_AUTO=0 (balik manual approve di GUI).
// 📄 Dok: FLowork_os/lock/approval-gate.md
package main

import (
	"context"
	"os"
	"strings"
	"time"

	"flowork-gui/internal/kernelhost"
)

func approvalAutoEnabled() bool {
	v := strings.TrimSpace(os.Getenv("FLOWORK_APPROVAL_AUTO"))
	return v == "" || v == "1" || strings.EqualFold(v, "true") // default ON (owner mau automatis)
}

func init() {
	RegisterFeature(Feature{Name: "approval-auto", Phase: PhaseSeed, Apply: func(d *Deps) {
		if d.Host == nil {
			return
		}
		go approvalAutoLoop(d.Ctx, d.Host)
	}})
}

func approvalAutoLoop(ctx context.Context, host *kernelhost.Host) {
	t := time.NewTicker(6 * time.Second) // lebih cepet dari notify (60s) → pending kelar sebelum notif
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if !approvalAutoEnabled() {
				continue
			}
			for _, id := range host.AgentIDs() {
				store, err := host.OpenAgentStore(id)
				if err != nil {
					continue
				}
				// skip agent tanpa tabel (pola wakeup_engine) — anti polusi.
				var tbl string
				if store.DB().QueryRow(
					"SELECT name FROM sqlite_master WHERE type='table' AND name='approval_queue'").
					Scan(&tbl) != nil {
					store.Close()
					continue
				}
				rows, lerr := store.ListApprovalQueue("pending", 100)
				if lerr == nil {
					for _, r := range rows {
						_ = store.DecideApproval(r.ID, "approved", "auto")
					}
				}
				store.Close()
			}
		}
	}
}
