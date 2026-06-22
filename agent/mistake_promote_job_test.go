package main

import (
	"testing"

	"flowork-gui/internal/agentdb"
)

// TestRecoveryClassKey — kunci-kelas (sumbu konvergensi INC-3) ke-extract bener dari
// title INC-2 "recovery: <tool>/<class>"; title aneh / kelas tak-dikenal → "".
func TestRecoveryClassKey(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"recovery: file_read/not-found", "not-found"},
		{"recovery: file_write/not-found", "not-found"}, // tool beda, kelas sama → konvergen
		{"recovery: exec_shell/permission", "permission"},
		{"recovery: telegram_send/timeout", "timeout"},
		{"recovery: x/already-exists", "already-exists"},
		{"recovery: y/invalid-input", "invalid-input"},
		{"recovery: z/blocked", "blocked"},
		{"recovery: w/error", "error"},
		{"recovery: tool/unknown-class", ""}, // kelas di luar whitelist → kosong
		{"some random mistake title", ""},    // bukan format recovery
		{"recovery: notool", ""},             // ga ada "/"
		{"", ""},
	}
	for _, c := range cases {
		got := recoveryClassKey(agentdb.Mistake{Title: c.title})
		if got != c.want {
			t.Errorf("recoveryClassKey(%q)=%q want %q", c.title, got, c.want)
		}
	}
}
