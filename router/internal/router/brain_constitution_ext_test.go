package router

import (
	"context"
	"testing"

	"github.com/flowork-os/flowork_Router/internal/brain"
)

func ccFixture() []brain.ConstitutionEntry {
	return []brain.ConstitutionEntry{
		{Section: "AOLA-001_IDENTITAS_TUJUAN", Content: "### AOLA-001 ..."},   // internal
		{Section: "AOLA-002_NAVIGASI_KEBENARAN", Content: "### AOLA-002 ..."},  // internal
		{Section: "AOLA-003_KEJUJURAN", Content: "### AOLA-003 ..."},           // universal
		{Section: "AOLA-011_KALKULASI", Content: "### AOLA-011 ..."},           // universal
		{Section: "AOLA-013_KERJA_PANJANG", Content: "### AOLA-013 ..."},       // universal
	}
}

func secs(rs []brain.ConstitutionEntry) map[string]bool {
	m := map[string]bool{}
	for _, r := range rs {
		m[aolaIDRe.FindString(r.Section)] = true
	}
	return m
}

// Internal caller (punya agent-id) → UTUH walau switch on.
func TestConstFilter_InternalUntouched(t *testing.T) {
	t.Setenv("FLOWORK_BRAIN_EXTERNAL_SCOPE", "1")
	ctx := WithAgentID(context.Background(), "mr-flow")
	got := externalConstitutionFilter(ctx, ccFixture())
	if len(got) != 5 {
		t.Errorf("internal caller harus utuh 5, dapet %d", len(got))
	}
}

// Switch OFF (default) → UTUH (no-op), walau caller eksternal.
func TestConstFilter_SwitchOffNoop(t *testing.T) {
	t.Setenv("FLOWORK_BRAIN_EXTERNAL_SCOPE", "0")
	got := externalConstitutionFilter(context.Background(), ccFixture())
	if len(got) != 5 {
		t.Errorf("switch off harus no-op (5), dapet %d", len(got))
	}
}

// Eksternal (id kosong) + switch ON → buang internal (001,002), simpan universal (003,011,013).
func TestConstFilter_ExternalDropsInternal(t *testing.T) {
	t.Setenv("FLOWORK_BRAIN_EXTERNAL_SCOPE", "1")
	got := secs(externalConstitutionFilter(context.Background(), ccFixture()))
	for _, drop := range []string{"AOLA-001", "AOLA-002"} {
		if got[drop] {
			t.Errorf("%s harus DI-DROP buat eksternal", drop)
		}
	}
	for _, keep := range []string{"AOLA-003", "AOLA-011", "AOLA-013"} {
		if !got[keep] {
			t.Errorf("%s universal harus TETEP buat eksternal", keep)
		}
	}
}
