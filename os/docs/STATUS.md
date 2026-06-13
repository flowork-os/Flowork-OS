# Flowork OS — Status & Verification Record

> Pipeline stage: **VERIFY / REVIEW**. This file is the evidence ledger for each phase.
> Update it every time the verification is re-run.

## P0 — PoC (QEMU VM, ephemeral) — ✅ COMPLETE & VERIFIED

Verified on an x86_64 host (Ubuntu, kernel 6.17), QEMU 8.2.2 + KVM, guest = Alpine 3.20 /
Linux 6.6.142-lts. One command builds (`build/build-flowork-os.sh`), one command boots and
checks (`build/run-qemu.sh`).

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| S1 | Builds reproducibly, docker-only (no host apk/root) | ✅ PASS | `build-flowork-os.sh` exits 0; artifacts in `out/` (see manifest). |
| S2 | Agent + router are fully static linux/amd64 | ✅ PASS | `file` → "statically linked" for both (18M / 14M). |
| S3 | Boots to userspace unattended | ✅ PASS | Serial: "Welcome to Alpine Linux 3.20 / Kernel 6.6.142-0-lts". |
| S4 | Agent serves the control panel on :1987 | ✅ PASS | Serial: `AGENT_UP`; the kiosk shows the live control panel. |
| S5 | Router answers on :2402 | ✅ PASS | Serial: `ROUTER_UP` (up ~within seconds of boot). |
| S6 | Chromium kiosk renders the Flowork UI | ✅ PASS | [`p0-kiosk.png`](p0-kiosk.png) = the "Flowork — Register / secure session, local only" panel, full-screen. DRM `/dev/dri/card0` present (virtio_gpu). |
| S7 | Ephemeral; no host-disk writes | ✅ PASS | Booted with `-kernel/-initrd`, no writable disk attached; root is the in-RAM initramfs. |

### Artifacts (`out/`, version 0.1.0-p0)
- `flowork-os-0.1.0-p0.vmlinuz` (11M) — Alpine linux-lts kernel.
- `flowork-os-0.1.0-p0.initramfs.gz` (~448M) — full ephemeral rootfs (the boot artifact).
- `flowork-os-0.1.0-p0.rootfs.squashfs` (~368M) — read-only rootfs (P1 persistence base).
- `flowork-os-0.1.0-p0.iso` (~470M) — GRUB bootable ISO (USB-shaped; boots the same initramfs).
- `flowork-os-0.1.0-p0.manifest.txt` — sizes + sha256.

### Boot path (verified)
`-kernel/-initrd` → busybox init (PID 1) reads `/etc/inittab` → OpenRC `sysinit`/`boot`/`default`
→ `udev` coldplug loads `virtio_gpu` → `seatd` + `dbus` → `flowork-router` (:2402) →
`flowork-agent` (:1987) → `flowork-kiosk` (cage + chromium → the panel) → `flowork-health`
prints `AGENT_UP`/`ROUTER_UP` to the serial console.

### Known trade-offs (acceptable at P0, addressed later)
- Chromium runs as root with `--no-sandbox` → hardened (non-root + sandbox) in **P3**.
- Software rendering (no GPU); UI paints in a few seconds → GPU accel in **P4**.
- Ephemeral only; all state is lost on power-off → persistence (LUKS DATA) in **P1**.
- No local LLM yet; router is up but has no model → Ollama + Qwen in **P2**.

## P1 — Persistence (LUKS DATA) — ✅ COMPLETE & VERIFIED

