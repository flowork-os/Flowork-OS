# 🔒 MESH LOCKBOX — FASE M/B (owner-gated, verified live 2026-07-02)

Status: LIVE. Owner di ruangan, ACC 3 keputusan: visibility default **PRIVATE**,
wall mesh↔evolusi **HARD (default aktif)**, tool manifest **consent gate ON**.
Semua via SEAM — **NOL buka file mesh tier-2 (super_scrit / super kramat)**.

## Prinsip: hardening + eksplisit, bukan bangun dari nol
Riset (2026-07-02) nemu mesh udah aman by-ISOLASI: knowledge mesh (`mesh_knowledge_inbox`)
TERPISAH dari knowledge lokal (`drawers`), tool manifest inert (discovery-only), LoRA
apply disabled, evolusi ga narik brain knowledge. FASE M/B bikin jaminan itu EKSPLISIT,
switchable, traceable, ke-tes — biar AI masa depan ga ngebocorin ga sengaja.

## FASE B — visibility publik-vs-privat (`internal/brain/drawer_visibility_ext.go`)
- Kolom `drawers.visibility` (default **'private'**, ALTER idempotent lazy — init.go frozen NOL disentuh).
- Guard tunggal `brain.IsShareable(db,id)` = HANYA 'public' boleh ke-mesh (fail-closed).
  Jalur share-ke-mesh manapun (sekarang/masa depan) WAJIB lewat sini.
- Endpoint: `GET/POST /api/brain/drawer/visibility`, `GET /api/brain/visibility/stats`.
- Verified live: 860.740 drawer lokal semua default private (nol bocor); set public↔private OK.

## FASE M — provenance + revoke (`internal/mesh/provenance_ext.go`)
- `ProvenanceByOrigin` = peta asal-usul knowledge federasi per peer (hitung per-status).
- `RevokeByOrigin(pubkey)` = peer jahat → invalidate SEMUA knowledge-nya sekali (status→
  invalidated, dropped di-skip) → drop dari graph projection sync berikut + penalti karma −0.2.
- Endpoint: `GET /api/mesh/provenance`, `GET /api/mesh/provenance/peer?pubkey=`,
  `POST /api/mesh/provenance/revoke`. Verified live.

## FASE M — verify-on-promote (`internal/mesh/verify_promote_ext.go`)
- Filter `verify-integrity` (RegisterMeshFilter, switch `FLOWORK_MESH_VERIFY_PROMOTE` default ON):
  auto-pipeline — node lokal ke-tamper (tier-2 `CoreClean()`=false) → REJECT (jangan promote).
- `VerifyPacketForPromote` + endpoint `POST /api/mesh/knowledge/approve-verified`: approve manual
  hardened — re-verify integritas node + tanda-tangan ed25519 paket asli dari `mesh_packets`
  SEBELUM promote. Verified live: paket ga ada → fail-closed (tolak, bukan diam-diam lolos).
- Approve lama (`/api/mesh/knowledge/approve`, frozen) tetep ada buat kompat; GUI arahkan ke -verified.

## FASE M — wall mesh↔evolusi (`internal/mesh/evolusi_wall_ext.go`)
- Switch `FLOWORK_MESH_EVOLUSI_ALLOW` **default OFF = tembok NAIK**: knowledge mesh
  (node id `mesh_know_*`) HARAM masuk konteks evolusi/codegen.
- `FilterOutMeshNodes(ids)` = jalur perakit konteks evolusi WAJIB buang node mesh dulu (tembok naik).
- `MeshInertInvariants()` = 4 invarian keamanan yg di-assert test (tool discovery-only,
  lora disabled, knowledge isolated, consent required). Kalau salah satu jadi false = bocor = bug.
- Endpoint observability: `GET /api/mesh/evolusi-wall`. Test `TestEvolusiWall` PASS.

## FASE M — consent gate tool manifest (`internal/mesh/tool_consent_ext.go`)
- Kolom `mesh_tool_manifests.consented` (default 0, migrasi additive `internal/store/mesh_lockbox_migration.go`).
- Manifest peer = PENDING sampai owner approve; discovery-consented cuma tampilin yang di-restui.
- Endpoint: `GET /api/mesh/tools/pending`, `GET /api/mesh/tools/consented`, `POST /api/mesh/tools/consent`.
- Verified live: approve 'weather' → pindah pending→consented.

## Switch (GUI fwswitch, kategori Mesh/Lockbox)
| Switch | Default | Efek |
|---|---|---|
| `FLOWORK_MESH_VERIFY_PROMOTE` | ON | promote cek integritas node dulu |
| `FLOWORK_MESH_EVOLUSI_ALLOW` | OFF | tembok mesh↔evolusi (OFF=tembok naik) |

## File (semua NON-FROZEN, deletable — delete-test PASS)
`router/handlers_mesh_lockbox_ext.go` · `router/internal/mesh/{provenance,verify_promote,evolusi_wall,tool_consent}_ext.go`
· `router/internal/brain/drawer_visibility_ext.go` · `router/internal/store/mesh_lockbox_migration.go`
· switch di `agent/internal/fwswitch/registry.go`. Unit test: `lockbox_ext_test.go` + `drawer_visibility_ext_test.go`.

## Batas yang disengaja
- Belum ada jalur share-DARI-drawers ke mesh (outgoing forward `mesh_packets`, bukan drawers).
  Visibility = guard forward-looking + API; begitu jalur share-drawer dibangun, WAJIB panggil `IsShareable`.
- Wall = guard + filter + test; enforcement penuh di recall-evolusi butuh jalur recall manggil
  `FilterOutMeshNodes` (dokumentasi kontrak; recall frozen belum di-wire — aman karena mesh ga di drawers).
- Freeze cluster lockbox SETELAH owner puas + live-tested lebih lama (belum di-chattr, sengaja).
