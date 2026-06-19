// cognitive_embed_backfill.go — Phase 3C: backfill embedding ke node yg dibikin
// struktural tanpa embedding (tool/agent/code awareness nodes). File BARU
// (cognitive_graph.go locked). Embed-nya di-INJECT dari caller (agentdb gak import
// routerclient) — di sini cuma list + set.

package agentdb

// EmbedTarget — node yg butuh embedding (label + properties buat teks embed).
type EmbedTarget struct {
	ID         string
	Label      string
	Type       string
	Properties string
}

// NodesNeedingEmbedding — node active yg embedding-nya NULL (kerangka struktural
// belum di-embed). limit<=0 → 1000.
func (s *Store) NodesNeedingEmbedding(limit int) ([]EmbedTarget, error) {
	if limit <= 0 {
		limit = 1000
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	rows, err := s.db.Query(`SELECT id, label, type, properties FROM cognitive_nodes
		WHERE status='active' AND (embedding IS NULL OR LENGTH(embedding)=0) ORDER BY id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmbedTarget
	for rows.Next() {
		var t EmbedTarget
		if rows.Scan(&t.ID, &t.Label, &t.Type, &t.Properties) == nil {
			out = append(out, t)
		}
	}
	return out, rows.Err()
}

// SetNodeEmbedding — tulis embedding (quantized) + flip flag needs_embedding di
// properties (kalau ada). Idempoten.
func (s *Store) SetNodeEmbedding(id string, emb []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE cognitive_nodes
		SET embedding=?, properties=REPLACE(properties,'"needs_embedding":true','"needs_embedding":false')
		WHERE id=?`, emb, id)
	return err
}
