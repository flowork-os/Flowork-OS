package agentdb

// router_model.go — set model router per-agent (kv.router_model) TANPA efek samping,
// pola sama SetPrompt (locked ensureSchema/Save ga disentuh). Dipakai seed
// self-evolution: tiap otak dewan pakai strong-cloud model. Idempotent (upsert).
func (s *Store) SetRouterModel(model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		"INSERT INTO kv(k, v) VALUES('router_model', ?) ON CONFLICT(k) DO UPDATE SET v=excluded.v",
		model,
	)
	return err
}
