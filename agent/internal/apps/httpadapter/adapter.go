// httpadapter — HTTP-Adapter Core (ROADMAP_REPO_TO_APP F5, kontrak HTTP).
// Buat repo yang = SERVER (web app / API: streamlit, fastapi, express, dll) — bukan CLI request-response.
// Adapter ini: spawn server repo (proses panjang) → tunggu port ready → jembatani tiap op ke HTTP call.
// Ngomong protokol core Flowork yang SAMA (stdio baris-JSON, lihat proc.go) biar engine ga usah tau apa-apa.
//
// Op khusus: "_url" balikin URL UI (buat GUI iframe). "_alive" status server. Op lain → HTTP call ke server.
// Server mati saat adapter ditutup (die-with-parent). White-label, multi-OS (pure-Go, no shell), CGO-off.
//
// cliadapter (kontrak CLI) FROZEN — ini adapter TERPISAH (kontrak baru = adapter baru, bukan bongkar beku).
package httpadapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ConfigName — file konfigurasi di folder app (cwd adapter).
const ConfigName = "httpadapter.json"

const (
	defaultStartTimeout = 120 * time.Second
	defaultCallTimeout  = 120 * time.Second
)

// OpSpec — satu operasi → HTTP request ke server. {key} di Path disubstitusi dari args.
type OpSpec struct {
	Method string `json:"method"` // GET|POST|PUT|DELETE (default GET)
	Path   string `json:"path"`   // mis "/api/v1/videos" (boleh {key})
	Body   string `json:"body"`   // "json" = kirim args sbg JSON body; "" = ga ada body
}

// Config — isi httpadapter.json.
type Config struct {
	Workdir         string            `json:"workdir"`           // subdir server jalan (default "repo")
	StartCmd        []string          `json:"start_cmd"`         // argv launch server (mis ["python","main.py"])
	Host            string            `json:"host"`              // default 127.0.0.1
	Port            int               `json:"port"`              // port server listen (wajib)
	ReadyPath       string            `json:"ready_path"`        // path cek ready (default "/"); "" = cuma cek TCP
	URLPath         string            `json:"url_path"`          // path UI manusia (default "/")
	StartTimeoutSec int               `json:"start_timeout_sec"` // tunggu port (default 120)
	Env             map[string]string `json:"env"`               // env tambahan (config/white-label dari sini)
	Ops             map[string]OpSpec `json:"ops"`               // op agent → HTTP
}

func (c *Config) host() string {
	if strings.TrimSpace(c.Host) == "" {
		return "127.0.0.1"
	}
	return c.Host
}

func (c *Config) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host(), c.Port)
}

// LoadConfig — baca httpadapter.json dari folder app.
func LoadConfig(baseDir string) (*Config, error) {
	raw, err := os.ReadFile(filepath.Join(baseDir, ConfigName))
	if err != nil {
		return nil, fmt.Errorf("baca %s: %w", ConfigName, err)
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ConfigName, err)
	}
	if len(c.StartCmd) == 0 || strings.TrimSpace(c.StartCmd[0]) == "" {
		return nil, errors.New(ConfigName + " tak punya 'start_cmd'")
	}
	if c.Port <= 0 {
		return nil, errors.New(ConfigName + " tak punya 'port' valid")
	}
	return &c, nil
}

// server — proses server repo yang diawasi adapter.
type server struct {
	cfg     *Config
	workdir string
	cmd     *exec.Cmd
	started bool
}

