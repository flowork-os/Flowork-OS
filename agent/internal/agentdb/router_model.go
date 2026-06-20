package agentdb

import "strings"

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

// LOCKED (soft, owner-approved 2026-06-20): sumber kebenaran model PER-AGENT.
// GetRouterModel — baca model per-agent (kv.router_model) yg di-set GUI Settings.
// "" kalau belum diset. Sumber kebenaran model per-agent (owner 2026-06-20: GUI, bukan env).
func (s *Store) GetRouterModel() string {
	kv, err := s.readKV("kv")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(kv["router_model"])
}
