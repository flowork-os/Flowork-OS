// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package main

import (
	"context"
	"hash/fnv"
	"log"
	"strings"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/routerclient"
)

const promoteMinHit = 3

// promoteRecoveryShadowMinHit — pola coarse recovery harus BERULANG segini (hit_count
// naik via dedup-embedding lintas recovery beda) baru shadow→active. Gerbang anti-
// degenerasi (KUNCI roadmap #1): instinct dari LLM-coarsen ga langsung aktif.
const promoteRecoveryShadowMinHit = 2

// PromoteRecurringMistakes — D32 INC-1+INC-3. Tiap tick (1 menit, ticker non-beku):
//
//	(1) raw recovery-mistake hit>=3 → GENERALISASI (Lapis A strip + Lapis B coarsen) →
//	    recovery-instinct SHADOW (logic di agentdb.GeneralizeRecovery).
//	(2) GERBANG: recovery-instinct SHADOW yg pola-nya berulang (hit>=N) → ACTIVE.
//
// Gerbang sengaja di SINI (ticker non-beku), BUKAN nyandar autodigest (default OFF) —
// kalau enggak, shadow nyangkut selamanya & ga pernah ke-recall.
func PromoteRecurringMistakes(ctx context.Context, host *kernelhost.Host) int {
	rc := routerclient.New("")
	embedFn := func(c context.Context, text string) ([]float32, error) {
		ec, cancel := context.WithTimeout(c, 30*time.Second)
		defer cancel()
		return rc.EmbedText(ec, "", text)
	}
	llmFn := dreamDigestLLM(host) // Lapis B coarsen via dream-digester (model Haiku, bukan 26B lokal)

	promoted := 0
	for _, id := range host.AgentIDs() {
		store, err := host.OpenAgentStore(id)
		if err != nil {
			continue
		}

		var tbl string
		if store.DB().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='mistakes_local'").Scan(&tbl) != nil {
			store.Close()
			continue
		}
		scope := "agent:" + id

		if eligible, lerr := store.ListMistakesEligibleForPromote(promoteMinHit, 50); lerr == nil {
			for _, m := range eligible {
				label := recoveryLabel(m)
				if agentdb.ContainsBrand(label) { // pre-check white-label (raw)
					_ = store.SetMistakePromoted(m.ID, 1)
					continue
				}
				// INC-3: strip(A) → coarsen(B) → SHADOW instinct. classKey (kelas-error stabil)
				// = sumbu konvergensi DETERMINISTIK (lintas tool) → hit naik → gerbang firable.
				gctx, cancel := context.WithTimeout(ctx, 90*time.Second)
				nodeID, ok, gerr := store.GeneralizeRecovery(gctx, scope, recoveryClassKey(m), label, embedFn, llmFn)
				cancel()
				if gerr != nil {
					continue // transien (embed/llm) → retry tick berikut, JANGAN tandai promoted
				}
				if !ok { // drop (kosong/brand/leak post-coarsen) → tandai biar ga di-proses ulang
					_ = store.SetMistakePromoted(m.ID, 1)
					continue
				}
				h := fnv.New64a()
				_, _ = h.Write([]byte(strings.ToLower(nodeID)))
				if perr := store.SetMistakePromoted(m.ID, int64(h.Sum64()>>1)); perr != nil {
					log.Printf("[mistake-promote] set promoted gagal (%s id=%d): %v", id, m.ID, perr)
				}
				promoted++
				log.Printf("[mistake-promote] %s mistake#%d (hit=%d) → recovery-instinct SHADOW", id, m.ID, m.HitCount)
			}
		}

		// GERBANG INC-3: shadow→active recovery yg pola-nya berulang. Dijalanin TIAP tick
		// per-agent (walau ga ada eligible baru) biar shadow dari tick lalu ga nyangkut.
		if n, perr := store.PromoteRecoveryShadows(promoteRecoveryShadowMinHit); perr == nil && n > 0 {
			promoted += n
			log.Printf("[mistake-promote] %s %d recovery-instinct shadow→active (gerbang repetisi)", id, n)
		}
		store.Close()
	}
	return promoted
}

// recoveryClassKey — kelas-error stabil (sumbu generalisasi) dari title INC-2
// "recovery: <tool>/<class>". Whitelist kelas (toolErrClass) biar title aneh ga jadi
// kunci. Kosong = mistake non-recovery → GeneralizeRecovery pakai hash-konten (no konvergen).
func recoveryClassKey(m agentdb.Mistake) string {
	t := strings.ToLower(strings.TrimSpace(m.Title))
	i := strings.LastIndex(t, "/")
	if i < 0 || i+1 >= len(t) {
		return ""
	}
	switch cls := strings.TrimSpace(t[i+1:]); cls {
	case "not-found", "permission", "timeout", "already-exists", "invalid-input", "blocked", "error":
		return cls
	default:
		return ""
	}
}

func recoveryLabel(m agentdb.Mistake) string {
	content := strings.TrimSpace(m.Content)
	title := strings.TrimSpace(m.Title)

	if strings.HasPrefix(strings.ToUpper(content), "WHEN ") && strings.Contains(content, "->") {
		return trimLen(content, 400)
	}

	lesson := content
	if lesson == "" {
		lesson = title
	}
	return trimLen("WHEN "+title+" -> "+lesson, 400)
}

func trimLen(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
