// suggest.go — AUTO-SARAN KONTRAK (ROADMAP_REPO_TO_APP). Deteksi framework server (web app/API)
// dari dependency repo → saranin kontrak HTTP + port + start_cmd, biar owner ga nebak manual ("setting dikit").
// Heuristik (bukan beku) — gampang ditambah framework baru. Default: CLI kalau ga kedeteksi server.
package adopt

import (
	"os"
	"path/filepath"
	"strings"
)

// Suggestion — saran kontrak buat GUI pre-fill. Owner tetap bisa ubah.
type Suggestion struct {
	Contract string   `json:"contract"`  // "cli" | "http"
	StartCmd []string `json:"start_cmd"` // saran perintah start server (HTTP)
	Port     int      `json:"port"`
	Reason   string   `json:"reason"`
}

// fw — 1 framework server: keyword di dep → port default + cara start.
type fw struct {
	key  string // substring dep (lowercase)
	port int
	kind string // streamlit|asgi|flask|gradio|node-next|node-vite|node-express
}

var pyFrameworks = []fw{
	{"streamlit", 8501, "streamlit"},
	{"gradio", 7860, "gradio"},
	{"fastapi", 8000, "asgi"},
	{"uvicorn", 8000, "asgi"},
	{"flask", 5000, "flask"},
}

var nodeFrameworks = []fw{
	{"next", 3000, "node-next"},
	{"vite", 5173, "node-vite"},
	{"express", 3000, "node-express"},
	{"fastify", 3000, "node-express"},
}

func venvBinTool(name string) string {
	if winhost() {
		return filepath.Join(".venv", "Scripts", name+".exe")
	}
	return filepath.Join(".venv", "bin", name)
}

// SuggestContract — periksa dep repo, saranin kontrak. det = hasil Detect (buat fallback run_cmd CLI).
func SuggestContract(repoDir string, det Detection) Suggestion {
	switch det.Runtime {
	case Python:
		if f, ok := matchFramework(readLower(repoDir, "requirements.txt")+readLower(repoDir, "pyproject.toml")+readLower(repoDir, "uv.lock"), pyFrameworks); ok {
			return pySuggest(repoDir, f)
		}
	case Node:
		if f, ok := matchFramework(readLower(repoDir, "package.json"), nodeFrameworks); ok {
			return nodeSuggest(f)
		}
	}
	return Suggestion{Contract: "cli", Reason: "ga kedeteksi framework server — default CLI"}
}

func matchFramework(hay string, list []fw) (fw, bool) {
	for _, f := range list {
		if strings.Contains(hay, f.key) {
			return f, true
		}
	}
	return fw{}, false
}

func pySuggest(repoDir string, f fw) Suggestion {
	s := Suggestion{Contract: "http", Port: f.port, Reason: "framework " + f.kind + " kedeteksi → server HTTP"}
	switch f.kind {
	case "streamlit":
		entry := findFirst(repoDir, "webui/Main.py", "streamlit_app.py", "Home.py", "app.py", "main.py")
		if entry == "" {
			entry = "app.py"
		}
		s.StartCmd = []string{venvBinTool("streamlit"), "run", entry, "--server.port", itoa(f.port), "--server.headless", "true", "--browser.gatherUsageStats", "false"}
	case "gradio":
		s.StartCmd = []string{venvBinTool("python"), findFirst(repoDir, "app.py", "main.py", "demo.py")}
	case "asgi":
		// fastapi/uvicorn — coba 'uvicorn <modul>:app'; fallback python main.py.
		mod := asgiModule(repoDir)
		if mod != "" {
			s.StartCmd = []string{venvBinTool("uvicorn"), mod, "--port", itoa(f.port), "--host", "127.0.0.1"}
		} else {
			s.StartCmd = []string{venvBinTool("python"), findFirst(repoDir, "main.py", "app.py", "server.py")}
		}
	case "flask":
		s.StartCmd = []string{venvBinTool("python"), findFirst(repoDir, "app.py", "main.py", "wsgi.py")}
	}
	return s
}

func nodeSuggest(f fw) Suggestion {
	s := Suggestion{Contract: "http", Port: f.port, Reason: "framework " + f.kind + " kedeteksi → server HTTP"}
	switch f.kind {
	case "node-next", "node-vite":
		s.StartCmd = []string{"npm", "run", "dev"}
	default:
		s.StartCmd = []string{"npm", "start"}
	}
	return s
}

// asgiModule — tebak 'modul:app' dari struktur umum (main.py, app/main.py, app.py).
func asgiModule(repoDir string) string {
	for _, m := range []struct{ file, mod string }{
		{"main.py", "main:app"}, {"app/main.py", "app.main:app"}, {"app.py", "app:app"}, {"server.py", "server:app"},
	} {
		if exists(repoDir, m.file) {
			return m.mod
		}
	}
	return ""
}

func readLower(dir, name string) string {
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	return strings.ToLower(string(b)) + "\n"
}

func findFirst(dir string, cands ...string) string {
	for _, c := range cands {
		if exists(dir, filepath.FromSlash(c)) {
			return c
		}
	}
	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
