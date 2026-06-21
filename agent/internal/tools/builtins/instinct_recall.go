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

func init() { tools.Register(&instinctRecallTool{}) }

const (
	instinctDefaultK = 6
	instinctMaxChars = 1400
)

type instinctRecallTool struct{}

func (instinctRecallTool) Name() string       { return "instinct_recall" }
func (instinctRecallTool) Capability() string { return "state:read" }
func (instinctRecallTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Tarik POLA INSTINCT coding+security (distilasi dari model kuat) yang relevan SEBELUM nulis/review code. Return fact-sheet ringkas 'WHEN trigger -> rule'. Pakai pas mulai task coding — apalagi yg nyentuh input/auth/network/crypto/kontrak (mata-hacker: sadar celah sebelum nulis).",
		Params: []tools.Param{
			{Name: "query", Type: tools.ParamString, Description: "deskripsi task coding / area kode (mis. 'parse user input ke SQL query')", Required: true},
			{Name: "k", Type: tools.ParamInt, Description: "max insting (default 6)", Required: false, Default: instinctDefaultK},
		},
		Returns: "{instincts: [\"WHEN ... -> ...\"], count, fact_sheet}",
	}
}

func (instinctRecallTool) Run(ctx context.Context, args map[string]any) (tools.Result, error) {
	store, ok := tools.FromStore(ctx)
	if !ok {
		return tools.Result{}, fmt.Errorf("agent store not in context")
	}
	query, _ := args["query"].(string)
	if strings.TrimSpace(query) == "" {
		return tools.Result{}, fmt.Errorf("query required")
	}
	k := instinctDefaultK
	switch v := args["k"].(type) {
	case float64:
		k = int(v)
	case int:
		k = v
	}
	if k <= 0 {
		k = instinctDefaultK
	}

	routerURL := routerclient.DefaultRouterURL
	if cfg, lerr := store.Load(); lerr == nil {
		if u, ok := cfg["router_url"].(string); ok && u != "" {
			routerURL = u
		}
	}

	ectx, cancel := context.WithTimeout(ctx, 30*time.Second)
	vec, eerr := routerclient.New(routerURL).EmbedText(ectx, "", query)
	cancel()
	if eerr != nil {
		return tools.Result{}, fmt.Errorf("instinct recall embed: %w", eerr)
	}

	var instincts []string
	var sb strings.Builder
	sb.WriteString("# Coding/security instincts — apply BEFORE writing code\n")
	for _, n := range store.SearchNodesByEmbedding("instinct", agentdb.Quantize(vec), k) {
		line := strings.TrimSpace(strings.ReplaceAll(n.Label, "\n", " "))
		if line == "" {
			continue
		}
		entry := "- " + line + "\n"
		if sb.Len()+len(entry) > instinctMaxChars {
			break
		}
		sb.WriteString(entry)
		instincts = append(instincts, line)
	}

	sheet := ""
	if len(instincts) > 0 {
		sheet = sb.String()
	}
	return tools.Result{Output: map[string]any{
		"instincts":  instincts,
		"count":      len(instincts),
		"fact_sheet": sheet,
	}}, nil
}
