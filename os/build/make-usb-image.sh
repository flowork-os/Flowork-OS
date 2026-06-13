#!/usr/bin/env bash
# === LOCKED FILE === STABLE — DO NOT MODIFY without owner approval (Aola Sahidin / Mr.Dev).
# Locked 2026-06-13 after the adaptive FAT sizing (fits the data-seed) was verified.
# 2026-06-13 (audit→fix): FAT slack raised 96M→256M so an in-place update (rewrites the ~170M state.db
#   before deleting the old copy) no longer overflows the partition.
# make-usb-image.sh — build ONE flashable Flowork OS disk image. DUAL-PURPOSE stick:
#   • BOOT it  → the Flowork OS appliance (GRUB BIOS+UEFI → A/B verity squashfs → LUKS DATA).
#   • PLUG it into a running Windows/Linux/macOS → a FAT "FLOWORK" partition (p1, visible
#     everywhere) carries the portable launchers (GUI / Background / Stop) — run without rebooting.
# Layout (GPT):  p1 FAT FLOWORK(launchers) · p2 bios-boot · p3 ESP · p4 FLOWORKAB(A/B+model) ·
#                p5 DATA(blank LUKS, auto-grows to fill the stick on first boot).
# Needs root (losetup/partition/grub/mount).
#   args: OUT_DIR TAG [MODEL] [SIZE]
set -euo pipefail
OUT="$1"; TAG="$2"; MODEL="${3:-qwen2.5:0.5b}"; SIZE="${4:-6G}"
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Slim/public knobs: FLOWORK_NO_MODEL=1 ships no bundled model (default LLM = Claude cloak;
# Ollama pulls a model on demand). FLOWORK_AB_SIZE sizes the A/B partition (2 slots + any model).
AB_SIZE="${FLOWORK_AB_SIZE:-3G}"
LAUNCH_SIZE="${FLOWORK_LAUNCHER_SIZE:-320M}"
SQ="$OUT/$TAG.rootfs.squashfs"; KERNEL="$OUT/$TAG.vmlinuz"; INITRD="$OUT/$TAG.initramfs-small.gz"
for f in "$SQ" "$KERNEL" "$INITRD"; do [ -f "$f" ] || { echo "missing $f — run build-flowork-os.sh first" >&2; exit 1; }; done
for t in sgdisk losetup mkfs.vfat mkfs.ext4 grub-install; do command -v "$t" >/dev/null || { echo "$t required" >&2; exit 1; }; done
SUDO=""; [ "$(id -u)" -ne 0 ] && SUDO="sudo"

# Portable launchers for p1 — build them if not already present.
# Portable launchers for p1. ALWAYS (re)build fresh unless an explicit FLOWORK_PORTABLE_DIR was
# passed — reusing a stale out/flowork-portable cache once shipped a USB with old launchers
# (missing the .desktop files). Cheap (~30s) vs the whole image build.
PORTABLE="${FLOWORK_PORTABLE_DIR:-}"
if [ -z "$PORTABLE" ]; then
	PORTABLE="$OUT/flowork-portable"
	echo "[usb] (re)building portable launchers fresh (Win/Linux/macOS · GUI/Background/Stop/Update + .desktop)…"
	bash "$SELF_DIR/../portable/make-portable.sh" "$PORTABLE" >/dev/null
fi

# Size the FLOWORK FAT partition to fit the portable bundle. A Full/dev build now carries a
# data-seed (owner DB + agents, ~130M+) so 320M no longer fits — grow to the bundle size + slack.
# Public builds (no data-seed, ~125M) stay at the 320M floor. An explicit FLOWORK_LAUNCHER_SIZE wins.
# Slack is 256M (not 96M): an IN-PLACE update (update-usb / rsync onto a live stick) rewrites the
# largest single file — the ~170M+ mr-flow state.db — BEFORE deleting the old copy, so it transiently
# needs old+newfile headroom. 96M was enough for a fresh mkfs+copy but overflowed in-place updates.
if [ -z "${FLOWORK_LAUNCHER_SIZE:-}" ] && [ -d "$PORTABLE" ]; then
	pmb=$(du -sm "$PORTABLE" 2>/dev/null | cut -f1)
	need=$(( pmb + 256 ))
	[ "$need" -lt 320 ] && need=320
	LAUNCH_SIZE="${need}M"
	echo "[usb] FLOWORK partition sized to $LAUNCH_SIZE (portable bundle ${pmb}M)"
fi

IMG="$OUT/$TAG.usb.img"
echo "[usb] $SIZE GPT image: FLOWORK(launchers) + bios-boot + ESP + FLOWORKAB + DATA(blank LUKS)"
rm -f "$IMG"; truncate -s "$SIZE" "$IMG"
sgdisk -Z "$IMG" >/dev/null
sgdisk -n1:0:+"$LAUNCH_SIZE" -t1:0700 -c1:FLOWORK     "$IMG" >/dev/null   # FAT, visible on Win/Mac
sgdisk -n2:0:+2M             -t2:ef02 -c2:biosboot    "$IMG" >/dev/null
sgdisk -n3:0:+256M           -t3:ef00 -c3:esp         "$IMG" >/dev/null
sgdisk -n4:0:+"$AB_SIZE"     -t4:8300 -c4:floworkab   "$IMG" >/dev/null
sgdisk -n5:0:0               -t5:8300 -c5:floworkdata "$IMG" >/dev/null

LOOP="$($SUDO losetup -fP --show "$IMG")"
LN="$(mktemp -d)"; M="$(mktemp -d)"; ESP="$(mktemp -d)"
cleanup(){ for d in "$LN" "$M" "$ESP"; do $SUDO umount "$d" 2>/dev/null||true; done; $SUDO losetup -d "$LOOP" 2>/dev/null||true; }
trap cleanup EXIT

$SUDO mkfs.vfat -F 32 -n FLOWORK "${LOOP}p1" >/dev/null
$SUDO mkfs.vfat -n FLOWORKESP "${LOOP}p3" >/dev/null
$SUDO mkfs.ext4 -q -L FLOWORKAB "${LOOP}p4"
# p5 (DATA) left blank — flowork-data LUKS-formats + auto-grows it on first boot.

# p1: portable launchers (run Flowork on a running OS, no reboot).
$SUDO mount "${LOOP}p1" "$LN"
$SUDO cp -r "$PORTABLE/." "$LN/"
$SUDO rm -rf "$LN/flowork-data"   # runtime dir — the launcher recreates it fresh on the stick
# README on the FAT partition: Windows/macOS can read THIS partition but NOT the
# Linux/LUKS ones — so the OS pops a scary "format this disk?" prompt for those.
# Explain it here so a user never formats their bootable stick.
$SUDO tee "$LN/README.txt" >/dev/null <<'RDM'
================  FLOWORK OS — bootable USB  ================

⚠️  DO NOT FORMAT / ERASE THIS DRIVE.

If Windows or macOS asks "You need to format the disk before you can use it"
— click CANCEL. That prompt is NORMAL: this USB also holds Linux + encrypted
partitions that Windows/macOS simply can't read. Formatting would DESTROY the
Flowork OS on the stick. (On Linux everything is readable, so no prompt.)

HOW TO USE FLOWORK
------------------
1) Run it on your CURRENT OS, no reboot — open this FLOWORK drive and launch:
     Windows : Start-Flowork.bat
     macOS   : Start-Flowork.command
     Linux   : ./start.sh   (or the .desktop launcher)
   Then open  http://127.0.0.1:1987  in your browser.

