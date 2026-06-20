// wakeup_engine.go — fire one-shot self-wakeups (the ScheduleWakeup tool).
//
// ScheduleWakeup writes a durable `wakeups` row (due_unix, prompt, reason) in the
// calling agent's own store, but it CANNOT fire itself: the WASM invoke model is
// synchronous, there is no per-agent timer running between turns. This kernel-side
// poller closes the loop. Each minute it scans every agent's store for due, unfired
// wakeups, re-invokes the agent on the saved prompt, delivers the reply to the owner
// via Telegram, and marks it fired. Mirrors the scheduler/trigger engines — reuses
// host.InvokeAgentMessage (action) + notifyOwnerTelegram (deliver). No WASM rebuild,
// no per-agent tick cost (kernel runs it once for all agents).
package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"flowork-gui/internal/kernelhost"
)

// RunDueWakeups fires every due, unfired wakeup across all agents. Returns count fired.
func RunDueWakeups(ctx context.Context, host *kernelhost.Host) int {
	now := time.Now().Unix()
	fired := 0
	for _, id := range host.AgentIDs() {
		store, err := host.OpenAgentStore(id)
		if err != nil {
			continue
		}
		db := store.DB()
		// Most agents never call ScheduleWakeup → no `wakeups` table. Skip cheaply
		// (don't pollute every agent DB with an empty table).
		var tbl string
		if db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='wakeups'").Scan(&tbl) != nil {
			store.Close()
			continue
		}
		type wk struct{ id, prompt, reason string }
		var due []wk
		rows, qerr := db.Query(
			"SELECT id, prompt, COALESCE(reason,'') FROM wakeups WHERE fired=0 AND due_unix<=? ORDER BY due_unix LIMIT 20", now)
		if qerr == nil {
			for rows.Next() {
				var w wk
				if rows.Scan(&w.id, &w.prompt, &w.reason) == nil {
					due = append(due, w)
				}
			}
			rows.Close()
		}
		for _, w := range due {
			// Mark fired FIRST: invoke bisa lama; mark dulu = anti double-ping. Ticker
			// di-serialize (Go ticker drop tick saat handler busy) → ga ada RunDueWakeups
			// overlap walau invoke panjang.
			if _, e := db.Exec("UPDATE wakeups SET fired=1 WHERE id=?", w.id); e != nil {
				continue
			}
			// LOCKED (soft, owner-approved 2026-06-20 anti-stuck): JANGAN turunin di bawah loopBudgetMs.
			// 300s = SAMA dengan turn normal (InvokeAgentMessage). FIX KRITIS (owner 2026-06-20):
			// dulu 90s < loopBudgetMs(200s) → tugas panjang yg di-RESUME via wakeup ke-cut SEBELUM
			// loop sempat checkpoint + jadwalin lanjutan → rantai auto-continue PUTUS = evolusi stuck
			// di tengah. Samain budget biar agent bisa nyelesain chunk + jadwalin wakeup berikutnya
			// (nyawa flowork: loop evolusi WAJIB nyambung lintas-wakeup, ga boleh mati di tengah).
			ictx, cancel := context.WithTimeout(ctx, 300*time.Second)
			reply, ierr := host.InvokeAgentMessage(ictx, id, w.prompt, "wakeup")
			cancel()
			if ierr != nil {
				reply = "(wakeup ke-fire tapi agent error: " + ierr.Error() + ")"
			}
			// Agent emits {"reply":"..."} — unwrap for a clean owner message.
			text := strings.TrimSpace(reply)
			var emitted map[string]any
			if json.Unmarshal([]byte(reply), &emitted) == nil {
				if rv, ok := emitted["reply"].(string); ok {
					text = strings.TrimSpace(rv)
				}
			}
			msg := "⏰ " + w.reason
			if text != "" {
				msg += "\n\n" + text
			}
			if nerr := notifyOwnerTelegram(ctx, msg); nerr != nil {
				log.Printf("[wakeup] notify owner gagal (%s): %v", id, nerr)
			}
			log.Printf("[wakeup] fired %s (agent %s)", w.id, id)
			fired++
		}
		store.Close()
	}
	return fired
}
