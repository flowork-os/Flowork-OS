// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md §6.3/§13
//
// recovery_generalize.go — D32 INC-3 brain-pathway: GENERALISASI recovery-instinct.
// Recovery MENTAH (INC-2 capture) → instinct UMUM + privacy-safe, SEBELUM jadi instinct
// AKTIF. Dua lapis (deterministik dulu, baru LLM):
//
//   Lapis A — STRIP deterministik (no-LLM, WAJIB): buang path absolut, url/email/token,
//             id/hash panjang, dan nama personal owner (allowlist runtime) → JAMIN 0
//             data owner-spesifik walau Lapis B (LLM) meleset. Privacy-first, ke-test.
//   Lapis B — COARSEN via dream-digester (LLM): distil pola spesifik → pola coarse
//             reusable "WHEN <umum> -> <aksi>" → SHADOW instinct (where_domain=recovery).
//
// GERBANG (anti-degenerasi, KUNCI roadmap #1 & #5): instinct lahir SHADOW, promote ke
// ACTIVE cuma kalau pola-nya BERULANG (hit_count>=N) lintas recovery beda.
// ⚠️ IDENTITAS/DEDUP pakai KUNCI DETERMINISTIK (kelas-error, mis. "not-found"), BUKAN
// embedding output LLM. Sebab kebukti e2e: LLM coarsen NON-DETERMINISTIK (input sama →
// teks beda tiap call) → dedup-embedding ga reliable → hit ga naik → instinct nyangkut
// shadow selamanya. Kelas-error stabil → semua recovery kelas-sama (lintas tool) nyatu
// ke 1 node → hit naik deterministik → gerbang firable. Teks LLM tetap jadi LABEL +
// EMBEDDING (buat recall by-makna). Promote dijalanin ticker non-beku 1-menit
// (PromoteRecurringMistakes) — BUKAN nyandar autodigest (default OFF).
//
// ⚠️ PRIVASI D8 (LANTAI KERAS): JANGAN PERNAH hardcode nama/data personal owner di file
// ini (ke-track repo). Pola di-strip by-REGEX (generic); nama personal di-strip lewat
// allowlist yang di-LOAD RUNTIME dari graph (type=person) — bukan literal di kode.

package agentdb

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// ───────────────────────── Lapis A — strip deterministik ─────────────────────────

// Pola deterministik (urutan apply penting: yg lebih spesifik/menelan dulu).
var (
	// URL penuh (http/https) — telan dulu sebelum path generic (url punya //host/path).
	reGenURL = regexp.MustCompile(`(?i)\bhttps?://[^\s"'<>)]+`)
	// Email.
	reGenEmail = regexp.MustCompile(`(?i)[\w.+\-]+@[\w\-]+\.[\w.\-]+`)
	// Token rahasia umum (github/slack/openai-style) — sebelum long-hex biar ga ke-mangle.
	reGenToken = regexp.MustCompile(`(?i)\b(?:gh[posru]_[A-Za-z0-9]{8,}|xox[baprs]-[A-Za-z0-9\-]{8,}|sk-[A-Za-z0-9]{12,})\b`)
	// Path Windows (C:\Users\..., D:\foo\bar).
	reGenWinPath = regexp.MustCompile(`(?i)\b[a-z]:\\[^\s"'<>]*`)
	// Path home/tilde (~/x/y, /home/x, /Users/x, /root/x, /var/.../x) — owner-identifying.
	reGenHomePath = regexp.MustCompile(`(?i)(?:~|/(?:home|users|root|var|tmp|opt|mnt|media))(?:/[\w.\-]+)+/?`)
	// Path absolut unix generic (>=2 segmen) — privacy-first (rela over-strip drpd bocor).
	reGenUnixPath = regexp.MustCompile(`(?:/[\w.\-]+){2,}/?`)
	// Id/hash hex panjang (>=16 hex) — sha/uuid-tanpa-dash/blob id.
	reGenLongHex = regexp.MustCompile(`(?i)\b[0-9a-f]{16,}\b`)
	// Rapikan spasi sisa strip.
	reGenSpaces = regexp.MustCompile(`[ \t]{2,}`)
)

