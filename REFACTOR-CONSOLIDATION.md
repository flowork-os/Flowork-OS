# ROADMAP — KONSOLIDASI FLOWORK ("CABUT, BUKAN TAMBAL")

> **Filosofi (owner):** sakit gigi dicabut, bukan ditambal. Tiap refactor dikerjakan
> TUNTAS + tested, satu cabutan per langkah — bukan 5 tambalan setengah jadi.
> **Tanggal mulai:** 2026-06-15 · **Backup GitHub:** commit `d9ce235` (last-good PRA-refactor).

## ATURAN KERJA (WAJIB)
- **GitHub = jaring pengaman.** State pra-refactor udah ke-push (`d9ce235`). Saat refactor:
  **commit LOKAL doang, JANGAN push.** Kalau gagal → `git reset --hard origin/main` (restore last-good).
- **Push HANYA** setelah 1 cabutan TUNTAS + tested + owner OK.
- **Test-gate tiap langkah:** sebelum lanjut, pastiin call LLM normal + jalur kena-dampak masih jalan.
- **⚠️ CEK DAMPAK GUI tiap langkah (owner: "gw sering luput GUI").** Tiap cabutan WAJIB dinilai:
  nyentuh GUI gak? (web/tabs/*.js, web/js/*, i18n, label, layout). Karena GUI **auth-gated** (gw gak
  bisa render/screenshot), kalau nyentuh GUI → gw: (a) verifikasi embed di binary + i18n key kebaca,
  (b) **LOGIN sendiri** (pass GUI di `<file-secrets-owner>`, JANGAN ditampilin; pakai cookie
  session) → cek **endpoint auth-gated** (data yg GUI tampilkan, mis. /api/apps, /api/groups) bener,
  (c) **LAPOR owner buat cek VISUAL via LOGIN + SCREENSHOT** (layout/render gw gak bisa lihat) sebelum
  langkah dianggap selesai. Tiap R di bawah ada tag **[GUI: ya/tidak]**.
- **LOCKED** boleh diedit (refactor) → re-lock + catatan bertanggal. **FREEZE (kernel) JANGAN disentuh.**
- Tiap cabutan punya **ROLLBACK** jelas di bawah.

## KENAPA REFACTOR (akar masalah)
Flowork dibangun cepat → tiap fitur jadi **sistem paralel sendiri** → fragmentasi:
3 gudang skill, 2 orchestrator, schedule vs trigger, crew-app vs fwapp, taksonomi "app" ambigu.
Akibat: bingung, rawan bug (cth nyata: "app ngak muncul", "team peramal → repo-reviewer"),
susah dirawat. **Cabut = satukan ke 1 model kanonik per konsep.**

## 🎯 FORMULA TARGET (1 sumber kebenaran per konsep)
| Konsep | SATU sumber (target) |
|---|---|
| Cipta | **AI Studio (architect chat)** — bikin semua *(sudah)* |
| Orchestrator | **`mr-flow-next`** (pegang Telegram + ask_group + groups) — `mr-flow` legacy dipensiunin |
| Skill | **brain** (embedded base + dynamic dir) · 1 inject (`allSkills`) · 1 kontrak tulis (endpoint) |
| Otomasi | **trigger engine** (time+event) · Schedule = view time-trigger |
| Eksekusi | **Agent** (1 wasm) → **Group** (koloni) · **App** (program) · **Skill** (playbook) — buang "category/crew" |

---

## R1 — SATUKAN SKILL (3 gudang → 1) + kontrak tulis · **[GUI: tidak]**
**Kenapa:** skill ke-author ke disk, tapi inject cuma baca embedded; DB-skill mati; coupling
agent-nulis-dir-router-baca rapuh. → 1 jalur jelas.
**Target:** `brain.allSkills()` = satu-satunya sumber inject (embedded + dynamic dir). Tulis skill
lewat **endpoint router** `/api/brain/skills/upsert` (router yg punya storage), bukan nulis dir lintas-proses.
**File disentuh:**
- `router/internal/brain/skills.go` (LOCKED) — allSkills (udah) + pastiin 1 jalur.
- `router/internal/router/routes.go` + handler baru — endpoint upsert skill → tulis ke DynamicSkillsDir.
- `agent/internal/routerclient/skills.go` — tambah UpsertSkill (POST ke router).
- `agent/architect.go` (LOCKED) — `authorSkill` panggil routerclient (ganti nulis-dir langsung).
- (opsional) pensiunin `router/internal/store/skills.go` jalur inject kalau gak dipakai.
**RESIKO:** 🟡 sedang — endpoint additive (aman); `skills.go` di jalur enrich (tapi fail-open).
**MITIGASI/TEST:** endpoint dites kirim+baca; enrich normal masih jalan (gate); revert = `git checkout`.
**ROLLBACK:** `git checkout router/ agent/architect.go agent/internal/routerclient/`.

## R2 — TAKSONOMI + BUANG "CATEGORY/CREW" · **[GUI: YA → owner cek login+screenshot]**
**Kenapa:** "app" ambigu (crew vs program); "category" (taskflow crew) tumpang-tindih Group.
→ definisi tegas, 1 jalur cipta+registry.
**Target:** Agent/Group/App/Skill/Schedule/Trigger jelas. "category/crew" lama dipensiunin
(diganti Group). Architect + GUI pakai istilah konsisten.
**File disentuh:**
- `doc/` — 1 dok taksonomi (sumber kebenaran istilah).
- `agent/architect_chat.go` (LOCKED) — system prompt + deskripsi tool pakai taksonomi.
- `agent/web/tabs/*.js` + i18n — label konsisten.
- (hati-hati) `agent/coder.go` / taskflow — kalau "category" mau dipensiunin dari UI.
**RESIKO:** 🟢 rendah (mostly prompt/label/doc) · 🟡 kalau buang taskflow-category dari backend.
**MITIGASI/TEST:** ubah label+prompt dulu (aman); pensiun backend belakangan + verifikasi gak ada yg manggil.
**ROLLBACK:** `git checkout` file terkait.

## R3 — MERGE ORCHESTRATOR (2 → 1: `mr-flow-next`)  ⚠️ CABUTAN TERDALAM · **[GUI: tidak — tapi test /api/chat + Telegram]**
**Kenapa:** `mr-flow` (legacy, classifyRoute→category) + `mr-flow-next` (Telegram, ask_group→group)
= 2 otak, logika dobel, bingung mana yg otoritatif.
**Target:** `mr-flow-next` jadi SATU-satunya. Logika berguna dari `mr-flow` (classify→route ke
group/app/architect) dipindah/diserap; `mr-flow` legacy dipensiunin. `/api/chat` default → `mr-flow-next`.
**File disentuh:**
- `agent/agents/mr-flow-next.fwagent/main.go` — serap routing (ke group/app/AI Studio).
- `agent/agents/mr-flow/main.go` (LOCKED) — pensiun (atau jadi thin-shim ke next).
- `agent/chat.go` — default agent `mr-flow` → `mr-flow-next`.
- `agent/main.go` — daftar daemon/orchestrator id.
- wasm rebuild (wasip1) + deploy ke `~/.flowork/agents/`.
**RESIKO:** 🔴 TINGGI — Telegram bot LIVE + 2 wasm + `/api/chat`. Salah = bot/chat mati.
**MITIGASI/TEST:** (1) backup wasm lama. (2) migrasi bertahap: next nyerep dulu, mr-flow masih ada.
(3) test 2 jalur: `/api/chat` + Telegram (kirim pesan beneran). (4) baru pensiun mr-flow setelah next terbukti.
**ROLLBACK:** restore wasm lama dari git/backup + `git reset --hard origin/main`.

## R4 — EXTENSION POINTS (anti-edit-core) · **[GUI: tidak]**
**Kenapa:** tiap extend kudu ngutak file LOCKED jalur-panas. → bikin titik-extend resmi.
**Target:** registry "deliver" (telegram/chat/…) plug-able; interface skill-provider. Extend = daftar, bukan edit core.
**File disentuh:** `agent/internal/triggers/engine.go` (LOCKED) — deliver jadi registry · `router/internal/brain/` — provider iface.
**RESIKO:** 🟡 sedang (refactor jalur-panas trigger). **MITIGASI:** dikerjakan SAMBIL R1/R3 (bukan terpisah). **ROLLBACK:** git checkout.

## R5 — FAILOVER + MODEL LOKAL · **[GUI: tidak (kecuali nambah kontrol Settings → owner screenshot)]**
**Kenapa:** opus 429 → jatuh ke LOKAL (Gemma ngawur), bukan haiku (cloud bagus). → urutan salah.
**Target:** priority provider: opus → **haiku** → lokal (cloud-bagus dulu sebelum lokal). + prompt architect/classifier ramping buat lokal.
**File disentuh:** setelan **priority provider** (Settings/router DB `data.sqlite` providers) — bukan kode, atau `router/internal/router/dispatcher.go` (LOCKED) kalau perlu urutan default.
**RESIKO:** 🟢 rendah (config). **TEST:** matiin opus → cek failover ke haiku (bukan lokal). **ROLLBACK:** balikin urutan.

---

## URUTAN EKSEKUSI (tiap satu TUNTAS + test sebelum lanjut)
1. **R1 Skill → 1** (paling dikuasai, low-risk, beres 2 masalah) — *commit lokal, test, JANGAN push.*
2. **R2 Taksonomi** (label/prompt/doc dulu; pensiun backend belakangan).
3. **R5 Failover** (config cepat, ngilangin sakit 429).
4. **R3 Merge orchestrator** (terdalam — backup, migrasi bertahap, test 2 jalur, revert siap).
5. **R4 Extension points** (sambil R1/R3).
→ Setelah SEMUA tuntas + tested + owner OK → **baru push** (GitHub jadi state pasca-refactor).

## CHECKPOINT
- Pra-refactor (last-good): GitHub `d9ce235`.
- Kalau gagal kapan pun: `cd FLowork_os && git reset --hard origin/main` → balik last-good.
