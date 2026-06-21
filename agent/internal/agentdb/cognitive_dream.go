// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"context"
	"fmt"
	"strings"
)

type EmbedFunc func(ctx context.Context, text string) ([]float32, error)

type LLMFunc func(ctx context.Context, prompt string) (string, error)

type DigestDeps struct {
	LLM        LLMFunc
	Embed      EmbedFunc
	AgentScope string
	Tier       int

	SourceKind string

	SourceRef string
}

type DigestStats struct {
	NodesAdded  int
	EdgesAdded  int
	Quarantined int
	Tensions    int
	Dropped     int
}

func (s *Store) DigestText(ctx context.Context, text string, dep DigestDeps) (DigestStats, error) {
	var st DigestStats
	if dep.LLM == nil {
		return st, fmt.Errorf("DigestDeps.LLM wajib")
	}
	if strings.TrimSpace(text) == "" {
		return st, nil
	}
	scope := dep.AgentScope
	if scope == "" {
		scope = "agent:local"
	}
	srcRef := dep.SourceRef
	if srcRef == "" {
		srcRef = "digest"
	}

	skOf := func(extracted string) string {
		if dep.SourceKind != "" {
			return dep.SourceKind
		}
		return extracted
	}

	raw, err := dep.LLM(ctx, BuildExtractPrompt(text))
	if err != nil {
		return st, fmt.Errorf("llm extract: %w", err)
	}
	res, err := ParseExtraction(raw)
	if err != nil {
		return st, fmt.Errorf("parse: %w", err)
	}
	st.Dropped = len(res.Dropped)

	antibodies, _ := s.LoadAntibodyPatterns()

	labelToID := map[string]string{}
	for _, n := range res.Nodes {
		id, emb := s.resolveNodeID(ctx, dep, scope, n.Label, n.Type)
		status, reason := GateStatus(n.Label+" "+n.Why, n.Confidence, antibodies)
		if dep.Tier <= 1 && status == "active" {
			status = "shadow"
		}
		node := CogNode{
			ID: id, Label: n.Label, Type: n.Type, Why: n.Why, Who: n.Who,
			WhereDomain: n.WhereDomain, WhenValid: n.WhenValid,
			SourceKind: skOf(n.SourceKind), SourceRef: srcRef, Confidence: n.Confidence,
			Status: status, Embedding: emb,
		}
		if status == "quarantined" {
			node.Status = "quarantined"
		}
		added, uerr := s.UpsertNode(node)
		if uerr != nil {
			continue
		}
		if status == "quarantined" {
			st.Quarantined++
			_ = s.setNodeQuarantineReason(id, reason)
		}
		if added {
			st.NodesAdded++
		}
		labelToID[strings.ToLower(n.Label)] = id
	}

	for _, e := range res.Edges {
		fromID := s.edgeEndpointID(ctx, dep, scope, labelToID, e.FromLabel)
		toID := s.edgeEndpointID(ctx, dep, scope, labelToID, e.ToLabel)
		if fromID == "" || toID == "" {
			continue
		}
		if old, conflict := s.DetectEdgeContradiction(fromID, e.RelationType, toID); conflict {
			_ = s.RecordTension(fromID, e.RelationType, old, toID,
				fmt.Sprintf("digest: %s already -[%s]-> %s, now -> %s", fromID, e.RelationType, old, toID))
			st.Tensions++
			continue
		}
		status := "active"
		if dep.Tier <= 1 {
			status = "shadow"
		}
		if uerr := s.UpsertEdge(CogEdge{
			FromID: fromID, ToID: toID, RelationType: e.RelationType,
			Confidence: e.Confidence, SourceKind: skOf(e.SourceKind), SourceRef: srcRef, Status: status,
		}); uerr == nil {
			st.EdgesAdded++
		}
	}
	return st, nil
}

func (s *Store) resolveNodeID(ctx context.Context, dep DigestDeps, scope, label, typ string) (string, []byte) {
	var q []byte

	if id, ok := s.resolveCanonicalIdentity(scope, label, typ); ok {
		if dep.Embed != nil {
			if vec, err := dep.Embed(ctx, label); err == nil {
				q = Quantize(vec)
			}
		}
		return id, q
	}
	if dep.Embed != nil {
		if vec, err := dep.Embed(ctx, label); err == nil {
			q = Quantize(vec)
			if id, _, found := s.ResolveByEmbedding(typ, q, 0); found {
				return id, q
			}
		}
	}
	return scope + "/" + typ + "/" + slug(label), q
}

func (s *Store) edgeEndpointID(ctx context.Context, dep DigestDeps, scope string, labelToID map[string]string, label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}
	if id, ok := labelToID[strings.ToLower(label)]; ok {
		return id
	}
	id, emb := s.resolveNodeID(ctx, dep, scope, label, "concept")
	_, _ = s.UpsertNode(CogNode{ID: id, Label: label, Type: "concept", Status: "shadow", SourceRef: "digest-edge", Embedding: emb})
	labelToID[strings.ToLower(label)] = id
	return id
}

func (s *Store) setNodeQuarantineReason(id, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE cognitive_nodes SET reason_quarantine=? WHERE id=?`, reason, id)
	return err
}

func (s *Store) DigestPendingInteractions(ctx context.Context, dep DigestDeps, limit int) (DigestStats, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	type row struct {
		id      int64
		content string
	}
	s.mu.Lock()
	s.ensureCognitiveGraphSchema()
	rows, err := s.db.Query(`
		SELECT i.id, i.content FROM interactions i
		LEFT JOIN cognitive_digest_log d ON d.interaction_id = i.id
		WHERE d.interaction_id IS NULL AND i.deleted_at IS NULL AND TRIM(i.content) <> ''
		ORDER BY i.id ASC LIMIT ?`, limit)
	if err != nil {
		s.mu.Unlock()
		return DigestStats{}, 0, err
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.content); err == nil {
			pending = append(pending, r)
		}
	}
	rows.Close()
	s.mu.Unlock()

	var total DigestStats
	var b strings.Builder
	var ids []int64
	for _, r := range pending {
		b.WriteString(r.content)
		b.WriteString("\n")
		ids = append(ids, r.id)
	}
	if len(pending) == 0 {
		return total, 0, nil
	}

	st, derr := s.DigestText(ctx, b.String(), dep)
	if derr != nil {
		return total, 0, derr
	}
	total = st

	s.mu.Lock()
	for _, id := range ids {
		_, _ = s.db.Exec(
			`INSERT OR IGNORE INTO cognitive_digest_log (interaction_id, nodes_added, edges_added, status)
			 VALUES (?,?,?, 'ok')`, id, st.NodesAdded, st.EdgesAdded)
	}
	s.mu.Unlock()
	return total, len(ids), nil
}

func (s *Store) PromoteShadows(minHits int) (int, error) {
	if minHits < 2 {
		minHits = 2
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	rn, _ := s.db.Exec(`UPDATE cognitive_nodes SET status='active' WHERE status='shadow' AND hit_count>=?`, minHits)
	re, _ := s.db.Exec(`UPDATE cognitive_edges SET status='active' WHERE status='shadow' AND strength>=?`, float64(minHits))
	n, e := int64(0), int64(0)
	if rn != nil {
		n, _ = rn.RowsAffected()
	}
	if re != nil {
		e, _ = re.RowsAffected()
	}
	return int(n + e), nil
}

func slug(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	var b strings.Builder
	prevDash := false
	for _, r := range label {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "x"
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return out
}
