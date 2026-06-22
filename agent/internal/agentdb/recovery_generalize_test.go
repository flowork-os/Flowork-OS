package agentdb

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
)

// fakeEmbed — deterministik: teks identik → vektor identik (cosine 1.0).
func fakeEmbed(_ context.Context, text string) ([]float32, error) {
	v := make([]float32, 32)
	for i, r := range text {
		v[i%32] += float32(int(r)%17) + 1
	}
	return v, nil
}

func fakeLLM(out string, err error) LLMFunc {
	return func(_ context.Context, _ string) (string, error) { return out, err }
}

// leakRe = pola yang HARAM tersisa setelah strip (path absolut / token / hex panjang).
// NB: fixture test SENGAJA pakai data GENERIK (user1/example.com) — JANGAN data owner (D8).
var leakRe = regexp.MustCompile(`(?i)(/home/|/users/|/root/|[a-z]:\\|ghp_|https?://|@[\w\-]+\.[\w]|[0-9a-f]{16,})`)

func TestStripDeterministic_NoLeak(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"linux-home", "WHEN file_read gagal (not-found) di /home/user1/Documents/x.go -> RECOVERED"},
		{"linux-var", "gagal baca /var/lib/app/state.db -> retry"},
		{"tilde", "config ~/.app/agents/worker rusak -> rebuild"},
		{"windows", `path C:\Users\Dev\Desktop\file.txt not found -> fix`},
		{"url", "fetch https://api.internal.example/v1/keys?token=abc -> 404"},
		{"email", "kirim ke user@example.com gagal -> retry"},
		{"ghtoken", "auth gagal pakai ghp_AbCd1234EfGh5678IjKl -> refresh"},
		{"longhex", "blob 9f8e7d6c5b4a39281706fedcba987654 hilang -> re-fetch"},
		{"mixed", "tool file_write ke /home/user1/.ssh/id_rsa (sk-abcdefghijkl12345) -> denied"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := StripDeterministic(c.in)
			if loc := leakRe.FindString(out); loc != "" {
				t.Fatalf("LEAK kebobol: %q tersisa di output %q (input %q)", loc, out, c.in)
			}
			if strings.Contains(out, "/home/") || strings.Contains(out, "user1") {
				t.Fatalf("path masih ada: %q", out)
			}
		})
	}
}

func TestStripDeterministic_Placeholders(t *testing.T) {
	out := StripDeterministic("gagal di /home/user1/x.go dan https://a.b/c serta blob 1234567890abcdef12")
	for _, want := range []string{"<path>", "<url>", "<id>"} {
		if !strings.Contains(out, want) {
			t.Errorf("placeholder %s ilang dari %q", want, out)
		}
	}
}

func TestStripDeterministic_Idempotent(t *testing.T) {
	in := "WHEN file_read gagal (not-found) di /home/user1/Documents/x.go -> RECOVERED"
	once := StripDeterministic(in)
	twice := StripDeterministic(once)
	if once != twice {
		t.Fatalf("tidak idempotent:\n once=%q\n twice=%q", once, twice)
	}
}

// Teks recovery coarse (TANPA data spesifik) HARUS lolos utuh — jangan over-strip makna.
func TestStripDeterministic_PreservesCoarse(t *testing.T) {
	in := "WHEN file_read gagal (not-found) -> RECOVERED: agent berhasil di percobaan berikutnya (perbaiki argumen / ganti pendekatan)."
	out := StripDeterministic(in)
	for _, must := range []string{"WHEN file_read", "not-found", "RECOVERED", "ganti pendekatan"} {
		if !strings.Contains(out, must) {
			t.Errorf("makna coarse ke-strip: %q ilang dari %q", must, out)
		}
	}
}

func TestRedactNames(t *testing.T) {
	names := []string{"Alpha", "Bravo", "Charlie"}
	in := "owner alpha minta bravo cek, bukan bravado kuat" // "bravado" != "bravo" (word-boundary)
	out := RedactNames(in, names)
	if strings.Contains(strings.ToLower(out), "alpha") || regexp.MustCompile(`(?i)\bbravo\b`).MatchString(out) {
		t.Fatalf("nama bocor: %q", out)
	}
	if !strings.Contains(out, "bravado") {
		t.Errorf("word-boundary salah: 'bravado' ke-strip padahal beda kata: %q", out)
	}
	if strings.Count(out, "<name>") != 2 {
		t.Errorf("harusnya 2 <name> (alpha, bravo), dapat: %q", out)
	}
}

func TestRedactNames_EmptyAllowlist(t *testing.T) {
	in := "WHEN tool gagal -> RECOVERED"
	if out := RedactNames(in, nil); out != in {
		t.Errorf("allowlist kosong harus no-op, dapat %q", out)
	}
}

func TestStripRecoveryContent_FullNet(t *testing.T) {
	names := []string{"Alpha"}
	in := "WHEN file_read gagal buka /home/user1/alpha-notes.txt milik Alpha (404) -> RECOVERED"
	out := StripRecoveryContent(in, names)
	if leakRe.FindString(out) != "" || strings.Contains(strings.ToLower(out), "alpha") {
		t.Fatalf("jaring privasi bobol: %q", out)
	}
	if !strings.Contains(out, "WHEN file_read") || !strings.Contains(out, "RECOVERED") {
		t.Errorf("makna ilang: %q", out)
	}
}