Verified with `build/verify-p1.sh` (boots the same encrypted DATA disk twice). The boot-counter
lives *inside* `/root/.flowork`, so reaching 2 proves the encrypted volume was unlocked, mounted,
and survived the reboot. See `docs/SPEC-P1.md` / `docs/PLAN-P1.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| P1.1 | Blank DATA disk initialized as LUKS on first boot | ✅ PASS | boot#1 serial: "DATA: initializing LUKS on /dev/vda (first boot)". |
| P1.2 | `/root/.flowork` backed by the encrypted DATA mount before agent/router | ✅ PASS | serial: "DATA: persistent /root/.flowork mounted (encrypted)". |
| P1.3 | State on boot N present on boot N+1 (real reboot) | ✅ PASS | boot#1 `boot_count=1` → boot#2 `boot_count=2` (same disk). |
| P1.4 | `flowork.db` persists across reboot | ✅ PASS | boot#2 serial: `flowork_db=yes`. |
| P1.5 | No DATA disk → still boots (P0 unbroken) | ✅ PASS | no-disk run: "DATA: no DATA disk — running EPHEMERAL"; `AGENT_UP`/`ROUTER_UP` still reached. |

### How it works
`flowork-data` (boot runlevel, before agent/router) polls for the DATA block device, then:
first boot → `cryptsetup luksFormat` (luks2, pbkdf2, keyfile) + `mkfs.ext4`; later boots →
`cryptsetup open`. It mounts the volume, seeds agent definitions from `/usr/share/flowork/agents-seed`,
and bind-mounts it over `/root/.flowork`. With no disk it falls back to tmpfs (ephemeral).

### Known trade-off (recorded honestly)
- **Key custody is PoC-grade.** The LUKS key is a per-build random file baked into the image
  (`/etc/flowork/data.key`, never committed). This proves the *encryption path* and protects DATA
  against someone who has the disk but not the image. It does **not** yet give cross-version
  persistence (a new image = a new key) nor at-rest secrecy against someone who has the image.
  Real custody — passphrase at boot / TPM-sealed key, and a key stable across updates — is **P3/P4**
  (verified boot). P1's claim ("state persists across reboots, on an encrypted volume") is met.

## P3a — App Sandbox (bubblewrap) — ✅ COMPLETE & VERIFIED

Verified with `build/verify-p3a.sh` (boots with `flowork.selftest=1`; runs the Python and Go
sample apps under bwrap and checks confinement). See `docs/SPEC-P3a.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| P3a.1 | `bubblewrap` + `python3` present | ✅ PASS | serial: "bwrap=bubblewrap 0.10.0 \| Python 3.12.13". |
| P3a.2 | Python sample runs sandboxed | ✅ PASS | serial: "py app -> py-hello: ran ok (cwd=/app)". |
| P3a.3 | Go sample (static) runs sandboxed | ✅ PASS | serial: "go app -> go-hello: ran ok (cwd=/app)". |
| P3a.4 | Sandboxed app **cannot read** `/root/.flowork` | ✅ PASS | both apps: "ISOLATION_OK: cannot reach /root/.flowork/flowork.db". |
| P3a.5 | App sees only its own dir | ✅ PASS | host pre-flight + in-guest: app's `/app` = its dir only; no host/owner paths visible. |
| P3a.6 | Self-test gated; normal boots unaffected | ✅ PASS | runs only with kernel cmdline `flowork.selftest=1`. |

### Boot-architecture change (made for P3a — re-verified P0 + P1)
bubblewrap needs `pivot_root`, which **fails on a ramfs/initramfs root**. So a stage-1 `/init`
now moves the system onto a real **tmpfs** root via `switch_root` before OpenRC starts. After
this, `/` is a proper filesystem and the sandbox works. **P0 (kiosk) and P1 (persistence) were
re-verified green** on the new tmpfs root. (The product build will later switch this to a
squashfs + overlay root — same idea, a real fs — see BACKLOG.)

### `flowork-app-run` (the sandbox primitive)
Runs `<cmd>` confined: read-only system, private `/tmp`, **only the app's own dir at `/app`**,
network unshared by default (`FLOWORK_APP_NET=1` to allow). No access to `/root/.flowork`, other
apps, or the host. This is the OS-level capability the agent's app subsystem will call (that
wiring is an agent-repo task — see BACKLOG).

