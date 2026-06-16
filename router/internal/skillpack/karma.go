// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval. Owner: Aola Sahidin (Mr.Dev).
// Locked 2026-06-17 · P2 fase-2a gerbang #3 (karma-gate publish), owner-approved, tested.
//
// karma.go — P2 A2 fase-2a gerbang #3: KARMA-GATE publish.
//
// Doktrin roadmap: skill mesti "kebukti bagus LOKAL dulu" baru boleh di-share/publish —
// network-effect-nya QUALITY, bukan spam. Track-record per-skill di tabel skill_karma
// (additive, CREATE IF NOT EXISTS di store DB — TIDAK nyentuh migration locked).
//
// "Kebukti bagus" = lolos content gate (wajib) DAN salah satu:
//   - di-ENDORSE owner (vonis manual "ini bagus, boleh publish"), ATAU
//   - track-record pemakaian cukup (uses >= floor) dgn rasio positif sehat.
// RecordSkillUse di-feed oleh feedback loop / hook pemakaian skill (endpoint terbuka;
// auto-hook ke SelectSkills nyusul — SelectSkills LOCKED, gak di-edit di sini).
package skillpack

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	publishMinUses     = 3   // minimal pemakaian sebelum dianggap "teruji" (tanpa endorse)
	publishMinPositive = 0.6 // rasio positif minimal
)

// SkillKarma — track-record satu skill.
type SkillKarma struct {
	Skill      string  `json:"skill"`
	Uses       int     `json:"uses"`
	Positive   int     `json:"positive"`
	Endorsed   bool    `json:"endorsed"`
	EndorsedBy string  `json:"endorsed_by,omitempty"`
	Score      float64 `json:"score"` // positive/uses (0 kalau belum dipakai)
	FirstSeen  string  `json:"first_seen,omitempty"`
	LastUsed   string  `json:"last_used,omitempty"`
}

func ensureKarmaSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS skill_karma (
		skill       TEXT PRIMARY KEY,
		uses        INTEGER DEFAULT 0,
		positive    INTEGER DEFAULT 0,
		endorsed    INTEGER DEFAULT 0,
		endorsed_by TEXT,
		first_seen  TEXT,
		last_used   TEXT
	)`)
	if err != nil {
		return fmt.Errorf("ensure skill_karma schema: %w", err)
	}
	return nil
}

func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

// RecordSkillUse mencatat satu pemakaian skill (positive = outcome bagus).
func RecordSkillUse(db *sql.DB, skill string, positive bool) error {
	if db == nil || skill == "" {
		return fmt.Errorf("db + skill required")
	}
	if err := ensureKarmaSchema(db); err != nil {
		return err
	}
	pos := 0
	if positive {
		pos = 1
	}
	now := nowRFC()
	_, err := db.Exec(`INSERT INTO skill_karma(skill, uses, positive, first_seen, last_used)
		VALUES(?, 1, ?, ?, ?)
		ON CONFLICT(skill) DO UPDATE SET
			uses = uses + 1,
			positive = positive + ?,
			last_used = ?`,
		skill, pos, now, now, pos, now)
	if err != nil {
		return fmt.Errorf("record skill use %q: %w", skill, err)
	}
	return nil
}

// EndorseSkill menandai skill "proven" oleh owner (vonis manual). by = identitas.
func EndorseSkill(db *sql.DB, skill, by string) error {
	if db == nil || skill == "" {
		return fmt.Errorf("db + skill required")
	}
	if err := ensureKarmaSchema(db); err != nil {
		return err
	}
	now := nowRFC()
	_, err := db.Exec(`INSERT INTO skill_karma(skill, endorsed, endorsed_by, first_seen)
		VALUES(?, 1, ?, ?)
		ON CONFLICT(skill) DO UPDATE SET endorsed = 1, endorsed_by = ?`,
		skill, by, now, by)
	if err != nil {
		return fmt.Errorf("endorse skill %q: %w", skill, err)
	}
	return nil
}

// GetSkillKarma mengambil track-record satu skill (zero-value kalau belum ada).
func GetSkillKarma(db *sql.DB, skill string) (SkillKarma, error) {
	k := SkillKarma{Skill: skill}
	if db == nil || skill == "" {
		return k, fmt.Errorf("db + skill required")
	}
	if err := ensureKarmaSchema(db); err != nil {
		return k, err
	}
	var endorsed int
	var endorsedBy, firstSeen, lastUsed sql.NullString
	err := db.QueryRow(`SELECT uses, positive, endorsed, endorsed_by, first_seen, last_used
		FROM skill_karma WHERE skill = ?`, skill).
		Scan(&k.Uses, &k.Positive, &endorsed, &endorsedBy, &firstSeen, &lastUsed)
	if errors.Is(err, sql.ErrNoRows) {
		return k, nil // belum ada track-record (zero)
	}
	if err != nil {
		return k, err
	}
	k.Endorsed = endorsed == 1
	k.EndorsedBy = endorsedBy.String
	k.FirstSeen = firstSeen.String
	k.LastUsed = lastUsed.String
	if k.Uses > 0 {
		k.Score = float64(k.Positive) / float64(k.Uses)
	}
	return k, nil
}

// ListSkillKarma mengembalikan track-record semua skill (cap limit).
func ListSkillKarma(db *sql.DB, limit int) ([]SkillKarma, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	if db == nil {
		return nil, fmt.Errorf("nil db")
	}
	if err := ensureKarmaSchema(db); err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT skill, uses, positive, endorsed, COALESCE(endorsed_by,''),
		COALESCE(first_seen,''), COALESCE(last_used,'') FROM skill_karma
		ORDER BY endorsed DESC, uses DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SkillKarma
	for rows.Next() {
		var k SkillKarma
		var endorsed int
		if err := rows.Scan(&k.Skill, &k.Uses, &k.Positive, &endorsed, &k.EndorsedBy, &k.FirstSeen, &k.LastUsed); err != nil {
			return nil, err
		}
		k.Endorsed = endorsed == 1
		if k.Uses > 0 {
			k.Score = float64(k.Positive) / float64(k.Uses)
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// CanPublish — gerbang #3: boleh publish/share kalau konten bersih DAN kebukti bagus
// lokal (endorsed owner ATAU track-record cukup). Return (ok, reason).
func CanPublish(db *sql.DB, skill, content string) (bool, string) {
	if flags := VerifyContent(content); len(flags) > 0 {
		return false, "unsafe content: " + flags[0]
	}
	k, err := GetSkillKarma(db, skill)
	if err != nil {
		return false, "karma lookup error: " + err.Error()
	}
	if k.Endorsed {
		return true, "endorsed by owner"
	}
	if k.Uses >= publishMinUses && k.Score >= publishMinPositive {
		return true, fmt.Sprintf("proven locally (%d uses, %.0f%% positive)", k.Uses, k.Score*100)
	}
	return false, fmt.Sprintf("not proven locally yet (uses=%d/%d, positive=%.0f%%); endorse it or let it earn track-record first",
		k.Uses, publishMinUses, k.Score*100)
}
