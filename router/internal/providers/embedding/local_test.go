package embedding

import (
	"context"
	"math"
	"os"
	"testing"
)

func TestLocalBaseURL(t *testing.T) {
	if got := LocalEmbedBaseURL("http://x:1/v1/"); got != "http://x:1/v1" {
		t.Errorf("override: %s", got)
	}
	os.Setenv("FLOWORK_LOCAL_EMBED_URL", "http://env:2/v1/")
	if got := LocalEmbedBaseURL(""); got != "http://env:2/v1" {
		t.Errorf("env: %s", got)
	}
	os.Unsetenv("FLOWORK_LOCAL_EMBED_URL")
	if got := LocalEmbedBaseURL(""); got != "http://127.0.0.1:11434/v1" {
		t.Errorf("default: %s", got)
	}
}

func TestLocalRegistered(t *testing.T) {
	if Get("local") == nil {
		t.Fatal("provider 'local' not registered")
	}
}

// TestLocalEmbedLive — integrasi: butuh Ollama bge-m3 lokal. Skip kalau gak ada (portable).
func TestLocalEmbedLive(t *testing.T) {
	p := localProvider{}
	res, err := p.Embed(context.Background(), Request{Input: []string{"honeypot scam token detection"}})
	if err != nil {
		t.Skipf("embedder lokal unreachable (ollama + bge-m3?): %v", err)
	}
	if len(res.Data) != 1 || len(res.Data[0].Embedding) != 1024 {
		t.Fatalf("want 1x1024 dim, got %d x %d", len(res.Data), len(res.Data[0].Embedding))
	}
	var ss float64
	for _, v := range res.Data[0].Embedding {
		ss += v * v
	}
	if ss == 0 {
		t.Fatal("zero embedding")
	}
	t.Logf("local bge-m3 embed OK: dim=%d norm=%.2f", len(res.Data[0].Embedding), math.Sqrt(ss))
}
