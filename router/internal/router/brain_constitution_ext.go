// brain_constitution_ext.go — GROWTH-POINT (NON-frozen). #10 brain-as-service: scope doktrin
// buat caller EKSTERNAL (agent luar ber-jiwa-Aola via :2402/v1). Hook dipanggil dari
// maybeInjectConstitution (brain_constitution.go) — default NO-OP (hook nil sampai init ini set).
//
// AKAR: 12 doktrin AOLA mayoritas Flowork-internal (identitas Mr.Flow/Aola, navigasi tool internal,
// preferensi Mr.Dev, aturan FLowork OS). Kalau di-inject ke AI LUAR → bingung/halu (dia bukan Mr.Flow,
// ga punya tool internal). Buat eksternal: kasih cuma doktrin UNIVERSAL (reasoning/etika), skip internal.
//
// SWITCH: sama dgn instinct external-scope → FLOWORK_BRAIN_EXTERNAL_SCOPE (default OFF). OFF atau
// caller internal (X-Agent-ID ada) → rules UTUH (zero perubahan). Lihat lock/scoped-instinct.md.
package router

import (
	"context"
	"regexp"
	"strings"

	"github.com/flowork-os/flowork_Router/internal/brain"
)

// constitutionFilterHook — dipanggil maybeInjectConstitution (frozen) abis ListConstitution.
// nil = no-op (default). Di-set init() ext ini. Filter di-skala AgentID + switch eksternal.
var constitutionFilterHook func(context.Context, []brain.ConstitutionEntry) []brain.ConstitutionEntry

func init() { constitutionFilterHook = externalConstitutionFilter }

// internalDoctrines — AOLA yang Flowork-INTERNAL (skip buat eksternal): identitas, navigasi-tool,
// preferensi-owner, idempotent-file, ephemeral-mem, hierarki-FLowork-OS. Sisanya (003 kejujuran,
// 004 siklus-kognitif, 005 5W1H, 009 transparansi, 010 anti-delusi, 011 blast-radius, 013 loop) =
// UNIVERSAL → tetep buat eksternal. Tunable: tambah/kurang ID di sini (NON-frozen).
var internalDoctrines = map[string]bool{
	"AOLA-001": true, // identitas Mr.Flow/Aola
	"AOLA-002": true, // navigasi tool internal Flowork
	"AOLA-006": true, // preferensi anti-mainstream Mr.Dev
	"AOLA-007": true, // idempotent file_write Flowork
	"AOLA-008": true, // ephemeral memory_set Flowork
	"AOLA-012": true, // hierarki freeze FLowork OS
}

var aolaIDRe = regexp.MustCompile(`AOLA-\d{3}`)

// externalConstitutionFilter — caller eksternal + switch ON → buang doktrin internal. Selainnya utuh.
func externalConstitutionFilter(ctx context.Context, rules []brain.ConstitutionEntry) []brain.ConstitutionEntry {
	if AgentIDFromContext(ctx) != "" || !brainExternalScopeEnabled() {
		return rules // internal caller / switch off → utuh (default path)
	}
	out := make([]brain.ConstitutionEntry, 0, len(rules))
	dropped := 0
	for _, r := range rules {
		id := aolaIDRe.FindString(r.Section)
		if id == "" {
			id = aolaIDRe.FindString(r.Content)
		}
		if id != "" && internalDoctrines[strings.ToUpper(id)] {
			dropped++
			continue // doktrin internal → skip buat eksternal (anti-halu)
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return rules // jaga2: jangan kosongin total
	}
	return out
}
