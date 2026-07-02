# 📦 PORTABLE v0.10.0 — SIDECAR SEED (verified terisolasi 2026-07-02)

Status: VERIFIED. Manifest kanonik: `flowork-secrets/sidecar.md` (list di
`os/portable/make-portable.sh` HARUS sinkron). Builder wajib baca `os/builder/bacaduluini.md`.

## Arsitektur
- **Stick layout**: `bin/<os>/` (binari cross-compile) · `sidecar/` (aset runtime:
  agents/apps/tools/brain/models/skills/docs/doc/…) · `data-seed/` (CUMA build dev/Full).
- **Launcher wiring (dulu BOLONG — public stick boot 0 agent + 0 tool)**, sekarang di
  `start-flowork.sh` + `Start-Flowork.bat` + `Start-Flowork-Background.bat` (PARITY):
  1. `sidecar/agents/*.fwagent` → seed **ONLY-IF-ABSENT** ke `$DATA/.flowork/agents`
     (agent lokal user MENANG; WAL/SHM stale dibuang — jebakan revert SQLite).
  2. `sidecar/tools` → copy ke work-dir lokal (stick FAT = noexec) + export
     `FLOWORK_TOOLS_DIR` (resolver `toolsidecar.ToolsDir` baca env ini duluan).
- **Guardian**: `FLOWORK_GUARDIAN_AUTO=0` di launcher (distributable JANGAN auto-arm —
  baseline pasti mismatch → SAFE-MODE false-positive). Owner arm di mesin final.

## Bukti verifikasi (terisolasi: agent :11987, router URL mati, HOME sendiri)
```
kernel: agent scan complete: 1 accepted, 0 rejected      ← mr-flow.fwagent ke-seed (N>0)
tools-sidecar: 8 tool ke-register dari <work>/tools      ← FLOWORK_TOOLS_DIR jalan
```
Stack live :1987/:2402 TIDAK tersentuh selama tes.

## Fix nebeng (akar, multi-OS)
- `agent/feature_health_doctor_ext.go` pake `syscall.Statfs` TANPA build-tag →
  cross-compile Windows/mac GAGAL (= portable ga kebuild). Split per-OS:
  `feature_health_doctor_disk_unix.go` (linux|darwin) + `_other.go` (stub fail-open).
  Verified: GOOS=windows + darwin/arm64 build OK.

## Batas yang disengaja
- Binary sidecar tools masih HOST-build (linux) — Windows/mac exec tool = TODO
  cross-OS per-target (udah dicatat di make-portable.sh).
- Jalur img/appliance (`os/build/build-flowork-os.sh`) pakai mekanisme sendiri
  (copy agents/. ke rootfs) — ga lewat wiring launcher ini.
