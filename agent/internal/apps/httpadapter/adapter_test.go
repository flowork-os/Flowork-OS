package httpadapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func writeCfg(t *testing.T, dir string, cfg Config) {
	t.Helper()
	raw, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, ConfigName), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func runOps(t *testing.T, dir string, reqs ...string) []map[string]any {
	t.Helper()
	in := strings.NewReader(strings.Join(reqs, "\n") + "\n")
	var out bytes.Buffer
	if err := Run(in, &out, dir); err != nil {
		t.Fatalf("Run: %v", err)
	}
	var resps []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("resp bukan JSON: %q (%v)", line, err)
		}
		resps = append(resps, m)
	}
	return resps
}

// _url ga butuh server jalan — cuma kasih alamat.
func TestURLOpNoServer(t *testing.T) {
	dir := t.TempDir()
	writeCfg(t, dir, Config{
		StartCmd: []string{"false"}, Port: 12345, URLPath: "/app",
		Ops: map[string]OpSpec{},
	})
	resps := runOps(t, dir, `{"op":"_url","args":{}}`)
	res, _ := resps[0]["result"].(map[string]any)
	if res == nil || res["url"] != "http://127.0.0.1:12345/app" {
		t.Fatalf("_url = %v, mau http://127.0.0.1:12345/app", resps[0])
	}
}

// E2E: spawn server beneran (python http.server) → tunggu ready → op proxy ke HTTP.
func TestHTTPProxyRealServer(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 ga ada — skip E2E http proxy")
	}
	port := freePort(t)
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	// file biar GET / (directory listing) balik 200 + nyebut nama file.
	if err := os.WriteFile(filepath.Join(repo, "MARKER-OK.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCfg(t, dir, Config{
		Workdir:  "repo",
		StartCmd: []string{"python3", "-m", "http.server", fmt.Sprint(port)},
		Port:     port, ReadyPath: "/", StartTimeoutSec: 30,
		Ops: map[string]OpSpec{"list": {Method: "GET", Path: "/"}},
	})
	resps := runOps(t, dir, `{"op":"list","args":{}}`)
	if e, ok := resps[0]["error"]; ok {
		t.Fatalf("op list error: %v", e)
	}
	res, _ := resps[0]["result"].(map[string]any)
	if res == nil {
		t.Fatalf("ga ada result: %v", resps[0])
	}
	if st, _ := res["status"].(float64); st != 200 {
		t.Fatalf("status = %v, mau 200", res["status"])
	}
	bodyStr, _ := res["body"].(string)
	if !strings.Contains(bodyStr, "MARKER-OK.txt") {
		t.Fatalf("body ga nyebut file (proxy GET / gagal?): %q", bodyStr)
	}
}

func TestBadConfig(t *testing.T) {
	dir := t.TempDir()
	// no start_cmd
	writeCfg(t, dir, Config{Port: 8080, Ops: map[string]OpSpec{}})
	var out bytes.Buffer
	if err := Run(strings.NewReader(`{"op":"_url"}`+"\n"), &out, dir); err == nil {
		t.Fatal("mau error config (start_cmd kosong), dapet nil")
	}
}
