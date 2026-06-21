// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentmgr

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/floworkdb"
	"flowork-gui/internal/httpx"
	"flowork-gui/internal/routerclient"
)

const cgmDigestMaxTokens = 4096

func cgmRouterClient() *routerclient.Client {
	return routerclient.New(strings.TrimSpace(os.Getenv("ROUTER_DEFAULT_URL")))
}

func cgmModel() string {
	return floworkdb.DefaultModelShared()
}

func buildDigestDeps(scope string, tier int) agentdb.DigestDeps {
	rc := cgmRouterClient()
	model := cgmModel()
	return agentdb.DigestDeps{
		AgentScope: scope,
		Tier:       tier,
		LLM: func(ctx context.Context, prompt string) (string, error) {

			if DigestLLMOverride != nil {
				if out, err := DigestLLMOverride(ctx, prompt); err == nil && strings.TrimSpace(out) != "" {
					return out, nil
				}
			}
			c, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			return rc.ChatComplete(c, model, prompt, cgmDigestMaxTokens)
		},
		Embed: func(ctx context.Context, text string) ([]float32, error) {
			c, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			return rc.EmbedText(c, "", text)
		},
	}
}

func DigestAgent(agentID string, tier int) (agentdb.DigestStats, int, error) {
	store, err := openAgentStore(agentID)
	if err != nil {
		return agentdb.DigestStats{}, 0, err
	}
	defer store.Close()

	if tier < 1 || tier > 2 {
		tier = 2
	}
	scope := "agent:" + agentID
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	stats, n, derr := store.DigestPendingInteractions(ctx, buildDigestDeps(scope, tier), 100)
	if derr != nil {
		return stats, n, derr
	}
	if tier >= 2 {

		_, _ = store.PromoteShadows(2)
	}
	return stats, n, nil
}

func DigestAllAgents(agentIDs []string) int {
	if strings.TrimSpace(os.Getenv("FLOWORK_CGM_AUTODIGEST")) != "1" {
		return 0
	}
	total := 0
	for _, id := range agentIDs {
		func() {
			defer func() { _ = recover() }()
			if _, n, err := DigestAgent(id, 2); err == nil {
				total += n
			}
		}()
	}
	return total
}

func CognitiveDigestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteJSON(w, map[string]any{"error": "method not allowed (POST)"})
		return
	}
	agentID := strings.TrimSpace(r.URL.Query().Get("id"))
	if agentID == "" {
		httpx.WriteJSON(w, map[string]any{"error": "agent id required"})
		return
	}
	tier := parseLimitOr(r.URL.Query().Get("tier"), 2)
	stats, n, err := DigestAgent(agentID, tier)
	if err != nil {
		httpx.WriteJSON(w, map[string]any{"error": err.Error(), "digested": n})
		return
	}
	httpx.WriteJSON(w, map[string]any{
		"ok": true, "agent": agentID, "tier": tier, "digested": n,
		"nodes_added": stats.NodesAdded, "edges_added": stats.EdgesAdded,
		"quarantined": stats.Quarantined, "tensions": stats.Tensions, "dropped": stats.Dropped,
	})
}
