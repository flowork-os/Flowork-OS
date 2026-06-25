# SLEEP-DRIVEN (#7) — kerja-LAMA via tidur→bangun (bukan loop panas)

> Owner: Aola Sahidin (Mr.Dev) · 2026-06-26. Akar #7 ("kurang insting kapan tidur/bangun + design")
> SUDAH KEISI. Stack lengkap + ke-wire. Detail tool: `lock/list_filtur.md` §WAIT/WAKE/LOOP.

## STACK LENGKAP (semua udah ADA + ke-wire)
| Lapis | Komponen | Status |
|---|---|---|
| **Tidur bangun sendiri** | `ScheduleWakeup` (tool, claude_tools.go) → nulis row `wakeups` (due_unix) | ✅ |
| **FIRE wakeup** | `wakeup_engine.go RunDueWakeups` — poller **tiap 1 menit** (main.go ticker) baca tiap agent-store, due+fired=0 → `host.InvokeAgentMessage(prompt,"wakeup")` → mark fired → notify owner | ✅ |
| **Jadwal rutin (cron)** | `scheduler_schedule_add/list/next` + `RunDueSchedules` (poller sama) | ✅ |
| **Kerja gede lintas-sesi** | `agent_run` (create/checkpoint/pause/resume/stop) — resume dari checkpoint terakhir | ✅ |
| **Auto-continue in-turn** | agentkit `loopBudgetMs ~200s` lewat → auto `ScheduleWakeup(5s, lanjutan)`, bounded `maxAutoContinue=50` | ✅ |
| **Anti-stall** | ghost-guard maksa `ScheduleWakeup` kalau model "nunggu" tanpa tool (bukan diem) | ✅ |
| **INSTING/DOKTRIN "kapan"** | **AOLA-013_KERJA_PANJANG** (constitution, live inject) — ngajarin kapan pakai tiap tool | ✅ 2026-06-26 |

## KESIMPULAN
#7 ga butuh kode BARU — mekanismenya udah komplit + ke-wire (poller 1-mnt jalan di main loop).
Yang DULU kurang = GUIDANCE (agent ga tau kapan/gimana tidur-bangun) → keisi **AOLA-013** (verified
inject: "constitution: injected 13 sacred rule(s)"). Jadi #7 = CLOSE.

## CATATAN VERIFIKASI
- Mekanisme FIRE = **code-verified** (`RunDueWakeups`/`RunDueSchedules` di poller 1-mnt, kode lurus:
  query due → invoke → mark fired → notify).
- **Live-fire-test SENGAJA di-SKIP**: fire wakeup → notify owner via Telegram. Owner lagi istirahat
  (3 hari ga tidur) → JANGAN di-spam tes tengah malam. Empiris fire bisa dicek kapan2 owner bangun:
  `grep "wakeup: .* di-fire\|\[wakeup\] fired" /tmp/flowork-gui.log`. Ga ada kode yg diubah buat #7
  (zero risk), jadi ga perlu freeze apa2.
