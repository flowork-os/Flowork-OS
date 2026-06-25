# KEPUTUSAN ROADMAP — sesi autonomous 2026-06-26 (owner istirahat)

> Owner pasrah penuh + istirahat. Gw (AI) ambil keputusan "paling stabil & permanen" (cabut-akar,
> bukan tambal). Yang AMAN + teruji → gw KERJAIN + freeze. Yang beresiko-tinggi unattended → gw
> TUNDA dengan PLAN konkret di sini, biar owner eksekusi pas ada (jangan rusakin core 3am).

## ✅ SELESAI sesi ini (detail di lock masing2)
- **#8 KV-cache** → `lock/kv-cache.md`: `-np` parallel slots + `--slot-save-path` wired (opt-in switch),
  prompt-order udah favorable. runtime.go non-frozen.
- **#10 brain-as-service** → `lock/scoped-instinct.md`: constitution external-scope (skip doktrin
  internal buat AI luar). Hook + filter + switch + 3 test. brain_constitution.go re-frozen.
- **#7 sleep-driven** → `lock/sleep-driven.md`: stack komplit (ScheduleWakeup+poller 1mnt+scheduler+
  agent_run+auto-continue+AOLA-013). Akar (insting "kapan") keisi AOLA-013. Zero kode baru.
- **#9 intent-gated tools** (sebelumnya) + **AOLA-013 loop doctrine** + **fwswitch GUI settings**.

## ⏸️ DITUNDA (keputusan + alasan + plan)

### #3 Error-edukasi Layer-2 (DIGEST filter) — TUNDA (owner-present)
**Kenapa tunda:** butuh 3 edit di FROZEN brain-core sekaligus: (a) `cognitive_dream.go`
`DigestPendingInteractions` query → skip `outcome=failed`; (b) `mr-flow/main.go` tag failure-path
logInteraction `outcome=failed`; (c) `agentkit.go` idem buat worker. Tagging-path SUSAH di-unit-test
(perlu trigger failure live, lambat+flaky). Blast-radius = MEMORI PERMANEN (salah filter → graph
berhenti belajar / korup). Unattended 3am = ga sesuai "paling stabil". Layer-1 anti-anchor UDAH
nyaring read-path (gejala langsung ke-mitigasi).
**Plan siap eksekusi (owner ada):** interactions UDAH punya kolom `metadata TEXT` (JSON) →
1. tag di failure-path: `logInteraction(..., map{"outcome":"failed"})` pas ghost-max / flail-escalate /
   auto-continue-giveup (mr-flow/main.go + agentkit.go).
2. filter digest: tambah `AND json_extract(i.metadata,'$.outcome') IS NOT 'failed'` di query
   `DigestPendingInteractions` (cognitive_dream.go).
3. unit-test: insert interaction metadata outcome=failed → DigestPendingInteractions SKIP.
4. unfreeze→edit→re-hash→KERNEL_FREEZE→re-freeze ke-3 file. Verify Rule-9 bahasa-manusia + digest test.
Detail tambahan: `lock/ERROR_EDUKASI.md §5`.

### #5 #2B re-digest insting Mr.Flow — TUNDA (Rule 8, butuh kurasi owner)
**Kenapa:** nge-digest 750 pesan chat jadi INSTING = NAMBAH ISI BRAIN (Rule 8: ubah brain/doktrin/
persona butuh ACC owner). KRITIS: constraint keamanan (kill-switch / AI-pindah-server) JANGAN jadi
insting umum — butuh JUDGMENT owner mana pola yg jadi insting, mana yg literal/skip. Auto-digest
unattended bisa polusi brain dgn insting jelek = susah dibalik. Mekanisme (`/api/brain/ingest/submit`)
udah ADA → owner tinggal kurasi manual via GUI (Brain → Instincts) atau sesi bareng.

### #6 Rebuild ~24 agent template-lama → agentkit — TUNDA (no-source)
**Kenapa:** ~24 agent itu DEPLOY-ONLY di `~/.flowork` (ga ada source di repo). Ga bisa rebuild tanpa
source. Mereka fails-open AMAN sementara (scoped-instinct/defer fallback). Konversi butuh source
(owner punya di tempat lain / di-regenerate). Bukan task yg bisa gw kerjain dari repo.

### #9 #5 binary-vector recall — TUNDA (premature, scale-only)
**Kenapa:** roadmap sendiri bilang "worth pas korpus GEDE". Korpus sekarang kecil (~ribuan drawer) →
int8 vecindex udah cepet (recall@1=1.0). Binary (XNOR+popcount) = optimasi SKALA; bikin sekarang =
premature-optimization + risiko recall@k turun, ga ada gain nyata. "Cabut akar" ≠ nambah kompleksitas
sebelum perlu. Revisit pas korpus → jutaan. Lokasi: `internal/brain/vecindex/ann.go`.

### RESIDUAL Telegram getUpdates 409 — ENVIRONMENTAL (aksi owner)
**Kenapa:** 409 = poller GANDA (bot token dipakai instance/mesin LAIN barengan). Akar = OPERASIONAL
(di luar kode): matiin instance/token ganda. Kode bisa kasih backoff/single-flight tapi ga nyabut akar
(token tetep kepake 2 tempat). Owner: pastiin cuma 1 instance jalan / rotate token. `lock/mrflow.md §6b`.

## CATATAN
- Push sesi ini ke **flowork-base (private) DOANG** (perintah owner) → public = titik rollback kalau halu.
- Disiplin baru (QC GUI, test-before-freeze, switch-before-freeze) → `lock/DISIPLIN.md`.
