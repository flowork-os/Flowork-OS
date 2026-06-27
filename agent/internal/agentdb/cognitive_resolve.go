// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"math"
	"os"
	"strconv"
	"strings"
)

const DefaultResolveThreshold = 0.86

// resolveThreshold — ambang similaritas entity-resolution (dedup node graph). SWITCH
// FLOWORK_RESOLVE_MINSCORE (0..1), default 0.86. Terlalu rendah → entitas beda ke-merge.
func resolveThreshold() float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("FLOWORK_RESOLVE_MINSCORE")), 64); err == nil && v > 0 && v <= 1 {
		return v
	}
	return DefaultResolveThreshold
}

func Quantize(vec []float32) []byte {
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return nil
	}
	out := make([]byte, len(vec))
	for i, v := range vec {
		q := math.Round(float64(v) / norm * 127)
		if q > 127 {
			q = 127
		} else if q < -127 {
			q = -127
		}
		out[i] = byte(int8(q))
	}
	return out
}

func CosineQ(a, b []byte) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot int64
	for i := range a {
		dot += int64(int8(a[i])) * int64(int8(b[i]))
	}
	return float64(dot) / (127.0 * 127.0)
}

func (s *Store) ResolveByEmbedding(typ string, queryEmb []byte, threshold float64) (id string, score float64, found bool) {
	if len(queryEmb) == 0 {
		return "", 0, false
	}
	if threshold <= 0 {
		threshold = resolveThreshold()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()

	rows, err := s.db.Query(
		`SELECT id, embedding FROM cognitive_nodes
		 WHERE type=? AND status='active' AND embedding IS NOT NULL`, typ)
	if err != nil {
		return "", 0, false
	}
	defer rows.Close()

	bestScore := -1.0
	bestID := ""
	for rows.Next() {
		var nid string
		var emb []byte
		if err := rows.Scan(&nid, &emb); err != nil {
			continue
		}
		sc := CosineQ(queryEmb, emb)
		if sc > bestScore {
			bestScore, bestID = sc, nid
		}
	}
	if bestID != "" && bestScore >= threshold {
		return bestID, bestScore, true
	}
	return "", bestScore, false
}
