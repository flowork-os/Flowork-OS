# Flowork OS — P1 Specification (DEFINE): Persistence

> Phase **P1 — Persistence**. Builds on P0 (do not start until P0 is green in STATUS.md).
> Pipeline stage: **DEFINE**.

## Goal
Flowork state survives a power-off: a **separate, LUKS-encrypted DATA partition** holds
`/root/.flowork` (the agent/router state — `flowork.db`, agent loket DBs, settings). The system
root stays **ephemeral/read-only**; only DATA persists. This is the design's *hybrid* mode
(RO/ephemeral system + encrypted data overlay).

## Why this shape
- **DATA separate from ROOT** → updating the OS (swapping ROOT) never erases user data. (Design §5.)
- **LUKS** → a lost/stolen stick reveals nothing *without the key*. (P1 proves the encryption
  mechanism; secure key custody — TPM/passphrase — is **P3/P4**, see non-goals.)
- **Keep P0's proven ephemeral root** → lowest-risk increment; we add a persistence layer, we do
  not rewrite the boot path.

## Success criteria (binary, testable)
| # | Criterion | How verified |
|---|-----------|--------------|
| P1.1 | A blank DATA disk is initialized as **LUKS** on first boot (header written, ext4 made) | serial: "initializing LUKS"; `cryptsetup isLuks` true afterwards |
| P1.2 | `/root/.flowork` is backed by the encrypted DATA mount before agent/router start | serial: "persistent /root/.flowork mounted (encrypted)"; mount shows mapper device |
| P1.3 | State written on boot N is present on boot N+1 (across a real reboot) | a boot-counter in `/root/.flowork` reads **2** on the second boot of the same DATA disk |
| P1.4 | `flowork.db` created by the agent persists across reboot | serial reports `flowork_db=yes` on the second boot |
| P1.5 | With **no** DATA disk attached, the system still boots ephemerally (P0 unbroken) | P0 run (no DATA img) still reaches AGENT_UP/ROUTER_UP |

## Non-goals (deferred)
- **Secure key custody.** P1 bakes a per-build random keyfile into the image to prove the
  encryption path. That protects DATA against someone who has the *disk* but not the *image*.
  Real custody (passphrase at boot / TPM-sealed key) is **P3/P4** (verified boot).
- Disk-image/GRUB packaging with squashfs-on-disk ROOT + A/B partitions → **P4** (the squashfs
  artifact already exists; P1 keeps the ephemeral initramfs root).
- LUKS for the ROOT itself (ROOT is non-secret, read-only) — not needed.

## Deliverables
- `build/make-data-disk.sh` — create a blank DATA disk image.
- `rootfs-overlay/.../flowork-data` service + `flowork-data-setup` — init/unlock/mount DATA,
  populate the agent seed, bind it over `/root/.flowork`.
- `build/verify-p1.sh` — boot twice on the same DATA disk and assert the counter reaches 2.
- STATUS.md updated with the P1 verification record.
