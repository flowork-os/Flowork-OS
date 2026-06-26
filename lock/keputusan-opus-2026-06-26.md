# KEPUTUSAN AUTONOMOUS — Opus 2026-06-26 (owner istirahat, titip rumah)

> Owner pasrah penuh + off. Gw ambil keputusan "paling stabil & permanen" (cabut-akar). Prinsip:
> yang AMAN+teruji → KERJAIN; yang beresiko-tinggi unattended (data-loss / buka banyak frozen) →
> TUNDA dengan ALASAN konkret biar owner eksekusi pas hadir. Backup penuh: `Pictures/flowok_backup/FLOWORKOS4`.

## ✅ SELESAI (dibangun + di-test + live)
1. **DreamGraph auto-populate + auto-update** (`lock/dreamgraph.md`). Graph router 0/0 → 16 node/15
   edge, boot-sync + ticker + switch GUI (`FLOWORK_DREAMGRAPH_AUTOSYNC/_SYNC_MIN`) + endpoint manual.
   Verified live + build/vet PASS. Nyelesain keluhan owner "graph kosong".
2. **Codemap enrich VERIFIED berfungsi** (`lock/codemap-enrich.md`). Audit: enrich sebelumnya 0 row
   = belum pernah dijalanin (bukan rusak). Dijalanin → row terisi via AGENT `codemap-enricher`
   (provenance kebukti, bukan fallback). Full-populate 411 file jalan (background).
3. **M1 codemap incremental by-hash** (`lock/codemap-enrich.md`). Kolom `content_hash` → file BERUBAH
   = re-enrich (akar staleness dicabut). Agent rebuilt+restart, TestKernelFreeze PASS.

## 🔎 AUDIT (temuan, buat owner)
- **CGM agent (#cognitive)**: graph hidup (701 node/496 edge) + GUI render OK. **TEMUAN: orphan
  REGRESI ~204 (29%)** — klaim "orphan 0%" (`CognitiveGraph.md` backfill 2026-06-23) udah basi krn
  graph tumbuh & backfill TIDAK periodik. Plus **49 open-tension** numpuk (nunggu owner decide).
  → fix = backfill periodik (lihat TUNDA #5).
- **Mesh & Policy Console (#mesh-console)**: SEMUA endpoint 200, fungsional di level API. Tapi mesh
  masih **PHASE-1 by-design** (`stack/overview`: "single-owner no real mesh traffic, multi-host
  phase 2") → data peer/packet/karma/tool-manifest = SEED. Policy/pricing/provider/localai jalan.
  Bukan rusak; phase-2 (multi-host) = fase tersendiri.

## ⏸️ DITUNDA (ALASAN — owner eksekusi pas hadir)
1. **Memory → DreamGraph (dream-cycle rebuild)**: butuh ganti mock-extractor → LLM asli + digest-log
   aman. Blast-radius = MEMORI PERMANEN (bug lama = data-loss). Unattended 3am = ga sesuai "paling
   stabil". Plan lengkap: `opus_roadmap.md` Bagian 1 Fase D.
2. **Skill Central (copy→reference)**: nyentuh file LOCKED (`router_skills.go`, `agents_router_skills.js`)
   + core `mr-flow/main.go`. Perubahan kontrak besar; butuh review owner. Plan: roadmap Bagian 2.
3. **Codemap auto-enrich saat kode berubah (M4)**: hook self-evolve = core sensitif. M1 (fondasi
   hash) udah jalan; auto-trigger butuh owner-attended. Plan: roadmap Bagian 6 M2–M4.
4. **DreamGraph: knowledge-hub (860k) + instincts (365) projection**: additive tapi butuh fungsi baru
   di package `brain` (paket router) + uji skala. Aman tapi belum kekejar sesi ini. Roadmap Bagian 1 B/C.
5. **CGM orphan periodic backfill**: file CGM = FROZEN brain-core. Backfill 2026-06-23 manual one-time;
   bikin periodik = nyentuh frozen / job baru. Tunda biar ga buru-buru buka frozen unattended.
6. **Dropdown mem_type unify (Bagian 3)**: polish prioritas-rendah (owner sendiri bilang "nanti
   rapiin") + edit embedded dashboard HTML 408KB = risiko > nilai buat sesi unattended. Plan ready
   di roadmap Bagian 3.

## ❄️ KEPUTUSAN FREEZE (alasan)
Prinsip owner: "switch sebelum freeze, freeze CORE biarin seam". Perubahan gw SEMUA = **seam +
switch-protected** (evolution-safe by-design):
- `dreamgraph_autosync.go` → **DI-FREEZE** (logika stabil, lengkap, switch FLOWORK_DREAMGRAPH_* =
  escape; gak ada kerja lanjutan). Kunci di `KERNEL_FREEZE.md`.
- `main.go`/`routes.go` (router) → **TIDAK** (harus terbuka buat nambah route/boot-hook).
- `registry.go` (fwswitch) → **TIDAK** (extension-point switch, by-design non-frozen).
- codemap files (`codemap_semantic.go` store+handler) → **TIDAK** (masih ada M2–M4 terencana; freeze
  sekarang = mateng evolusi belum selesai). Aman: extension additive, gak sentuh frozen.
TestKernelFreeze tetap PASS (gak ada frozen lama yg ke-sentuh).

## 📦 PUSH
Ke **base (private) DOANG** (perintah owner) → public = titik rollback kalau halu. Audit secret +
path `/home/` sebelum push.
