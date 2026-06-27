package triggers

import (
	"testing"
	"time"
)

func TestIdleShouldFire(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	cd := 30 * time.Minute
	zero := time.Time{}

	cases := []struct {
		name             string
		load, threshold  float64
		last             time.Time
		want             bool
	}{
		{"busy → no fire", 80, 60, zero, false},
		{"idle pertama kali → fire", 20, 60, zero, true},
		{"idle tapi cooldown belum lewat → no fire", 20, 60, now.Add(-10 * time.Minute), false},
		{"idle + cooldown lewat → fire", 20, 60, now.Add(-31 * time.Minute), true},
		{"pas di ambang (load==threshold) → no fire", 60, 60, zero, false},
	}
	for _, c := range cases {
		got := idleShouldFire(c.load, c.threshold, c.last, now, cd)
		if got != c.want {
			t.Errorf("%s: idleShouldFire(load=%.0f,thr=%.0f)=%v, mau %v", c.name, c.load, c.threshold, got, c.want)
		}
	}
}

func TestIdleTypeRegistered(t *testing.T) {
	found := false
	for _, ty := range ListTypes() {
		if ty.ID() == "idle" {
			found = true
			if ty.Mode() != "poll" {
				t.Errorf("idle Mode = %q, mau poll", ty.Mode())
			}
		}
	}
	if !found {
		t.Error("trigger tipe 'idle' ga ke-register")
	}
}
