// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

type EmbedTarget struct {
	ID         string
	Label      string
	Type       string
	Properties string
}

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

func (s *Store) SetNodeEmbedding(id string, emb []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE cognitive_nodes
		SET embedding=?, properties=REPLACE(properties,'"needs_embedding":true','"needs_embedding":false')
		WHERE id=?`, emb, id)
	return err
}
