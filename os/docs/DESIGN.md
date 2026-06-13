# Flowork OS — Design (public)

> Turn Flowork from "an app you install" into "a sovereign computer in your pocket."
> Plug a USB into any PC → it boots straight into Flowork → unplug → zero trace on the host.

This is the public technical blueprint. Each decision states **what** and **why**, so the next
engineer (or AI) understands the reasoning, not just the result.

## Locked decisions
| Topic | Decision | Why (short) |
|-------|----------|-------------|
| Appliance type | **Kiosk (boots to a screen)** | Plug USB → the Flowork UI appears on the host's display. A self-contained box. |
| Browser | **One chromium, dual-mode** (kiosk + headless) | One engine: shows the UI *and* does agent web-automation. Smaller image. |
| App languages | **Python + Go only** | Go = static binary (zero runtime). Python = one interpreter + per-app venv. Tiny base, no multi-runtime container engine. |
| Router | **Bundled on the same appliance, separate binary/repo** | Sovereignty (local LLM) + clean architecture + independent updates. |
| Auto-update | **A/B partitions, SIGNED images, pulled from our GitHub Releases** | Atomic + rollback-able + anti-MITM + we own the channel. |
| Linux base | **Alpine** (PoC) → **Buildroot** (locked product) | Small + fast first; reproducible + verified-boot when it ships. |

## Anatomy (bottom → top)
```
KIOSK     cage + chromium --kiosk → localhost:1987      what the user sees
APPS      Go binary + Python venv, bwrap-sandboxed      apps, confined to their room
FLOWORK   agent (:1987) + router (:2402) + Ollama        the brain, autostart services
USERLAND  minimal Linux (Alpine→Buildroot) + CPython+bwrap+cage
KERNEL    Linux (wifi/gpu/usb drivers)
BOOT      GRUB/systemd-boot (from USB) + host Secure Boot
   ▲ everything on the flash drive. The host (Windows) disk is never touched.
```

## Why a browser at all, and only one
Flowork needs a browser for two different jobs — kept separate:
- **Show the UI** (kiosk): `cage` (a single-window Wayland compositor) + chromium full-screen on
  `localhost:1987`. No desktop environment → hundreds of MB and attack surface avoided.
- **Agent web access:** ~90% is plain HTTP (fetch READMEs, post to X/Dev.to/Telegram/Graph API) and
  needs **no browser** — the Go host does it. Only JS-heavy scraping/cookie automation uses chromium
  **headless** via CDP. One chromium binary serves both roles.

## Persistence model (P1+)
```
USB: [EFI] bootloader · [ROOT-A] read-only rootfs · [ROOT-B] spare (A/B update) · [DATA] LUKS-encrypted
```
DATA is separate from ROOT so **updating the OS never erases user data** (brain/agents/model). LUKS
means a lost/stolen stick reveals nothing. Three modes: ephemeral (RAM) · persistent · hybrid (RO
system + encrypted data overlay ← product target).

## Sovereign router (why local, why co-located)
```
Agent (:1987) ──LLM──▶ Router (:2402) ──▶ Ollama local (GGUF)   default: sovereign, offline-capable
                                       └──▶ Cloud API (opt-in, owner key)
```
If the router lived in someone else's cloud, prompts would leave the box → not sovereign. Local
router + local model = **zero data egress**. Collective "wisdom" sync is possible later, but
**opt-in**; sovereignty is the default.

## Verified boot / anti-tamper (the moat)
Wealthy clients put their whole mind in it; enterprises put secrets in it — they buy **trust**, and
trust = "this OS can't be hijacked." Staged: squashfs RO (P0) → dm-verity (P3, every block hashed) →
signed images (P3, Secure Boot chain) → A/B update (P4). The cloud giants can't copy this because
their business model is to *collect* your data.

## Auto-update (pulled from our releases)
```
CHECK GitHub Releases → DOWNLOAD rootfs to the idle partition → VERIFY signature vs baked-in key
→ SWITCH bootloader to it → REBOOT; health-check passes → commit, else AUTO-ROLLBACK.
```
A/B (not overwrite-in-place) means a power cut mid-update can never brick the box. Signing means a
hijacked GitHub or a MITM can't feed it a poisoned image. DATA is never touched by an update.

## Roadmap
| Phase | Goal | Why the tech |
|-------|------|--------------|
| **P0** | Alpine + Flowork (agent+router) + chromium kiosk boots in a **QEMU VM**, ephemeral. | VM = seconds-fast iteration, zero risk to real hardware. *(this release)* |
| **P1** | + LUKS DATA partition; state survives reboots. | Encryption at rest + safe updates that don't wipe data. |
| **P2** | + Ollama + local router + a small **Qwen** GGUF; runs **offline**. | Local inference = the core sovereignty pitch. Qwen = Apache-2.0, multilingual, strong tool-use. |
| **P3** | + CPython + **bubblewrap** app sandbox (Python venv / Go binary); + dm-verity + signed images. | 2 languages + controlled runtime → venv+sandbox beats a container engine. Anti-tamper = the moat. |
| **P4** | + A/B auto-update + certified hardware + GPU + branding. | Bundled hardware = guaranteed drivers + "plug-and-go". GPU = serious local LLMs. |

## Hard realities (kept honest)
- wifi/GPU drivers vary per PC → Alpine + firmware covers ~90%; the rest → certified hardware (P4).
- Secure Boot is on by default → PoC disables it in BIOS; the product uses a signed shim.
- No-GPU LLM only runs small models → a serious appliance has a GPU (mini-PC), not just a stick.
- Phones (locked ARM bootloaders) are **not** this USB track — that's the Android-app/remote track.
