// Package sidecar — SUMBER KEBENARAN PATH tunggal untuk Flowork (roadmap_sidecar.md,
// owner 2026-06-20). Tujuan: cabut akar "konten/data nyebar di repo + ~/.flowork +
// ~/.flow_router + exe-dir". Satu anchor, dua laci:
//
//	$FLOWORK_SIDECAR/
//	├── content/   🟢 shippable (apps bawaan, brain-seed, models, skills bawaan, bin)
//	└── data/      🔴 user (agents, workspace, media, brain-memori, flowork.db, …)
//
// Baca overlay: data MENANG atas content (app installed override bawaan).
//
// KOMPAT (penting): kalau sidecar BELUM aktif (env FLOWORK_SIDECAR kosong), SEMUA
// resolver fall back ke lokasi LEGACY (~/.flowork/*, $FLOWORK_AGENTS_DIR, dst) →
// mesin existing + build portable sekarang ZERO keganggu. Build baru (Fase 4) yang
// set FLOWORK_SIDECAR buat ngaktifin layout baru.
//
// CATATAN MODUL: agent (flowork-gui) & router (flowork_Router) modul terpisah & sengaja
// independen — paket ini diduplikat di tiap modul. Yg di-share KONTRAK layout (content/
// data di bawah root), bukan kode. Jaga dua salinan tetap sinkron.
package sidecar

import (
	"os"
	"path/filepath"
	"strings"
)

// Root — akar sidecar kalau aktif, "" kalau legacy (env FLOWORK_SIDECAR kosong).
// SENGAJA belum baca FLOWORK_HOME: env itu masih dipakai make-portable sbg data-home
// FLAT ($FLOWORK_HOME/.flowork). Ngalias-in sekarang = mecahin portable. Fase 4 yg
// nyambungin FLOWORK_HOME→FLOWORK_SIDECAR bareng update build.
func Root() string {
	return strings.TrimSpace(os.Getenv("FLOWORK_SIDECAR"))
}

// Active — true kalau layout sidecar dipakai (env di-set).
func Active() bool { return Root() != "" }

// ContentDir — <root>/content/<parts...> kalau aktif, else "".
func ContentDir(parts ...string) string {
	r := Root()
	if r == "" {
		return ""
	}
	return filepath.Join(append([]string{r, "content"}, parts...)...)
}

// DataDir — <root>/data/<parts...> kalau aktif, else "".
func DataDir(parts ...string) string {
	r := Root()
	if r == "" {
		return ""
	}
	return filepath.Join(append([]string{r, "data"}, parts...)...)
}

// exeDir — folder binary (buat content bawaan yg dikirim di sebelah exe).
func exeDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return ""
}

// ── RESOLVER RESOURCE AGENT (legacy-default; sidecar kalau aktif) ─────────────
//
// Tiap fungsi: kalau sidecar aktif → laci baru; else → PERSIS path legacy lama
// (zero behavior change). Call-site lama dipindah ke sini bertahap per fase.

// AgentsDir — folder <id>.fwagent (state.db, workspace, loket per agent). 🔴 data.
// Legacy = loader.AgentsDir() PERSIS: $FLOWORK_AGENTS_DIR → ~/.flowork/agents →
// /tmp/flowork-agents (last resort biar headless smoke test punya target writable).
func AgentsDir() string {
	if d := DataDir("agents"); d != "" {
		return d
	}
	if v := strings.TrimSpace(os.Getenv("FLOWORK_AGENTS_DIR")); v != "" {
		return v
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".flowork", "agents")
	}
	return "/tmp/flowork-agents"
}

// FloworkDB — path flowork.db (owner-level: auth, settings, wallet). 🔴 data.
// Legacy = floworkdb.Path() PERSIS: $FLOWORK_DATA_DIR/flowork.db → ~/.flowork/
// flowork.db → /tmp/flowork/flowork.db.
func FloworkDB() string {
	if d := DataDir(); d != "" {
		return filepath.Join(d, "flowork.db")
	}
	if v := strings.TrimSpace(os.Getenv("FLOWORK_DATA_DIR")); v != "" {
		return filepath.Join(v, "flowork.db")
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".flowork", "flowork.db")
	}
	return filepath.Join("/tmp", "flowork", "flowork.db")
}

// AppsDataDir — app INSTALLED user (overlay yg menang). 🔴 data.
// Legacy = Dir(AgentsDir())/apps PERSIS (appsMgr lama).
func AppsDataDir() string {
	if d := DataDir("apps"); d != "" {
		return d
	}
	return filepath.Join(filepath.Dir(AgentsDir()), "apps")
}

// AppsContentDir — app BAWAAN (shippable, read-only). 🟢 content. "" kalau ga ada.
// Legacy/dev: <exe-dir>/apps (portable) atau "" (dev jalan dari repo — bawaan udah
// ke-seed ke data-home). Sidecar: <root>/content/apps.
func AppsContentDir() string {
	if d := ContentDir("apps"); d != "" {
		return d
	}
	if ed := exeDir(); ed != "" {
		if p := filepath.Join(ed, "apps"); dirExists(p) {
			return p
		}
	}
	return ""
}

func dirExists(p string) bool {
	if p == "" {
		return false
	}
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
