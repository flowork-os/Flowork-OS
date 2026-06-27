package main

import (
	"testing"
	"time"

	"flowork-gui/internal/worklog"
)

func TestDeadairDecide(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-10 * time.Minute).Format(time.RFC3339)
	stale := now.Add(-90 * time.Minute).Format(time.RFC3339)

	// kosong → bukan anomali (idle sah).
	if a, _ := deadairDecide(nil, now, 60); a {
		t.Error("papan kosong mestinya BUKAN anomali")
	}
	// ada aktif + paling baru masih fresh → bukan anomali (masih gerak).
	items := []worklog.Item{{Updated: stale}, {Updated: fresh}}
	if a, _ := deadairDecide(items, now, 60); a {
		t.Error("ada yg fresh (<60mnt) mestinya BUKAN anomali")
	}
	// semua aktif beku > ambang → ANOMALI.
	items = []worklog.Item{{Updated: stale}, {Updated: now.Add(-120 * time.Minute).Format(time.RFC3339)}}
	if a, newest := deadairDecide(items, now, 60); !a {
		t.Error("semua beku >60mnt mestinya ANOMALI")
	} else if newest.IsZero() {
		t.Error("newest mestinya keisi")
	}
	// timestamp ngaco → fail-safe bukan anomali.
	if a, _ := deadairDecide([]worklog.Item{{Updated: "bukan-waktu"}}, now, 60); a {
		t.Error("timestamp invalid mestinya fail-safe (bukan anomali)")
	}
}