// StripDeterministic — Lapis A inti (pola, no nama): redaksi path/url/email/token/id
// owner-spesifik jadi placeholder bermakna (<path>/<url>/<email>/<token>/<id>) supaya
// MAKNA recovery kejaga ("gagal buka <path>") tanpa bocor nilainya. Deterministik &
// idempotent (jalanin 2x = sama). TIDAK butuh data personal apa pun.
func StripDeterministic(s string) string {
	s = reGenURL.ReplaceAllString(s, "<url>")
	s = reGenEmail.ReplaceAllString(s, "<email>")
	s = reGenToken.ReplaceAllString(s, "<token>")
	s = reGenWinPath.ReplaceAllString(s, "<path>")
	s = reGenHomePath.ReplaceAllString(s, "<path>")
	s = reGenUnixPath.ReplaceAllString(s, "<path>")
	s = reGenLongHex.ReplaceAllString(s, "<id>")
	s = reGenSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// RedactNames — strip nama personal owner (word-boundary, case-insensitive) → <name>.
// names = allowlist privat yang di-LOAD RUNTIME (jangan hardcode di repo, D8). names
// kosong = no-op (mis. di unit-test publik). Entri kosong/whitespace di-skip.
func RedactNames(s string, names []string) string {
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(n) + `\b`)
		s = re.ReplaceAllString(s, "<name>")
	}
	return strings.TrimSpace(reGenSpaces.ReplaceAllString(s, " "))
}

// StripRecoveryContent — Lapis A LENGKAP: pola deterministik + nama personal (allowlist
// runtime). Ini JARING PRIVASI yg JAMIN 0 data owner-spesifik di recovery content,
// dipanggil SEBELUM Lapis B (LLM coarsen) DAN atas output-nya (defense in depth).
func StripRecoveryContent(s string, names []string) string {
	return RedactNames(StripDeterministic(s), names)
}

// ───────────────────────── Brand-filter (white-label) ─────────────────────────

// brandRe — nama produk/model AI yg HARAM masuk instinct (white-label). Satu sumber
// kebenaran: dipakai pre-check (raw) di host + post-coarsen di GeneralizeRecovery.
var brandRe = regexp.MustCompile(`(?i)\b(claude|anthropic|fable|gemini|opus|sonnet|haiku|chatgpt|openai|gpt|llama|mistral|deepseek|qwen)\b`)

// ContainsBrand — true kalau s nyebut nama produk/model AI (gagal white-label).
func ContainsBrand(s string) bool { return brandRe.MatchString(s) }

// ───────────────────────── Lapis B — coarsen via LLM ─────────────────────────

func recoveryCoarsenPrompt(stripped string) string {
	return "TUGAS (abaikan skema ekstraksi default): ringkas SATU pelajaran recovery ini " +
		"jadi SATU pola UMUM yang bisa dipakai ulang lintas-konteks.\n" +
		"FORMAT WAJIB tepat satu baris: \"WHEN <kondisi-umum> -> <aksi-perbaikan-umum>\".\n" +
		"Buang detail spesifik. JANGAN sebut nama produk/model AI. JANGAN ada path/nama orang/angka-id.\n" +
		"Output HANYA satu baris pola, tanpa penjelasan/markdown/code-fence.\n\nINPUT:\n" + stripped
}

// looksLikePattern — output coarsen valid kalau berbentuk "WHEN ... -> ...".
func looksLikePattern(s string) bool {
	return strings.Contains(strings.ToUpper(s), "WHEN") && strings.Contains(s, "->")
}

// extractPatternLine — ambil 1 baris pola dari output LLM (buang markdown/code-fence/bullet).
func extractPatternLine(raw string) string {
	best := ""
	for _, ln := range strings.Split(raw, "\n") {
		ln = strings.TrimSpace(strings.Trim(strings.TrimSpace(ln), "`"))
		ln = strings.TrimSpace(strings.TrimPrefix(ln, "- "))
		if !strings.Contains(ln, "->") {
			continue
		}
		if strings.Contains(strings.ToUpper(ln), "WHEN") {
			return ln
		}
		if best == "" {
			best = ln
		}
	}
	return best
}

// coarsenRecovery — Lapis B. LLM gagal/nil/format-salah → FALLBACK ke input (stripped)
// yang udah privacy-safe. Jadi Lapis B cuma nambah kualitas; ga pernah nurunin privasi.
func coarsenRecovery(ctx context.Context, llm LLMFunc, stripped string) string {
	if llm == nil {
		return stripped
	}
	out, err := llm(ctx, recoveryCoarsenPrompt(stripped))
	if err != nil {
		return stripped
	}
	if cand := extractPatternLine(out); looksLikePattern(cand) {
		return cand
	}
	return stripped
}

