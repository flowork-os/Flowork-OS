// Owner: Mr.Dev · github.com/flowork-os/Flowork-OS · floworkos.com
// ⚠️ FROZEN brain-core — jangan edit tanpa unfreeze owner. Arsitektur & alasan: lihat lock/brain.md

package main

import (
	"context"
	"os"
	"path/filepath"

	"flowork-gui/internal/agentdb"
	"flowork-gui/internal/agentmgr"
	"flowork-gui/internal/kernel/loader"
	"flowork-gui/internal/kernelhost"
)

const digesterID = "dream-digester"

const digesterPersona = "Kamu CGM EXTRACTOR Flowork. Pesan user berisi instruksi + skema LENGKAP " +
	"buat ekstrak entitas/relasi (node & edge) dari teks. IKUTI instruksi & skema itu PERSIS. " +
	"Balas HANYA output yang diminta (JSON sesuai skema) — tanpa markdown, tanpa code-fence, " +
	"tanpa prosa/penjelasan/sapaan. Output only."

func seedDreamDigester() {
	agentsDir := loader.AgentsDir()
	dir := filepath.Join(agentsDir, digesterID+".fwagent")
	if _, e := os.Stat(dir); e != nil {
		tplWasm, err := os.ReadFile(filepath.Join("templates", "agent-template", "agent.wasm"))
		if err != nil || len(tplWasm) == 0 {
			return
		}
		if mk := os.MkdirAll(filepath.Join(dir, "workspace"), 0o755); mk != nil {
			return
		}
		_ = os.WriteFile(filepath.Join(dir, "manifest.json"), evoMemberManifest(digesterID, "Dream Digester"), 0o644)
		_ = os.WriteFile(filepath.Join(dir, "agent.wasm"), tplWasm, 0o644)
		_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("agent.wasm\nworkspace/*.db\nworkspace/*.db-*\n"), 0o644)
	}

	if st, e := agentdb.Open(agentdb.Resolve(digesterID, dir)); e == nil {
		_ = st.SetPrompt(digesterPersona)

		if st.GetRouterModel() == "" {
			_ = st.SetRouterModel("claude-haiku-4-5")
		}
		_ = st.Close()
	}
	agentmgr.ProvisionAgentDNA(digesterID)
}

func dreamDigestLLM(host *kernelhost.Host) func(context.Context, string) (string, error) {
	return func(ctx context.Context, prompt string) (string, error) {
		raw, err := host.InvokeAgentMessage(ctx, digesterID, prompt, "cgm-digest")
		if err != nil {
			return "", err
		}
		return extractReply(raw), nil
	}
}
