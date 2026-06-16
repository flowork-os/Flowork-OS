// === LOCKED FILE (soft) === Status: STABLE — owner-approved 2026-06-16 (LOCKED ≠ FREEZE).
// AI lain: JANGAN otak-atik tanpa izin owner. Teruji live (TestLocalEmbedLive: dim 1024).
//
// Vendor: local — embed lewat server OpenAI-compat LOKAL (Ollama :11434 atau llama-server).
//
// TUJUAN (buat AI lain): embedder SOVEREIGN buat semantic-RAG brain (bge-m3). Query DAN drawer
// di-embed pakai ENGINE SAMA (llama.cpp di balik Ollama/llama-server) → vektor align → recall
// gak jebol diam-diam (= anti-halu; kalau embedder query beda dari yang bikin drawer, hasil
// retrieval ngaco tanpa error). Portable (ikut OS flashdisk), NO API key (lokal), NO cloud.
//
// Base URL di-resolve: req.BaseURL > env FLOWORK_LOCAL_EMBED_URL > default Ollama OpenAI-compat.
// Normalisasi L2 SENGAJA gak di sini — dilakukan sisi pemakai (brain) biar 1 tempat kontrol,
// konsisten antara index-build & query. Respons = OpenAI-shape (doEmbedRequest yang decode).
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func init() { Register(&localProvider{}) }

type localProvider struct{}

func (localProvider) Name() string { return "local" }

// LocalEmbedBaseURL — base URL embedder lokal (TANPA suffix /embeddings). Dipakai provider ini
// + pemakai lain yang mau tau endpoint embed lokal. Default = Ollama OpenAI-compat.
func LocalEmbedBaseURL(override string) string {
	if s := strings.TrimSpace(override); s != "" {
		return strings.TrimRight(s, "/")
	}
	if s := strings.TrimSpace(os.Getenv("FLOWORK_LOCAL_EMBED_URL")); s != "" {
		return strings.TrimRight(s, "/")
	}
	return "http://127.0.0.1:11434/v1"
}

// DefaultLocalEmbedModel — model embed lokal default (env FLOWORK_LOCAL_EMBED_MODEL > bge-m3).
// bge-m3: 1024-dim multilingual, dipake bikin drawer brain → WAJIB sama buat query.
func DefaultLocalEmbedModel() string {
	if s := strings.TrimSpace(os.Getenv("FLOWORK_LOCAL_EMBED_MODEL")); s != "" {
		return s
	}
	return "bge-m3"
}

func (localProvider) Embed(ctx context.Context, req Request) (*Result, error) {
	base := LocalEmbedBaseURL(req.BaseURL)
	model := defaultStr(req.Model, DefaultLocalEmbedModel())
	payload := map[string]any{"model": model, "input": req.Input}
	body, _ := json.Marshal(payload)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	// lokal: gak butuh API key. doEmbedRequest pakai embedHTTPClient (timeout 2m).
	return doEmbedRequest(r)
}
