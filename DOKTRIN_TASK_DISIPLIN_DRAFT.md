# DOKTRIN AOLA-014 TASK-DISIPLIN (Worklog) — ACC DEV ✅

> STATUS (2026-06-27): di-ACC owner. **SEEDED** ke `router/internal/brain/doctrine_seed.json`
> (constitution, amp 999999) → kepakai di brain FRESH / install baru. Pondasi: `lock/worklog.md`.
> ⚠️ **Brain LIVE belum** (13 entri): inject runtime ke-blok editionGate(free) + safety-classifier.
> Owner mau live ("inject sekalian") → butuh keputusan jalur (lihat akhir file).

## KENAPA
MANDOR (idle-supervisor) cuma bisa rekonsiliasi kerjaan kalau tiap agent NYATET kerjaannya ke
`agent_run`. Ga nyatet = kerjaan ga keliatan di worklog = Mandor buta = "PC nganggur" ga ke-deteksi.

## TEKS (yang di-inject kalau di-acc)

```
### AOLA-014_TASK_DISIPLIN · amp 999999
Tiap mulai/lanjut kerja PANJANG: WAJIB `agent_run` (action create→checkpoint→complete). Catat task + todo, JANGAN cuma di kepala — ga kecatat = ngilang = dosa. Kerjaan kepotong/nunggu: `checkpoint` dulu biar bisa di-`resume`, JANGAN ngulang dari nol. Kelar → `complete`. Mandor baca worklog dari sini; kalau lo ga nyatet, kerjaan lo ke-anggap NYANGKUT.
```

## CATATAN
- Cuma buat kerja PANJANG/lintas-turn (bukan tiap balasan sepele) — biar ga noise + hemat token.
- Selaras AOLA-013 (kerja panjang: tidur-bangun) — 013 = KAPAN nunggu/bangun, 014 = CATAT biar ga ilang.
- Verifikasi pasca-acc: ukur token vs baseline (jangan nambah halu); cek mr-flow bikin agent_run via /api/worklog.