## P4a — Signed-Update Verification (anti-MITM) — ✅ COMPLETE & VERIFIED

The security-critical core of auto-update (design Section 8): the appliance refuses any OS image
that is not signed by the owner's key. Verified with `build/verify-p4a.sh` (host; the crypto is
identical in the appliance). See `docs/SPEC-P4a.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| P4a.1 | Validly-signed artifact ACCEPTED | ✅ PASS | `flowork-update verify` exit 0 on a good signature. |
| P4a.2 | Tampered artifact REJECTED | ✅ PASS | flip one byte → verify exits non-zero. |
| P4a.3 | Wrong-key signature REJECTED | ✅ PASS | verify against another pubkey → non-zero. |
| P4a.4 | Pubkey + version + tooling baked into the image | ✅ PASS | staged rootfs has `/etc/flowork/update-pub.pem`, `/etc/flowork/version`, `/usr/local/bin/flowork-update`, `openssl`. |

EC P-256 + SHA-256 detached signatures. `build/make-update-keys.sh` makes the keypair (private
**gitignored/offline**, public committed+baked); `build/sign-release.sh` signs a release;
`flowork-update {verify|check|run}` runs in the appliance. The atomic **A/B switch + reboot/
rollback** is deferred (needs the disk-image boot path) — `run` stages the verified image.

## P3b — Anti-Tamper (dm-verity) — ✅ PRIMITIVE COMPLETE & VERIFIED

The integrity moat (design Section 7): every block of the read-only root is hashed into a Merkle
tree, so any tamper is detected. Verified with `build/verify-p3b.sh` (host, root). See `docs/SPEC-P3b.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| P3b.1 | Hash tree built; root hash produced | ✅ PASS | `veritysetup format` → 256-bit root hash. |
| P3b.2 | Intact image VERIFIES | ✅ PASS | `veritysetup verify` exit 0. |
| P3b.3 | Tampered image REJECTED | ✅ PASS | flip one byte → verify non-zero ("Verification failed at position ..."). |
| P3b.4 | Wrong root hash REJECTED | ✅ PASS | bogus hash → non-zero. |

`build/make-verity.sh <squashfs> [--sign]` builds the hash tree + root hash, and `--sign` signs the
root hash → the full chain: **signed root hash (P4a) → dm-verity-verified squashfs (P3b)**. Wiring
this into BOOT (refuse to mount a tampered root) needs the squashfs-on-disk boot path — deferred.

## Squashfs-Root Boot — ✅ COMPLETE & VERIFIED (the product/USB boot shape)

