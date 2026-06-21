// === LOCKED FILE (soft) === Status: STABLE — owner-approved 2026-06-21 (3E/D13 + AI-IN-AGENT). LOCKED ≠ FREEZE (boleh diedit dgn izin owner + re-lock + changelog).
package scanapi

// distill_model.go — AI-IN-AGENT (owner 2026-06-21 G6): model distiller (topics →
// nuclei template) dulu HARDCODE literal `distillModelDefault` ("claude-haiku-4-5").
// Sekarang resolusi default-nya di-inject dari main (DistillModelResolver = model
// AGENT scan-distiller, kv router_model GUI) → owner bisa ganti model fitur ini di
// Settings. nil = belum di-wire → fallback literal (robust). Override ?model= di body
// tetap menang (dicek di handler sebelum manggil distillDefaultModel).

// DistillModelResolver — di-set main package: model GUI agent scan-distiller.
var DistillModelResolver func() string

// distillDefaultModel — model default distiller: dari AGENT scan-distiller (GUI) kalau
// resolver ke-wire & non-kosong, else literal distillModelDefault (last-resort).
func distillDefaultModel() string {
	if DistillModelResolver != nil {
		if m := DistillModelResolver(); m != "" {
			return m
		}
	}
	return distillModelDefault
}