2) Boot the WHOLE PC into Flowork OS — restart, open the boot menu (F12/F2/Esc,
   varies by PC), pick this USB, and Flowork OS starts on its own (Secure Boot
   off). Your data partition is encrypted; nothing is written to the host PC.

Open source:  https://github.com/flowork-os/Flowork-OS
===========================================================
RDM
echo "[usb] launchers + README.txt (anti-format) on FLOWORK partition (Win·Linux·macOS)"

$SUDO mount "${LOOP}p3" "$ESP"
$SUDO mount "${LOOP}p4" "$M"

# Kernel + initramfs + grub live on the ESP (FAT, which GRUB reads natively for BIOS+UEFI).
$SUDO mkdir -p "$ESP/boot/grub"
$SUDO cp "$KERNEL" "$ESP/boot/vmlinuz"
$SUDO cp "$INITRD" "$ESP/boot/initramfs.gz"
$SUDO tee "$ESP/boot/grub/grub.cfg" >/dev/null <<'CFG'
set timeout=3
set default=0
menuentry "Flowork OS" {
    linux  /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 net.ifnames=0
    initrd /boot/initramfs.gz
}
menuentry "Flowork OS (safe graphics / nomodeset)" {
    linux  /boot/vmlinuz console=tty0 console=ttyS0,115200 loglevel=4 nomodeset net.ifnames=0
    initrd /boot/initramfs.gz
}
CFG

