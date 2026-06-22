// tool_install.go — install/uninstall/list TOOL-PACK plug-and-play.
//
//	POST /api/tools/install   (multipart "file" = .fwpack, kind:tool) → extract
//	     wasm tool-agent (hot-load) + register WasmTool + marker. Loopback-only.
//	POST /api/tools/uninstall?tool=<name>  → Unregister + hapus dir agent.
//	GET  /api/tools/installed              → daftar tool plugin (vs builtin).
//
// .fwpack tool layout (zip): plugin.json {kind:"tool", tool:{...spec, agent_id}} +
// agents/<agent_id>/{agent.wasm, manifest.json}. REUSE pola extract task install.

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"flowork-gui/internal/kernel/loader"
	"flowork-gui/internal/kernelhost"
	"flowork-gui/internal/tools"
)

type toolPackManifest struct {
	ID   string   `json:"id"`
	Kind string   `json:"kind"`
	Tool toolSpec `json:"tool"`
}

// installToolPack — CORE install tool-pack. Balik (body, status). status 0 = ok.
func installToolPack(host *kernelhost.Host, raw []byte) (map[string]any, int) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return map[string]any{"error": "not a valid zip: " + err.Error()}, http.StatusBadRequest
	}
	// 1) plugin.json → tool spec
	var manRaw []byte
	for _, f := range zr.File {
		base := strings.TrimPrefix(f.Name, "./")
		if base == "plugin.json" || strings.HasSuffix(base, "/plugin.json") {
			rc, e := f.Open()
			if e == nil {
				manRaw, _ = io.ReadAll(io.LimitReader(rc, 1<<20))
				rc.Close()
			}
			break
		}
	}
	if manRaw == nil {
		return map[string]any{"error": "plugin.json missing from pack"}, http.StatusBadRequest
	}
	var man toolPackManifest
	if err := json.Unmarshal(manRaw, &man); err != nil {
		return map[string]any{"error": "plugin.json parse: " + err.Error()}, http.StatusBadRequest
	}
	if man.Kind != "tool" {
		return map[string]any{"error": "kind bukan 'tool' (ini bukan tool-pack)"}, http.StatusBadRequest
	}
	spec := man.Tool
	if !toolNameRe.MatchString(spec.Name) {
		return map[string]any{"error": "tool.name invalid (^[a-z][a-z0-9_]{1,39}$)"}, http.StatusBadRequest
	}
	if !pluginIDRe.MatchString(spec.AgentID) {
		return map[string]any{"error": "tool.agent_id invalid"}, http.StatusBadRequest
	}
	if tools.IsBuiltinName(spec.Name) {
		return map[string]any{"error": "tool.name bentrok builtin: " + spec.Name}, http.StatusConflict
	}
	// SECURITY: a tool pack takes the kind-dispatch path, which skips the agent
	// caps-consent gate. A tool has no business with exec:/secret:/fs:shared/
	// rpc:agent-invoke — refuse them here (the channel path does the same).
	if danger := scanPackCaps(zr); len(danger) > 0 {
		return map[string]any{
			"error":          "tool pack requests dangerous capabilities — refused",
			"dangerous_caps": danger,
		}, http.StatusForbidden
	}

	// 2) extract wasm → staging → atomic rename (hot-load). [shared: pack_extract.go]
	markerRaw, _ := json.MarshalIndent(spec, "", "  ")
	if eb, st := extractWasmAgentPack(zr, spec.AgentID, ".toolpack-staging-", "tool.json", markerRaw); st != 0 {
		return eb, st
	}

	// 3) smoke: tunggu hot-load + invoke sekali (compute test).
	smoke := smokeTestSynth(host, spec.AgentID)

	// 4) register (persist=false — marker udah ke-tulis pas staging).
	if err := registerWasmTool(host, spec, false); err != nil {
		return map[string]any{"error": "register tool: " + err.Error()}, http.StatusInternalServerError
	}
	return map[string]any{
		"ok": true, "tool": spec.Name, "agent_id": spec.AgentID,
		"smoke": smoke, "params": len(spec.Params),
		"next": "tool LIVE — agent bisa pake via tool_search / tools/run.",
	}, 0
}

func toolInstallHandler(host *kernelhost.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tfWriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "parse form: " + err.Error()})
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "missing file field"})
			return
		}
		defer file.Close()
		raw, err := io.ReadAll(io.LimitReader(file, 128<<20))
		if err != nil {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "read: " + err.Error()})
			return
		}
		body, status := installToolPack(host, raw)
		tfWriteJSON(w, status, body)
	}
}

// toolPackNames — map nama→agent_id tool yg di-install sebagai .fwpack TOOL-PACK
// (punya marker tool.json di dir agent-nya). Pembeda dari tool App (prefix app_)
// & MCP (prefix mcp_) yang JUGA lewat RegisterDynamic tapi TANPA marker tool.json
// — mereka dikelola di tab Apps / Connections, bukan tab Tools. Sumber kebenaran
// "tool plugin mana yang upload-an" buat /api/tools/installed + uninstall.
func toolPackNames() map[string]string {
	out := map[string]string{}
	root := loader.AgentsDir()
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, rerr := os.ReadFile(filepath.Join(root, e.Name(), "tool.json"))
		if rerr != nil {
			continue
		}
		var spec toolSpec
		if json.Unmarshal(raw, &spec) == nil && spec.Name != "" {
			out[spec.Name] = spec.AgentID
		}
	}
	return out
}

// findToolAgent — agent_id pemilik tool-pack bernama `name` ("" kalau bukan tool-pack).
func findToolAgent(name string) string { return toolPackNames()[name] }

func toolUninstallHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tfWriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST only"})
			return
		}
		name := strings.TrimSpace(r.URL.Query().Get("tool"))
		if name == "" {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": "tool required"})
			return
		}
		if err := tools.Unregister(name); err != nil {
			tfWriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		agentID := findToolAgent(name)
		if agentID != "" {
			_ = os.RemoveAll(filepath.Join(loader.AgentsDir(), agentID+".fwagent"))
		}
		tfWriteJSON(w, 0, map[string]any{"ok": true, "uninstalled": name, "agent_removed": agentID})
	}
}

// toolInstalledHandler — GET daftar TOOL-PACK (.fwpack upload-an) buat tab Tools.
// HANYA tool yg punya marker tool.json — tool App (app_) & MCP (mcp_) disaring
// (mereka dynamic juga tapi dikelola di tab Apps/Connections, bukan di sini).
func toolInstalledHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packs := toolPackNames()
		out := []map[string]any{}
		for _, n := range tools.DynamicNames() {
			agentID, isPack := packs[n]
			if !isPack {
				continue // bukan tool-pack upload-an (app_/mcp_/dll) — skip
			}
			item := map[string]any{"name": n, "agent_id": agentID}
			if t, ok := tools.Lookup(n); ok {
				s := t.Schema()
				item["description"] = s.Description
				item["capability"] = t.Capability()
				item["params"] = len(s.Params)
			}
			out = append(out, item)
		}
		tfWriteJSON(w, 0, map[string]any{"installed": out, "count": len(out)})
	}
}
