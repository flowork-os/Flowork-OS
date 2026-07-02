# Gerbang approval interaktif — F-B v1 (2026-07-02)

## Arsitektur
Mesin approval udah ada dari dulu di frozen core (JANGAN bikin ulang):
- `internal/tools/sandbox_v3.go` (FROZEN): chokepoint `requiresApproval` → enqueue
  `approval_queue` (per tool+args_hash, approved berlaku 1 jam) + sentinel `ErrPendingApprove`.
  Urutan: sensitive-args → ReadOnlyClassifier (exempt) → sensitiveTools → **ExtraGatePolicy**.
- Endpoint (frozen, udah dari dulu): GET `/api/agents/protector/approval/queue?id=<agent>` ·
  POST `.../approve_pending?id=&queue_id=` · POST `.../reject_pending` — **butuh login GUI**
  (bukan loopback-allowlist; cuma owner yang bisa mutusin).

F-B v1 = NGISI colokan yang udah disiapin (semua NON-frozen, deletable):
- `internal/tools/builtins/permission_policy.go` — `tools.ExtraGatePolicy = approvalGatePolicy`
  (mode-aware) + interceptor `approval-mode-agent` (per-agent `approval_mode`='plan' di config
  agent → agent read-only; per-agent CUMA bisa memperketat, ga bisa relaksasi).
- `internal/tools/builtins/cmdsem.go` — git SADAR-SUBCOMMAND: dulu semua `git` dianggap
  read-only → `git push`/`commit` lolos exempt; sekarang cuma subcommand baca (status/log/
  diff/show/rev-parse/…; branch/remote/tag cuma kalau semua arg flag). Test: cmdsem_test.go.
- `internal/fwswitch/registry.go` — switch GUI `FLOWORK_APPROVAL_MODE` (string, default
  `default`, kategori Security / Approval). Live tanpa restart (policy baca os.Getenv host-side).

## Mode (global, switch GUI) — DEFAULT = `bypass` (owner 2026-07-02: "Flowork sebebas
## mungkin, mandiri termasuk keamanan" → evolusi ga nunggu manusia; gerbang interaktif OPT-IN)
| Mode | Perilaku |
|---|---|
| `bypass` | (DEFAULT) Tanpa gerbang interaktif. Keamanan MANDIRI tetap aktif: protector baseline immutable, cmdsem structural block, caps, sandbox workspace, ARM power. |
| `default` | Aksi DESTRUKTIF (shell mutasi, termasuk git push/commit) → antrian approval. Read-only + file-tool workspace auto-allow. |
| `acceptEdits` | Alias `default` (edit file workspace emang udah auto-allow di Flowork). |
| `plan` | SEMUA non-read-only → antrian approval. |

`system_power` TIDAK diurus di sini — udah punya gerbang sendiri (cap `exec:power` + ARM +
`FLOWORK_POWER_REQUIRE_APPROVAL`).

## Verifikasi 2026-07-02
E2E bahasa manusia: minta mr-flow `mkdir` → ke-hold `pending owner approval queue_id=1`
(pending pertama itu masih nunggu keputusan owner di GUI), reply edukatif, agent ga stuck.
`git status` (read-only) tetap jalan tanpa approval. Unit: `TestClassifyCommand_GitSubcommand`
+ `TestApprovalGatePolicy_Modes` PASS. Build/vet/test/TestKernelFreeze hijau.

## Tambahan 2026-07-02 (sesi sore-2, semua FROZEN perintah owner)
- **Notif Telegram pending** KELAR: `agent/feature_approval_notify.go` — poller 60s
  (pola deadair/wakeup, skip agent tanpa tabel, dedup per proses, batch 1 pesan).
  Switch `FLOWORK_APPROVAL_NOTIFY` (default ON).
- **Seam allowlist auth** (Rule #7): `internal/floworkauth/allow_seam.go` —
  `RegisterLoopbackPublic(path, methods...)` dari file non-frozen; invarian DIPAKSA
  terpusat (loopback + anti cross-site + method match) → ext salah pun ga bisa buka
  celah remote. `handlers.go` dibuka-sadar 1x (fallback `loopbackAllowExt`), re-lock.
- **`/api/health`** KELAR (F-F): `feature_health.go` + `allow_health_ext.go` —
  loopback GET tanpa sesi; payload doctor-ringan: status/version/agents_loaded/router_ok.
- **Seed settings user service** (F-F): `os/lockbox/setup.sh` stage 4 — copy
  `flowork_settings.json` ke home user `flowork` KALAU BELUM ADA (ga nimpa).

## Panel GUI (KELAR 2026-07-02)
- Endpoint AGREGAT `/api/agents/protector/approval/pending-all` (feature_approval_panel.go,
  NON-frozen — iterate host.AgentIDs pola approval_notify). GUI: card "Pending Approvals"
  di tab Autonomy (otonomi.js) — tabel agent/tool/reason/requested + tombol Approve/Reject
  (POST approve_pending/reject_pending per-agent frozen). i18n en+id.

## Sisa F-B (buat penerus)
- Mode per-agent penuh (relaksasi per-agent SENGAJA ga dibikin — keputusan keamanan).
