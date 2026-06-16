// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval. Owner: Aola Sahidin (Mr.Dev).
// Locked 2026-06-17 · P2 fase-2a gerbang #1+#3 endpoints, owner-approved, E2E tested.
//
// handlers_skillpack.go — P2 A2 fase-2a gerbang #1 (signing/provenance) + #3 (karma-gate
// publish). Endpoint router (skill files + identity privkey + store DB ada di router):
//
//   GET  /api/skills/pack/export-signed[?all=1] — bundle SKILL.md → SIGNED pack (ed25519).
//        Default cuma skill yang LOLOS karma-gate (kebukti bagus lokal); ?all=1 = override owner.
//   POST /api/skills/pack/verify                 — verify signature + content gate per-skill.
//   GET  /api/skills/karma                       — track-record semua skill.
//   POST /api/skills/karma/record?skill=&positive=1  — catat 1 pemakaian (feedback).
//   POST /api/skills/karma/endorse?skill=        — owner vonis "proven" (X-Voter-ID audit).
//
// Engine: internal/skillpack (types + content gate + karma), internal/mesh/sign.go (crypto).

package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flowork-os/flowork_Router/internal/brain"
	"github.com/flowork-os/flowork_Router/internal/mesh"
	"github.com/flowork-os/flowork_Router/internal/skillpack"
	"github.com/flowork-os/flowork_Router/internal/store"
)

const maxSkillPackBody = 32 << 20 // 32MB

// loopbackOnly menjaga endpoint SENSITIF (exercise signing key / mutate publish-eligibility)
// supaya cuma bisa dari localhost. Router default bind 127.0.0.1 + authEnforce single-owner
// (default off di loopback) → ini defense-in-depth kalau addr suatu saat di-bind non-loopback.
// Owner-action sebenarnya lewat GUI (browser localhost → tetap loopback). Gerbang owner-approve
// penuh (session) nyusul di fase-2b GUI publish.
func loopbackOnly(w http.ResponseWriter, r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	writeJSON(w, http.StatusForbidden, map[string]any{"error": "loopback-only endpoint (owner/local action)"})
	return false
}

// readSkillFiles membaca SKILL.md dari dir (flat <name>.md atau <name>/SKILL.md).
// Mirror brain.loadDynamicSkills, tapi balikin (name, RAW content) untuk sign/verify.
func readSkillFiles(dir string) []skillpack.SignedSkill {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []skillpack.SignedSkill
	for _, e := range entries {
		var path, name string
		if e.IsDir() {
			path = filepath.Join(dir, e.Name(), "SKILL.md")
			name = e.Name()
		} else if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			path = filepath.Join(dir, e.Name())
			name = strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		} else {
			continue
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil || len(raw) == 0 {
			continue
		}
		out = append(out, skillpack.SignedSkill{Name: name, Content: string(raw)})
	}
	return out
}

// skillPackExportSignedHandler — GET /api/skills/pack/export-signed[?all=1]
func skillPackExportSignedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET only"})
		return
	}
	if !loopbackOnly(w, r) {
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	all := r.URL.Query().Get("all") == "1"
	skills := readSkillFiles(brain.DynamicSkillsDir())

	pack := skillpack.SignedPack{Kind: "fwskill-signed", Version: 1, Source: "flowork",
		SignedAt: time.Now().UTC().Format(time.RFC3339)}
	skipped := make([]map[string]any, 0)
	for _, s := range skills {
		if !all {
			ok, reason := skillpack.CanPublish(db, s.Name, s.Content) // gerbang #3
			if !ok {
				skipped = append(skipped, map[string]any{"name": s.Name, "reason": reason})
				continue
			}
		}
		sig, pub, serr := mesh.SignData(db, []byte(s.Content)) // gerbang #1
		if serr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "sign: " + serr.Error()})
			return
		}
		pack.AuthorPubkey = pub
		pack.Skills = append(pack.Skills, skillpack.SignedSkill{Name: s.Name, Content: s.Content, Sig: sig})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pack": pack, "signed": len(pack.Skills), "skipped": skipped,
		"note": "Signed dgn identity instance (provenance). ?all=1 = override karma-gate (owner).",
	})
}

// skillPackVerifyHandler — POST /api/skills/pack/verify {pack:{...}} atau {SignedPack}
func skillPackVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSkillPackBody)
	var wrap struct {
		Pack *skillpack.SignedPack `json:"pack"`
	}
	var pack skillpack.SignedPack
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&wrap); err == nil && wrap.Pack != nil {
		pack = *wrap.Pack
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid signed pack (expect {pack:{...}})"})
		return
	}
	results := make([]map[string]any, 0, len(pack.Skills))
	validCount := 0
	for _, s := range pack.Skills {
		sigValid := mesh.VerifyData(pack.AuthorPubkey, []byte(s.Content), s.Sig)
		flags := skillpack.VerifyContent(s.Content)
		safe := len(flags) == 0
		if sigValid && safe {
			validCount++
		}
		results = append(results, map[string]any{
			"name": s.Name, "sig_valid": sigValid, "content_safe": safe, "flags": flags,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"author_pubkey": pack.AuthorPubkey, "total": len(pack.Skills),
		"trusted": validCount, "results": results,
	})
}

// skillKarmaListHandler — GET /api/skills/karma
func skillKarmaListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	items, err := skillpack.ListSkillKarma(db, 0)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// skillKarmaRecordHandler — POST /api/skills/karma/record?skill=&positive=1
func skillKarmaRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	if !loopbackOnly(w, r) {
		return
	}
	skill := strings.TrimSpace(r.URL.Query().Get("skill"))
	if skill == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "skill required"})
		return
	}
	// Trust-gate input: WAJIB eksplisit positive=1|0 (jangan default positif —
	// param yang lupa di-set bisa nge-inflate track-record lewat karma-gate).
	p := r.URL.Query().Get("positive")
	if p != "0" && p != "1" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "positive required (1 or 0)"})
		return
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if err := skillpack.RecordSkillUse(db, skill, p == "1"); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	k, _ := skillpack.GetSkillKarma(db, skill)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "karma": k})
}

// skillKarmaEndorseHandler — POST /api/skills/karma/endorse?skill=  (X-Voter-ID audit)
func skillKarmaEndorseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
		return
	}
	if !loopbackOnly(w, r) {
		return
	}
	skill := strings.TrimSpace(r.URL.Query().Get("skill"))
	if skill == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "skill required"})
		return
	}
	by := strings.TrimSpace(r.Header.Get("X-Voter-ID"))
	if by == "" {
		by = "owner"
	}
	db, err := store.Open()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if err := skillpack.EndorseSkill(db, skill, by); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	k, _ := skillpack.GetSkillKarma(db, skill)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "karma": k})
}
