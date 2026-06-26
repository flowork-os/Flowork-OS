// feature_autosync.go — SELESAIIN 2 sisa roadmap (jalur AMAN, NON-frozen seam RegisterFeature):
//   (A) Skill Central AUTO-SYNC: tiap interval re-pull skill ter-link dari Router → update agent
//       (edit skill di router NYEBAR otomatis, gak cuma tombol manual). = reference-model praktis.
//   (B) M4 AUTO-ENRICH codemap: tiap interval panggil enrich (hash-skip M1 → MURAH pas stabil,
//       re-enrich file BERUBAH). Panggil handler langsung (bypass middleware, internal).
// Semua switch GUI (kebenaran di GUI). Goroutine best-effort, gak blok serve.
package main

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"

	"flowork-gui/internal/agentmgr"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/routerclient"
)

func init() {
	RegisterFeature(Feature{Name: "autosync", Phase: PhaseRoute, Apply: func(d *Deps) {
		go skillResyncLoop(d)
		go codemapAutoEnrichLoop(d)
	}})
}

func asOn(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v != "0" && v != "false" && v != "off" && v != "no"
}

func asMin(key string, def int) time.Duration {
	m := def
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		if n, e := strconv.Atoi(s); e == nil && n > 0 {
			m = n
		}
	}
	return time.Duration(m) * time.Minute
}

// (A) Skill Central auto-sync — re-pull skill ter-link dari router, update kalau beda. Switch
// FLOWORK_SKILL_AUTOSYNC (default ON), interval FLOWORK_SKILL_AUTOSYNC_MIN (default 30).
func skillResyncLoop(d *Deps) {
	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()
	last := time.Time{}
	for {
		select {
		case <-d.Ctx.Done():
			return
		case <-t.C:
			if !asOn("FLOWORK_SKILL_AUTOSYNC", true) {
				continue
			}
			if time.Since(last) < asMin("FLOWORK_SKILL_AUTOSYNC_MIN", 30) {
				continue
			}
			last = time.Now()
			if n := runSkillResync(d.Host); n > 0 {
				log.Printf("[skill-autosync] %d skill di-update dari router catalog", n)
			}
		}
	}
}

func runSkillResync(host *kernelhost.Host) int {
	rc := routerclient.New("") // DefaultRouterURL (127.0.0.1:2402) — router lokal
	updated := 0
	for _, id := range host.AgentIDs() {
		st, err := host.OpenAgentStore(id)
		if err != nil {
			continue
		}
		rows, qerr := st.DB().Query("SELECT id, COALESCE(instructions,'') FROM skills WHERE COALESCE(archived,0)=0")
		if qerr != nil {
			st.Close()
			continue
		}
		type sk struct{ id, instr string }
		var list []sk
		for rows.Next() {
			var s sk
			if rows.Scan(&s.id, &s.instr) == nil && strings.TrimSpace(s.id) != "" {
				list = append(list, s)
			}
		}
		rows.Close()
		for _, s := range list {
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			doc, e := rc.GetSkill(ctx, s.id)
			cancel()
			if e == nil && strings.TrimSpace(doc.Body) != "" && doc.Body != s.instr {
				if _, ue := st.DB().Exec("UPDATE skills SET instructions=? WHERE id=?", doc.Body, s.id); ue == nil {
					updated++
				}
			}
		}
		st.Close()
	}
	return updated
}

// (B) M4 codemap auto-enrich — panggil enrich handler langsung tiap interval. Hash-skip (M1) bikin
// MURAH pas kode stabil (0 LLM); file BARU/BERUBAH di-enrich. Switch FLOWORK_CODEMAP_AUTOENRICH
// (default ON), interval FLOWORK_CODEMAP_AUTOENRICH_MIN (default 30).
func codemapAutoEnrichLoop(d *Deps) {
	h := agentmgr.CodemapEnrichHandler(codemapSemanticSummarizer(d.Host))
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	last := time.Time{}
	for {
		select {
		case <-d.Ctx.Done():
			return
		case <-t.C:
			if !asOn("FLOWORK_CODEMAP_AUTOENRICH", true) {
				continue
			}
			if time.Since(last) < asMin("FLOWORK_CODEMAP_AUTOENRICH_MIN", 30) {
				continue
			}
			last = time.Now()
			// panggil handler internal (bypass middleware auth — trigger sistem, bukan user).
			req := httptest.NewRequest("POST", "/api/codemap/enrich?limit=20", nil)
			h(httptest.NewRecorder(), req)
		}
	}
}
