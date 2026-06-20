# 🧬 STANDAR WARGA FLOWORK — Kontrak & Alur Tiap Agent

**Sumber:** roadmap_agent.md (§5/§5b/§6.3) + roadmap_opus8.md (D7/D11/D12) — didistilasi & **divalidasi live** (Opus 4.8 [1m], 2026-06-20).
**Untuk:** template tiap agent baru (incl. hasil AI Studio kelak). Ini "DNA" yang HARUS dimiliki tiap entitas "mikir" di Flowork.
**Aturan emas:** plug-and-play · terisolasi · DB-based (NO .md) · pure-Go no-cgo · multi-OS · **cabut akar bukan tambal** · no claim tanpa test.

---

## 0. RINGKAS — 8 SYARAT WAJIB tiap agent lahir
1. **Config DB-based** (persona di DB, BUKAN .md). [§1]
2. **DNA sacred** auto-seed via `ProvisionAgentDNA(id)` (konstitusi + edu-errors + antibody + schema cognitive graph). [§2]
3. **Pipa Otak Bersama** (genome): `brain_search_shared` + `instinct_recall` + `graph_recall`. [§3]
4. **Cognitive Graph LOKAL** per-agent (twin/relasi sendiri). [§4]
5. **Identitas penting masuk BRAIN/recall**, bukan cuma persona. [§5]
6. **Capability `rpc:router:brain`** di manifest. [§6]
7. **Autonomy: loop + sleep/bangun + anti-ghosting** (sadar kapan pakai apa). [§7]
8. **GUI standar** (Inggris + i18n, login+screenshot QC, display JUJUR). [§8]

---

## 1. CONFIG = DATABASE, BUKAN .md
- Persona/prompt → DB: `config.prompt` + `self_prompt` (slot versioned) + `constitution` (sacred). **HARAM** baca persona dari file `prompt.md`/`doktrin.md` (jebakan agent lama → mismatch DB-vs-file).
- Reference benar: `agents/mr-flow/main.go` `buildSystemPrompt()` (3-tier: STABLE persona+aturan / KONTEKS self_prompt+skill / VOLATILE waktu+memory).

## 2. DNA SACRED — `ProvisionAgentDNA(agentID)` (idempotent)
Dipanggil saat **boot loop** + **install path** → agent lahir warga penuh TANPA restart. Isi: konstitusi sacred + edu-errors + antibody + ensure schema cognitive graph.
**Konstitusi sacred wajib** (`constitution.go` `sacredSeed`, amplitude 999999, always-inject):
- `5w1h-gate` · `identity-guard` · `anti-halu` · `sync-honest` (anti janji-background palsu) · `recall-first` (recall brain/twin dulu sebelum nyuruh owner ngetik ulang).

## 3. PIPA OTAK BERSAMA (genome — di-REFERENSI, bukan copy)
`coreExposedTools` (tool_specs.go) tiap agent auto punya:
- `graph_recall` (state:read) — recall cognitive graph SENDIRI (twin/relasi).
- `instinct_recall` + `brain_search_shared` (rpc:router:brain) — insting coding/security + pengetahuan kolektif (859K) di SHARED brain router.
Update shared sekali → semua agent lihat (no drift). Cap-denied = graceful.

## 4. COGNITIVE GRAPH LOKAL per-agent
`cognitive_nodes`/`cognitive_edges` di `state.db` sendiri. Diisi via digestion/dream (percakapan→node+relasi), dibaca via `graph_recall`. Owner-facing → twin (sejarah/keluarga); agent lain → relasi domainnya. **Ini memori relasional agent.**

## 5. IDENTITAS MASUK BRAIN (temuan TERPENTING)
Model 26B percaya **BRAIN (recall) = "DATA"**, persona/doktrin = "role". Ditanya "siapa owner" → dia cek "ada di brain?". Kalau cuma di persona → bisa NOLAK ("itu halu"). → **Fakta identitas WAJIB masuk drawer/recall**, bukan cuma persona.

## 6. CAPABILITY (manifest `capabilities_required`)
Kapabilitas ASLI = manifest, BUKAN `config.tools` (override coarse, sering kosong). GUI Tools section WAJIB tampilin caps manifest (jujur — jangan keliatan "0 tool" padahal punya). Minimal coordinator: `state:read/write`, `time:read`, `fs:read/write`, `net:fetch:<endpoint spesifik>`, `rpc:router:brain`, `rpc:router:skill`, `rpc:taskflow`, `rpc:agent-invoke`.

## 7. AUTONOMY — SADAR kapan loop / tidur / bangun (VALIDATED LIVE 2026-06-20)
Mekanisme SEMUA real (bukan shim):

