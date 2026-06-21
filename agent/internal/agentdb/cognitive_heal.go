// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import "strings"

type SelfHealOpts struct {
	DecayFactor  float64
	MinStrength  float64
	ProtectTypes []string
}

type SelfHealStats struct {
	EdgesDecayed  int64
	EdgesPruned   int64
	NodesOrphaned int64
}

var defaultProtectedTypes = []string{"person", "trait", "preference", "persona", "doctrine"}

func (s *Store) SelfHeal(opt SelfHealOpts) (SelfHealStats, error) {
	if opt.DecayFactor <= 0 || opt.DecayFactor > 1 {
		opt.DecayFactor = 0.95
	}
	if opt.MinStrength <= 0 {
		opt.MinStrength = 0.1
	}
	protect := opt.ProtectTypes
	if len(protect) == 0 {
		protect = defaultProtectedTypes
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()

	var st SelfHealStats

	if r, err := s.db.Exec(`UPDATE cognitive_edges SET strength = strength * ? WHERE status='active'`, opt.DecayFactor); err == nil {
		st.EdgesDecayed, _ = r.RowsAffected()
	}

	if r, err := s.db.Exec(`DELETE FROM cognitive_edges WHERE strength < ?`, opt.MinStrength); err == nil {
		st.EdgesPruned, _ = r.RowsAffected()
	}

	ph := strings.TrimRight(strings.Repeat("?,", len(protect)), ",")
	args := make([]any, len(protect))
	for i, t := range protect {
		args[i] = t
	}
	q := `DELETE FROM cognitive_nodes
	      WHERE id NOT IN (SELECT from_id FROM cognitive_edges)
	        AND id NOT IN (SELECT to_id FROM cognitive_edges)
	        AND type NOT IN (` + ph + `)`
	if r, err := s.db.Exec(q, args...); err == nil {
		st.NodesOrphaned, _ = r.RowsAffected()
	}
	return st, nil
}
