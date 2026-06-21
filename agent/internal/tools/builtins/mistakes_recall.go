// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package builtins

import (
	"context"
	"fmt"
	"strings"

	"flowork-gui/internal/tools"
)

type mistakeRecallTool struct{}

func (mistakeRecallTool) Name() string       { return "mistake_recall" }
func (mistakeRecallTool) Capability() string { return "state:read" }
func (mistakeRecallTool) Schema() tools.Schema {
	return tools.Schema{
		Description: "Cek apakah lo PERNAH salah di konteks mirip (recall mistakes journal lo sendiri). Panggil SEBELUM ngerjain hal beresiko / yang pernah bermasalah, biar ga ngulang error yang sama. Balik daftar 'dulu lo salah X (Nx), solusinya Y' urut paling sering keulang.",
		Params: []tools.Param{
			{Name: "context", Type: tools.ParamString, Description: "deskripsi singkat situasi/tugas sekarang (kata kunci)", Required: true},
			{Name: "limit", Type: tools.ParamInt, Description: "max hasil (default 5, max 20)"},
		},
		Returns: "{count, warnings:[{title, remediation, hit_count, category}]}",
	}
}

func (mistakeRecallTool) Run(ctx context.Context, args map[string]any) (tools.Result, error) {
	store, ok := tools.FromStore(ctx)
	if !ok {
		return tools.Result{}, fmt.Errorf("agent store not in context")
	}
	q, _ := args["context"].(string)
	if strings.TrimSpace(q) == "" {
		return tools.Result{}, fmt.Errorf("context required")
	}
	limit := 0
	switch v := args["limit"].(type) {
	case float64:
		limit = int(v)
	case int:
		limit = v
	}
	hits, err := store.SearchMistakes(q, limit)
	if err != nil {
		return tools.Result{}, fmt.Errorf("mistake_recall: %w", err)
	}
	warnings := make([]map[string]any, 0, len(hits))
	for _, m := range hits {
		warnings = append(warnings, map[string]any{
			"title":       m.Title,
			"remediation": m.Content,
			"hit_count":   m.HitCount,
			"category":    m.Category,
		})
	}
	note := ""
	if len(warnings) > 0 {
		note = "ada riwayat kesalahan mirip — baca remediation biar ga ngulang"
	}
	return tools.Result{Output: map[string]any{
		"count":    len(warnings),
		"warnings": warnings,
	}, Note: note}, nil
}