// ───────────────────────── Pipeline + gerbang ─────────────────────────

// ownerNameAllowlist — nama personal (graph type=person) buat di-redaksi (defense in
// depth atas output LLM). Best-effort; label generik non-identifying di-skip. Aman D8:
// data LOKAL dipakai cuma buat MEMBUANG, ga pernah ditulis ke instinct.
func (s *Store) ownerNameAllowlist() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	rows, err := s.db.Query(`SELECT DISTINCT label FROM cognitive_nodes WHERE type='person' AND label<>'' LIMIT 200`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	skip := map[string]bool{"user": true, "owner": true, "agent": true, "the user": true, "mr.dev": true}
	var out []string
	for rows.Next() {
		var l string
		if rows.Scan(&l) != nil {
			continue
		}
		l = strings.TrimSpace(l)
		if len(l) < 3 || skip[strings.ToLower(l)] {
			continue
		}
		out = append(out, l)
	}
	return out
}

// recoveryNodeID — IDENTITAS DETERMINISTIK instinct. classKey (kelas-error stabil, mis.
// "not-found") → 1 node per-kelas (konvergen lintas tool → hit naik → gerbang firable).
// classKey kosong (mistake non-recovery) → hash stripped-raw (stabil per-konten; ga
// konvergen tapi ga ngerusak). SENGAJA ga pakai embedding LLM (non-deterministik).
func recoveryNodeID(scope, classKey, stripped string) string {
	if classKey != "" {
		return scope + "/instinct/recov-" + classKey
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(stripped)))
	return fmt.Sprintf("%s/instinct/recov%016x", scope, h.Sum64())
}

// GeneralizeRecovery — INTI INC-3: raw recovery-label → instinct UMUM privacy-safe SHADOW.
// Alur: strip(Lapis A) → coarsen(Lapis B) → strip-lagi + brand-check → embed → UpsertNode
// SHADOW dgn ID deterministik (classKey) → dedup otomatis (ON CONFLICT id, hit_count++).
// Return (nodeID, ok, err):
//   - ok=false, err=nil → DROP (kosong/brand/leak) → caller tandai mistake promoted (buang).
//   - err!=nil           → transien (embed/llm) → caller skip, retry tick berikut.
//
// Status node SHADOW; aktivasi lewat PromoteRecoveryShadows (gerbang repetisi).
func (s *Store) GeneralizeRecovery(ctx context.Context, scope, classKey, rawLabel string, embed EmbedFunc, llm LLMFunc) (string, bool, error) {
	names := s.ownerNameAllowlist()
	stripped := StripRecoveryContent(rawLabel, names)
	if stripped == "" {
		return "", false, nil
	}
	coarse := StripRecoveryContent(coarsenRecovery(ctx, llm, stripped), names) // Lapis B + re-strip
	coarse = trimLenG(coarse, 400)
	if coarse == "" || ContainsBrand(coarse) {
		return "", false, nil
	}
	if embed == nil {
		return "", false, fmt.Errorf("embed func nil")
	}
	vec, err := embed(ctx, coarse)
	if err != nil {
		return "", false, fmt.Errorf("embed coarse: %w", err)
	}
	id := recoveryNodeID(scope, classKey, stripped)
	if _, uerr := s.UpsertNode(CogNode{
		ID: id, Label: coarse, Type: "instinct",
		WhereDomain: "recovery", SourceKind: "verified", SourceRef: "recov-gen",
		Confidence: 0.85, Status: "shadow", Embedding: Quantize(vec),
	}); uerr != nil {
		return "", false, uerr
	}
	return id, true, nil
}

// PromoteRecoveryShadows — GERBANG: recovery-instinct SHADOW yg hit_count>=minHits (pola
// coarse-nya BERULANG lintas recovery) → ACTIVE (baru bisa ke-recall). Dijalanin ticker
// non-beku tiap menit. Return jumlah yg ke-promote. minHits<2 dipaksa 2 (anti-degenerasi).
func (s *Store) PromoteRecoveryShadows(minHits int) (int, error) {
	if minHits < 2 {
		minHits = 2
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureCognitiveGraphSchema()
	res, err := s.db.Exec(
		`UPDATE cognitive_nodes SET status='active'
		 WHERE type='instinct' AND where_domain='recovery' AND status='shadow' AND hit_count>=?`,
		minHits)
	if err != nil {
		return 0, fmt.Errorf("promote recovery shadows: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func trimLenG(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
