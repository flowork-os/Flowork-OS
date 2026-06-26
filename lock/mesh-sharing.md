# Mesh Sharing — share on/off + approve manual/auto + 2-tier security

> Owner: Aola Sahidin (Mr.Dev) · floworkos.com. Switch GUI: fwswitch (Mesh/Sharing). Tier-2: integrity.md.

## 2 tingkat keamanan (konfirmasi owner 2026-06-27)
- **Tingkat 1 — FREEZE (anti-edit):** `chattr +i` + hash `KERNEL_FREEZE.md` (enforce `TestKernelFreeze`).
  Jaga semua file frozen biar AI lain tak ngubah tanpa sadar. Manifest privat (gitignored).
- **Tingkat 2 — INTEGRITY GATE (jantung):** manifest `flowork-secrets/super_scrit.md` (PRIVAT, di luar
  repo, gitignored, `chattr +i`, BUKAN di-upload). 32 file jalur trust mesh. Kalau salah satu BERUBAH
  (root-hash mismatch) → node tampered → TOLAK semua pengetahuan mesh. Gate: `internal/mesh/integrity.go`
  baca `../super_scrit.md` (fallback KERNEL_FREEZE.md; env `FLOWORK_KERNEL_MANIFEST` buat relokasi/USB).

## Share / Approve (semua di GUI — kebenaran di GUI)
Switch fwswitch (tab Switch Fitur, kategori Mesh/Sharing), live lintas-proses:
- **`FLOWORK_MESH_SHARE`** (bool, default **ON**). OFF → otomatis TIDAK nerima & TIDAK relay (adil/reciprocity).
- **`FLOWORK_MESH_APPROVE`** (string, default **manual**). manual = pengetahuan masuk ditahan (quarantine),
  owner approve di GUI. auto = promote otomatis kalau lolos filter.

## Logic
- Gate via seam `RegisterMeshFilter` (filter `share-policy`, FROZEN) di `RunFilterPipeline`:
  - `!share` → **reject** (drop) — reciprocity: tak share = tak terima.
  - `approve != auto` → **flag** → `StatusQuarantine` (antrian approve).
  - else → **pass** (promote kalau lapis lain lolos).
- Relay-out: `gossip.tick` skip kalau `!meshShareEnabled()` (tak nyebar paket pas share OFF).
- Approve manual (GUI), via seam `RegisterExtraRoute`:
  - `GET  /api/mesh/knowledge/pending` → antrian quarantine.
  - `POST /api/mesh/knowledge/approve {packet_id}` → promote (terima).
  - `POST /api/mesh/knowledge/reject  {packet_id}` → drop (tolak).

## File yang dilewati
- `router/internal/mesh/policy.go` (helper share/approve), `filter_meshpolicy.go` (gate), `gossip.go` (relay gate),
  `integrity.go` (tier-2 root-hash), `pipeline.go`+`knowledge.go` (status quarantine/promoted/dropped) — semua FROZEN tier-2.
- `router/handlers_mesh_approve_ext.go` (endpoint approve, FROZEN) via `routes_ext.go`.
- `agent/internal/fwswitch/registry.go` (switch GUI, non-frozen).
- `flowork-secrets/super_scrit.md` (tier-2 manifest, PRIVAT).

## Teknologi
- Switch env lintas-proses (fwswitch file `~/.flowork/flowork_settings.json`, watcher 3 dtk, live).
- Mesh filter pipeline (decision pass/flag/reject) + status inbox (shadow/quarantine/promoted/dropped).
- sha256 root-hash tier-2 (integrity.md).

## Status (live 2026-06-27)
- `/api/integrity` checked:32 clean (tier-2). `/api/mesh/knowledge/pending` jalan. Default share ON + manual approve.
- Test: `TestMeshSharePolicy` (share off→reject, manual→flag, auto→pass) + `TestCoreIntegrityCleanAndTamper` PASS.

## Catatan keamanan tier-2 (super_scrit.md)
- Di luar repo + gitignored + chattr +i → **tak ke-upload**, AI tak bisa forge tanpa sudo.
- Mau "hilang dari PC" total: simpan di USB/eksternal, set env `FLOWORK_KERNEL_MANIFEST` ke path-nya
  (gate baca dari sana). Kalau benar-benar absen → gate fail-open (clean=true) — opsi hardening:
  embed root-hash tier-2 sbg const di binary (anchor) biar gate jalan tanpa file. Belum dibangun (opsional).
