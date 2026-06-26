# DREAMGRAPH (router Knowledge Graph) — auto-populate + auto-update

> Router cognitive graph yang tampil di dashboard `:2402` tab Brain → "FLowork Knowledge Graph
> (DreamGraph)". BEDA dari CGM agent (`lock/CognitiveGraph.md`, per-agent `state.db`). Owner: Mr.Dev.

## APA INI
Graph relasi entitas-inti Flowork (node `cognitive_nodes` + edge `cognitive_edges` di
`router/brain/flowork-brain.sqlite`), cermin dari sumber: **constitution + persona + skill + agent**
→ kesadaran-diri sistem (aturan, identitas, kemampuan).

## MASALAH (sebelum, akar)
Graph KOSONG (0 node/0 edge) → canvas blank. Akar: tabel `cognitive_nodes/edges` cuma keisi via
`SyncGraphToRAG` (kepanggil pas CRUD node/edge manual doang) atau `RunDreamCycle` (disabled: mock +
data-loss). GAK ADA trigger boot/terjadwal → selamanya kosong.

## FIX (cabut-akar)
File baru NON-frozen `router/dreamgraph_autosync.go`:
- **Boot-populate**: `startDreamGraphAutoSync(ctx)` dipanggil di `router/main.go` (deket ticker
  fresh-index) → sekali sync saat boot → graph langsung keisi.
- **Auto-update**: loop poll 30 dtk, sync tiap interval (GUI switch, default 5 menit) → graph
  selalu cermin sumber tanpa aksi manual. Loop RE-BACA switch tiap siklus → ganti di GUI langsung
  kepakai (gak perlu restart).
- **Manual**: `POST /api/brain/graph/sync` (tombol "Sync Now") → balas `{ok,nodes,edges}`.
- Mekanisme = `brain.SyncGraphToRAG` (idempotent, MIRROR-only: constitution/persona/skill/agent →
  graph; TIDAK hapus memory). Serialized via mutex (anti SQLite-lock bentrok ticker vs manual).

## SWITCH GUI (kebenaran di GUI, bukan hardcode)
Di `agent/internal/fwswitch/registry.go` (kategori "Brain / Graph"), muncul di tab "🎛️ Switch Fitur":
- `FLOWORK_DREAMGRAPH_AUTOSYNC` (bool, default `true`) — ON/OFF auto-sync.
- `FLOWORK_DREAMGRAPH_SYNC_MIN` (int, default `5`) — interval menit.
Lintas-proses via `~/.flowork/flowork_settings.json` → router baca live (≤3 dtk).

## VERIFIKASI (2026-06-26)
- Boot log: `dreamgraph: boot sync OK`.
- `cognitive_graph_stats` 0/0 → **16 node / 15 edge** (13 constitution + flowork + agent_mr.flow +
  1 persona; semua edge `governs`/`member_of`/`persona_of` → `flowork`). Angka kecil = CERMIN sumber
  saat ini (snapshot lama 44/57 itu STALE, bukan target).
- `POST /api/brain/graph/sync` → `{"edges":15,"nodes":16,"ok":true}`.
- Build+vet router PASS. Idempotent (sync ulang stabil).

## KEPUTUSAN FREEZE
File-file ini = **orchestration/extension seam + switch-protected**, SENGAJA non-frozen:
- `main.go`/`routes.go` → harus tetap terbuka buat nambah route/boot-hook (freeze = mateng evolusi).
- `registry.go` → extension-point switch (by-design non-frozen).
- `dreamgraph_autosync.go` → switch (FLOWORK_DREAMGRAPH_AUTOSYNC) = pengaman evolusi; tuning lewat
  GUI, bukan edit kode. Logika inti graph (`SyncGraphToRAG`) ada di `dream_cycle.go` (soft-lock).
Alasan: prinsip "freeze CORE, biarin seam" → switch sudah cukup lindungi dari AI lain ngerusak.

## SISA (belum, sengaja ditunda — lihat opus_roadmap.md)
- Projeksi KNOWLEDGE corpus (860k drawer) sbg hub-per-wing → graph "pengetahuan baru" lebih literal.
- Projeksi INSTINCTS (365 drawer) sbg node.
- MEMORY via dream-cycle rebuild (real extractor + safe digest) = HIGH-RISK data-loss → butuh sesi
  owner-attended (alasan: blast-radius memori permanen).
