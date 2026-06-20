# Flowork Apps — App ↔ Agent Connector (STANDAR)

App Flowork itu **dwi-guna**: dipakai **user** (GUI) DAN **agent** (tool). Dokumen
ini standar wajib biar tiap app bisa dipakai keduanya — bukan cuma pajangan GUI.

---

## 1. Dua lokasi app (kenapa app "jam" ga ada di sini)

| Lokasi | Isi | Di-commit? |
|---|---|---|
| `agent/apps/` (repo ini) | App **bawaan** yang ikut dikirim ke semua user clone (`flowalpha`, `notepad`, `scam-detector`). | YA |
| `~/.flowork/apps/` (data home) | App **terinstall runtime** di mesin user (lewat tab Apps / `.fwpack`). Mis. `jam-digital-gelap`. | TIDAK (Pilar 6: data di `$HOME`, bukan repo) |

Jadi `jam-digital-gelap` ada di `~/.flowork/apps/`, **bukan** di repo — itu app buatan
owner yang diinstall lokal. App bawaan repo di-load + app data-home di-load, dua-duanya
muncul di GUI. **Sumber kebenaran daftar app = folder, bukan JSON** → install = folder
lahir → muncul; uninstall = folder ilang → hilang. Otomatis, zero-sync.

---

## 2. Anatomi app

```
<app-id>/
  manifest.json     # identitas + daftar operasi (WAJIB)
  core.py           # CORE: handler operasi (kalau runtime:process)
  ui/index.html     # GUI buat user (gui_entry)
  ui/icon.svg
  state.json        # opsional: persistensi punya app sendiri
```

`manifest.json`:
```json
{
  "id": "jam-digital-gelap",
  "kind": "app",
  "name": "Jam Digital",
  "runtime": "process",
  "core_entry": "python3 core.py",
  "gui_entry": "ui/index.html",
  "operations": [
    { "name": "get_time", "tool": true, "gui": true, "mutates": false,
      "description": "Waktu sekarang (24j/12j/ISO/tanggal/epoch/zona).",
      "input_schema": { "type": "object", "properties": {} } }
  ]
}
```

**CORE protokol** (bahasa bebas — bukti lintas-bahasa): baca `{"op","args"}` per baris
di stdin, balas `{"result": ...}` atau `{"error": ...}` per baris di stdout. Host
spawn core sekali, reuse; cwd = folder app (buat persistensi sendiri). Lihat
`notepad/core.py` & `jam-digital-gelap/core.py` sebagai contoh.

---

## 3. Cara app NEMPEL ke agent (konektor)

Tiap operasi dengan **`"tool": true`** = **konektor agent**. Saat boot, host daftarin
operasi itu sebagai tool:

- **Nama tool** = `app_<id>_<op>`, char non-alfanumerik (mis. `-`) jadi `_`.
  Contoh: app `jam-digital-gelap` op `get_time` → tool **`app_jam_digital_gelap_get_time`**.
- **Capability** = `app:<id>` (pakai id asli, tetap pake `-`). Contoh: `app:jam-digital-gelap`.
- Agent nemu + pake lewat `tool_search` → `/api/agents/tools/run` → `SandboxRunV3`.

Operasi tanpa `tool:true` = cuma GUI (user-only), **agent ga bisa pake**.

---

## 4. Izin: GUI = KEBENARAN

Kepunyaan tool ≠ boleh-eksekusi. Capability `app:<id>` digate **per-agent** lewat
fitur **"Allowed Apps"** di GUI agent (tab Agents → Setting → section Tools →
"Aplikasi yang boleh dipakai"):

- Centang app → `app:<id>` di-grant ke broker **LIVE** (langsung bisa, tanpa restart).
- Uncentang → cap dicabut → agent ga bisa pake app itu.
- Disimpen di tabel `app_grants` per agent (`state.db`), BUKAN JSON. Backend:
  `app_grants_handler.go` + `internal/agentdb/app_grants.go`.

---

## 5. STANDAR WAJIB bikin app baru

1. **WAJIB punya ≥1 operasi `tool:true`** (konektor agent). App tanpa konektor =
   user-only — ga lolos standar. (Contoh retrofit: `jam-digital-gelap` dikasih
   `get_time` biar agent bisa pakai, bukan cuma user lihat jam.)
2. **WAJIB punya `gui_entry`** (guna buat user) — app dwi-guna.
3. `runtime` + `core_entry` diisi kalau ada operasi (core yang handle).
4. Tiap operasi tool: `description` jelas + `input_schema` valid (LLM baca ini).
5. `mutates: true` buat operasi yang ngubah state (GUI auto-sync ke perubahan agent).
6. Persistensi (kalau perlu) tanggung jawab app sendiri (file/db di folder app).

> Intinya: **app dibikin buat agent DAN user**. Kalau cuma salah satu, belum lengkap.
