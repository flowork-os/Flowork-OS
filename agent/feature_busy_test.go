package main

import (
	"testing"
	"time"
)

func TestBusyShouldAlert(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	cd := 30 * time.Minute
	zero := time.Time{}
	cases := []struct {
		name            string
		load, threshold float64
		ok              bool
		last            time.Time
		want            bool
	}{
		{"sepi → no alert", 40, 90, true, zero, false},
		{"berat pertama → alert", 95, 90, true, zero, true},
		{"berat tapi cooldown → no", 95, 90, true, now.Add(-10 * time.Minute), false},
		{"berat + cooldown lewat → alert", 95, 90, true, now.Add(-31 * time.Minute), true},
		{"pas di ambang → no", 90, 90, true, zero, false},
		{"ga bisa baca load → no (fail-safe)", 99, 90, false, zero, false},
	}
	for _, c := range cases {
		if got := busyShouldAlert(c.load, c.ok, c.threshold, c.last, now, cd); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