The boot path now matches the real USB product: a read-only **squashfs root + tmpfs overlay**,
booted from the media via GRUB — instead of unpacking the whole rootfs into RAM. Verified with
`build/verify-squashroot.sh` (QEMU `-cdrom`). See `docs/SPEC-SQUASHROOT.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| SR.1 | ISO boots via GRUB; stage-1 init runs | ✅ PASS | serial: "FLOWORK stage-1 init (squashfs-root)". |
| SR.2 | rootfs.squashfs found + mounted as overlay root | ✅ PASS | serial: "boot media: /dev/sr0 (iso9660)" + "overlay root ready (squashfs lower + tmpfs upper)". |
| SR.3 | OpenRC + services up on the overlay root | ✅ PASS | serial: `ROUTER_UP` + `AGENT_UP` (booted in ~36s). |
| SR.4 | Functionally equal to the PoC | ✅ PASS | kiosk renders (`out/screen-squash.png`); **sandbox** self-test OK on the overlay root; **persistence** (LUKS DATA on `/dev/vda`) boot_count 1→2 across reboot. |

Small stage-1 initramfs (84M: busybox-static + kernel modules + `initramfs-src/init`) loads the
SATA/AHCI + squashfs + overlay modules, finds the media holding `flowork/rootfs.squashfs`, builds
the overlay, and `switch_root`s in. Artifact: `flowork-os-<ver>.squashroot.iso`.

## dm-verity AT BOOT — ✅ COMPLETE & VERIFIED (the moat, now enforced)

The squashfs system root is now integrity-protected **at boot**: stage-1 `/init` sets up a
dm-verity device over the squashfs (hash tree + root hash carried on the media) and mounts THAT.
A tampered system root is detected by the kernel and the boot is **refused**. Verified with
`build/verify-verity.sh` (both a legit boot and a tampered boot). See `docs/SPEC-P3b.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| DV.1 | Legit image boots with integrity enforced | ✅ PASS | serial: "dm-verity: integrity device active (root hash enforced on every read)" + `AGENT_UP`/`ROUTER_UP`. |
| DV.2 | Tampered system root is DETECTED and boot REFUSED | ✅ PASS | flip 1 byte in the squashfs (keep the hash tree) → kernel: "device-mapper: verity: ... data block 0 is corrupted" → "ROOT INTEGRITY CHECK FAILED — refusing to boot" → no `AGENT_UP`. |
| DV.3 | Build produces signed root hash (chain with P4a) | ✅ PASS | `build-flowork-os.sh` runs `make-verity.sh --sign` → `rootfs.squashfs.{verity,roothash,roothash.sig}` on the media. |
| DV.4 | Graceful: no hash sidecars → still boots (dev/test) | ✅ PASS | `/init` falls back to a plain squashfs mount when the sidecars are absent. |

The full chain of trust now exists: **signed root hash (P4a) → dm-verity-verified squashfs root
(P3b) → bubblewrap-sandboxed apps (P3a) → encrypted DATA (P1)**. The small initramfs carries
`veritysetup` + its musl library closure; the squashfs `.verity`/`.roothash` sidecars ride on the
boot media next to the squashfs.

## A/B Auto-Update — ✅ COMPLETE & VERIFIED (atomic switch + auto-rollback)

