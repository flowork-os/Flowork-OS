package builtins

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"flowork-gui/internal/tools"
)

func resetEditStats() {
	editStatMu.Lock()
	editStats = map[string]*editStat{}
	editStatMu.Unlock()
}

func TestLineDiffStatLCS(t *testing.T) {
	cases := []struct {
		old, new           string
		wantAdd, wantRemov int
	}{
		{"", "a\nb\n", 2, 0},                     // file baru 2 baris
		{"a\nb\nc\n", "", 0, 3},                  // hapus semua
		{"a\nb\nc\n", "a\nX\nc\n", 1, 1},         // 1 baris ganti
		{"a\nb\nc\n", "a\nc\n", 0, 1},            // hapus 1 baris
		{"a\nc\n", "a\nb\nc\n", 1, 0},            // sisip 1 baris
		{"a\nb\nc\n", "a\nb\nc\n", 0, 0},         // sama = 0
	}
	for i, c := range cases {
		a, r := lineDiffStat(c.old, c.new)
		if a != c.wantAdd || r != c.wantRemov {
			t.Errorf("case %d: got +%d/-%d, want +%d/-%d", i, a, r, c.wantAdd, c.wantRemov)
		}
	}
}

func TestLinesChangedInterceptorAccumulates(t *testing.T) {
	resetEditStats()
	dir := t.TempDir()
	ctx := tools.WithAgent(tools.WithSharedDir(context.Background(), dir), "agent-x")

	// file_write file baru 3 baris → added 3.
	if err := (linesChangedInterceptor{}).Before(ctx, fileWriteTool{}, map[string]any{
		"file_path": "a.go", "content": "package a\n\nfunc F(){}\n",
	}); err != nil {
		t.Fatal(err)
	}
	// Tulis file itu ke disk (biar edit berikutnya punya isi lama).
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n\nfunc F(){}\n"), 0o644)

	// edit: ganti 'func F(){}' jadi 2 baris → added 1, removed 1 (net LCS).
	if err := (linesChangedInterceptor{}).Before(ctx, editTool{}, map[string]any{
		"file_path": "a.go", "old_string": "func F(){}", "new_string": "func F() {\n\treturn\n}",
	}); err != nil {
		t.Fatal(err)
	}

	total, perAgent := EditStatsSnapshot()
	if total["added"] < 3 {
		t.Errorf("total added = %d, harusnya >=3", total["added"])
	}
	if total["edits"] != 2 {
		t.Errorf("edits = %d, harusnya 2", total["edits"])
	}
	ax := perAgent["agent-x"]
	if ax == nil || ax["files"] != 1 {
		t.Errorf("agent-x files = %v, harusnya 1 file", ax)
	}
}

func TestLinesChangedIgnoresNonEditTools(t *testing.T) {
	resetEditStats()
	ctx := tools.WithAgent(tools.WithSharedDir(context.Background(), t.TempDir()), "a")
	// bash bukan tool edit → ga dihitung.
	_ = (linesChangedInterceptor{}).Before(ctx, bashTool{}, map[string]any{"command": "ls"})
	total, _ := EditStatsSnapshot()
	if total["edits"] != 0 {
		t.Errorf("tool non-edit harusnya ga keitung, edits=%d", total["edits"])
	}
}
