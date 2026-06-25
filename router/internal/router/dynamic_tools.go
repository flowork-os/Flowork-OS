// ⚠️ FROZEN 2026-06-26 — intent-gated tool filter (#9), stabil+Rule-9. Extend lewat ENV/GUI
// (FLOWORK_DYNAMIC_TOOLS/_TOPK/_MINSCORE), BUKAN edit file ini. lock/intent-gated-tools.md
package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/flowork-os/flowork_Router/internal/providers/embedding"
	"github.com/flowork-os/flowork_Router/internal/store"
)

// requestTool - representasi tool yang masuk dalam request
type requestTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

// maybeFilterTools - memfilter tools secara dinamis menggunakan embedding lokal
// envIntDefault / envFloatDefault — baca ENV (di-override GUI Switch Fitur lewat fwswitch),
// fallback default kalau kosong/invalid. Tunable runtime tanpa rebuild.
func envIntDefault(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// truthyEnv — true kalau ENV key = 1/on/true/yes.
func truthyEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "on", "true", "yes":
		return true
	}
	return false
}

func envFloatDefault(key string, def float64) float64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			return f
		}
	}
	return def
}

func maybeFilterTools(ctx context.Context, req OpenAIRequest, settings *store.Settings) OpenAIRequest {
	// 1. Cek switch. PRIMARY: FLOWORK_DYNAMIC_TOOLS (dikelola GUI Switch Fitur lewat fwswitch —
	// prefix FLOWORK_). FALLBACK: legacy FLOW_ROUTER_DYNAMIC_TOOLS. Default OFF.
	if !truthyEnv("FLOWORK_DYNAMIC_TOOLS") && os.Getenv("FLOW_ROUTER_DYNAMIC_TOOLS") != "1" {
		return req
	}

	// Jika tidak ada tools dalam request, skip
	if len(req.Tools) == 0 {
		return req
	}

	// 2. Parse tools dari request
	var toolsList []requestTool
	if err := json.Unmarshal(req.Tools, &toolsList); err != nil {
		log.Printf("flow_router dynamic_tools: fail to parse tools: %v", err)
		return req
	}
	if len(toolsList) == 0 {
		return req
	}

	// 3. Ambil query terakhir user
	query := lastUserText(req.Messages)
	if query == "" {
		return req
	}

	// 4. Deteksi tool yang pernah dipanggil (always-keep based on history) + ESCAPE-HATCH.
	// Escape-hatch WAJIB lolos: kalau tool yg dibutuhin ke-prune, model masih bisa NEMU/AMBIL
	// balik lewat tool_search/tool_lookup (#2C) → pruning jadi AMAN (anti salah-gate). +
	// StructuredOutput (kontrak output) + ScheduleWakeup (loop/wait) selalu ada.
	alwaysKeep := map[string]bool{
		"structured_output": true,
		"StructuredOutput":  true,
		"tool_search":       true,
		"tool_lookup":       true,
		"ScheduleWakeup":    true,
	}
	for _, msg := range req.Messages {
		if msg.Role == "tool" && msg.Name != "" {
			alwaysKeep[msg.Name] = true
		}
		if len(msg.ToolCalls) > 0 {
			var calls []struct {
				Function struct {
					Name string `json:"name"`
				} `json:"function"`
			}
			if err := json.Unmarshal(msg.ToolCalls, &calls); err == nil {
				for _, c := range calls {
					if c.Function.Name != "" {
						alwaysKeep[c.Function.Name] = true
					}
				}
			}
		}
	}

	// 5. Hitung embedding query user
	embedder := embedding.Get("local")
	if embedder == nil {
		log.Printf("flow_router dynamic_tools: local embedder not registered, fail-open")
		return req
	}

	queryEmbedding, err := embedQueryLocal(ctx, embedder, query)
	if err != nil {
		log.Printf("flow_router dynamic_tools: query embedding fail: %v, fail-open", err)
		return req
	}

	// 6. Hitung kemiripan untuk setiap tool
	type toolWithScore struct {
		Tool  requestTool
		Score float64
	}
	var scoredTools []toolWithScore
	var toolsToEmbed []string
	var toolsToEmbedIndices []int

	for i, t := range toolsList {
		if alwaysKeep[t.Function.Name] {
			scoredTools = append(scoredTools, toolWithScore{Tool: t, Score: 2.0}) // score 2.0 ensures it's always kept
			continue
		}
		desc := t.Function.Description
		if desc == "" {
			desc = t.Function.Name
		}
		toolsToEmbed = append(toolsToEmbed, desc)
		toolsToEmbedIndices = append(toolsToEmbedIndices, i)
	}

	// Batch embed deskripsi tool agar efisien
	if len(toolsToEmbed) > 0 {
		embeddings, err := embedBatchLocal(ctx, embedder, toolsToEmbed)
		if err != nil {
			log.Printf("flow_router dynamic_tools: batch embedding fail: %v, fail-open", err)
			return req
		}
		for idx, emb := range embeddings {
			origIdx := toolsToEmbedIndices[idx]
			sim := cosineSimilarity(queryEmbedding, emb)
			scoredTools = append(scoredTools, toolWithScore{
				Tool:  toolsList[origIdx],
				Score: sim,
			})
		}
	}

	// 7. Saring tools berdasarkan skor. Default lebih AMAN dari versi awal (top-K 5 kekecilan
	// buat orkestrator yg butuh tool beragam) → 12. Tunable RUNTIME tanpa rebuild (+GUI Switch
	// Fitur): FLOW_ROUTER_DYNAMIC_TOOLS_TOPK / _MINSCORE. Escape-hatch (always-keep) DI LUAR top-K.
	topK := envIntDefault("FLOWORK_DYNAMIC_TOOLS_TOPK", 12)
	threshold := envFloatDefault("FLOWORK_DYNAMIC_TOOLS_MINSCORE", 0.30)

	var filtered []requestTool
	// Masukkan yang alwaysKeep dulu
	for _, st := range scoredTools {
		if st.Score >= 2.0 {
			filtered = append(filtered, st.Tool)
		}
	}

	// Sort sisanya berdasarkan skor descending
	for i := 0; i < len(scoredTools); i++ {
		for j := i + 1; j < len(scoredTools); j++ {
			if scoredTools[i].Score < scoredTools[j].Score {
				scoredTools[i], scoredTools[j] = scoredTools[j], scoredTools[i]
			}
		}
	}

	// Tambahkan tool yang memenuhi syarat (skor tinggi atau masuk top-K)
	addedCount := 0
	for _, st := range scoredTools {
		if st.Score >= 2.0 {
			continue // sudah dimasukkan
		}
		if st.Score >= threshold && addedCount < topK {
			filtered = append(filtered, st.Tool)
			addedCount++
		}
	}

	// 8. Marshal kembali ke request
	filteredBytes, err := json.Marshal(filtered)
	if err != nil {
		log.Printf("flow_router dynamic_tools: marshal filtered tools fail: %v", err)
		return req
	}

	req.Tools = json.RawMessage(filteredBytes)
	log.Printf("flow_router dynamic_tools: filtered tools from %d down to %d", len(toolsList), len(filtered))
	return req
}

// embedQueryLocal - pembungkus lokal untuk embedder query
func embedQueryLocal(ctx context.Context, embedder embedding.EmbeddingProvider, input string) ([]float64, error) {
	res, err := embedder.Embed(ctx, embedding.Request{Input: []string{input}})
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 || len(res.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return res.Data[0].Embedding, nil
}

// embedBatchLocal - pembungkus lokal untuk batch embedding
func embedBatchLocal(ctx context.Context, embedder embedding.EmbeddingProvider, inputs []string) ([][]float64, error) {
	res, err := embedder.Embed(ctx, embedding.Request{Input: inputs})
	if err != nil {
		return nil, err
	}
	if len(res.Data) != len(inputs) {
		return nil, fmt.Errorf("embedding batch size mismatch")
	}
	embeddings := make([][]float64, len(inputs))
	for i, d := range res.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, nil
}

// cosineSimilarity - menghitung cosine similarity antara dua vektor float64
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
