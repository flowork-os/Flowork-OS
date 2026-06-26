# FROZEN CORE — mass-freeze 2026-06-26 (LOCK → FREEZE)

> Dev: Aola Sahidin. Owner: "gw maunya freez bukan lock" + "scanner juga freeze". Semua file yang
> dulu cuma ber-header `// === LOCKED FILE ===` (soft-lock, gak ke-enforce) sekarang **FREEZE
> beneran** = 2-lapis: hash di `KERNEL_FREEZE.md` (di-cek `TestKernelFreeze`) + `chattr +i`
> (immutable OS). Tujuan: Flowork TAHAN BANTING dari AI eksternal & internal — boleh evolusi, tapi
> hasil evolusi GAK BISA robohin yang udah jalan.

## ANGKA
- **665 file .go** total frozen (chattr +i + manifest). Naik dari 206 → +459 file (lock→freeze sesi ini).
- Cakupan: agent (agentmgr/agentdb/tools/scanner/connections/slashcmd/...) + router (handlers/
  internal/{brain,store,translator,executors,providers,mitm,router,rtk,quotalive,...}).

## STANDAR FREEZE (dipakai semua file frozen baru)
1. **Strip komentar** (go/parser, token-identik — 0 perubahan kode; build+TestKernelFreeze bukti).
   File ber-directive (`//go:build`/`//go:embed`/`//+build`) → directive DI-PRESERVE (strip dir-safe).
2. **Header white-label 4-baris** (no path /home; relative `os/`):
   ```
   // Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
   // Cara kerja sistem: lihat os/.  ⚠️ FROZEN — jangan edit file ini.
   // Nambah/ubah fitur TANPA buka frozen: pakai SEAM non-frozen + SWITCH
   // (internal/fwswitch/registry.go). Pola lengkap: lock/<doc>
   ```
3. **Hash** `sha256` → `KERNEL_FREEZE.md` (privat, gitignored). **`chattr +i`** (immutable).

## CARA EVOLUSI TANPA BUKA FROZEN (prinsip inti)
Existing stable code = beku. Nambah fitur = **FILE BARU + mekanisme seam** yang udah ada:
- **Registry init-append**: `tools.Register`, `triggers.Register`/`RegisterDeliverer`,
  `RegisterSkillProvider`, `RegisterInstinctSelector`, `RegisterGraphProjection`, scanner `Auditors[x]=fn`,
  **`RegisterExtraRoute` (router endpoint HTTP baru — seam `router/routes_ext.go`)**.
  → tipe/tool/channel/auditor/proyeksi/ENDPOINT baru = file `*_<x>.go` baru, `init()` daftar ke registry.
- **Switch GUI**: tambah entri di `internal/fwswitch/registry.go` (NON-frozen extension point).
- **`*_ext.go` sibling**: hook tambahan tanpa sentuh core.
- **DATA**: persona/skill/instinct/constitution/jadwal/event-type = baris DB/JSON, bukan kode.

## NON-FROZEN (sengaja — seam, gak boleh dibekuin)
`internal/fwswitch/registry.go` (switch), `web/tabs/*.js` + `web/static/*` (GUI),
**`router/routes_ext.go`** (seam route: `RegisterExtraRoute`), `feature_*.go`, semua file
`*_ext.go`, file `*_test.go`.
> KOREKSI 2026-06-26: `routes.go` & `main.go` SEKARANG **FROZEN** (mass-freeze; owner: "semua
> logic router frozen"). Dulu dok ini bilang non-frozen — itu tak sinkron sama realita. Akar
> dicabut: ditambah seam **`routes_ext.go`** (NON-frozen) + hook `registerExtraRoutes(mux)` di
> `registerRoutes` (routes.go). Jadi nambah endpoint TANPA buka frozen → file `handlers_<x>_ext.go`
> baru + `init(){ RegisterExtraRoute(func(m){ m.HandleFunc(...) }) }`. Bukti: `TestRouteSeamWired`.
> Status: **71/71 handler router frozen** (chat_learn, ssrf_guard, pentest, brain_wing ikut).

### Router freeze-sweep 2026-06-26 — FROZEN CORE SELF-SUFFICIENT
Konversi 15 logic stabil (soft-LOCK→FREEZE): `cmd/brain-{buildindex,reembed,search}`,
`localai_autostart`, `vecindex/{ann,builder,vecindex}`, `brain/{fresh_index,skill_provider}`,
`providers/embedding/local`, `router/modellock`, `sidecar/sidecar`, `creds/{login,refresh,save}`.

**+ Self-sufficiency fix (owner: "hapus non-frozen → router error = cacat arsitektur").**
Uji empiris (worktree, hapus file 1-1) buktiin 12 file non-frozen **load-bearing**: frozen core
manggil simbol yg didefinisikan di non-frozen → hapus = build patah. AKAR: file itu sebenarnya
**MEKANISME seam (infra stabil), bukan ekstensi yg mau dihapus** — ekstensi = file SIBLING baru.
Maka 11 di-FREEZE (mekanisme jadi immutable, evolusi tetap via sibling+switch):
`mesh/filter_ext` (RegisterMeshFilter), `vecindex/binary_ext` (method), `brain/graph_extras_ext`
(RegisterGraphProjection), `router/{toolcall_recover_ext,brain_constitution_ext,agentctx_ext,instinctenrich_ext2}`,
`edition_gate`, `routes_ext` (RegisterExtraRoute), `brain/mem_type_registry` (type MemType dipakai
frozen write/ingest), `brain/dream_cycle` (syncCoreEntities, audit-clean; legacy dream tetap di-fence env).
**HASIL terverifikasi:** hapus SEMUA non-frozen non-test non-GUI → `go build ./...` **exit 0**.
Sisa non-frozen sah: `brain/{reclassify,reclassify_rules}` (pair self-contained, frozen core tak
butuh), `llm_idle_sleep_ext` (observer registry), semua sibling `*_ext` baru, GUI, `*_test.go`.

> Evolusi pasca-freeze: edisi/scoped-instinct via SWITCH `FLOWORK_EDITION`/`FLOWORK_INSTINCT_SCOPED`
> (GUI fwswitch, bukan edit kode). Lapis filter/route/proyeksi BARU via file sibling + `Register*`.
> Rebuild dream-cycle (Phase 0) = satu-satunya yg perlu unfreeze `dream_cycle.go` saat dikerjakan.

## KALAU BENERAN HARUS UBAH FILE FROZEN (mis. migrasi schema baru)
Arsitektur cacat = idealnya kasih seam. Kalau terpaksa: ikut CARAFREEZE.MD —
`sudo chattr -i <file>` → edit → re-hash `sha256sum` → update `KERNEL_FREEZE.md` →
`sudo chattr +i` → `TestKernelFreeze` PASS. **Wajib izin DEV.**

## VERIFIKASI 2026-06-26
Strip token-identik (sample re-strip HEAD == current). Agent+router build OK. TestKernelFreeze PASS
(665 hash). Append ke file frozen → "Operation not permitted". Service hidup (:1987/:2402).
Detail per-subsistem: `lock/{threat-radar,trigger-schedule,code-progress}.md` + doc lain.
