// 📄 Dok: FLowork_os/lock/ERROR_EDUKASI.md
//
// edu_errors_seam.go — MEKANISME BEKU (POLA-B) buat upsert edu-errors override. Method
// SeedEduErrorsExt + hook default ADA DI SINI biar caller (provision_dna.go) self-sufficient:
// hapus edu_errors_ext.go (DATA override, deletable) → hook balik ke default (nil = nol extra),
// SeedEduErrorsExt seed kosong, build TETEP OK (delete-test §6.4 lulus). DATA override tetep di
// edu_errors_ext.go (non-frozen) — di-pasang ke hook lewat init(). Tambah/ubah pelajaran: edit _ext.
package agentdb

import "time"

// extraEduErrors — hook DATA override edu-errors. Default nil (nol extra) = aman. File SIBLING
// edu_errors_ext.go (non-frozen) nimpa lewat init() → ExtraEduErrors().
var extraEduErrors = func() []EduError { return nil }

// SeedEduErrorsExt — UPSERT extraEduErrors() (DO UPDATE = override entri basi by code). Idempotent.
// Return jumlah baris ke-insert/-update. Dipanggil ProvisionAgentDNA SETELAH SeedEduErrors.
func (s *Store) SeedEduErrorsExt() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ts := time.Now().UTC().Format(time.RFC3339)
	n := 0
	for _, e := range extraEduErrors() {
		res, err := s.db.Exec(
			`INSERT INTO educational_errors_cache(code, category, title, explanation, remediation, synced_at)
			 VALUES(?, ?, ?, ?, ?, ?)
			 ON CONFLICT(code) DO UPDATE SET
			   category=excluded.category, title=excluded.title,
			   explanation=excluded.explanation, remediation=excluded.remediation,
			   synced_at=excluded.synced_at, deleted_at=NULL`,
			e.Code, e.Category, e.Title, e.Explanation, e.Remediation, ts,
		)
		if err != nil {
			return n, err
		}
		if c, _ := res.RowsAffected(); c > 0 {
			n++
		}
	}
	return n, nil
}
