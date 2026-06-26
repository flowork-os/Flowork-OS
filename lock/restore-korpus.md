# RESTORE KORPUS 800rb (#4) — security/CVE corpus balik, ZERO re-embed

> Owner: Aola Sahidin (Mr.Dev) · 2026-06-26. Korpus hacking/crypto/CVE (859k drawer) yg ke-trim
> dari brain live (tinggal 421) di-restore dari backup, TANPA re-embed (reuse vindex prebuilt).

## AKAR
Brain router live = 421 drawer (4.8MB) — korpus security ke-reset/trim. Backup
`/home/mrflow/Videos/FLowork_os/router/brain/` (Jun-21) punya: `flowork-brain.sqlite` 7.2GB
(859,808 drawer: trickest_cve_*, cve_*, mitre_attack, php_webapps, nuclei, pentesting, dll) +
**`brain.vindex` 854MB prebuilt** (Jun-17, vektor int8 korpus). Kunci: vindex pakai hash-id sama
dgn drawer → **restore = import drawer + pasang vindex, ga perlu re-embed 800k (yg makan berhari-hari).**

## AKSI (reversible)
1. Backup brain+vindex current → scratchpad.
2. Copy current → `flowork-brain.NEW.sqlite`; `ATTACH backup (immutable=1)` (bypass WAL-lock) →
   `INSERT OR IGNORE INTO drawers SELECT ... WHERE deleted_at IS NULL` (859,808 row, ~11 detik,
   journal=OFF). Hasil: 421 + 859,808 = **860,229 drawer**. Dedup natural by hash-id (TEXT PK).
3. Stop router → swap: `flowork-brain.NEW.sqlite` → `flowork-brain.sqlite`, backup `brain.vindex`
   → `router/brain/brain.vindex` (854MB). Old disimpen `*.OLD-421.sqlite` / `*.OLD-390k` (reversible).
4. Router respawn → load vindex 854MB (859k vektor) + 860k drawer.

## VERIFIKASI (live)
- Search korpus relevansi TINGGI: "Log4Shell JNDI"→0.70 trickest_cve_2021; "SQL inj auth bypass"
  →0.77 src_pentesting_web; "reverse shell privesc"→0.66 src_windows_hardening.
- **NO REGRESI recent**: instinct injection jalan + domain-tepat ("private key crypto" → injected
  [crypto tool crypto]); constitution 13 rules; instinct (365) + constitution (13) recent UTUH.
- Total drawer API: 860,149 · integrity quick_check OK.

## CATATAN
- `flowork-brain.sqlite` + `brain.vindex` = **gitignored** (data, GEDE — JANGAN commit; auto-update
  user ga bawa korpus ini, itu per-mesin/opsional). Re-digest=DATA op lokal.
- **Coverage gap kecil**: vindex Jun-17, drawer Jun-21 → ~4 hari drawer baru korpus + insting recent
  BELUM di main-vindex → di-cover fresh-index (recent) ATAU butuh full rebuild (`cmd/brain-reembed`
  + `brain-buildindex`) kalau mau 100% vector-coverage. Bulk korpus (Jun-17) udah ke-cover → search jalan.
- Reversible: `router/brain/*.OLD-*` + scratchpad backup. Kalau halu → swap balik.
