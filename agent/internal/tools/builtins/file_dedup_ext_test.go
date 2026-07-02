// file_dedup_ext_test.go — bukti dedup file-read (gape1 §C).
package builtins

import (
	"context"
	"strings"
	"testing"

	"flowork-gui/internal/tools"
)

func dedupCtx() context.Context {
	return tools.WithAgent(context.Background(), "test-agent")
}

func resetDedup() {
	fileDedupMu.Lock()
	fileDedupSeen = map[string]fileDedupEntry{}
	fileDedupMu.Unlock()
}

func TestFileDedup_FirstReadFull_SecondStub(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "")
	resetDedup()
	ctx := dedupCtx()
	content := strings.Repeat("x", 2000)
	// baca #1 → full (no stub)
	if _, ok := fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 2000, content); ok {
		t.Fatal("baca pertama harus FULL, bukan stub")
	}
	// baca #2, mtime+size sama → stub
	out, ok := fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 2000, content)
	if !ok {
		t.Fatal("baca kedua unchanged harus STUB")
	}
	if out["unchanged"] != true {
		t.Fatal("stub harus nandain unchanged=true")
	}
	head, _ := out["content_head"].(string)
	if len([]rune(head)) != fileDedupHeadRune {
		t.Fatalf("head harus %d rune, dapet %d", fileDedupHeadRune, len([]rune(head)))
	}
}

func TestFileDedup_SmallFileAlwaysFull(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "")
	resetDedup()
	ctx := dedupCtx()
	small := strings.Repeat("y", fileDedupMinRune) // <= ambang → full terus
	fileReadDedup(ctx, map[string]any{}, "s.txt", 111, 10, small)
	if _, ok := fileReadDedup(ctx, map[string]any{}, "s.txt", 111, 10, small); ok {
		t.Fatal("file kecil harus SELALU full (stub+note lebih gede dari isi)")
	}
}

func TestFileDedup_MtimeChanged_Full(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "")
	resetDedup()
	ctx := dedupCtx()
	fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 10, "isi")
	if _, ok := fileReadDedup(ctx, map[string]any{}, "a.txt", 222, 10, "isi2"); ok {
		t.Fatal("mtime berubah harus FULL")
	}
}

func TestFileDedup_ForceBypass(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "")
	resetDedup()
	ctx := dedupCtx()
	fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 10, "isi")
	if _, ok := fileReadDedup(ctx, map[string]any{"force": true}, "a.txt", 111, 10, "isi"); ok {
		t.Fatal("force=true harus FULL walau unchanged")
	}
}

func TestFileDedup_SwitchOff(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "off")
	resetDedup()
	ctx := dedupCtx()
	fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 10, "isi")
	if _, ok := fileReadDedup(ctx, map[string]any{}, "a.txt", 111, 10, "isi"); ok {
		t.Fatal("switch OFF harus selalu FULL")
	}
}

func TestFileDedup_PerAgentIsolated(t *testing.T) {
	t.Setenv("FLOWORK_FILE_DEDUP", "")
	resetDedup()
	a := tools.WithAgent(context.Background(), "agent-a")
	b := tools.WithAgent(context.Background(), "agent-b")
	fileReadDedup(a, map[string]any{}, "a.txt", 111, 10, "isi")
	if _, ok := fileReadDedup(b, map[string]any{}, "a.txt", 111, 10, "isi"); ok {
		t.Fatal("agent beda harus FULL (cache per-agent)")
	}
}
