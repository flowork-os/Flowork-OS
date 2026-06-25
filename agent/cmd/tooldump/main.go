// tooldump — throwaway: dump semua tool ter-register (name/cap/desc/params) ke JSON.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"flowork-gui/internal/tools"
	"flowork-gui/internal/tools/builtins"
)

func main() {
	builtins.Init()
	type p struct {
		Name, Type, Desc string
		Req              bool
	}
	type t struct {
		Name, Cap, Desc, Returns string
		Params                   []p
	}
	var out []t
	for _, tl := range tools.List() {
		s := tl.Schema()
		var ps []p
		for _, pr := range s.Params {
			ps = append(ps, p{pr.Name, string(pr.Type), pr.Description, pr.Required})
		}
		out = append(out, t{tl.Name(), tl.Capability(), s.Description, s.Returns, ps})
	}
	fmt.Fprintf(os.Stderr, "TOTAL tools: %d\n", len(out))
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
