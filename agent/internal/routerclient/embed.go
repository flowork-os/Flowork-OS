// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package routerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultEmbedModel = "bge-m3"

type embedReq struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResp struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) EmbedText(ctx context.Context, model, text string) ([]float32, error) {
	if c == nil {
		return nil, fmt.Errorf("router client nil")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("text kosong")
	}
	if model == "" {
		model = DefaultEmbedModel
	}

	body, err := json.Marshal(embedReq{Model: model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out embedResp
	if uerr := json.Unmarshal(raw, &out); uerr != nil {
		return nil, fmt.Errorf("decode (status %d): %w", resp.StatusCode, uerr)
	}
	if resp.StatusCode >= 400 {
		if out.Error != nil {
			return nil, fmt.Errorf("router %d: %s", resp.StatusCode, out.Error.Message)
		}
		return nil, fmt.Errorf("router status %d", resp.StatusCode)
	}
	if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding (model %q — configured di Settings?)", model)
	}

	vec := make([]float32, len(out.Data[0].Embedding))
	for i, f := range out.Data[0].Embedding {
		vec[i] = float32(f)
	}
	return vec, nil
}
