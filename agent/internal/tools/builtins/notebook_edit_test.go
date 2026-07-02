package builtins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"flowork-gui/internal/tools"
)

func writeSampleNB(t *testing.T, dir string) string {
	t.Helper()
	nb := map[string]any{
		"cells": []any{
			map[string]any{"cell_type": "markdown", "metadata": map[string]any{}, "source": []any{"# Judul\n"}},
			map[string]any{"cell_type": "code", "metadata": map[string]any{}, "execution_count": float64(3),
				"outputs": []any{map[string]any{"output_type": "stream", "text": []any{"hasil lama\n"}}},
				"source":  []any{"print('lama')\n"}, "id": "cell-abc"},
		},
		"metadata":       map[string]any{"kernelspec": map[string]any{"name": "python3"}},
		"nbformat":       float64(4),
		"nbformat_minor": float64(5),
	}
	b, _ := json.MarshalIndent(nb, "", " ")
	p := filepath.Join(dir, "nb.ipynb")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func readNB(t *testing.T, p string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var nb map[string]any
	if err := json.Unmarshal(raw, &nb); err != nil {
		t.Fatalf("hasil bukan JSON valid: %v", err)
	}
	return nb
}

func TestNotebookEditRegistered(t *testing.T) {
	found := false
	for _, n := range tools.ListNames() {
		if n == "notebook_edit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("notebook_edit ga ke-register di papan tools (init() plug-and-play gagal)")
	}
}

func TestNotebookEditReplaceResetsOutputsAndKeepsMeta(t *testing.T) {
	dir := t.TempDir()
	writeSampleNB(t, dir)
	ctx := tools.WithSharedDir(context.Background(), dir)

	_, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "nb.ipynb", "edit_mode": "replace", "cell_index": float64(1),
		"new_source": "print('baru')\nx = 1\n",
	})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	nb := readNB(t, filepath.Join(dir, "nb.ipynb"))
	// nbformat fields harus kepreserve (anti-korup).
	if nb["nbformat"].(float64) != 4 || nb["nbformat_minor"].(float64) != 5 {
		t.Errorf("nbformat metadata ilang: %v / %v", nb["nbformat"], nb["nbformat_minor"])
	}
	if _, ok := nb["metadata"].(map[string]any)["kernelspec"]; !ok {
		t.Error("kernelspec metadata ilang")
	}
	cell := nb["cells"].([]any)[1].(map[string]any)
	if got := nbSourceToString(cell["source"]); got != "print('baru')\nx = 1\n" {
		t.Errorf("source ga keupdate: %q", got)
	}
	// Code cell diedit → output & execution_count reset.
	if outs, _ := cell["outputs"].([]any); len(outs) != 0 {
		t.Errorf("outputs harus reset kosong, dapet %v", outs)
	}
	if cell["execution_count"] != nil {
		t.Errorf("execution_count harus null, dapet %v", cell["execution_count"])
	}
}

func TestNotebookEditInsertAndDelete(t *testing.T) {
	dir := t.TempDir()
	writeSampleNB(t, dir)
	ctx := tools.WithSharedDir(context.Background(), dir)

	// Insert markdown di posisi 1.
	if _, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "nb.ipynb", "edit_mode": "insert", "cell_index": float64(1),
		"cell_type": "markdown", "new_source": "## Sisipan\n",
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	nb := readNB(t, filepath.Join(dir, "nb.ipynb"))
	cells := nb["cells"].([]any)
	if len(cells) != 3 {
		t.Fatalf("harusnya 3 cell, dapet %d", len(cells))
	}
	if ct := cells[1].(map[string]any)["cell_type"]; ct != "markdown" {
		t.Errorf("cell[1] harusnya markdown, dapet %v", ct)
	}

	// Delete by cell_id (cell-abc, code cell asli).
	if _, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "nb.ipynb", "edit_mode": "delete", "cell_id": "cell-abc",
	}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	nb = readNB(t, filepath.Join(dir, "nb.ipynb"))
	cells = nb["cells"].([]any)
	if len(cells) != 2 {
		t.Fatalf("harusnya 2 cell abis delete, dapet %d", len(cells))
	}
	for _, c := range cells {
		if id, _ := c.(map[string]any)["id"].(string); id == "cell-abc" {
			t.Error("cell-abc harusnya udah kehapus")
		}
	}
}

func TestNotebookEditRejectsNonIpynbAndBadRange(t *testing.T) {
	dir := t.TempDir()
	writeSampleNB(t, dir)
	os.WriteFile(filepath.Join(dir, "x.txt"), []byte("bukan notebook"), 0o644)
	ctx := tools.WithSharedDir(context.Background(), dir)

	if _, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "x.txt", "edit_mode": "replace", "cell_index": float64(0), "new_source": "y",
	}); err == nil {
		t.Error("harus nolak non-.ipynb")
	}
	if _, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "nb.ipynb", "edit_mode": "replace", "cell_index": float64(99), "new_source": "y",
	}); err == nil {
		t.Error("harus nolak cell_index di luar rentang")
	}
	// Isolasi: path '..' harus ditolak resolver.
	if _, err := (notebookEditTool{}).Run(ctx, map[string]any{
		"file_path": "../escape.ipynb", "edit_mode": "insert", "new_source": "z",
	}); err == nil {
		t.Error("harus nolak file_path '..' (workspace escape)")
	}
}
