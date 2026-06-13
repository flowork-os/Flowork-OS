package router

import (
	"encoding/json"
	"testing"
)

func respFromJSON(t *testing.T, raw string) *OpenAIResponse {
	t.Helper()
	var r OpenAIResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return &r
}

// Feedback detection: kategori non-kanonik kedeteksi, kanonik lewat, non-task_run aman.
func TestDetectNonCanonicalTaskRun_HaluDetected(t *testing.T) {
	resp := respFromJSON(t, `{"choices":[{"message":{"tool_calls":[
		{"type":"function","function":{"name":"task_run","arguments":"{\"category\":\"analysis\",\"subject\":\"BBCA\"}"}}
	]}}]}`)
	if got := detectNonCanonicalTaskRun(resp); got != "analysis" {
		t.Fatalf("harusnya deteksi 'analysis', dapet %q", got)
	}
}

func TestDetectNonCanonicalTaskRun_CanonicalClean(t *testing.T) {
	resp := respFromJSON(t, `{"choices":[{"message":{"tool_calls":[
		{"type":"function","function":{"name":"task_run","arguments":"{\"category\":\"saham\",\"subject\":\"BBCA\"}"}}
	]}}]}`)
	if got := detectNonCanonicalTaskRun(resp); got != "" {
		t.Fatalf("kategori kanonik 'saham' harusnya bersih, dapet %q", got)
	}
}

func TestDetectNonCanonicalTaskRun_NonTaskRunIgnored(t *testing.T) {
	resp := respFromJSON(t, `{"choices":[{"message":{"tool_calls":[
		{"type":"function","function":{"name":"web_search","arguments":"{\"category\":\"whatever\"}"}}
	]}}]}`)
	if got := detectNonCanonicalTaskRun(resp); got != "" {
		t.Fatalf("tool non-task_run harusnya diabaikan, dapet %q", got)
	}
}

func TestDetectNonCanonicalTaskRun_NoToolCalls(t *testing.T) {
	resp := respFromJSON(t, `{"choices":[{"message":{"content":"halo bro"}}]}`)
	if got := detectNonCanonicalTaskRun(resp); got != "" {
		t.Fatalf("tanpa tool_calls harusnya bersih, dapet %q", got)
	}
	if got := detectNonCanonicalTaskRun(nil); got != "" {
		t.Fatalf("nil resp harusnya bersih, dapet %q", got)
	}
}