Two OS **slots** (A/B) + an `active` pointer on writable media. The appliance boots the active
slot (dm-verity-verified); an update writes the new image to the **inactive** slot and flips the
pointer atomically; a bad or unconfirmed slot **auto-rolls-back**. Verified with `build/verify-ab.sh`
(8/8). See `docs/SPEC-AB.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| AB.1 | Boots the active slot (verified) | ✅ PASS | serial: "A/B mode: active slot = a" + `AGENT_UP`. |
| AB.2 | A pointer switch boots the other slot | ✅ PASS | flip active→b → "active slot = b" + `AGENT_UP`. |
| AB.3 | A healthy boot is confirmed | ✅ PASS | serial: "FLOWORK_AB slot=a confirmed"; `slot-a.confirmed` written. |
| AB.4 | A bad slot (verity fail) auto-rolls-back | ✅ PASS | corrupt slot-b → "rolling back b -> a" → pointer reverts → slot a boots. |
| AB.5 | An unconfirmed slot rolls back after N tries | ✅ PASS | preset slot-b.tries=3 → boot → tries>3 → revert to a. |
| AB.6 | **Signed install** accepts a valid update, rejects a wrong-key one | ✅ PASS | `build/verify-update.sh`: `flowork-update install` writes+activates the inactive slot on a valid signature; a wrong-key signature is refused (slot + pointer unchanged). |

Stage-1 `/init` reads `flowork/active`, runs the rollback guard (tries counter), verity-verifies +
mounts the active slot, and keeps the writable media at `/flowork-media`. `flowork-health` writes
`slot-<x>.confirmed` on a healthy boot (clearing the counter). `flowork-update {install|switch|status}`
writes the inactive slot (P4a-verifying a signed root hash) and flips the pointer. Atomicity = the
single pointer write; the running slot is never overwritten.

## P2 — Local AI (Ollama + Qwen) — ✅ COMPLETE & VERIFIED (the sovereign brain, offline)

The appliance runs an LLM **on-device, offline**. Ollama (`127.0.0.1:11434`) serves a local Qwen
model; the model rides on a writable disk (label `FLOWORKMODELS`), the runtime in the squashfs.
Verified with `build/verify-ollama.sh` — booted with **no NIC**, so a completion can only come from
the on-device model. See `docs/SPEC-P2.md`.

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| P2.1 | Ollama serves on `:11434` | ✅ PASS | serial: "OLLAMA serve starting on 127.0.0.1:11434". |
| P2.2 | Local model present + listed | ✅ PASS | serial: "first model = qwen2.5:0.5b". |
| P2.3 | **Offline** inference returns a completion | ✅ PASS | VM has no network; `/api/generate` → "inference response = [FLOWORK OK]" → "OLLAMA_AI RESULT: OK". |
| P2.4 | Loopback-only (sovereign) | ✅ PASS | Ollama binds `127.0.0.1` only. |
| P2.5 | Coexists with verified boot / sandbox / persistence | ✅ PASS | same image still boots the dm-verity squashroot. |
| **P2b** | **Router → Ollama routing** | ✅ PASS | a `/v1/chat/completions` to the router (`:2402`) returns a completion from local Qwen → "ROUTER_AI RESULT: OK". The appliance router embeds the **sovereign seed** (local-llama active, `qwen*` matched). |

**The musl fix:** Ollama is a glibc/cgo binary; the appliance is musl. A **glibc sidecar** (loader +
libs at the standard multiarch paths, ~42 MB) lets the glibc ollama + its `llama-server` runner run
alongside musl. GPU libs are stripped → the ollama runtime is ~27 MB. Model store: `make-model-disk.sh`.

## Single-USB Image — ✅ COMPLETE & VERIFIED (the real flashable product)

One disk image (`make-usb-image.sh`) holds **everything**: a GPT layout with a BIOS-boot
partition + an **ESP** (GRUB EFI + kernel + initramfs) + an **ext4 FLOWORKAB** partition (the A/B
slots + verity sidecars + the local-AI model) + a **blank DATA** partition (LUKS-formatted on
first boot). GRUB is installed for **BIOS and UEFI**. `dd` it to a stick → it boots a real PC.
Verified with `build/verify-usb.sh` (QEMU, both firmwares).

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| USB.1 | Boots via **GRUB (BIOS)** into the A/B system | ✅ PASS | serial: GRUB → "boot media: /dev/vda3 (ext4)" → "A/B mode: active slot = a" → `AGENT_UP`. |
| USB.2 | Boots via **GRUB (UEFI/OVMF)** too | ✅ PASS | same path under OVMF firmware → `AGENT_UP`. |
| USB.3 | dm-verity still enforced from the stick | ✅ PASS | serial: "dm-verity: integrity device active" on the USB slot. |
| USB.4 | One image carries A/B + model + DATA | ✅ PASS | GPT: biosboot + ESP + FLOWORKAB(slots+model) + DATA(blank); model embedded. |

`flowork-data` is now **safe on a single disk** — it only ever LUKS-formats a device that is
explicitly `FLOWORKDATA`, an existing LUKS, or genuinely blank (never the boot disk / ESP / FLOWORKAB).

## Deferred / remaining — see [`BACKLOG.md`](BACKLOG.md)
Recorded so nothing is forgotten (owner's principle: do what's executable now, log the rest).
- **Physical-USB boot on real hardware** — `dd` the image + boot a real PC (Secure Boot off). The
  image is verified in QEMU (BIOS+UEFI); the final hardware boot is the owner's step (no physical
  stick here).
- **Live download wiring** — `flowork-update run` end-to-end against a real two-slot GitHub release.
- **Real key custody** (TPM/passphrase), Secure Boot signed shim, certified hardware + GPU —
  product stage (hardware needed).
