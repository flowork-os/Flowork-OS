// feature_compat.go — FASE-B: reference-GUI compat shim + self-evolution + finance/codemap/
// warga-caps/commits. Self-register (PhaseRoute). Dipindah dari main.go biar main.go agnostic.
package main

import (
	"net/http"

	"flowork-gui/internal/agentmgr"
)

func init() {
	RegisterFeature(Feature{Name: "compat", Phase: PhaseRoute, Apply: func(d *Deps) {
		// Legacy reference GUI compat shim → agent-scoped (default mr-flow).
		d.Mux.HandleFunc("/api/finance/snapshot", agentmgr.FinanceSnapshotCompatHandler)
		// R8 self-finance fase-1: wallet EVM (privkey lokal terenkripsi).
		d.Mux.HandleFunc("/api/finance/wallet", financeWalletHandler())
		d.Mux.HandleFunc("/api/brain/prompt-templates", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				agentmgr.PromptTemplatesUpsertCompatHandler(w, r)
				return
			}
			agentmgr.PromptTemplatesListCompatHandler(w, r)
		})
		d.Mux.HandleFunc("/api/brain/prompt-templates/detail", agentmgr.PromptTemplatesDetailCompatHandler)
		d.Mux.HandleFunc("/api/brain/prompt-templates/update", agentmgr.PromptTemplatesUpsertCompatHandler)
		d.Mux.HandleFunc("/api/brain/prompt-templates/delete", agentmgr.PromptTemplatesDeleteCompatHandler)
		d.Mux.HandleFunc("/api/protector", agentmgr.ProtectorListCompatHandler)
		d.Mux.HandleFunc("/api/protector/add", agentmgr.ProtectorAddCompatHandler)
		d.Mux.HandleFunc("/api/protector/remove", agentmgr.ProtectorRemoveCompatHandler)
		d.Mux.HandleFunc("/api/protector/toggle", agentmgr.ProtectorToggleCompatHandler)
		d.Mux.HandleFunc("/api/protector/test", agentmgr.ProtectorTestCompatHandler)
		d.Mux.HandleFunc("/api/codemap/graph", agentmgr.CodemapGraphCompatHandler)
		d.Mux.HandleFunc("/api/codemap/status", agentmgr.CodemapStatusCompatHandler)
		d.Mux.HandleFunc("/api/codemap/zombies", agentmgr.CodemapZombiesCompatHandler)
		d.Mux.HandleFunc("/api/codemap/reindex", agentmgr.CodemapReindexCompatHandler)
		d.Mux.HandleFunc("/api/codemap/roots", agentmgr.CodemapRootsCompatHandler)
		d.Mux.HandleFunc("/api/codemap/docs", agentmgr.CodemapDocsCompatHandler)
		// R6 self-map semantik (LLM inject).
		d.Mux.HandleFunc("/api/codemap/enrich", agentmgr.CodemapEnrichHandler(codemapSemanticSummarizer(d.Host)))
		d.Mux.HandleFunc("/api/codemap/semantic", agentmgr.CodemapSemanticHandler)
		// R7 self-evolution.
		d.Mux.HandleFunc("/api/evolve/reflect", agentmgr.EvolveReflectHandler(evolveProposer()))
		d.Mux.HandleFunc("/api/evolve/proposals", agentmgr.EvolveProposalsHandler)
		d.Mux.HandleFunc("/api/evolve/config", agentmgr.EvolveConfigHandler(evolveGateDeps()))
		d.Mux.HandleFunc("/api/evolve/eval", evolveEvalHandler())
		d.Mux.HandleFunc("/api/evolve/apply", agentmgr.EvolveApplyHandler(evolveGateDeps(), evolveApplier(d.Host, d.FDB, d.GroupsAPI)))
		d.Mux.HandleFunc("/api/evolve/core-apply", agentmgr.EvolveCoreApplyHandler(evolveGateDeps(), evolveCoreApplier(d.Host)))
		d.Mux.HandleFunc("/api/evolve/council", agentmgr.EvolveCouncilHandler(evolveCouncilJudgeViaGroup(d.Host)))
		d.Mux.HandleFunc("/api/evolve/proposal/delete", agentmgr.EvolveProposalDeleteHandler)
		d.Mux.HandleFunc("/api/evolve/stages", agentmgr.EvolveStagesHandler)
		d.Mux.HandleFunc("/api/evolve/stage-action", agentmgr.EvolveStageActionHandler(evolveGateDeps(), evolveStageCommitter()))
		d.Mux.HandleFunc("/api/evolve/push-config", evolvePushConfigHandler())
		d.Mux.HandleFunc("/api/evolve/rollback-log", evolveRollbackLogHandler())
		d.Mux.HandleFunc("/api/evolve/schedule", evolveScheduleHandler(d.Host, d.FDB, d.GroupsAPI))
		startEvolveScheduler(d.Ctx, d.Host, d.FDB, d.GroupsAPI) // goroutine: cron refleksi
		d.Mux.HandleFunc("/api/agents/protector/approval/queue", agentmgr.ApprovalQueueHandler)
		d.Mux.HandleFunc("/api/agents/protector/approve_pending", agentmgr.ApproveHandler)
		d.Mux.HandleFunc("/api/agents/protector/reject_pending", agentmgr.RejectHandler)
		d.Mux.HandleFunc("/api/agents/tool-audit", agentmgr.ToolAuditHandler)
		d.Mux.HandleFunc("/api/settings/educational-errors", agentmgr.EduErrorsCompatHandler)
		d.Mux.HandleFunc("/api/warga-caps/warga", agentmgr.WargaListCompatHandler)
		d.Mux.HandleFunc("/api/warga-caps/catalog", agentmgr.WargaCapsCatalogCompatHandler)
		d.Mux.HandleFunc("/api/warga-caps/effective", agentmgr.WargaCapsEffectiveCompatHandler)
		d.Mux.HandleFunc("/api/warga-caps/override", agentmgr.WargaCapsOverrideCompatHandler)
		d.Mux.HandleFunc("/api/warga-caps/seed", agentmgr.WargaCapsSeedCompatHandler)
		d.Mux.HandleFunc("/api/commits", agentmgr.CommitsCompatHandler)
	}})
}
