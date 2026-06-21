// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package builtins

import (
	"context"
	"fmt"

	"flowork-gui/internal/routerclient"
	"flowork-gui/internal/tools"
)

const (
	defaultBrainSearchK = 5
	maxBrainSearchK     = 10
)

type brainSearchTool struct{}

func (brainSearchTool) Name() string       { return "brain_search_shared" }
func (brainSearchTool) Capability() string { return "rpc:router:brain" }
func (brainSearchTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Cari di korpus pengetahuan SHARED di Router (5jt drawers: security/training/dll) via BM25/FTS. Buat pengetahuan LUAS yang bukan pengalaman pribadi lo. Remote (butuh router up). Buat brain PRIBADI lo, pakai brain_search (lokal).",
		Params: []tools.Param{
			{Name: "query", Type: tools.ParamString, Description: "search query (natural language atau keyword)", Required: true},
			{Name: "k", Type: tools.ParamInt, Description: "max hits (default 5, max 10)", Required: false, Default: defaultBrainSearchK},
		},
		Returns: "{query, hits: [{wing, room, content, score, drawer_id}], count}",
	}
}

func (brainSearchTool) Run(ctx context.Context, args map[string]any) (tools.Result, error) {
	store, ok := tools.FromStore(ctx)
	if !ok {
		return tools.Result{}, fmt.Errorf("agent store not in context")
	}

	query, _ := args["query"].(string)
	if query == "" {
		return tools.Result{}, fmt.Errorf("query required")
	}

	k := defaultBrainSearchK
	switch v := args["k"].(type) {
	case float64:
		k = int(v)
	case int:
		k = v
	}
	if k <= 0 {
		k = defaultBrainSearchK
	}
	if k > maxBrainSearchK {
		k = maxBrainSearchK
	}

	routerURL := routerclient.DefaultRouterURL
	if cfg, lerr := store.Load(); lerr == nil {
		if u, ok := cfg["router_url"].(string); ok && u != "" {
			routerURL = u
		}
	}

	client := routerclient.New(routerURL)
	resp, err := client.SearchBrain(ctx, query, k)
	if err != nil {
		return tools.Result{}, fmt.Errorf("search brain: %w", err)
	}
	return tools.Result{Output: map[string]any{
		"query": resp.Query,
		"hits":  resp.Hits,
		"count": resp.Count,
	}}, nil
}
