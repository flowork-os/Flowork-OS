// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN — jangan edit. Logic stabil + teruji. Dok: lock/worklog.md
//
// Package journal — JURNAL PENGALAMAN (roadmap P2): SURFACE apa yang agent UDAH PELAJARI.
//
// Owner: experience journal — "bro jangan ulangi cara ini, dulu gagal di X". Mesinnya UDAH ADA
// (mistakes_local, brain_drawers wing=eureka, brain_antibody, skills self-authored, instinct). Yang
// kurang = KEBUKTIAN. Aggregator READ-only per-agent (pola internal/worklog). Murni + testable.
// Dipakai feature_journal.go (HTTP) + builtin tool `journal` (agent self-refleksi). Dok: lock/worklog.md.
package journal

import (
	"database/sql"
	"sort"
)

type Lesson struct {
	Title   string `json:"title"`
	Hits    int    `json:"hits"`
	LastHit string `json:"last_hit"`
}

type Summary struct {
	Agent     string   `json:"agent"`
	Mistakes  int      `json:"mistakes"`  // pelajaran dari kegagalan
	Eureka    int      `json:"eureka"`    // insight hasil dream→eureka
	Antibody  int      `json:"antibody"`  // pola anti-halu
	Skills    int      `json:"skills"`    // skill self-authored (aktif)
	Instincts int      `json:"instincts"` // insting (cognitive_nodes)
	Lessons   []Lesson `json:"lessons"`   // top pelajaran (by hit)
}

func countQuery(db *sql.DB, q string, args ...any) int {
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		return 0 // tabel ga ada / error → 0 (fail-safe)
	}
	return n
}

// Collect — ringkas pembelajaran tiap agent. openDB nil → skip. lessonsPerAgent = top-N pelajaran.
// Cuma agent yang PUNYA pembelajaran (total > 0) yang masuk (biar ga noise).
func Collect(ids []string, openDB func(agentID string) *sql.DB, lessonsPerAgent int) []Summary {
	out := []Summary{}
	for _, id := range ids {
		db := openDB(id)
		if db == nil {
			continue
		}
		s := Summary{
			Agent:     id,
			Mistakes:  countQuery(db, `SELECT COUNT(*) FROM mistakes_local WHERE deleted_at IS NULL`),
			Eureka:    countQuery(db, `SELECT COUNT(*) FROM brain_drawers WHERE deleted_at IS NULL AND wing='eureka'`),
			Antibody:  countQuery(db, `SELECT COUNT(*) FROM brain_antibody`),
			Skills:    countQuery(db, `SELECT COUNT(*) FROM skills WHERE archived=0`),
			Instincts: countQuery(db, `SELECT COUNT(*) FROM cognitive_nodes WHERE type='instinct'`),
		}
		if s.Mistakes+s.Eureka+s.Antibody+s.Skills+s.Instincts == 0 {
			continue
		}
		if lessonsPerAgent > 0 {
			rows, err := db.Query(`SELECT COALESCE(title,''), COALESCE(hit_count,0), COALESCE(last_hit_at,'')
				FROM mistakes_local WHERE deleted_at IS NULL ORDER BY hit_count DESC, id DESC LIMIT ?`, lessonsPerAgent)
			if err == nil {
				for rows.Next() {
					var l Lesson
					if rows.Scan(&l.Title, &l.Hits, &l.LastHit) == nil && l.Title != "" {
						s.Lessons = append(s.Lessons, l)
					}
				}
				rows.Close()
			}
		}
		out = append(out, s)
	}
	// Urut: agent paling "kaya pengalaman" dulu (total learning desc).
	sort.Slice(out, func(i, j int) bool {
		ti := out[i].Mistakes + out[i].Eureka + out[i].Antibody + out[i].Skills + out[i].Instincts
		tj := out[j].Mistakes + out[j].Eureka + out[j].Antibody + out[j].Skills + out[j].Instincts
		return ti > tj
	})
	return out
}

// Totals — agregat lintas-agent (buat ringkasan GUI).
func Totals(items []Summary) map[string]int {
	t := map[string]int{"mistakes": 0, "eureka": 0, "antibody": 0, "skills": 0, "instincts": 0, "agents": len(items)}
	for _, s := range items {
		t["mistakes"] += s.Mistakes
		t["eureka"] += s.Eureka
		t["antibody"] += s.Antibody
		t["skills"] += s.Skills
		t["instincts"] += s.Instincts
	}
	return t
}
