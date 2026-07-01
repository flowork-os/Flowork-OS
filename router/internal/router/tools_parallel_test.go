package router

import (
	"encoding/json"
	"testing"
)

// TestMergeConsecutiveAnthropic membuktikan AKAR fix parallel-tool-calls:
// beberapa tool_result beruntun (parallel) HARUS jadi 1 user message berisi
// array tool_result — bukan N user message beruntun (yg bikin Anthropic 400).
func TestMergeConsecutiveAnthropic(t *testing.T) {
	in := []map[string]any{
		{"role": "user", "content": "kerjain 3 hal"},
		{"role": "assistant", "content": []map[string]any{
			{"type": "tool_use", "id": "a"},
			{"type": "tool_use", "id": "b"},
			{"type": "tool_use", "id": "c"},
		}},
		{"role": "user", "content": []map[string]any{{"type": "tool_result", "tool_use_id": "a"}}},
		{"role": "user", "content": []map[string]any{{"type": "tool_result", "tool_use_id": "b"}}},
		{"role": "user", "content": []map[string]any{{"type": "tool_result", "tool_use_id": "c"}}},
	}
	out := mergeConsecutiveAnthropic(in)

	// Role WAJIB alternate: user, assistant, user (3 pesan, bukan 5).
	if len(out) != 3 {
		t.Fatalf("expected 3 merged messages, got %d", len(out))
	}
	roles := []string{"user", "assistant", "user"}
	for i, want := range roles {
		if out[i]["role"] != want {
			t.Fatalf("msg %d role = %v, want %s", i, out[i]["role"], want)
		}
	}
	// Pesan user terakhir HARUS berisi 3 blok tool_result.
	blocks, ok := out[2]["content"].([]map[string]any)
	if !ok || len(blocks) != 3 {
		t.Fatalf("last user message should hold 3 tool_result blocks, got %#v", out[2]["content"])
	}
	for _, b := range blocks {
		if b["type"] != "tool_result" {
			t.Fatalf("block type = %v, want tool_result", b["type"])
		}
	}
}

// TestMergeUserTextAfterToolResult — teks user setelah tool_result (mis. flail
// nudge) juga digabung ke user message yg sama (string → text block).
func TestMergeUserTextAfterToolResult(t *testing.T) {
	in := []map[string]any{
		{"role": "assistant", "content": []map[string]any{{"type": "tool_use", "id": "a"}}},
		{"role": "user", "content": []map[string]any{{"type": "tool_result", "tool_use_id": "a"}}},
		{"role": "user", "content": "koreksi: jangan ulang tool yg sama"},
	}
	out := mergeConsecutiveAnthropic(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(out))
	}
	blocks, ok := out[1]["content"].([]map[string]any)
	if !ok || len(blocks) != 2 {
		t.Fatalf("expected tool_result + text block, got %#v", out[1]["content"])
	}
	if blocks[0]["type"] != "tool_result" || blocks[1]["type"] != "text" {
		t.Fatalf("unexpected block order: %#v", blocks)
	}
	// Pastikan seluruh body tetap marshalable (JSON valid buat Anthropic).
	if _, err := json.Marshal(out); err != nil {
		t.Fatalf("merged messages not marshalable: %v", err)
	}
}

// TestMergeSingleToolUnchanged — kasus umum (1 tool/turn) TIDAK berubah bentuk:
// tetap 1 user message berisi 1 tool_result (block array asli dipertahankan).
func TestMergeSingleToolUnchanged(t *testing.T) {
	in := []map[string]any{
		{"role": "user", "content": "halo"},
		{"role": "assistant", "content": []map[string]any{{"type": "tool_use", "id": "a"}}},
		{"role": "user", "content": []map[string]any{{"type": "tool_result", "tool_use_id": "a"}}},
	}
	out := mergeConsecutiveAnthropic(in)
	if len(out) != 3 {
		t.Fatalf("expected 3 messages unchanged, got %d", len(out))
	}
}
