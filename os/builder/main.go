// === LOCKED FILE === STABLE — DO NOT MODIFY without owner approval (Aola Sahidin / Mr.Dev).
// Locked 2026-06-13 after the osDir-normalization + version-fallback fix was verified.
// Flowork OS Creator — a small local web GUI to build + flash a Flowork OS USB stick.
// Pick the Flowork OS folder, pick the target flashdisk, pick a profile (Full / Default),
// then Build and Flash. Stdlib only.
//
//	Full     = everything is baked in (the owner's database + settings + tokens). Ready to use.
//	           Keep private. (-> make-distributable.sh dev)
//	Default  = clean: NO database/settings ship; secrets become `change_this_token`. Shareable.
//	           (-> make-distributable.sh public)
//
// Run it from the OS folder (or pass -os). Flashing needs root, so launch with sudo.
package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed web
var webFS embed.FS

var (
	addr      = flag.String("addr", "127.0.0.1:8799", "listen address")
	osDir     = flag.String("os", "", "Flowork OS folder (default: auto-detect from this binary's location)")
	noBrowser = flag.Bool("no-browser", false, "do not auto-open a browser (the launcher opens it as the real user)")
	dataDef   = flag.String("data-default", "", "default Full-profile data source (default: <home>/.flowork; set by the launcher to the real user's home when running elevated)")
)

