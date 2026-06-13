# Flowork OS — P0 Build Plan (PLAN)

> Pipeline stage: **PLAN** (follows DEFINE in `SPEC-P0.md`).
> This is the concrete recipe the build script implements. Each step says *what* and *why*.

## Repo split (carried from design §3)
- **flowork-ai-agent_os_version** (this repo) = the **appliance/OS builder**. It assembles the
  bootable image: Linux + Flowork agent + router + chromium kiosk. The OS releases ship from here.
- **Flowork-Router_personal** = the **sovereign router** (local-first, Ollama-default config +
  appliance service unit). Built as a separate static binary, co-located at runtime (`:2402`).
- Canonical *application* source stays in the main repos (`Flowork_Agent`, `flowork_Router`).
  The OS builder compiles binaries from those sources; it does not fork the full app source.

## Source resolution (so a fresh clone still builds)
`build-flowork-os.sh` resolves each source tree in order:
1. Env override (`AGENT_SRC`, `ROUTER_SRC`).
2. Sibling checkout (`../Flowork_Agent`, `../Flowork-Router_personal`, `../flowork_Router`).
3. `git clone` from GitHub into a build cache (`.cache/`).
First match wins. This keeps the appliance buildable from a bare clone while reusing local trees.

## Build pipeline (build-flowork-os.sh)
```
STEP 1  BINARIES (host, native linux/amd64 — no cross-compile needed)
        CGO_ENABLED=0 go build  ->  flowork-agent (static), flowork-router (static)
        WHY static: musl-based Alpine has no glibc; a static Go binary runs anywhere.
        WHY CGO off works: deps are pure-Go (modernc.org/sqlite). Verified.

STEP 2  ROOTFS (docker, no host root/apk)
        docker build FROM alpine:3.20:
          apk add: openrc, busybox, kernel (linux-lts), firmware, mesa-dri-gallium,
                   cage, chromium, seatd, eudev, dbus, font, ca-certificates, curl, util-linux
        WHY docker: build the Alpine userland with apk without needing apk/root on the host.
        WHY linux-lts (not linux-virt): we need the virtio-gpu DRM driver for the kiosk; the
        minimal "virt" kernel often omits DRM. lts is bigger but renders.
        Export the container filesystem to a staging dir.

STEP 3  OVERLAY
        Copy rootfs-overlay/* over the staging rootfs:
          /etc/init.d/flowork-agent, flowork-router, flowork-kiosk   (OpenRC services)
          /etc/flowork/flowork.env                                   (ports, FLOWORK_HOME)
          /usr/local/bin/flowork-kiosk-launch, flowork-health        (helpers)
        Copy the two static binaries to /usr/local/bin/.
        Pre-seed agent definitions: source agents/*.fwagent -> /root/.flowork/agents/ (read model).
        Enable services in the boot + default runlevels.

STEP 4  KERNEL + INITRAMFS (extract from the rootfs)
        Pull /boot/vmlinuz-lts and the kernel modules out of the staging rootfs.
        EPHEMERAL boot model for P0: pack the ENTIRE rootfs as a single initramfs (cpio.gz).
          WHY: kernel unpacks it into a tmpfs root -> runs OpenRC -> services start. No disk
          mount, no iso9660/squashfs/overlay module juggling, no bootloader. Maximally robust
          and exactly matches P0 "ephemeral (RAM, zero trace)". Direct `-kernel/-initrd` boot.
        ALSO produce the "closer-to-USB" artifact: mksquashfs the rootfs + grub-mkrescue ISO
          (kernel + a small initramfs that mounts the squashfs read-only with a tmpfs overlay).
          This is the shape a real USB will take in P1; built and boot-attempted, not the gate.

STEP 5  ARTIFACTS  ->  out/
          flowork-os-<ver>.vmlinuz        (kernel)
          flowork-os-<ver>.initramfs.gz   (full ephemeral rootfs — primary boot artifact)
          flowork-os-<ver>.iso            (grub + squashfs — USB-shaped artifact)
          flowork-os-<ver>.manifest.txt   (sizes, sha256, build inputs)
```

## Boot + kiosk bring-up (inside the guest)
```
kernel -> /sbin/init (OpenRC) ->
  sysinit/boot runlevel:  mount proc/sys/dev, udev (eudev) coldplug, load virtio_gpu,
                          seatd, dbus, set XDG_RUNTIME_DIR
  default runlevel:       flowork-router  (start :2402)
                          flowork-agent   (start :1987, depends on router being up-ish)
                          flowork-kiosk   (cage -- chromium --kiosk http://localhost:1987)
                          flowork-health  (curl :1987 / :2402, print AGENT_UP/ROUTER_UP to console)
```

## Verification (run-qemu.sh)
- Boot `-kernel vmlinuz -initrd initramfs.gz` with: `-serial file:serial.log`,
  `-device virtio-vga`, `-display none`, QMP monitor socket, no writable disk (S7), 4G RAM.
- Poll `serial.log` for `AGENT_UP` and `ROUTER_UP` (S3/S4/S5).
- Via QMP, `screendump out/screen.ppm` -> convert to PNG -> visual check of the kiosk (S6).
- Record pass/fail per criterion in `docs/STATUS.md`.

## Risk register
- **virtio-gpu DRM in guest** (S6): if the kiosk can't get a DRM node, S1–S5 still pass; record
  S6 blocker honestly. Mitigation: linux-lts kernel + explicit `modprobe virtio_gpu` + `-device virtio-vga`.
- **cage needs a seat**: start `seatd` + set `XDG_RUNTIME_DIR=/run/user/0` before cage.
- **initramfs size** (full rootfs incl. chromium ~300–600MB): give the VM 4GB RAM. Fine for PoC.
- **agent binds 127.0.0.1**: kiosk is same-host, so localhost works; no exposure off-box (good).
