# CONNECTIONS — channel / MCP / native connector (GUI 🔌 Connections)

> Dev: Aola Sahidin. 2026-06-26. Pintu masuk/keluar Flowork: channel (Telegram/Discord/Slack/WA
> via WASM connector), MCP server (tool eksternal), native built-in (CLI/MCP). Cara kerja: `os/`.

## JALUR
```
connections.js (GUI)
  → /api/connections{,/toggle,/config,/uninstall} + /api/plugins/install (channel WASM)
  → /api/mcp{,/install,/enable,/disable} (MCP server)
    (connections/handlers.go + mcphub/handlers.go)
  → connections.go (lifecycle) · central.go (sentralisasi secret → floworkdb) · native.go (CLI/MCP builtin)
  → mcphub.go (MCP runtime: spawn server, register tool dinamis)
  → loket: manifest.go (validasi kind:channel + cap-gate) → dispatcher.go (routing pesan) →
    providers_registry.go (registry provider channel)
```

## FROZEN — pipa/jalur (chattr +i + KERNEL_FREEZE + TestKernelFreeze)
`connections/{connections,handlers,central}.go`, `mcphub/{mcphub,handlers}.go`, loket core 12 file
(`contract/dispatcher/manifest/ratelimit/service/store/providers*`). Strip + header white-label.

## SEAM — nambah koneksi TANPA buka frozen (plug-and-play, DATA/plugin-driven)
- **Channel baru (WASM)** = pack `.fwpack` (manifest kind:channel) → `POST /api/plugins/install` →
  loket `manifest.Validate` (cap-gate: tolak exec/secret/fs:shared/rpc:agent-invoke) → hot-load.
  Bikin dari `templates/connector-template`. **Zero edit frozen.**
- **MCP server baru** = paste `mcpServers` JSON di GUI → `POST /api/mcp/install` → `mcphub` spawn +
  `tools.RegisterDynamic` runtime. **Zero edit frozen, zero restart.**
- **Provider-type channel baru** = `loket/providers_registry.go` pattern (registry).
- **Native connector baru** (CLI/MCP builtin) = `connections/native.go` (**NON-frozen** by design —
  ini seam in-file buat native; edit/tambah native connector di sini).

## SWITCH — kenapa GAK ada FLOWORK_ switch global (jujur)
Connections = **DATA/plugin-driven**, BUKAN behavior-flag. Enable/disable per-connector & per-MCP =
DATA (marker-file / toggle GUI), bukan switch global. Evolusi dijamin lewat plugin-install + MCP-
register + native.go + loket registry (di atas). Switch global (mis. matiin semua MCP) = gak perlu +
bakal gate kode frozen (mundur). Kalau owner mau toggle spesifik nanti → taro di `native.go`
(non-frozen) atau registry, bukan di file frozen.

## NON-FROZEN (seam, sengaja)
`connections/native.go` (native connector in-file seam), `web/tabs/connections.js` (GUI),
`feature_*.go` (route reg), `templates/connector-template` (cetakan channel baru).

## VERIFIKASI 2026-06-26
QC live (login :1987): `/api/connections` → 2 connector (cli+mcp, enabled). `/api/mcp` → 0 server
(belum ada yang di-install, by-design). Jalur frozen (TestKernelFreeze PASS). native.go open
(seam). Catatan audit: MCP enabled-flag in-memory (re-enable on boot) — known, non-blocking.
