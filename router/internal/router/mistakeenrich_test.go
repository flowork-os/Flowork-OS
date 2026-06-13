package router

import (
	"strings"
	"testing"
	"time"

	"github.com/flowork-os/flowork_Router/internal/brain"
)

// Pakai waktu acuan tetap biar test deterministik.
var refNow = time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)

func fresh() string { return refNow.Format(time.RFC3339) }
func daysAgo(d int) string {
	return refNow.Add(-time.Duration(d) * 24 * time.Hour).Format(time.RFC3339)
}

// Antibody ranking: karma(hit_count) × relevansi(overlap) × decay(recency).
func TestRankAntibodies_KarmaTimesRelevance(t *testing.T) {
	ms := []brain.Mistake{
		{ID: 1, Category: "logic", Title: "task_run kategori saham", Content: "kategori wajib kanonik", HitCount: 20, UpdatedAt: fresh()},
		{ID: 2, Category: "safety", Title: "FK constraint", Content: "foreign key", HitCount: 12, UpdatedAt: fresh()},
		{ID: 3, Category: "ux", Title: "tombol warna", Content: "kontras kurang", HitCount: 3, UpdatedAt: fresh()},
		{ID: 4, Category: "logic", Title: "saham subject murni", Content: "tanpa suffix JK", HitCount: 4, UpdatedAt: fresh()},
	}
	got := rankAntibodies(ms, "tolong analisa saham BBCA", antibodyMaxInject, refNow)
	if len(got) == 0 {
		t.Fatal("ranking kosong — harusnya ada antibodi relevan")
	}
	if got[0].ID != 1 {
		t.Fatalf("antibodi #1 harusnya ID=1 (relevan+karma tinggi), dapet ID=%d", got[0].ID)
	}
	for _, m := range got {
		if m.ID == 3 {
			t.Fatal("ID=3 (noise: ga relevan + karma rendah) bocor ke hasil")
		}
	}
	if len(got) > antibodyMaxInject {
		t.Fatalf("melebihi cap MAX %d: dapet %d", antibodyMaxInject, len(got))
	}
}

// DECAY: dua antibodi karma+relevansi SAMA, yang updated_at lebih baru menang.
func TestRankAntibodies_RecencyDecay(t *testing.T) {
	ms := []brain.Mistake{
		{ID: 10, Category: "logic", Title: "saham x", Content: "y", HitCount: 15, UpdatedAt: daysAgo(120)}, // tua → pudar
		{ID: 11, Category: "logic", Title: "saham x", Content: "y", HitCount: 15, UpdatedAt: fresh()},      // fresh → menang
	}
	got := rankAntibodies(ms, "saham", 2, refNow)
	if len(got) != 2 || got[0].ID != 11 {
		t.Fatalf("antibodi fresh (ID=11) harusnya rank atas decay yg tua, dapet %+v", got)
	}
}

func TestRecencyFactor_DecayCurve(t *testing.T) {
	if f := recencyFactor(fresh(), refNow); f < 0.99 {
		t.Fatalf("fresh harusnya ~1.0, dapet %.3f", f)
	}
	half := recencyFactor(daysAgo(30), refNow) // 1 half-life → ~0.5
	if half < 0.45 || half > 0.55 {
		t.Fatalf("30 hari (1 half-life) harusnya ~0.5, dapet %.3f", half)
	}
	floor := recencyFactor(daysAgo(3650), refNow) // sangat tua → floor
	if floor < antibodyRecencyFloor-0.001 || floor > antibodyRecencyFloor+0.001 {
		t.Fatalf("sangat tua harusnya floor %.2f, dapet %.3f", antibodyRecencyFloor, floor)
	}
	if bad := recencyFactor("bukan-tanggal", refNow); bad != antibodyRecencyFloor {
		t.Fatalf("format invalid harusnya floor, dapet %.3f", bad)
	}
}

func TestRankAntibodies_CapEnforced(t *testing.T) {
	var ms []brain.Mistake
	for i := 0; i < 10; i++ {
		ms = append(ms, brain.Mistake{ID: int64(i), Category: "logic", Title: "saham x", Content: "y", HitCount: 15, UpdatedAt: fresh()})
	}
	if got := rankAntibodies(ms, "saham", antibodyMaxInject, refNow); len(got) != antibodyMaxInject {
		t.Fatalf("cap tidak ditegakkan: expected %d, got %d", antibodyMaxInject, len(got))
	}
}

func TestBuildAntibodySystem_FormatAndEmpty(t *testing.T) {
	if s := buildAntibodySystem(nil); s != "" {
		t.Fatalf("kosong harus return '', dapet %q", s)
	}
	s := buildAntibodySystem([]brain.Mistake{
		{Category: "logic", Title: "kategori kanonik", Content: "pakai saham", HitCount: 20},
	})
	for _, want := range []string{"Antibodi", "logic", "kuat 20", "kategori kanonik", "pakai saham"} {
		if !strings.Contains(s, want) {
			t.Fatalf("system msg ga ngandung %q:\n%s", want, s)
		}
	}
}

func TestTokenSet_StopwordAndShort(t *testing.T) {
	ts := tokenSet("analisa saham yang dan ok")
	if _, ok := ts["yang"]; ok {
		t.Fatal("stopword 'yang' harusnya kebuang")
	}
	if _, ok := ts["ok"]; ok {
		t.Fatal("token pendek 'ok' (<3) harusnya kebuang")
	}
	if _, ok := ts["saham"]; !ok {
		t.Fatal("token valid 'saham' harusnya ada")
	}
}
