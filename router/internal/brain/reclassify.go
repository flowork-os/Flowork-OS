// reclassify.go — Background goroutine daemon untuk re-klasifikasi mem_type.
//
// Daemon ini berjalan otomatis sebagai goroutine background (mirip dream cycle).
// Secara periodik men-scan drawers yang punya mem_type legacy/non-canonical,
// lalu meng-UPDATE ke tipe canonical menggunakan rules engine.
//
// SAFETY:
//   - Sacred types (user, doctrine, antibody) TIDAK akan pernah di-reclassify.
//   - Menggunakan ClassifyContentStrict: hanya reclassify jika ada rule match.
//   - Batch processing dengan limit untuk menghindari DB lock berkepanjangan.
//   - Graceful shutdown via context cancellation.
//
// ZERO modifikasi file FROZEN. File ini standalone, cuma panggil OpenRW() dari write.go.
//
// Dibuat: 2026-06-23 — Phase 2 Memory Typed System.

package brain

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Daemon Config
// ──────────────────────────────────────────────────────────────────────────────

// ReclassifyConfig mengatur perilaku daemon reclassify.
type ReclassifyConfig struct {
	Interval   time.Duration // interval antar batch scan (default: 30 menit)
	BatchSize  int           // jumlah drawers per batch (default: 100)
	DryRun     bool          // jika true, hanya log tanpa UPDATE
}

// DefaultReclassifyConfig mengembalikan konfigurasi default.
func DefaultReclassifyConfig() ReclassifyConfig {
	return ReclassifyConfig{
		Interval:  30 * time.Minute,
		BatchSize: 100,
		DryRun:    false,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Daemon State
// ──────────────────────────────────────────────────────────────────────────────

// ReclassifyStats melacak statistik daemon.
type ReclassifyStats struct {
	TotalScanned  int64
	TotalMigrated int64
	TotalSkipped  int64
	TotalErrors   int64
	LastRunAt     time.Time
	Running       int32 // atomic: 1 = sedang jalan
}

var reclassifyStats ReclassifyStats

// GetReclassifyStats mengembalikan snapshot statistik daemon.
func GetReclassifyStats() ReclassifyStats {
	return ReclassifyStats{
		TotalScanned:  atomic.LoadInt64(&reclassifyStats.TotalScanned),
		TotalMigrated: atomic.LoadInt64(&reclassifyStats.TotalMigrated),
		TotalSkipped:  atomic.LoadInt64(&reclassifyStats.TotalSkipped),
		TotalErrors:   atomic.LoadInt64(&reclassifyStats.TotalErrors),
		LastRunAt:     reclassifyStats.LastRunAt,
		Running:       atomic.LoadInt32(&reclassifyStats.Running),
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Background Daemon
// ──────────────────────────────────────────────────────────────────────────────

// StartReclassifyDaemon memulai goroutine background yang periodik men-scan
// dan re-klasifikasi drawers. Goroutine akan berhenti saat ctx di-cancel.
//
// Contoh penggunaan (di router main atau startup):
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	brain.StartReclassifyDaemon(ctx, brain.DefaultReclassifyConfig())
func StartReclassifyDaemon(ctx context.Context, cfg ReclassifyConfig) {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Minute
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}

	go func() {
		log.Printf("[reclassify] daemon started: interval=%v batch=%d dryRun=%v",
			cfg.Interval, cfg.BatchSize, cfg.DryRun)

		// Jalankan sekali saat startup (setelah delay singkat agar DB ready)
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			log.Printf("[reclassify] daemon stopped before first run")
			return
		}

		runOnce(ctx, cfg)

		// Loop periodik
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[reclassify] daemon stopped (context cancelled)")
				return
			case <-ticker.C:
				runOnce(ctx, cfg)
			}
		}
	}()
}