func main() {
	flag.Parse()
	if *osDir == "" {
		*osDir = detectOSDir()
	}
	*osDir = normalizeOSDir(*osDir)
	mux := http.NewServeMux()
	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/info", apiInfo)
	mux.HandleFunc("/api/devices", apiDevices)
	mux.HandleFunc("/api/build", apiBuild)
	mux.HandleFunc("/api/flash", apiFlash)

	fmt.Printf("Flowork OS Creator → http://%s   (OS folder: %s)\n", *addr, *osDir)
	if os.Geteuid() != 0 {
		fmt.Println("note: flashing needs root — re-run with sudo if Flash fails.")
	}
	if !*noBrowser {
		go openBrowser("http://" + *addr)
	}
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// detectOSDir: the OS folder is the one holding build/make-distributable.sh. The builder lives
// in <os>/builder, so the parent is the default; fall back to the cwd.
func detectOSDir() string {
	if exe, err := os.Executable(); err == nil {
		p := filepath.Dir(filepath.Dir(exe)) // <os>/builder/<bin> -> <os>
		if isOSDir(p) {
			return p
		}
	}
	if wd, err := os.Getwd(); err == nil {
		for _, c := range []string{wd, filepath.Dir(wd)} {
			if isOSDir(c) {
				return c
			}
		}
	}
	return "."
}

func isOSDir(p string) bool {
	_, err := os.Stat(filepath.Join(p, "build", "make-distributable.sh"))
	return err == nil
}

// normalizeOSDir tolerates being pointed at the monorepo ROOT instead of the os/ subdir:
// if p itself isn't the OS folder but p/os is, descend into it. Prevents an empty VERSION
// (and a wrong out/ path) when -os points one level too high — the cause of the bogus
// "flowork-os--dev.usb.img" lookup.
func normalizeOSDir(p string) string {
	if isOSDir(p) {
		return p
	}
	if sub := filepath.Join(p, "os"); isOSDir(sub) {
		return sub
	}
	return p
}

// osVersion reads VERSION from the OS folder, defaulting to 0.0.0 so a missing file never
// yields an empty tag like "flowork-os--dev.usb.img".
func osVersion() string {
	ver := strings.TrimSpace(readFile(filepath.Join(*osDir, "VERSION")))
	if ver == "" {
		ver = "0.0.0"
	}
	return ver
}

// ── info ────────────────────────────────────────────────────────────────────
func apiInfo(w http.ResponseWriter, r *http.Request) {
	ver := osVersion()
	tag := "flowork-os-" + ver
	out := filepath.Join(*osDir, "out")
	dataSrc := *dataDef
	if dataSrc == "" {
		home, _ := os.UserHomeDir()
		dataSrc = filepath.Join(home, ".flowork")
	}
	writeJSON(w, map[string]any{
		"osDir":             *osDir,
		"version":           ver,
		"isRoot":            os.Geteuid() == 0,
		"defaultDataSource": dataSrc,
		"images": map[string]any{
			"full":    statImage(filepath.Join(out, tag+"-dev.usb.img")),
			"default": statImage(filepath.Join(out, tag+"-public.usb.img")),
		},
	})
}

func statImage(p string) map[string]any {
	if fi, err := os.Stat(p); err == nil {
		return map[string]any{"path": p, "present": true, "size": fi.Size()}
	}
	return map[string]any{"path": p, "present": false}
}

// ── devices ───────────────────────────────────────────────────────────────
// List only safe-to-flash targets: removable / USB whole disks. NEVER the system disk.
func apiDevices(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("lsblk", "-J", "-b", "-o", "NAME,SIZE,TYPE,TRAN,RM,MODEL,MOUNTPOINT,PATH")
	raw, err := cmd.Output()
	if err != nil {
		writeJSON(w, map[string]any{"devices": []any{}, "error": err.Error()})
		return
	}
	var parsed struct {
		Blockdevices []struct {
			Name, Type, Tran, Model, Mountpoint, Path string
			Size                                      int64
			Rm                                        bool
			Children                                  []struct{ Mountpoint string }
		} `json:"blockdevices"`
	}
	_ = json.Unmarshal(raw, &parsed)
	devs := []map[string]any{}
	for _, b := range parsed.Blockdevices {
		if b.Type != "disk" {
			continue
		}
		removable := b.Rm || b.Tran == "usb"
		if !removable {
			continue // SAFETY: only removable/USB targets are offered
		}
		path := b.Path
		if path == "" {
			path = "/dev/" + b.Name
		}
		devs = append(devs, map[string]any{
			"path": path, "size": b.Size, "model": strings.TrimSpace(b.Model),
			"tran": b.Tran, "removable": removable,
		})
	}
	writeJSON(w, map[string]any{"devices": devs})
}

// ── build ─────────────────────────────────────────────────────────────────
func apiBuild(w http.ResponseWriter, r *http.Request) {
	var req struct{ Mode, DataSource string }
	_ = json.NewDecoder(r.Body).Decode(&req)
	mode := normMode(req.Mode)
	if mode == "" {
		http.Error(w, "mode must be full|default", 400)
		return
	}
	// make-distributable.sh speaks public|dev; the GUI speaks default|full.
	scriptMode := map[string]string{"full": "dev", "default": "public"}[mode]
	cmd := exec.Command("bash", filepath.Join(*osDir, "build", "make-distributable.sh"), scriptMode)
	cmd.Env = os.Environ()
	if mode == "full" && strings.TrimSpace(req.DataSource) != "" {
		// FULL profile bakes the owner's whole live state from this folder.
		cmd.Env = append(cmd.Env, "FLOWORK_STATE_SEED="+req.DataSource)
	}
	streamCmd(w, cmd)
}

// ── flash ─────────────────────────────────────────────────────────────────
func apiFlash(w http.ResponseWriter, r *http.Request) {
	var req struct{ Mode, Device string }
	_ = json.NewDecoder(r.Body).Decode(&req)
	mode := normMode(req.Mode)
	if mode == "" || req.Device == "" {
		http.Error(w, "need mode + device", 400)
		return
	}
	// SAFETY: re-check the device is a removable/USB whole disk right before writing.
	if !isFlashSafe(req.Device) {
		http.Error(w, "refusing: "+req.Device+" is not a removable/USB disk", 400)
		return
	}
	ver := osVersion()
	suffix := map[string]string{"full": "dev", "default": "public"}[mode]
	img := filepath.Join(*osDir, "out", "flowork-os-"+ver+"-"+suffix+".usb.img")
	if _, err := os.Stat(img); err != nil {
		http.Error(w, "image not built yet for this profile — run Build first ("+img+")", 400)
		return
	}
	args := []string{"if=" + img, "of=" + req.Device, "bs=4M", "status=progress", "conv=fsync"}
	cmd := exec.Command("dd", args...)
	if os.Geteuid() != 0 {
		cmd = exec.Command("sudo", append([]string{"dd"}, args...)...)
	}
	streamCmd(w, cmd)
}

func isFlashSafe(dev string) bool {
	out, err := exec.Command("lsblk", "-dno", "RM,TRAN,TYPE", dev).Output()
	if err != nil {
		return false
	}
	f := strings.Fields(string(out))
	// expect: RM TRAN TYPE
	rm, tran, typ := "", "", ""
	if len(f) == 3 {
		rm, tran, typ = f[0], f[1], f[2]
	} else if len(f) == 2 { // TRAN may be blank
		rm, typ = f[0], f[1]
	}
	return typ == "disk" && (rm == "1" || tran == "usb")
}

// ── helpers ───────────────────────────────────────────────────────────────
func normMode(m string) string {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "full", "dev":
		return "full"
	case "default", "public":
		return "default"
	}
	return ""
}

// streamCmd runs cmd and streams combined stdout/stderr to the client as text/event-stream,
// one SSE "data:" line per output line, then a final "data: __DONE__ <rc>".
func streamCmd(w http.ResponseWriter, cmd *exec.Cmd) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no streaming", 500)
		return
	}
	cmd.Stderr = cmd.Stdout // fold; we set Stdout via pipe below
	pr, pw, _ := os.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "data: ERROR: %s\n\n", err)
		fl.Flush()
		return
	}
	pw.Close() // child holds the write end
	sc := bufio.NewScanner(pr)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		fmt.Fprintf(w, "data: %s\n\n", sc.Text())
		fl.Flush()
	}
	rc := 0
	if err := cmd.Wait(); err != nil {
		rc = 1
	}
	fmt.Fprintf(w, "data: __DONE__ %d\n\n", rc)
	fl.Flush()
}

func readFile(p string) string { b, _ := os.ReadFile(p); return string(b) }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) {
	for _, c := range [][]string{{"xdg-open", url}, {"open", url}} {
		if _, err := exec.LookPath(c[0]); err == nil {
			_ = exec.Command(c[0], c[1:]...).Start()
			return
		}
	}
}
