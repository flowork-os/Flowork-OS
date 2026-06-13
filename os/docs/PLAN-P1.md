# Flowork OS — P1 Build Plan (PLAN): Persistence

## What changes vs P0
1. **Rootfs packages:** add `cryptsetup` + `e2fsprogs` (LUKS open + ext4 mkfs/mount in-guest).
2. **Agent seed moves** from `/root/.flowork/agents` to `/usr/share/flowork/agents-seed`
   (a read-only template). The new `flowork-data` service populates the live state dir from it.
   Reason: the live `/root/.flowork` becomes a mount point (DATA or tmpfs); seeding it directly
   would be masked by the mount.
3. **New service `flowork-data`** (boot runlevel, before agent/router):
   ```
   DEV=/dev/vdb (override FLOWORK_DATA_DEV); KEY=/etc/flowork/data.key
   if no block device DEV: copy seed -> /root/.flowork (tmpfs)  # ephemeral, P0 mode unbroken
   else:
     if !isLuks(DEV): luksFormat(DEV, KEY, pbkdf2); open; mkfs.ext4   # first boot
     else:            open(DEV, KEY)                                   # later boots
     mount mapper -> /mnt/flowork-data
     if /mnt/flowork-data/flowork empty: copy seed into it
     bind-mount /mnt/flowork-data/flowork -> /root/.flowork
   ```
   `--pbkdf pbkdf2` (not argon2id) keeps LUKS fast/light inside the VM.
4. **agent/router depend** gains `after flowork-data` so state is mounted before they run.
5. **Keyfile:** `build-flowork-os.sh` writes 4KB of random bytes to the staging rootfs at
   `/etc/flowork/data.key` (per build, **never committed**). PoC-only; custody hardened in P3/P4.
6. **flowork-health** also prints a persistence line: increments `/root/.flowork/.boot-count`
   and reports it + whether `flowork.db` exists.

## New build artifacts / scripts
- `build/make-data-disk.sh OUT SIZE` → `out/flowork-data.img` (blank raw, default 1G).
- `build/run-qemu.sh` → attaches `out/flowork-data.img` as a virtio disk **iff it exists**
  (so the P0 ephemeral run, with no DATA img, is unchanged → P1.5).
- `build/verify-p1.sh`:
  ```
  make a fresh blank DATA disk
  boot #1  -> expect boot_count=1, "initializing LUKS", flowork_db becomes yes
  boot #2 (same disk) -> expect boot_count=2, isLuks=true, flowork_db=yes  => PERSIST OK
  ```

## Verification (headless, deterministic)
The boot-counter living *inside* the encrypted DATA dir is the proof: it can only reach 2 on the
second boot if the encrypted volume was unlocked, mounted, and its contents survived. Read from
the serial log; no display needed. (The kiosk/UI is unchanged from P0.)

## Risk register
- **argon2 OOM/slow in VM** → forced `pbkdf2`. 
- **dm-crypt/ext4 modules** → present in linux-lts; `cryptsetup` modprobes dm-crypt; ext4 is loaded on mount.
- **seed masking** → solved by seeding from `/usr/share/flowork/agents-seed` after the mount.
- **first-boot vs later-boot branch** → keyed on `cryptsetup isLuks`; idempotent.
