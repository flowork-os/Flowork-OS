// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"regexp"
	"strings"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/routerclient"
)

const promoteMinHit = 3

var promoteBrandRe = regexp.MustCompile(`(?i)\b(claude|anthropic|fable|gemini|opus|sonnet|haiku|chatgpt|openai)\b`)

func PromoteRecurringMistakes(ctx context.Context, host *kernelhost.Host) int {
	rc := routerclient.New("")
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
		eligible, lerr := store.ListMistakesEligibleForPromote(promoteMinHit, 50)
		if lerr != nil || len(eligible) == 0 {
			store.Close()
			continue
		}
		scope := "agent:" + id
		for _, m := range eligible {
			label := recoveryLabel(m)
			if promoteBrandRe.MatchString(label) {

				_ = store.SetMistakePromoted(m.ID, 1)
				continue
			}
			ectx, cancel := context.WithTimeout(ctx, 30*time.Second)
			vec, eerr := rc.EmbedText(ectx, "", label)
			cancel()
			if eerr != nil {
				continue
			}
			h := fnv.New64a()
			h.Write([]byte(strings.ToLower(label)))
			hid := h.Sum64()
			nodeID := fmt.Sprintf("%s/instinct/recov%016x", scope, hid)
			if _, ue := store.UpsertNode(agentdb.CogNode{
				ID: nodeID, Label: label, Type: "instinct",
				WhereDomain: "recovery", SourceKind: "verified", SourceRef: fmt.Sprintf("recov%016x", hid),
				Confidence: 0.85, Status: "active", Embedding: agentdb.Quantize(vec),
			}); ue != nil {
				continue
			}
			if perr := store.SetMistakePromoted(m.ID, int64(hid>>1)); perr != nil {
				log.Printf("[mistake-promote] set promoted gagal (%s id=%d): %v", id, m.ID, perr)
			}
			promoted++
			log.Printf("[mistake-promote] %s mistake#%d (hit=%d) → recovery-instinct", id, m.ID, m.HitCount)
		}
		store.Close()
	}
	return promoted
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
