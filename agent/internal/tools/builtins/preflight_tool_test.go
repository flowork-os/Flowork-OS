package builtins

import (
	"strings"
	"testing"
)

func TestPreflightAdvice(t *testing.T) {
	if got := preflightAdvice(95, true, true, "deadly"); !strings.Contains(got, "DEADLY") {
		t.Errorf("deadly: %q", got)
	}
	if got := preflightAdvice(0, false, false, "normal"); !strings.Contains(got, "ga ke-baca") {
		t.Errorf("no-load: %q", got)
	}
	if got := preflightAdvice(95, true, true, "hemat"); !strings.Contains(got, "tunda") {
		t.Errorf("busy+hemat: %q", got)
	}
	if got := preflightAdvice(20, true, false, "normal"); !strings.Contains(got, "aman") {
		t.Errorf("idle+normal: %q", got)
	}
}
