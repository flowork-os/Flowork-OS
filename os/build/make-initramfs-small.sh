#!/usr/bin/env bash
# make-initramfs-small.sh — build the small stage-1 initramfs for the squashfs-root boot:
# busybox-static + the kernel modules needed to find/mount the boot media, squashfs and
# overlay, plus our init (initramfs-src/init). Much smaller than packing the whole rootfs.
#   args: ROOTFS_DIR OUT_FILE INIT_SCRIPT
set -euo pipefail
ROOTFS="$1"; OUT="$2"; INIT="$3"
KVER="$(ls "$ROOTFS/lib/modules" | head -1)"
[ -n "$KVER" ] || { echo "no kernel modules in $ROOTFS" >&2; exit 1; }

W="$(mktemp -d)"; trap 'rm -rf "$W"' EXIT
mkdir -p "$W"/{bin,sbin,proc,sys,dev,lower,over,newroot,media,run,lib/modules}

# busybox (prefer the static build so the initramfs needs no libc)
if [ -f "$ROOTFS/bin/busybox.static" ]; then
	cp "$ROOTFS/bin/busybox.static" "$W/bin/busybox"
else
	cp "$ROOTFS/bin/busybox" "$W/bin/busybox"
	# dynamic busybox: bring musl loader + libs
	mkdir -p "$W/lib"
	cp -a "$ROOTFS"/lib/ld-musl-*.so* "$W/lib/" 2>/dev/null || true
	cp -a "$ROOTFS"/lib/libc.musl-*.so* "$W/lib/" 2>/dev/null || true
fi
chmod 0755 "$W/bin/busybox"
# Create RELATIVE applet symlinks (each "applet -> busybox"). NOTE: `busybox --install -s`
# would make ABSOLUTE symlinks pointing at the build-time temp dir, which break at runtime
# (kernel then can't find /bin/sh to run /init -> "Failed to execute /init"). Relative is correct.
for a in $("$W/bin/busybox" --list 2>/dev/null); do
	ln -sf busybox "$W/bin/$a"
done
for a in sh mount umount mkdir modprobe switch_root find sleep mdev insmod uname cat echo ls blkid; do
	[ -e "$W/bin/$a" ] || ln -sf busybox "$W/bin/$a"
done

# --- veritysetup + its shared-library closure (for dm-verity at boot) ----------------
# Copy a dynamic ELF and, recursively, every DT_NEEDED library found in the rootfs lib
# dirs, mirroring paths so the musl loader resolves them. Plus the musl loader itself.
ROOTFS_LIBS="$ROOTFS/lib $ROOTFS/usr/lib $ROOTFS/lib64 $ROOTFS/usr/lib64"
copy_elf() {
	local src="$1"
	[ -f "$src" ] || return 0
	local rel="${src#$ROOTFS}" ; local dst="$W$rel"
	[ -f "$dst" ] && return 0
	mkdir -p "$(dirname "$dst")" ; cp -aL "$src" "$dst"
	local need d
	for need in $(readelf -d "$src" 2>/dev/null | sed -n 's/.*(NEEDED).*\[\(.*\)\]/\1/p'); do
		for d in $ROOTFS_LIBS; do
			[ -e "$d/$need" ] && { copy_elf "$d/$need"; break; }
		done
	done
}
if [ -f "$ROOTFS/sbin/veritysetup" ] && command -v readelf >/dev/null; then
	cp -aL "$ROOTFS"/lib/ld-musl-*.so.* "$W/lib/" 2>/dev/null || true
	copy_elf "$ROOTFS/sbin/veritysetup"
	echo "  + veritysetup + lib closure (dm-verity at boot)"
fi

# kernel modules (whole tree → deps resolve cleanly via modules.dep)
cp -a "$ROOTFS/lib/modules/$KVER" "$W/lib/modules/"

# init
cp "$INIT" "$W/init"; chmod 0755 "$W/init"

( cd "$W" && find . -print0 | cpio --null --create --format=newc --owner=+0:+0 --quiet ) \
	| gzip -9 > "$OUT"
echo "small initramfs -> $OUT ($(du -h "$OUT" | cut -f1))"
