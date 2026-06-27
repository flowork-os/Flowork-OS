# SETTINGS — Plug-and-Play + Cleanup (2026-06-27)

> Owner: Aola Sahidin (Mr.Dev) · floworkos.com. Prinsip: engine settings KRAMAT (frozen), nambah
> integrasi API = DATA (nol buka frozen). Arsitektur: lock/ARSITEKTUR.md · plug-and-play.md.

## PLUG-AND-PLAY (cara nambah API baru TANPA buka frozen)
- **Engine FROZEN:** `agent/internal/settingsapi/settingsapi.go` (`KeysHandler` + secret store generik).
- **Nambah API apapun (Facebook/email/social/dll) = DATA:** Settings → API Keys → tambah key
  `UPPER_SNAKE_CASE` (mis. `FB_PAGE_TOKEN`, `RESEND_API_KEY`) + value → `SetSecret` ke flowork.db,
  di-inject sbg env var. **NOL kode, NOL unfreeze.** Hint/chip per-service = `KEY_PRESETS` di
  `web/tabs/settings.js` (GUI non-frozen — nambah hint = edit GUI, bukan engine).
- Secret store: `floworkdb.SetSecret/GetSecret` (generik, masked di GUI, lowercase ga ke-env-inject).

## CLEANUP (2026-06-27)
- **YouTube DICABUT akar:** `settingsapi/youtube.go` dihapus · 5 route (`feature_platform.go`) dicabut ·
  YT watcher (`start.sh`) dihapus · `renderYouTube`+poll (`settings.js`) dihapus. Build clean, nol referensi fungsional.
- **Finance/wallet (crypto) ripped fungsional:** route `/api/finance/*` + `/api/agents/finance/*` dicabut ·
  `finance_wallet.go` dihapus · 5 tool finance (`finance_summary/log/budgets/budget_set/ledger_list`) UN-REGISTERED
  di v6/v7/v8/v14_extras. **Dormant-frozen (inert):** struct tool + `agentmgr/finance.go` + `agentdb/finance.go`
  + `legacy_compat` snapshot handler — di-biarin beku (un-registered = ga kepake; gut fisik = follow-up bersih).
  **DI-KEEP:** `marketdata/quote.go` (itu quote saham buat stock-analyst-squad, BUKAN wallet) + pillar "ekonomi"
  evolve (visi self-sustain, bukan crypto-wallet).
- **GUI full ENGLISH:** semua string Indonesia (Auto-Compact/Auto-Push/status/label) → English; app.js?v=17 (cache-bust).
  GUI selalu load locale `en` (app.js). Terjemahan `id/*.json` tetap di kamus (dormant, ga ke-serve).

## VERIFIKASI (2026-06-27)
`go build ./...` OK · `TestKernelFreeze` PASS (8 file finance re-frozen) · finance endpoint mati ·
binary embed app.js?v=17 · `settingsapi.go` frozen. GUI = non-frozen seam (sengaja, biar tumbuh).
