package tools

import (
	"context"
	"strings"
	"testing"
)

// TestInterceptorsEducational — confinement HARUS jalan (blok ..), DAN errornya
// edukatif (memeluk + petunjuk + framing "ini aturan"), sesuai hukum owner:
// jangan marahin agent. Sekaligus bukti klaim workspace-confinement (tutorial).
func TestInterceptorsEducational(t *testing.T) {
	wp := workspacePathInterceptor{}
	// 1) traversal '..' diblok
	err := wp.Before(context.Background(), nil, map[string]any{"path": "../../etc/passwd"})
	if err == nil {
		t.Fatal("traversal '..' harus diblok")
	}
	if !isEducational(err.Error()) {
		t.Errorf("error traversal harus edukatif, dapat: %q", err.Error())
	}
	// 2) lokasi sistem terlindungi diblok
	err = wp.Before(context.Background(), nil, map[string]any{"file": "/etc/shadow"})
	if err == nil {
		t.Fatal("path /etc/ harus diblok")
	}
	if !isEducational(err.Error()) {
		t.Errorf("error protected-path harus edukatif, dapat: %q", err.Error())
	}
	// 3) path bersih DI DALAM workspace → lolos
	if err := wp.Before(context.Background(), nil, map[string]any{"path": "document/catatan.txt"}); err != nil {
		t.Errorf("path workspace bersih harus lolos, malah: %v", err)
	}
	// 4) file sensitif diblok + edukatif
	sf := sensitiveFileInterceptor{}
	err = sf.Before(context.Background(), nil, map[string]any{"name": "id_rsa"})
	if err == nil || !isEducational(err.Error()) {
		t.Errorf("blok file sensitif harus edukatif, dapat: %v", err)
	}
}

// isEducational — pesan "memeluk": ada framing bukan-salah + ada petunjuk.
func isEducational(msg string) bool {
	m := strings.ToLower(msg)
	hug := strings.Contains(m, "bukan salahmu") || strings.Contains(m, "ini aturan")
	hint := strings.Contains(m, "petunjuk") || strings.Contains(m, "coba") || strings.Contains(m, "pakai") || strings.Contains(m, "kalau")
	return hug && hint
}
