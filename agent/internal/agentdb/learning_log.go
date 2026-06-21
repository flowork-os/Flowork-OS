// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package agentdb

func (s *Store) ensureLearningLogSchema() {
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS learning_record_log (
		recording_id INTEGER PRIMARY KEY,
		processed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
}

func (s *Store) LearningRecordSeen(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLearningLogSchema()
	var x int
	return s.db.QueryRow(`SELECT 1 FROM learning_record_log WHERE recording_id=?`, id).Scan(&x) == nil
}

func (s *Store) MarkLearningRecord(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLearningLogSchema()
	_, err := s.db.Exec(`INSERT OR IGNORE INTO learning_record_log(recording_id) VALUES(?)`, id)
	return err
}

func (s *Store) CountLearningProcessed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLearningLogSchema()
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM learning_record_log`).Scan(&n)
	return n
}
