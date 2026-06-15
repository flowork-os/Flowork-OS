package agentdb

import (
	"strings"
	"testing"
)

func TestClassifyPillars(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		wantFit   bool
		wantHas   string // pilar id yg WAJIB ada (kalau wantFit)
	}{
		{"bounty-ekonomi", "tambah agent Bounty Hunter buat submit ke HackerOne cari duit", true, "ekonomi"},
		{"sandbox-keamanan", "perketat sandbox + audit capability biar gak ada RCE", true, "keamanan"},
		{"selfheal-mandiri", "tambah watchdog auto-restart biar Flowork tetap hidup tanpa owner", true, "mandiri"},
		{"warga-mudahin", "bikin onboarding biar user/warga lain gampang pakai", true, "warga"},
		{"cerdas-evolusi", "tingkatin reasoning + RAG biar jawaban lebih akurat", true, "kecerdasan"},
		{"multi-pilar", "fitur monetisasi yang aman (sandbox) + mudahin user", true, "ekonomi"},
		{"ngelantur-1", "tambah easter egg gambar kucing pelangi pas tahun baru", false, ""},
		{"ngelantur-2", "ganti warna tombol jadi ungu biar lucu", false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ClassifyPillars(c.text)
			fit := PillarFit(c.text)
			if fit != c.wantFit {
				t.Fatalf("PillarFit(%q)=%v want %v (pilar=%v)", c.text, fit, c.wantFit, got)
			}
			if c.wantFit {
				found := false
				for _, p := range got {
					if p == c.wantHas {
						found = true
					}
				}
				if !found {
					t.Fatalf("ClassifyPillars(%q)=%v, harus mengandung %q", c.text, got, c.wantHas)
				}
				// label gak boleh kosong kalau fit
				if PillarsLabel(got) == "" {
					t.Fatalf("PillarsLabel kosong padahal fit, pilar=%v", got)
				}
			}
		})
	}
}

// pilar id harus kanonik + unik
func TestPillarsCanonical(t *testing.T) {
	seen := map[string]bool{}
	want := "ekonomi keamanan warga kecerdasan mandiri"
	got := []string{}
	for _, p := range EvolvePillars {
		if seen[p.ID] {
			t.Fatalf("pilar duplikat: %s", p.ID)
		}
		seen[p.ID] = true
		if len(p.kw) == 0 {
			t.Fatalf("pilar %s gak punya keyword", p.ID)
		}
		got = append(got, p.ID)
	}
	if strings.Join(got, " ") != want {
		t.Fatalf("urutan/isi pilar = %q, want %q", strings.Join(got, " "), want)
	}
}
