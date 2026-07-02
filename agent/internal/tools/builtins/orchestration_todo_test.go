package builtins

import (
	"context"
	"path/filepath"
	"testing"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/tools"
)

func todoCtx(t *testing.T) context.Context {
	t.Helper()
	st, err := agentdb.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return tools.WithStore(context.Background(), st)
}

func run(t *testing.T, ctx context.Context, args map[string]any) []todoItem {
	t.Helper()
	res, err := (todoTool{}).Run(ctx, args)
	if err != nil {
		t.Fatalf("todo %v: %v", args, err)
	}
	return res.Output.(map[string]any)["items"].([]todoItem)
}

func TestTodo_InProgressSingleActive(t *testing.T) {
	ctx := todoCtx(t)
	run(t, ctx, map[string]any{"op": "add", "content": "tugas A"})
	run(t, ctx, map[string]any{"op": "add", "content": "tugas B"})

	items := run(t, ctx, map[string]any{"op": "list"})
	if len(items) != 2 || items[0].Status != "pending" {
		t.Fatalf("add harus status pending: %+v", items)
	}

	// doing t1 → in_progress
	items = run(t, ctx, map[string]any{"op": "doing", "id": "t1"})
	if items[0].Status != "in_progress" {
		t.Fatalf("t1 harus in_progress: %+v", items[0])
	}
	// doing t2 → t2 in_progress, t1 balik pending (rule §3: 1 in-progress).
	items = run(t, ctx, map[string]any{"op": "doing", "id": "t2"})
	var s1, s2 string
	for _, it := range items {
		if it.ID == "t1" {
			s1 = it.Status
		}
		if it.ID == "t2" {
			s2 = it.Status
		}
	}
	if s1 != "pending" || s2 != "in_progress" {
		t.Fatalf("cuma 1 in-progress: t1=%q t2=%q", s1, s2)
	}

	// done t2 → status done + Done bool (kompat).
	items = run(t, ctx, map[string]any{"op": "done", "id": "t2"})
	for _, it := range items {
		if it.ID == "t2" && (it.Status != "done" || !it.Done) {
			t.Fatalf("done harus set status+bool: %+v", it)
		}
	}
}

func TestTodo_BackfillLegacyStatus(t *testing.T) {
	ctx := todoCtx(t)
	st, _ := tools.FromStore(ctx)
	// Simulasi data LAMA: cuma Done bool, tanpa Status.
	_ = st.SetToolMemory(keyTodo, `[{"id":"t1","content":"lama done","done":true},{"id":"t2","content":"lama pending","done":false}]`)
	items, err := loadTodos(st)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	for _, it := range items {
		m[it.ID] = it.Status
	}
	if m["t1"] != "done" || m["t2"] != "pending" {
		t.Fatalf("backfill status salah: %+v", m)
	}
}
