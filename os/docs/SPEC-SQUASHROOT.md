# Flowork OS — Squashfs-Root Boot (DEFINE)

> The boot-path change that unblocks the rest of hardening. The PoC boots an in-RAM/tmpfs root;
> this switches to a **read-only squashfs root + tmpfs overlay**, booted from the media (CD/USB)
> via GRUB. That real, immutable system layer is exactly what **dm-verity (P3b)** and
> **A/B auto-update (P4)** need to attach to. Pipeline stage: **DEFINE**.

## Goal
Boot the same Flowork OS, but with the system root coming from `rootfs.squashfs` on the boot
media (read-only), unioned with a tmpfs upper via overlayfs — instead of unpacking the whole
rootfs into RAM. Functionally identical (kiosk + agent + router + sandbox), product-shaped boot.

## Why
- **Immutable system layer** → dm-verity can hash/verify it; A/B update can swap it atomically.
- **Lower RAM** → only the squashfs (paged from media) + tmpfs writes, not the whole rootfs in RAM.
- **USB shape** → a hybrid ISO that `dd`s to a stick and boots a real PC.

## Architecture
```
GRUB (on the media) → kernel + small stage-1 initramfs
  initramfs /init:  modprobe {sr_mod,isofs,loop,squashfs,overlay,virtio_blk,...}
                    find the media holding flowork/rootfs.squashfs
                    mount squashfs (ro) = lower ; tmpfs = upper
                    mount -t overlay → /newroot ; switch_root /newroot /sbin/init
  → OpenRC boots exactly as before (router, agent, kiosk, data, health)
```
The small initramfs = busybox-static + the kernel module tree + `initramfs-src/init`. The full
rootfs is NOT in the initramfs anymore — it lives in the squashfs on the media.

## Success criteria (binary — `build/verify-squashroot.sh`, QEMU `-cdrom`)
| # | Criterion | How verified |
|---|-----------|--------------|
| SR.1 | The ISO boots via GRUB; stage-1 init runs | serial: "FLOWORK stage-1 init (squashfs-root)". |
| SR.2 | rootfs.squashfs is found + mounted as an overlay root | serial: "boot media: ..." + "overlay root ready". |
| SR.3 | OpenRC + services come up on the overlay root | serial: `AGENT_UP` + `ROUTER_UP`. |
| SR.4 | Functionally equal to the PoC (kiosk renders; sandbox + persistence still work) | screenshot; re-run verify-p3a / verify-p1 against this boot path. |

## Non-goals (next, now unblocked)
- Wire **dm-verity** at mount (refuse a tampered squashfs) — uses the P3b primitive on this lower.
- Wire **A/B** (two squashfs slots + switch/rollback) — uses the P4a verify on the downloaded slot.
- Persistent DATA on the SAME stick as a real partition (P1 currently uses a separate virtio disk).