# A/B slots + verity sidecars + model on the ext4 data partition (FLOWORKAB).
$SUDO mkdir -p "$M/flowork/slot-a" "$M/flowork/slot-b" "$M/flowork/ollama-models"
for s in a b; do
	$SUDO cp "$SQ" "$M/flowork/slot-$s/rootfs.squashfs"
	for e in verity roothash roothash.sig; do [ -f "$SQ.$e" ] && $SUDO cp "$SQ.$e" "$M/flowork/slot-$s/rootfs.squashfs.$e"; done
done
printf 'a' | $SUDO tee "$M/flowork/active" >/dev/null

# Embed the local-AI model store if it's on the build host — UNLESS this is a slim/public build
# (FLOWORK_NO_MODEL=1): default LLM is Claude cloak, Ollama pulls a model on demand instead.
SRC=/usr/share/ollama/.ollama/models
name="${MODEL%%:*}"; mtag="${MODEL#*:}"; [ "$mtag" = "$MODEL" ] && mtag=latest
MF="$SRC/manifests/registry.ollama.ai/library/$name/$mtag"
if [ -n "${FLOWORK_NO_MODEL:-}" ]; then
	echo "[usb] slim build: no bundled model (Claude cloak default; Ollama pulls on demand)"
elif $SUDO test -f "$MF"; then
	MD="$M/flowork/ollama-models/models"
	$SUDO mkdir -p "$MD/manifests/registry.ollama.ai/library/$name" "$MD/blobs"
	$SUDO cp "$MF" "$MD/manifests/registry.ollama.ai/library/$name/$mtag"
	for d in $($SUDO grep -ohE 'sha256:[a-f0-9]+' "$MF" | sort -u); do
		$SUDO cp "$SRC/blobs/sha256-${d#sha256:}" "$MD/blobs/sha256-${d#sha256:}"
	done
	echo "[usb] embedded model: $MODEL"
else
	echo "[usb] model $MODEL not on host — USB ships without a model"
fi

# GRUB: BIOS (core embedded into the bios-boot partition) + UEFI (removable /EFI/BOOT/BOOTX64.EFI).
GMODS="part_gpt ext2 fat search normal linux echo all_video gzio"
$SUDO grub-install --target=i386-pc --boot-directory="$ESP/boot" --modules="$GMODS" "$LOOP" >/dev/null 2>&1 \
	|| echo "[usb] WARN: BIOS grub-install failed"
$SUDO grub-install --target=x86_64-efi --efi-directory="$ESP" --boot-directory="$ESP/boot" \
	--removable --no-nvram --modules="$GMODS" >/dev/null 2>&1 \
	|| echo "[usb] WARN: UEFI grub-install failed"

sync
echo "[usb] DONE -> $IMG"
echo "[usb] flash:  sudo dd if=$IMG of=/dev/sdX bs=4M status=progress conv=fsync   (Secure Boot off)"
