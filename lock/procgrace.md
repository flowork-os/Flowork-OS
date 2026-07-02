# Cascading shutdown proses anak (F-G) — 2026-07-02

## Kenapa
Kill langsung (SIGKILL) ke proses anak (app/MCP/http-adapter) = ga sempat
flush/tutup file/lepas port → data corrupt / port nyangkut. Ganti jadi bertahap.

## Arsitektur
- `agent/internal/procgrace/procgrace.go` (FROZEN 2026-07-02): `Stop(cmd)` naik-tahap
  SIGINT → (tunggu grace) → SIGTERM → (tunggu grace) → SIGKILL, lalu REAP via Wait
  internal (ga ninggalin zombie). Aman dgn cmd/Process nil. MULTI-OS: cuma
  os.Interrupt + syscall.SIGTERM + Process.Kill (SIGTERM yg ga didukung Windows
  di-ignore, Kill nutup). Grace per-tahap = switch `FLOWORK_PROC_GRACE_MS`
  (default 2500, min 100, cap 30000).
- Call-site (3 file FROZEN, unlock-sadar + re-lock, delegasi ke helper — logika
  tuning ada di procgrace, bukan di call-site = prinsip switch):
  - `internal/apps/proc.go` `close()`
  - `internal/mcpclient/mcpclient.go` `Close()` (buang Wait manual — Stop udah reap)
  - `internal/apps/httpadapter/adapter.go` `stop()`

## QC 2026-07-02
Unit (nil-safe, terminate sleep, switch+cap) PASS · build/vet/TestKernelFreeze
hijau · test paket apps/mcpclient/httpadapter PASS.