// ───────────────────────── A2 — Lapis B + gerbang ─────────────────────────

func TestContainsBrand(t *testing.T) {
	if !ContainsBrand("WHEN claude api fails -> retry") {
		t.Error("harusnya kedeteksi brand")
	}
	if ContainsBrand("WHEN file not-found -> check new location") {
		t.Error("false positive brand")
	}
}

func TestExtractPatternLine(t *testing.T) {
	cases := map[string]string{
		"```\nWHEN a -> b\n```":           "WHEN a -> b",
		"- WHEN x -> y":                   "WHEN x -> y",
		"prosa dulu\nWHEN c -> d\nlanjut": "WHEN c -> d",
		"cuma prosa tanpa panah":          "",
	}
	for in, want := range cases {
		if got := extractPatternLine(in); got != want {
			t.Errorf("extractPatternLine(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCoarsenRecovery_Fallback(t *testing.T) {
	ctx := context.Background()
	in := "WHEN file_read gagal (not-found) -> RECOVERED"
	if got := coarsenRecovery(ctx, fakeLLM("", errors.New("boom")), in); got != in {
		t.Errorf("err harusnya fallback ke input, dapat %q", got)
	}
	if got := coarsenRecovery(ctx, fakeLLM("prosa tanpa panah", nil), in); got != in {
		t.Errorf("garbage harusnya fallback, dapat %q", got)
	}
	if got := coarsenRecovery(ctx, nil, in); got != in {
		t.Errorf("nil llm harusnya fallback, dapat %q", got)
	}
	if got := coarsenRecovery(ctx, fakeLLM("WHEN not-found -> cek lokasi baru", nil), in); got != "WHEN not-found -> cek lokasi baru" {
		t.Errorf("valid harusnya pakai LLM, dapat %q", got)
	}
}

// Inti A2 (cabut-akar): dedup pakai KUNCI KELAS deterministik → TAHAN variance LLM.
// 2 raw recovery BEDA (file_read vs codemap), kelas SAMA (not-found), LLM coarsen ke teks
// BEDA tiap call → tetap HARUS dedup (kunci=classKey, bukan embedding teks LLM).
func TestGeneralizeRecovery_DedupAndGate(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	id1, ok1, err1 := s.GeneralizeRecovery(ctx, "agent:test", "not-found",
		"WHEN file_read gagal (not-found) di /home/user1/x.go -> RECOVERED", fakeEmbed,
		fakeLLM("WHEN resource not-found -> retry dgn target yg dibetulin", nil))
	if err1 != nil || !ok1 {
		t.Fatalf("gen#1: ok=%v err=%v", ok1, err1)
	}
	n1, _, _ := s.GetNode(id1)
	if n1.Status != "shadow" {
		t.Fatalf("node baru harus SHADOW, dapat %q", n1.Status)
	}
	if p, _ := s.PromoteRecoveryShadows(2); p != 0 {
		t.Fatalf("gerbang bocor: promote=%d padahal baru 1×", p)
	}

	id2, ok2, err2 := s.GeneralizeRecovery(ctx, "agent:test", "not-found",
		"WHEN codemap_search gagal (not-found) -> RECOVERED", fakeEmbed,
		fakeLLM("WHEN missing resource on lookup -> verify location, alternative method", nil)) // teks BEDA
	if err2 != nil || !ok2 {
		t.Fatalf("gen#2: ok=%v err=%v", ok2, err2)
	}
	if id2 != id1 {
		t.Fatalf("dedup gagal: id1=%q id2=%q (kunci-kelas harus dedup walau teks LLM beda)", id1, id2)
	}
	n2, _, _ := s.GetNode(id1)
	if n2.HitCount != 2 {
		t.Fatalf("hit_count=%d want 2 (dedup lintas recovery)", n2.HitCount)
	}
	if p, _ := s.PromoteRecoveryShadows(2); p != 1 {
		t.Fatalf("promote=%d want 1 (hit=2 lolos gerbang)", p)
	}
	n3, _, _ := s.GetNode(id1)
	if n3.Status != "active" {
		t.Fatalf("setelah gerbang harus ACTIVE, dapat %q", n3.Status)
	}
}

// Defense-in-depth: LLM output yang RE-introduce path → re-strip wajib bersih.
func TestGeneralizeRecovery_RestripLLMLeak(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	leaky := fakeLLM("WHEN file di /home/user1/secret.txt not-found -> cek lokasi", nil)
	id, ok, err := s.GeneralizeRecovery(ctx, "agent:test", "not-found",
		"WHEN file_read gagal (not-found) -> RECOVERED", fakeEmbed, leaky)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	n, _, _ := s.GetNode(id)
	if strings.Contains(n.Label, "/home/") || strings.Contains(n.Label, "user1") {
		t.Fatalf("LLM leak lolos re-strip: %q", n.Label)
	}
}

// Brand di output coarse → DROP (ok=false), ga bikin node.
func TestGeneralizeRecovery_BrandDrop(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	_, ok, err := s.GeneralizeRecovery(ctx, "agent:test", "error",
		"WHEN tool gagal -> RECOVERED", fakeEmbed, fakeLLM("WHEN claude api fails -> retry", nil))
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if ok {
		t.Fatal("brand harusnya DROP (ok=false)")
	}
}
