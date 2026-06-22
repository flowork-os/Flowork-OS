package agentdb_test

// live_local_digest_test.go — BUKTI EMPIRIS (owner 2026-06-22): compact pakai model LOKAL
// (flowork-brain) beneran bisa digest pengalaman ke brain + trim, lewat chunking. Test ini
// HIT model lokal sungguhan via router :2402, jadi GATED env FLOWORK_LIVE_DIGEST=1 — ga ikut
// `go test ./...` biasa (suite frozen tetap hermetic). Pakai temp DB → ga sentuh data live.
//
// Jalanin:  FLOWORK_LIVE_DIGEST=1 go test ./internal/agentdb/ -run TestLiveLocalDigestCompact -v

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/routerclient"
)

func TestLiveLocalDigestCompact(t *testing.T) {
	if os.Getenv("FLOWORK_LIVE_DIGEST") != "1" {
		t.Skip("set FLOWORK_LIVE_DIGEST=1 (butuh router :2402 + flowork-brain :8088)")
	}

	s, err := agentdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	// Seed interaksi realistis (Indonesia, faktual → harusnya ke-extract jadi node). Total
	// ~12k char → > chunkCharBudget (6000) → MAKSA chunking jalan (>=2 chunk).
	facts := []string{
		"Owner namanya Aola, panggil dia dengan hormat. Dia owner project Flowork.",
		"Flowork itu sistem brain buat AI: cognitive graph, embedding bge-m3, recall 3 lapis.",
		"Owner minta semua roadmap kelar, stabil, lalu di-freeze biar bisa jalan tanpa AI.",
		"Model lokal Flowork namanya flowork-brain, jalan di port 8088 lewat llama-server.",
		"Auto-compact: kalau interaksi numpuk, agent digest pengalaman ke brain lalu trim.",
		"Chunking dipakai biar model ga choke pas konteks gede — digest per 6000 karakter.",
		"Default model digest compact itu lokal, biar gratis dan jalan tanpa langganan cloud.",
		"Router Flowork jalan di port 2402, routing model flowork-brain ke llama lokal 8088.",
	}
	const repeat = 4 // 8 fakta x 4 = 32 interaksi, cukup buat beberapa chunk
	seeded := 0
	for r := 0; r < repeat; r++ {
		for i, f := range facts {
			content := fmt.Sprintf("[sesi %d] %s (catatan tambahan biar panjang: poin %d penting buat memori jangka panjang organisme Flowork supaya recall-nya akurat).", r, f, i)
			if _, err := s.LogInteraction("chat", "in", "owner", content, nil); err != nil {
				t.Fatalf("log interaction: %v", err)
			}
			seeded++
		}
	}

	live0, undig0, chars0, _, err := s.CompactStats()
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	t.Logf("SEEDED: live=%d undigested=%d chars=%d", live0, undig0, chars0)

	// Wire deps PERSIS kayak buildDigestDepsModel(model="") → model LOKAL flowork-brain.
	rc := routerclient.New("") // default :2402
	deps := agentdb.DigestDeps{
		AgentScope: "agent:test-local",
		Tier:       2,
		LLM: func(ctx context.Context, prompt string) (string, error) {
			c, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			return rc.ChatComplete(c, "flowork-brain", prompt, 4096)
		},
		Embed: func(ctx context.Context, text string) ([]float32, error) {
			c, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			return rc.EmbedText(c, "", text)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	stats, marked, derr := s.DigestPendingInteractions(ctx, deps, 100)
	t.Logf("DIGEST(lokal): marked=%d nodes=%d edges=%d quarantined=%d tensions=%d err=%v",
		marked, stats.NodesAdded, stats.EdgesAdded, stats.Quarantined, stats.Tensions, derr)

	if derr != nil {
		t.Fatalf("DIGEST LOKAL GAGAL: %v (model lokal ga sanggup hasilin JSON valid — chunk perlu lebih kecil / model perlu lebih kapabel)", derr)
	}
	if marked == 0 {
		t.Fatalf("0 interaksi ke-mark digested — chunking/ digest ga jalan")
	}
	if stats.NodesAdded == 0 {
		t.Fatalf("0 node masuk brain — extraction lokal kosong")
	}

	// VERIFY: ga ada sisa undigested (mirip gate AutoCompactAgent sebelum trim).
	_, undig1, _, _, _ := s.CompactStats()
	t.Logf("ABIS DIGEST: undigested_sisa=%d", undig1)

	// TRIM: sisain 5 terbaru. Harus > 0 (banyak yg udah ke-brain).
	trimmed, terr := s.TrimDigestedInteractions(5)
	if terr != nil {
		t.Fatalf("trim: %v", terr)
	}
	t.Logf("TRIM: trimmed=%d (keepRecent=5)", trimmed)
	if trimmed == 0 {
		t.Fatalf("0 di-trim padahal udah di-digest — trim ga jalan")
	}

	live2, _, _, _, _ := s.CompactStats()
	t.Logf("HASIL AKHIR: live %d → %d (trim %d), node masuk brain %d ✓", live0, live2, trimmed, stats.NodesAdded)
}