| Mode | Tool/Mesin | Cara kerja |
|---|---|---|
| **Loop dalam-turn** | tool-loop `maxToolIters=12` | model panggil tool berkali-kali sampai jawab (serial 1/iter) |
| **Anti-GHOSTING** ⭐ | `ghost-guard` di tool-loop | teks-tanpa-tool yg nyinyalin niat-aksi ("tunggu bentar gw cek dulu") → DIPAKSA panggil tool / `ScheduleWakeup`, bounded `maxGhostNudges=2`. **Deterministik, model-agnostic.** ✅ proven: guard fire → agent ACT 6 tool-call |
| **Tidur→bangun** ⭐ | `ScheduleWakeup` + `RunDueWakeups` | tulis `wakeups` row (due+prompt) → poller kernel /menit fire yg due → agent kebangun + lanjut + notify owner. ✅ proven: fired ~19s |
| **Anti-anchor history** | `isAnchorNoise` di `fetchHistory` | skip reply gagal/denial LAMA (di luar `keepRecentTurns=4`) biar 26B ga ngechо pola basi |
| **Job ke crew** | `task_run` (notify balik) | analisa berat → trigger crew, hasil nyusul via notify (BUKAN ghosting) |
| **Schedule per-agent** ⭐ | `scheduler/engine` (auto-tick/menit) | cron per-agent → fire task ke agent. ✅ proven: auto-fire ~55s. (Beda dari global taskflow scheduler — KEEP dua-duanya) |

**ATURAN ANTI-GHOSTING (konstitusi):** "Kalau mau ngapain → LAKUIN di balasan SAMA (panggil tool). Butuh nunggu → `ScheduleWakeup`. Job berat ada timnya → `task_run`. Tiap janji 'nanti' WAJIB ada tool yg beneran nepatin. JANGAN ninggalin owner nunggu jawaban yg ga datang."

## 8. GUI STANDAR
- **String Inggris + i18n catalog** (`web/i18n/{en,id}/`). Komunikasi ke owner = Bahasa Indonesia.
- **Display JUJUR** — menu nampilin data ASLI (mis. Tools = caps manifest, bukan override kosong). No "pajangan".
- **QC WAJIB: jalanin app → LOGIN → SCREENSHOT → LIHAT PNG.** Tanpa itu = GUI belum diverifikasi = fase belum selesai. (Recipe: chrome headless + CDP via python websocket, set cookie `flowork_session`, navigate hash route.)
- Menu agent (verified work, bukan pajangan): Prompt(DB) · Tools(caps manifest) · Schedule(per-agent, auto) · Telegram · Diagnostics(Interactions/Decisions/Mistakes/Karma/DeathLetter/Workspace/ToolAudit/Slash) · Doktrin(edu-errors).

---

## 9. ALUR PESAN AGENT (end-to-end)
```
Pesan masuk (Telegram daemon / debug-chat / scheduler / wakeup)
  → validasi (token + chatID whitelist)
  → fetchHistory (window-capped + anti-anchor skip denial lama)
  → buildSystemPrompt (3-tier, budget-capped, recall-first)
  → tool-loop (maxToolIters):
      LLM → minta tool? → eksekusi → feed hasil → ulang
                       → jawab teks? → GHOST-GUARD cek:
                            niat-aksi tanpa tool → PAKSA lanjut (tool / ScheduleWakeup)
                            jawaban final → return
  → log interactions + decisions + karma
  → (deferred) ScheduleWakeup row → RunDueWakeups fire nanti → loop lagi
```

## 10. PANTANGAN (sacred)
- ❌ Persona di .md · ❌ claim selesai tanpa test · ❌ hardcode path · ❌ sentuh file LOCKED/FREEZE tanpa izin+alasan+re-lock · ❌ ghosting (janji tanpa tool) · ❌ display "pajangan" (menu gak nunjukin data asli) · ❌ data personal owner ke shared/mesh.
- ✅ cabut akar bukan tambal · ✅ test via debug-chat harness (jalur asli) + angka · ✅ GUI login+screenshot · ✅ isolasi penuh (agent A rusak → cuma folder A).

---

## 11. REFERENCE AGENT
**`agents/mr-flow/`** = reference build (DB-based, full DNA, autonomy validated). Build: `GOWORK=off GOOS=wasip1 GOARCH=wasm go build` (standard wasip1, ~3.5MB — BUKAN tinygo; `build-agent.sh` STALE). Deploy → `.flowork/agents/<id>.fwagent/agent.wasm` → restart.
