// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-20
// Reason: Fitur izin-app per-agent (GUI=truth). Tested end-to-end: migrasi seed
//   one-time, grant/revoke live (revoke→cap dicabut→tool ditolak, grant→bisa).
//   Lazy schema (locked ensureSchema ga disentuh), pola sama tool_subscriptions.go.

package agentdb

// app_grants.go — per-warga "izin pakai app" (owner 2026-06-20: "di GUI agent
// buat filtur baru, dia diijinin pake app apa saja"). Tabel ini simpan PILIHAN
// izin per-agent: app mana yang boleh dipakai. Capability app:<id> di-grant ke
// broker HANYA kalau app-nya ada di sini → GUI = kebenaran (centang = bisa,
// uncentang = ga bisa), bukan pajangan.
//
// PENTING (jawaban "gimana app masuk DB pas install, ilang pas uninstall"):
// sumber kebenaran "app apa yg ADA" itu BUKAN tabel ini — itu apps.Manager
// (folder ~/.flowork/apps). Tabel ini cuma nyimpen pilihan izin. App di-install
// → folder lahir → Manager.List() munculin → GUI munculin otomatis. App di-
// uninstall → folder ilang → Manager.List() ga keluarin → GUI ilang otomatis.
// Row sisa di sini (kalau app yg udah dihapus pernah diizinkan) INERT: cap
// app:<id> ga nyocokin tool mana pun + ga dirender (list digerakin Manager).
// Zero-sync, zero-JSON. Lazy schema (locked ensureSchema ga disentuh) — pola
// sama tool_subscriptions.go.

import (
	"fmt"
	"strings"
	"time"
)

func (s *Store) ensureAppGrantsSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS app_grants (
		  app_id     TEXT PRIMARY KEY,
		  granted_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("ensure app_grants: %w", err)
	}
	return nil
}

// GrantApp — izinkan agent pakai app. Idempotent (sudah ada = no-op).
func (s *Store) GrantApp(appID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureAppGrantsSchema(); err != nil {
		return err
	}
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return fmt.Errorf("app_id required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO app_grants (app_id, granted_at) VALUES (?, ?)
		 ON CONFLICT(app_id) DO NOTHING`, appID, now)
	return err
}

// RevokeApp — cabut izin. Idempotent (no-row = no-op).
func (s *Store) RevokeApp(appID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureAppGrantsSchema(); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM app_grants WHERE app_id = ?`, strings.TrimSpace(appID))
	return err
}

// ListAppGrants — daftar app_id yang diizinkan, urut. Cap 500.
func (s *Store) ListAppGrants() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureAppGrantsSchema(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT app_id FROM app_grants ORDER BY app_id LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if serr := rows.Scan(&id); serr != nil {
			return nil, serr
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// AppGrantsSeeded — penanda migrasi one-time (meta table). Biar seeding dari
// subscription cuma SEKALI; abis itu app_grants authoritative (kosong = ga ada
// app diizinkan, BUKAN trigger re-seed yg bikin app balik lagi pas owner
// uncentang semua).
func (s *Store) AppGrantsSeeded() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	var v string
	_ = s.db.QueryRow(`SELECT v FROM meta WHERE k='app_grants_seeded'`).Scan(&v)
	return v == "1"
}

// MarkAppGrantsSeeded — set penanda migrasi.
func (s *Store) MarkAppGrantsSeeded() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO meta(k, v) VALUES('app_grants_seeded', '1')
		 ON CONFLICT(k) DO UPDATE SET v='1'`)
	return err
}