// ensure — spawn server (sekali) lalu tunggu port ready. Idempotent.
func (s *server) ensure() error {
	if s.started {
		return nil
	}
	argv := s.cfg.StartCmd
	prog := argv[0]
	// program relatif ber-separator → resolve ke workdir (sama logika cliadapter).
	if !filepath.IsAbs(prog) && strings.ContainsAny(prog, `/\`) {
		prog = filepath.Join(s.workdir, filepath.FromSlash(prog))
	}
	cmd := exec.Command(prog, argv[1:]...) // #nosec — argv (no shell), start_cmd dari httpadapter.json owner-installed
	cmd.Dir = s.workdir
	cmd.Stdout = os.Stderr // log server ke stderr adapter (kebawa ke log host)
	cmd.Stderr = os.Stderr
	env := os.Environ()
	for k, v := range s.cfg.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server %q: %w", prog, err)
	}
	s.cmd = cmd
	s.started = true

	timeout := defaultStartTimeout
	if s.cfg.StartTimeoutSec > 0 {
		timeout = time.Duration(s.cfg.StartTimeoutSec) * time.Second
	}
	if err := waitReady(s.cfg, timeout); err != nil {
		s.stop()
		return err
	}
	return nil
}

func (s *server) stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		s.cmd = nil
	}
	s.started = false
}

// waitReady — poll sampai port kebuka (TCP) lalu (kalau ada ReadyPath) sampai HTTP balik.
func waitReady(cfg *Config, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(cfg.host(), strconv.Itoa(cfg.Port))
	// fase 1: TCP kebuka
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if time.Now().After(deadline) {
		return fmt.Errorf("server ga ready di %s (timeout %s)", addr, timeout)
	}
	// fase 2: HTTP ReadyPath balik apa pun (opsional)
	if strings.TrimSpace(cfg.ReadyPath) != "" {
		client := &http.Client{Timeout: 3 * time.Second}
		for time.Now().Before(deadline) {
			resp, err := client.Get(cfg.baseURL() + cfg.ReadyPath)
			if err == nil {
				resp.Body.Close()
				return nil
			}
			time.Sleep(300 * time.Millisecond)
		}
		return fmt.Errorf("server %s ga balas %s (timeout)", addr, cfg.ReadyPath)
	}
	return nil
}

type request struct {
	Op   string          `json:"op"`
	Args json.RawMessage `json:"args"`
}

// Run — loop core stdio: baca {op,args}, jembatani ke HTTP server. baseDir = folder app.
func Run(in io.Reader, out io.Writer, baseDir string) error {
	cfg, err := LoadConfig(baseDir)
	if err != nil {
		return err
	}
	wd := cfg.Workdir
	if strings.TrimSpace(wd) == "" {
		wd = "."
	}
	full := filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(wd)))
	if rel, e := filepath.Rel(baseDir, full); e != nil || strings.HasPrefix(rel, "..") {
		return errors.New("workdir di luar folder app: " + wd)
	}
	srv := &server{cfg: cfg, workdir: full}
	defer srv.stop()

	r := bufio.NewReaderSize(in, 1<<20)
	w := bufio.NewWriter(out)
	defer w.Flush()
	client := &http.Client{Timeout: defaultCallTimeout}
	var stateVer int64

	for {
		line, rerr := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			stateVer = handle(w, cfg, srv, client, line, stateVer)
			_ = w.Flush()
		}
		if rerr != nil {
			return nil
		}
	}
}

func handle(w io.Writer, cfg *Config, srv *server, client *http.Client, line []byte, stateVer int64) int64 {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		writeError(w, "request bukan JSON: "+err.Error())
		return stateVer
	}
	// op khusus "_url" — ga butuh server jalan dulu (cuma kasih alamat).
	if req.Op == "_url" {
		up := cfg.URLPath
		if up == "" {
			up = "/"
		}
		stateVer++
		writeResult(w, map[string]any{"url": cfg.baseURL() + up}, stateVer)
		return stateVer
	}
	// op lain butuh server hidup.
	if err := srv.ensure(); err != nil {
		writeError(w, err.Error())
		return stateVer
	}
	if req.Op == "_alive" {
		stateVer++
		writeResult(w, map[string]any{"alive": true, "base_url": cfg.baseURL()}, stateVer)
		return stateVer
	}
	spec, ok := cfg.Ops[req.Op]
	if !ok {
		writeError(w, "op tak terdaftar di "+ConfigName+": "+req.Op)
		return stateVer
	}
	args := map[string]any{}
	if len(req.Args) > 0 {
		_ = json.Unmarshal(req.Args, &args)
	}
	res, err := doCall(client, cfg, spec, args)
	if err != nil {
		writeError(w, err.Error())
		return stateVer
	}
	stateVer++
	writeResult(w, res, stateVer)
	return stateVer
}

// doCall — rakit + kirim HTTP request dari OpSpec + args.
func doCall(client *http.Client, cfg *Config, spec OpSpec, args map[string]any) (map[string]any, error) {
	method := strings.ToUpper(strings.TrimSpace(spec.Method))
	if method == "" {
		method = "GET"
	}
	path := spec.Path
	if path == "" {
		path = "/"
	}
	// substitusi {key} di path dari args.
	for k, v := range args {
		path = strings.ReplaceAll(path, "{"+k+"}", toStr(v))
	}
	var body io.Reader
	if strings.EqualFold(spec.Body, "json") {
		b, _ := json.Marshal(args)
		body = bytes.NewReader(b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultCallTimeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, method, cfg.baseURL()+path, body)
	if err != nil {
		return nil, fmt.Errorf("rakit request: %w", err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	out := map[string]any{"status": resp.StatusCode}
	// coba parse JSON; gagal → kasih mentah.
	var j any
	if json.Unmarshal(raw, &j) == nil {
		out["json"] = j
	} else {
		out["body"] = string(raw)
	}
	return out, nil
}

func toStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func writeResult(w io.Writer, result any, stateVer int64) {
	raw, err := json.Marshal(result)
	if err != nil {
		writeError(w, "marshal result: "+err.Error())
		return
	}
	resp, _ := json.Marshal(map[string]any{"result": json.RawMessage(raw), "state_version": stateVer})
	_, _ = w.Write(append(resp, '\n'))
}

func writeError(w io.Writer, msg string) {
	resp, _ := json.Marshal(map[string]any{"error": msg})
	_, _ = w.Write(append(resp, '\n'))
}
