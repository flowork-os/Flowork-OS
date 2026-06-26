// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// CLI tool plug-and-play (RegisterCLITool) → dok lock/plug-and-play.md  ⚠️ FROZEN — jangan edit.
// Nambah CLI tool: file sibling cli_<x>.go + init(){ RegisterCLITool(...) }. Cara: CARAFREEZE.MD
// (POLA A). Pola freeze: lock/frozen-core.md

package clitools

var extraCLITools []Tool

// cliDBSource — di-set file non-frozen (custom_db.go) buat baca CLI tool custom dari DB
// CALL-TIME → add/delete dari GUI langsung kebaca (nol restart). nil = ga ada DB source.
var cliDBSource func() []Tool

func RegisterCLITool(t Tool) {
	if t.ID == "" {
		return
	}
	extraCLITools = append(extraCLITools, t)
}

// allCLIExtras — gabungan code-registered (sibling) + DB custom (call-time). Dipakai All().
func allCLIExtras() []Tool {
	out := append([]Tool{}, extraCLITools...)
	if cliDBSource != nil {
		out = append(out, cliDBSource()...)
	}
	return out
}
