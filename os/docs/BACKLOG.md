# Flowork OS — Deferred Backlog

> Principle (owner): **do everything that can be executed now; anything that can't, move to the
> back — and write it down so nothing is forgotten.** This file is that record. Update it whenever
> something is deferred or unblocked.

## Deferred — can't fully execute autonomously *right now*

| Item | Phase | Why deferred | Unblock / when to do |
|------|-------|--------------|----------------------|
| **Physical-USB boot on real hardware** | P4 | The single-USB image (`make-usb-image.sh`) is built + verified in QEMU (BIOS + UEFI). The final step — `dd` to a real stick and boot a real PC (Secure Boot off) — needs a physical USB + PC, which isn't available in the build environment. | Owner step: `dd` the image, boot a PC with Secure Boot disabled. |
| **A/B live download (`flowork-update run`)** | P4 | The signed-install gate is DONE + verified (`verify-update.sh`: valid update installs+activates, wrong-key refused) and the boot-time A/B switch+rollback is verified. `run` only adds the GitHub-Releases **download** (curl) in front of the same `install`. | Cut a GitHub release with signed `rootfs.squashfs(.sig)` slot assets; test `flowork-update run` downloads → verifies → installs → reboots. |
| **Secure Boot signed shim** | P4 | dm-verity + signed root hash protect the squashfs; a trusted chain from firmware needs an MS-signed shim / key enrollment. | Product stage. |
| **A/B auto-update: atomic switch/reboot/rollback** | P4 | The signed-update **verify** core is DONE (P4a: check + download + signature-verify, anti-MITM). The remaining piece — the **atomic A/B partition switch + reboot + auto-rollback** — needs the disk-image + bootloader (GRUB) path, not the current `-kernel/-initrd` PoC boot. | Wire A/B once the squashfs-on-disk boot path exists (same path P3b/dm-verity needs). |
| **Physical USB flash + boot on real PC** | P4 | No physical USB / spare PC in this environment, and we must not risk the owner's machine. The artifacts (`.iso`, squashfs, kernel+initramfs) are produced and boot in QEMU. | When the owner has a spare stick + test PC: `dd` the ISO, boot with Secure Boot off. |
| **Certified hardware + GPU appliance** | P4 | Requires buying/bundling hardware; a serious local LLM needs a GPU. Not an autonomous task. | Product stage, with the owner. |
| **Real LUKS key custody (TPM / passphrase, stable across updates)** | P3/P4 | P1 uses a per-build random key baked in the image (PoC). Real custody needs TPM sealing or a boot passphrase, and a key that stays stable when the image updates (so DATA isn't lost on update). | P3b/P4 hardening. |
| **Wire the agent's `apps` subsystem to `flowork-app-run`** | — | The OS-level sandbox primitive exists + is proven (P3a). Making the agent launch apps through it is a change in the **agent** repo (kernel is frozen → needs careful unfreeze). | Separate agent-repo task; the OS side is ready. |
| **Buildroot base (reproducible) instead of Alpine** | P4 | Alpine is right for fast PoC iteration; Buildroot (reproducible, RO-by-design) is the locked-product base but slower to set up. | Product stage, once architecture is settled. |

## Waiting on an external clock (not engineering)
- **Facebook posting** — `pages_manage_posts` is in Meta's 24h review. When approved: owner pastes a
  valid `FB_PAGE_TOKEN`, says "aktifkan FB" → set kv `fb_active=on` + test a Page post. (Tracked in
  the agent's social pipeline, not the OS.)

## Done & shipped (for context)
- **P0** — boots to the Flowork kiosk in QEMU (ephemeral). ✅
- **P1** — encrypted DATA volume, state persists across reboot. ✅
- **P3a** — app sandbox (bubblewrap; Python + Go confined, owner state isolated). ✅
- **P4a** — signed-update verification (anti-MITM; valid accepted, tampered/wrong-key rejected). ✅
- **P3b** — dm-verity integrity primitive (intact verified; tampered + wrong-hash rejected). ✅
- **Squashfs-root boot** — read-only squashfs + tmpfs overlay, booted via GRUB; kiosk + sandbox + persistence all verified on it. ✅
- **dm-verity at boot** — system root integrity enforced at boot; a tampered squashfs is detected and the boot is refused. ✅
- **A/B auto-update** — two slots + atomic pointer switch + auto-rollback (bad-image and unconfirmed-boot), all verified. ✅
- **P2 local AI** — Ollama runs a Qwen model on-device, **offline** (glibc sidecar on musl), loopback-only; verified with no NIC. ✅
- **P2b router→ollama** — the appliance router (sovereign seed embedded) routes `/v1/chat/completions` to the local Qwen; verified. ✅
- **Single-USB product image** — one GPT image (GRUB BIOS+UEFI + A/B slots + verity + model + DATA) boots a real PC from one stick; verified in QEMU (both firmwares). ✅
- **Signed-update install gate** — `flowork-update install` accepts a validly-signed update (writes+activates the inactive slot) and refuses a wrong-key one; verified. ✅
