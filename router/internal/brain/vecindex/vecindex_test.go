package vecindex

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// unit makes a deterministic pseudo-unit vector from a seed.
func unit(dim, seed int) []float32 {
	v := make([]float32, dim)
	var ss float64
	x := uint64(seed*2654435761 + 1)
	for i := range v {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		f := float32(int64(x%2001)-1000) / 1000.0
		v[i] = f
		ss += float64(f) * float64(f)
	}
	n := float32(math.Sqrt(ss))
	for i := range v {
		v[i] /= n
	}
	return v
}

func TestMechanics(t *testing.T) {
	dim := 64
	ids := []string{}
	vecs := [][]float32{}
	for i := 0; i < 200; i++ {
		ids = append(ids, "d"+strconv.Itoa(i))
		vecs = append(vecs, unit(dim, i))
	}
	ix, err := Build(ids, vecs)
	if err != nil {
		t.Fatal(err)
	}
	if ix.Len() != 200 || ix.Dim() != 64 {
		t.Fatalf("len/dim wrong: %d %d", ix.Len(), ix.Dim())
	}
	// querying with an exact corpus vector must return itself as #1.
	for _, qi := range []int{0, 7, 50, 199} {
		hits := ix.Search(vecs[qi], 5)
		if len(hits) != 5 {
			t.Fatalf("want 5 hits, got %d", len(hits))
		}
		if hits[0].ID != ids[qi] {
			t.Errorf("query d%d: top hit = %s, want %s", qi, hits[0].ID, ids[qi])
		}
		// scores must be descending
		for i := 1; i < len(hits); i++ {
			if hits[i].Score > hits[i-1].Score {
				t.Errorf("scores not descending at %d", i)
			}
		}
	}
}

func TestSaveLoad(t *testing.T) {
	dim := 32
	ids := []string{}
	vecs := [][]float32{}
	for i := 0; i < 50; i++ {
		ids = append(ids, "x"+strconv.Itoa(i))
		vecs = append(vecs, unit(dim, i+1000))
	}
	ix, _ := Build(ids, vecs)
	p := filepath.Join(t.TempDir(), "idx.tv")
	if err := ix.Save(p); err != nil {
		t.Fatal(err)
	}
	ld, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if ld.Len() != ix.Len() || ld.Dim() != ix.Dim() {
		t.Fatalf("load mismatch")
	}
	a := ix.Search(vecs[10], 5)
	b := ld.Search(vecs[10], 5)
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Score != b[i].Score {
			t.Errorf("save/load search differs at %d: %+v vs %+v", i, a[i], b[i])
		}
	}
}

// readF32 reads a little-endian float32 raw file into rows of `dim`.
func readF32(t *testing.T, path string, dim int) [][]float32 {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("fixture %s absent: %v", path, err)
	}
	n := len(b) / 4 / dim
	out := make([][]float32, n)
	for i := 0; i < n; i++ {
		row := make([]float32, dim)
		for j := 0; j < dim; j++ {
			row[j] = math.Float32frombits(binary.LittleEndian.Uint32(b[(i*dim+j)*4:]))
		}
		out[i] = row
	}
	return out
}

// TestRealRecall validates recall against exact cosine on a REAL bge-m3 sample,
// IF the ephemeral fixture exists (exported by the prototype). Skips otherwise so
// the test stays portable (real brain data is never committed).
func TestRealRecall(t *testing.T) {
	dir := "/tmp/tvproto"
	meta, err := os.ReadFile(filepath.Join(dir, "fix_meta.txt"))
	if err != nil {
		t.Skip("no real-data fixture; skipping recall test")
	}
	parts := strings.Fields(string(meta))
	nCorp, _ := strconv.Atoi(parts[0])
	dim, _ := strconv.Atoi(parts[1])
	nQ, _ := strconv.Atoi(parts[2])
	K, _ := strconv.Atoi(parts[3])

	corpus := readF32(t, filepath.Join(dir, "fix_corpus.f32"), dim)
	queries := readF32(t, filepath.Join(dir, "fix_query.f32"), dim)
	if len(corpus) != nCorp || len(queries) != nQ {
		t.Fatalf("fixture shape mismatch")
	}
	// exact top-K (int32 indices)
	eb, _ := os.ReadFile(filepath.Join(dir, "fix_exact.i32"))
	exact := make([][]int, nQ)
	for i := 0; i < nQ; i++ {
		exact[i] = make([]int, K)
		for j := 0; j < K; j++ {
			exact[i][j] = int(int32(binary.LittleEndian.Uint32(eb[(i*K+j)*4:])))
		}
	}
	ids := make([]string, nCorp)
	for i := range ids {
		ids[i] = strconv.Itoa(i)
	}
	ix, err := Build(ids, corpus)
	if err != nil {
		t.Fatal(err)
	}
	var sum10, sum1 float64
	for i := 0; i < nQ; i++ {
		hits := ix.Search(queries[i], K)
		got := map[string]bool{}
		for _, h := range hits {
			got[h.ID] = true
		}
		exactSet := map[int]bool{}
		for _, e := range exact[i] {
			exactSet[e] = true
		}
		inter := 0
		for _, h := range hits {
			id, _ := strconv.Atoi(h.ID)
			if exactSet[id] {
				inter++
			}
		}
		sum10 += float64(inter) / float64(K)
		if id, _ := strconv.Atoi(hits[0].ID); id == exact[i][0] {
			sum1++
		}
	}
	r10 := sum10 / float64(nQ)
	r1 := sum1 / float64(nQ)
	t.Logf("REAL bge-m3 recall@10=%.3f recall@1=%.3f (corpus=%d dim=%d)", r10, r1, nCorp, dim)
	if r10 < 0.95 {
		t.Errorf("recall@10 %.3f < 0.95 — Go quantizer worse than numpy prototype (0.98)", r10)
	}
	if r1 < 0.95 {
		t.Errorf("recall@1 %.3f < 0.95", r1)
	}
}
