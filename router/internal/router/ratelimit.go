// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without explicit owner (Mr.Dev) approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-04
//
// ⚠️ PERINGATAN UNTUK AI MANAPUN (TERMASUK PASCA-COMPACT): jangan lemahin/buang
//   tanpa Mr.Dev minta eksplisit. Ini yang bikin armada agent ga kena 429 pas
//   nembak Claude barengan.
//
// ratelimit.go — ANTI-429: "agent jalan bergantian".
//
// MASALAH: subscription Claude rate-limit-nya didesain buat 1 user interaktif,
//   BUKAN armada agent. Pas crew + mr-flow nembak barengan (burst) → 429 →
//   dulu router langsung NYERAH (lock provider + "all providers exhausted").
//
// SOLUSI (2 lapis):
//   1. SEMAPHORE — max N request ke provider barengan (sisanya ANTRI = bergantian).
//   2. BACKOFF-RETRY — pas 429, TUNGGU (exponential) lalu ulang provider SAMA,
//      bukan langsung gagal. 429 jadi "pelan-pelan", bukan "mati".
//
// Tunable: env FLOW_ROUTER_MAX_CONCURRENCY (default 3).

package router

import (
	"os"
	"strconv"
	"time"
)

const (
	// 429 = throttle ritme jangka-pendek (BUKAN kuota — kuota subscription masih
	// banyak). Throttle kadang 1-2 menit, jadi retry sampai 6x (2+4+8+16+30+30 ≈
	// 90s) biar ke-ride, ga nyerah kecepetan. Kuota aman → nunggu lama ga rugi.
	maxRateLimitRetries        = 6
	defaultDispatchConcurrency = 3               // max request Claude barengan
	rateLimitBackoffCap        = 30 * time.Second // batas atas tiap tunggu
)

// claudeSem — semaphore "bergantian": cuma N forwardToProvider boleh in-flight
// barengan; sisanya block (antri) sampai ada slot. Size di-set sekali saat init.
var claudeSem = make(chan struct{}, dispatchConcurrency())

func dispatchConcurrency() int {
	if v := os.Getenv("FLOW_ROUTER_MAX_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultDispatchConcurrency
}

// acquireDispatchSlot/releaseDispatchSlot — dipakai forwardToProvider biar
// request ke provider antri rapi. Slot dilepas SEBELUM backoff-sleep (di loop
// dispatcher) supaya yang antri bisa jalan pas satu lagi nunggu rate-limit.
func acquireDispatchSlot() { claudeSem <- struct{}{} }
func releaseDispatchSlot() { <-claudeSem }

// backoffDuration — exponential: attempt 0→2s, 1→4s, 2→8s, 3→16s, cap 30s.
// Cukup buat burst (429 klaster detik-detikan) yang jadi akar masalah.
func backoffDuration(attempt int) time.Duration {
	d := time.Duration(2<<uint(attempt)) * time.Second
	if d > rateLimitBackoffCap {
		d = rateLimitBackoffCap
	}
	return d
}
