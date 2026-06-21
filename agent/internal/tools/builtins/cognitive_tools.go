// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package builtins

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/routerclient"
	"flowork-gui/internal/tools"
)

type graphRecallTool struct{}

func (graphRecallTool) Name() string       { return "graph_recall" }
func (graphRecallTool) Capability() string { return "state:read" }
func (graphRecallTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Recall grounding dari Cognitive Graph LOKAL (twin) — tarik subgraph relevan jadi fact-sheet ringkas (budget-capped). Pakai buat 'apa yang gw tau soal X' / 'gimana A nyambung ke B'. Beda dari brain_search (cari teks FTS): ini paham RELASI antar-entitas (Aola→prefers→X, dst).",
		Params: []tools.Param{
			{Name: "query", Type: tools.ParamString, Description: "topik / pertanyaan", Required: true},
			{Name: "max_chars", Type: tools.ParamInt, Description: "budget fact-sheet (default 1500)"},
		},
		Returns: "{query, fact_sheet, chars}",
	}
}

func (graphRecallTool) Run(ctx context.Context, args map[string]any) (tools.Result, error) {
	store, ok := tools.FromStore(ctx)
	if !ok {
		return tools.Result{}, fmt.Errorf("agent store not in context")
	}
	query, _ := args["query"].(string)
	if strings.TrimSpace(query) == "" {
		return tools.Result{}, fmt.Errorf("query required")
	}
	maxChars := 0
	switch v := args["max_chars"].(type) {
	case float64:
		maxChars = int(v)
	case int:
		maxChars = v
	}

	routerURL := routerclient.DefaultRouterURL
	if cfg, lerr := store.Load(); lerr == nil {
		if u, ok := cfg["router_url"].(string); ok && u != "" {
			routerURL = u
		}
	}
	rc := routerclient.New(routerURL)
	embed := func(ec context.Context, text string) ([]float32, error) {
		c, cancel := context.WithTimeout(ec, 30*time.Second)
		defer cancel()
		return rc.EmbedText(c, "", text)
	}
	sheet, err := store.RecallFactSheet(ctx, query, agentdb.RecallDeps{Embed: embed, MaxChars: maxChars})
	if err != nil {
		return tools.Result{}, fmt.Errorf("graph_recall: %w", err)
	}
	return tools.Result{Output: map[string]any{
		"query":      query,
		"fact_sheet": sheet,
		"chars":      len(sheet),
	}}, nil
}