func runOnce(ctx context.Context, cfg ReclassifyConfig) {
	if !atomic.CompareAndSwapInt32(&reclassifyStats.Running, 0, 1) {
		log.Printf("[reclassify] skipped: previous run still active")
		return
	}
	defer atomic.StoreInt32(&reclassifyStats.Running, 0)

	migrated, scanned, skipped, errs, err := RunReclassifyBatch(ctx, cfg.BatchSize, cfg.DryRun)
	reclassifyStats.LastRunAt = time.Now()

	if err != nil {
		log.Printf("[reclassify] batch error: %v", err)
		return
	}

	atomic.AddInt64(&reclassifyStats.TotalScanned, int64(scanned))
	atomic.AddInt64(&reclassifyStats.TotalMigrated, int64(migrated))
	atomic.AddInt64(&reclassifyStats.TotalSkipped, int64(skipped))
	atomic.AddInt64(&reclassifyStats.TotalErrors, int64(errs))

	if migrated > 0 || errs > 0 {
		log.Printf("[reclassify] batch done: scanned=%d migrated=%d skipped=%d errors=%d",
			scanned, migrated, skipped, errs)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Core Reclassify Logic
// ──────────────────────────────────────────────────────────────────────────────

// reclassifyRow menyimpan data drawer yang perlu dicek.
type reclassifyRow struct {
	id       string
	content  string
	wing     string
	room     string
	memType  string
}

// RunReclassifyBatch men-scan drawers dan re-klasifikasi yang perlu.
// Returns: (migrated, scanned, skipped, errors, err).
//
// Strategi scan:
//  1. Prioritas PERTAMA: drawers dengan mem_type legacy (tidak ada di canonical enum).
//  2. Prioritas KEDUA: drawers dengan mem_type canonical tapi mungkin salah kategori.
//
// Phase ini hanya handle prioritas PERTAMA (legacy migration).
// Prioritas kedua membutuhkan LLM analysis (future Phase 3).
func RunReclassifyBatch(ctx context.Context, batchSize int, dryRun bool) (migrated, scanned, skipped, errors int, err error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	db, err := OpenRW()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("reclassify open db: %w", err)
	}

	// ── Step 1: Temukan drawers dengan mem_type legacy ───────────────────
	rows, err := queryLegacyDrawers(ctx, db, batchSize)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("reclassify query: %w", err)
	}

	if len(rows) == 0 {
		return 0, 0, 0, 0, nil // tidak ada yang perlu di-reclassify
	}

	// ── Step 2: Classify dan Update ──────────────────────────────────────
	for _, row := range rows {
		scanned++

		// Gunakan ClassifyContentStrict: hanya reclassify jika ada rule match
		newType, changed := ClassifyContentStrict(row.content, row.wing, row.room, row.memType)

		if !changed {
			// Tidak ada rule match, tapi mungkin legacy → canonical mapping
			mapped := MapLegacy(row.memType)
			if mapped != MemType(row.memType) {
				newType = mapped
				changed = true
			}
		}

		if !changed {
			skipped++
			continue
		}

		if dryRun {
			log.Printf("[reclassify][dry-run] %s: %q → %q", row.id, row.memType, newType)
			migrated++
			continue
		}

		// UPDATE drawer
		if updateErr := updateDrawerMemType(ctx, db, row.id, string(newType)); updateErr != nil {
			log.Printf("[reclassify] error updating %s: %v", row.id, updateErr)
			errors++
			continue
		}

		migrated++
	}

	return migrated, scanned, skipped, errors, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Database Helpers (NEW code, zero sentuh file FROZEN)
// ──────────────────────────────────────────────────────────────────────────────

// queryLegacyDrawers mengambil drawers yang mem_type-nya BUKAN canonical.
func queryLegacyDrawers(ctx context.Context, db *sql.DB, limit int) ([]reclassifyRow, error) {
	// Build NOT IN clause dari canonical types
	placeholders := make([]string, len(AllMemTypes))
	args := make([]any, len(AllMemTypes)+1)
	for i, mt := range AllMemTypes {
		placeholders[i] = "?"
		args[i] = string(mt)
	}
	args[len(AllMemTypes)] = limit

	query := fmt.Sprintf(`
		SELECT id, content, wing, room, mem_type 
		FROM drawers 
		WHERE deleted_at IS NULL 
		  AND mem_type NOT IN (%s)
		ORDER BY rowid ASC
		LIMIT ?`,
		joinPlaceholders(placeholders))

	sqlRows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	var result []reclassifyRow
	for sqlRows.Next() {
		var r reclassifyRow
		if err := sqlRows.Scan(&r.id, &r.content, &r.wing, &r.room, &r.memType); err != nil {
			return nil, err
		}
		// Truncate content untuk analisis (hemat memori, cukup 500 char pertama)
		if len(r.content) > 500 {
			r.content = r.content[:500]
		}
		result = append(result, r)
	}
	return result, sqlRows.Err()
}

// updateDrawerMemType meng-UPDATE mem_type drawer tanpa sentuh field lain.
func updateDrawerMemType(ctx context.Context, db *sql.DB, drawerID, newMemType string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE drawers SET mem_type = ? WHERE id = ? AND deleted_at IS NULL`,
		newMemType, drawerID)
	return err
}

// joinPlaceholders menggabungkan slice string dengan koma.
func joinPlaceholders(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}
