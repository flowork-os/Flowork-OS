// selfevolve_coreapply_test.go — bukti deterministik guard core-apply (B1). Batas keamanan
// 🔴: path-traversal ditolak, file LOCKED kedeteksi, fence ke-strip, modul Go ketemu.

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// evolveCommitFile beneran nge-commit ke repo (B2) — diuji di temp git repo (no real-repo risk).
func TestEvolveCommitFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git ga ada")
	}
	repo := t.TempDir()
	ctx := context.Background()
	run := func(args ...string) {
		c := exec.Command("git", append([]string{"-C", repo}, args...)...)
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q")
	_ = os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("x"), 0o644)
	run("add", "-A")
	run("commit", "-qm", "seed")

	hash, err := evolveCommitFile(ctx, repo, "pkg/new.go", "package pkg\n", "evolve(auto): test")
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if hash == "" {
		t.Fatal("empty hash")
	}
	// File ada + ke-commit.
	if _, err := os.Stat(filepath.Join(repo, "pkg", "new.go")); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	logOut, _ := exec.Command("git", "-C", repo, "log", "-1", "--pretty=%an|%s").Output()
	got := strings.TrimSpace(string(logOut))
	if !strings.Contains(got, "Flowork Evolusi") || !strings.Contains(got, "evolve(auto): test") {
		t.Errorf("commit author/msg salah: %q", got)
	}
	// Path-scoped: file dirty lain di working tree TIDAK ikut ke-commit.
	_ = os.WriteFile(filepath.Join(repo, "dirty.txt"), []byte("y"), 0o644)
	h2, err := evolveCommitFile(ctx, repo, "pkg/two.go", "package pkg\n", "evolve(auto): two")
	if err != nil {
		t.Fatalf("commit2: %v", err)
	}
	if h2 == hash {
		t.Error("hash should advance")
	}
	st, _ := exec.Command("git", "-C", repo, "status", "--porcelain").Output()
	if !strings.Contains(string(st), "dirty.txt") {
		t.Error("dirty.txt should remain uncommitted (path-scoped commit)")
	}
}

func TestEvolveSafeRepoPath(t *testing.T) {
	root := "/repo"
	cases := []struct {
		rel     string
		wantOK  bool
		wantRel string
	}{
		{"agent/foo.go", true, "agent/foo.go"},
		{"internal/x/y.go", true, "internal/x/y.go"},
		{"../etc/passwd", false, ""},
		{"/etc/passwd", false, ""},
		{"", false, ""},
		{"a/../../b", false, ""},
		{"./a.go", true, "a.go"},
	}
	for _, c := range cases {
		got, ok := evolveSafeRepoPath(root, c.rel)
		if ok != c.wantOK {
			t.Errorf("rel=%q: ok=%v want %v", c.rel, ok, c.wantOK)
			continue
		}
		if ok && got != c.wantRel {
			t.Errorf("rel=%q: got %q want %q", c.rel, got, c.wantRel)
		}
	}
}

func TestEvolveStripFence(t *testing.T) {
	cases := map[string]string{
		"```go\npackage x\n```": "package x",
		"plain code":            "plain code",
		"```\nhello\n```":       "hello",
		"  ```js\na=1\n```  ":   "a=1",
	}
	for in, want := range cases {
		if got := evolveStripFence(in); got != want {
			t.Errorf("strip(%q)=%q want %q", in, got, want)
		}
	}
}

func TestEvolveFileLocked(t *testing.T) {
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked.go")
	_ = os.WriteFile(locked, []byte("// === LOCKED FILE (soft) ===\npackage x\n"), 0o644)
	plain := filepath.Join(dir, "plain.go")
	_ = os.WriteFile(plain, []byte("package x\n"), 0o644)
	if !evolveFileLocked(locked) {
		t.Error("locked file not detected")
	}
	if evolveFileLocked(plain) {
		t.Error("plain file falsely flagged locked")
	}
	if evolveFileLocked(filepath.Join(dir, "nope.go")) {
		t.Error("missing file should not be locked")
	}
}

func TestEvolveResolveTarget(t *testing.T) {
	root := t.TempDir()
	// root/agent/go.mod + root/agent/internal/agentdb (modul agent, package dir ada).
	_ = os.MkdirAll(filepath.Join(root, "agent", "internal", "agentdb"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "agent", "go.mod"), []byte("module x\n"), 0o644)
	// proposal target relatif modul agent → harus di-prefix "agent/".
	if got := evolveResolveTarget(root, "internal/agentdb/new.go"); got != filepath.Join("agent", "internal/agentdb/new.go") {
		t.Errorf("resolve agent-relative: got %q", got)
	}
	// udah repo-relatif (folder induk ada langsung) → as-is.
	if got := evolveResolveTarget(root, "agent/internal/agentdb/new.go"); got != "agent/internal/agentdb/new.go" {
		t.Errorf("resolve already-correct: got %q", got)
	}
	// folder bener-bener baru (ga ada di mana-mana) → as-is.
	if got := evolveResolveTarget(root, "brandnew/pkg/x.go"); got != "brandnew/pkg/x.go" {
		t.Errorf("resolve brand-new: got %q", got)
	}
}

func TestEvolveModuleDir(t *testing.T) {
	root := t.TempDir()
	modDir := filepath.Join(root, "agent")
	_ = os.MkdirAll(filepath.Join(modDir, "internal", "x"), 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module x\n"), 0o644)
	got := evolveModuleDir(root, "agent/internal/x/new.go")
	if got != modDir {
		t.Errorf("module dir=%q want %q", got, modDir)
	}
	// di luar modul (no go.mod) → "".
	if d := evolveModuleDir(root, "docs/readme.md"); d != "" {
		t.Errorf("no-module path should give empty, got %q", d)
	}
}
