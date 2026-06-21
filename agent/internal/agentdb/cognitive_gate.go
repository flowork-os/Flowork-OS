// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

import (
	"fmt"
	"strings"
)

var FunctionalRelations = map[string]bool{
	"is_a": true, "decides_by": true, "located_in": true, "created_by": true,
	"goal_is": true, "communicates_in_style": true,
}

func GateStatus(text string, confidence float64, antibodies []string) (status, reason string) {
	if hit := matchAntibody(text, antibodies); hit != "" {
		return "quarantined", "antibody match: " + hit
	}
	if confidence < quarantineConfidenceFloor {
		return "quarantined", fmt.Sprintf("low confidence %.2f < %.2f", confidence, quarantineConfidenceFloor)
	}
	return "active", ""
}

func (s *Store) LoadAntibodyPatterns() ([]string, error) {
	_, _ = s.SeedAntibodies()
	return s.loadAntibodies()
}

func (s *Store) DetectEdgeContradiction(fromID, relationType, newToID string) (oldToID string, conflict bool) {
	if !FunctionalRelations[strings.TrimSpace(relationType)] {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()

	err := s.db.QueryRow(
		`SELECT to_id FROM cognitive_edges
		 WHERE from_id=? AND relation_type=? AND status='active' AND to_id<>? LIMIT 1`,
		fromID, relationType, newToID).Scan(&oldToID)
	if err != nil || oldToID == "" {
		return "", false
	}
	return oldToID, true
}

func (s *Store) RecordTension(fromID, relationType, oldToID, newToID, detail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	_, err := s.db.Exec(
		`INSERT INTO cognitive_tension (from_id, relation_type, old_to_id, new_to_id, detail)
		 VALUES (?,?,?,?,?)`, fromID, relationType, oldToID, newToID, detail)
	if err != nil {
		return fmt.Errorf("record tension: %w", err)
	}
	return nil
}

type CogTension struct {
	ID           int64  `json:"id"`
	FromID       string `json:"from_id"`
	RelationType string `json:"relation_type"`
	OldToID      string `json:"old_to_id"`
	NewToID      string `json:"new_to_id"`
	Detail       string `json:"detail"`
	Status       string `json:"status"`
}

func (s *Store) ListOpenTensions(limit int) ([]CogTension, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()

	rows, err := s.db.Query(
		`SELECT id, from_id, relation_type, old_to_id, new_to_id, detail, status
		 FROM cognitive_tension WHERE status='open' ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CogTension
	for rows.Next() {
		var tn CogTension
		if err := rows.Scan(&tn.ID, &tn.FromID, &tn.RelationType, &tn.OldToID, &tn.NewToID, &tn.Detail, &tn.Status); err != nil {
			return nil, err
		}
		out = append(out, tn)
	}
	return out, rows.Err()
}

func (s *Store) ResolveTension(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	_, err := s.db.Exec(`UPDATE cognitive_tension SET status='resolved' WHERE id=?`, id)
	return err
}
