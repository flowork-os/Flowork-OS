package vecindex

import (
	"math"
	"testing"
)

// unit-normalize helper (vektor bge-m3 selalu unit).
func unitNorm(v []float32) []float32 {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	n := float32(math.Sqrt(s))
	if n == 0 {
		return v
	}
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = x / n
	}
	return out
}

// Cosine harus ~1.0 buat self-match, ~0 buat ortogonal, dan order-equivalent dgn
// cosine float asli. Ini lantai relevansi #3 (handler search) bergantung sama ini.
func TestCosineApproxAbsolute(t *testing.T) {
	a := unitNorm([]float32{1, 0, 0, 0})
	b := unitNorm([]float32{0, 1, 0, 0})              // ortogonal ke a
	c := unitNorm([]float32{0.9, 0.1, 0.05, 0})       // mirip a
	idx, err := Build([]string{"a", "b", "c"}, [][]float32{a, b, c})
	if err != nil {
		t.Fatal(err)
	}
	// self-match a: cosine harus mendekati 1.
	hits := idx.Search(a, 3)
	got := map[string]float32{}
	for _, h := range hits {
		got[h.ID] = idx.Cosine(h.Score)
	}
	if got["a"] < 0.9 {
		t.Errorf("self-match cosine a = %.3f, mau ~1.0", got["a"])
	}
	if got["a"] <= got["c"] || got["c"] <= got["b"] {
		t.Errorf("order salah: a=%.3f c=%.3f b=%.3f (harus a>c>b)", got["a"], got["c"], got["b"])
	}
	if got["b"] > 0.2 {
		t.Errorf("ortogonal cosine b = %.3f, mau ~0", got["b"])
	}
	// clamp: nil index aman.
	var nilIdx *Index
	if nilIdx.Cosine(123) != 0 {
		t.Error("nil index Cosine harus 0")
	}
}
