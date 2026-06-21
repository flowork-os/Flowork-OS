// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentmgr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"flowork-gui/internal/httpx"
	"flowork-gui/internal/routerclient"
)

const learningStoreAgent = "mr-flow"
const learningScope = "agent:" + learningStoreAgent

type learningRecord struct {
	ID        int64  `json:"id"`
	Model     string `json:"model"`
	Response  string `json:"response"`
	Agent     string `json:"agent"`
	BuildPass int64  `json:"build_pass"`
}

type LearningStats struct {
	Fetched   int `json:"fetched"`
	Processed int `json:"processed"`
	Skipped   int `json:"skipped"`
	Added     int `json:"added"`
	Promoted  int `json:"promoted"`
}

func isStrongModelForLearning(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		return false
	}
	return !strings.Contains(m, "flowork-brain") && !strings.Contains(m, "flowork")
}

func fetchRouterRecordings(limit int) ([]learningRecord, error) {
	base := routerclient.New(strings.TrimSpace(os.Getenv("ROUTER_DEFAULT_URL"))).BaseURL
	url := base + "/api/recordings?include_body=1&limit=" + itoaSmallLearn(limit)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	var out struct {
		Items []learningRecord `json:"items"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func DigestRecordings(limit int) (LearningStats, error) {
	var ls LearningStats
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	recs, err := fetchRouterRecordings(limit)
	if err != nil {
		return ls, err
	}
	ls.Fetched = len(recs)
	store, serr := openAgentStore(learningStoreAgent)
	if serr != nil {
		return ls, serr
	}
	defer store.Close()

	deps := buildDigestDeps(learningScope, 1)
	deps.SourceKind = "strong_model_unverified"
	deps.SourceRef = "learning"

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	for _, rc := range recs {
		if store.LearningRecordSeen(rc.ID) {
			continue
		}
		if !isStrongModelForLearning(rc.Model) || strings.TrimSpace(rc.Response) == "" {
			_ = store.MarkLearningRecord(rc.ID)
			ls.Skipped++
			continue
		}
		d := deps
		if rc.BuildPass == 1 {
			d.Tier = 2
		}
		st, derr := store.DigestText(ctx, rc.Response, d)
		if derr == nil {
			ls.Added += st.NodesAdded
		}
		_ = store.MarkLearningRecord(rc.ID)
		ls.Processed++
	}

	if n, _ := store.PromoteShadows(2); n > 0 {
		ls.Promoted = n
	}
	return ls, nil
}

func LearningDigestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteJSON(w, map[string]any{"error": "method not allowed (POST)"})
		return
	}
	limit := 0
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, e := strconv.Atoi(s); e == nil {
			limit = n
		}
	}
	ls, err := DigestRecordings(limit)
	if err != nil {
		httpx.WriteJSON(w, map[string]any{"error": err.Error()})
		return
	}
	httpx.WriteJSON(w, map[string]any{"ok": true, "stats": ls})
}

func itoaSmallLearn(n int) string {
	if n <= 0 {
		return "100"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
