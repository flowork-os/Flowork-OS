// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval (autonomy grant 2026-06-19).
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-19
// Reason: CGM self-heal (decay/prune/orphan, protect identity) — built + unit-tested. Extend = new file.
//
// cognitive_heal.go — Self-Healing buat cognitive graph (roadmap §4.7).
//
// Maintenance yang jalan di Tier-2 dream: (1) decay strength edge (yang ga
// diperkuat lama → memudar), (2) prune edge mati (strength < floor), (3) buang
// node yatim (ga punya edge). TAPI LINDUNGI node identitas (person/trait/
// preference/persona/doctrine) — jangan buang "Aola" cuma gara-gara edge memudar.
//
// Plug-and-play: file baru, operasi tabel cognitive_* (cognitive_graph.go LOCKED,
// di-extend lewat sini, BUKAN dimodif). Idempotent, aman dipanggil cron.

package agentdb

import "strings"

// SelfHealOpts — parameter maintenance (default aman).
type SelfHealOpts struct {
	DecayFactor  float64  // kali strength tiap heal (default 0.95)
	MinStrength  float64  // edge di bawah ini di-prune (default 0.1)
	ProtectTypes []string // type node yang JANGAN dibuang walau yatim
}

// SelfHealStats — hasil 1 pass.
type SelfHealStats struct {
	EdgesDecayed  int64
	EdgesPruned   int64
	NodesOrphaned int64
}

// defaultProtectedTypes — identitas owner/agent: jangan hilang walau yatim.
var defaultProtectedTypes = []string{"person", "trait", "preference", "persona", "doctrine"}

// SelfHeal jalanin decay → prune edge mati → buang node yatim (non-protected).
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

	// 1. Decay strength edge active.
	if r, err := s.db.Exec(`UPDATE cognitive_edges SET strength = strength * ? WHERE status='active'`, opt.DecayFactor); err == nil {
		st.EdgesDecayed, _ = r.RowsAffected()
	}

	// 2. Prune edge mati (strength terlalu kecil).
	if r, err := s.db.Exec(`DELETE FROM cognitive_edges WHERE strength < ?`, opt.MinStrength); err == nil {
		st.EdgesPruned, _ = r.RowsAffected()
	}

	// 3. Buang node yatim (ga ada edge masuk/keluar) KECUALI type protected.
	//    Placeholder ?,?,… buat protect list (parameterized, anti-injection).
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
