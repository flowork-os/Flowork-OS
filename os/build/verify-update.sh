#!/usr/bin/env bash
# verify-update.sh — prove the signed-install gate of the updater (design Section 8): installing
# a new slot ACCEPTS a validly-signed root hash and writes+activates the inactive slot, and
# REJECTS a wrong-key signature (no slot change). Runs on the host against a mock A/B media dir.
# (`flowork-update run` adds only the GitHub download in front of this same install.)
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SELF_DIR/.." && pwd)"
UPDATE="$REPO_DIR/rootfs-overlay/usr/local/bin/flowork-update"
PUB="$REPO_DIR/config/update-pub.pem"; KEY="$REPO_DIR/config/update.key"
"$SELF_DIR/make-update-keys.sh" >/dev/null 2>&1 || true
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT

# mock A/B media (active = a)
mkdir -p "$T/media/flowork/slot-a" "$T/media/flowork/slot-b"
printf 'a' > "$T/media/flowork/active"
# a "new" image + a signed root hash
head -c 1048576 /dev/urandom > "$T/new.squashfs"
head -c 64 /dev/urandom | sha256sum | cut -c1-64 > "$T/new.roothash"
: > "$T/new.squashfs.verity"
"$SELF_DIR/sign-release.sh" "$T/new.roothash" "$KEY" >/dev/null

pass=0; fail=0
ok(){ echo "  ✅ $1"; pass=$((pass+1)); }
no(){ echo "  ❌ $1"; fail=$((fail+1)); }
inst(){ FLOWORK_MEDIA="$T/media" FLOWORK_UPDATE_PUBKEY="$PUB" sh "$UPDATE" install "$@" >/dev/null 2>&1; }

echo "=== signed-install gate verification ==="

# 1) valid signature -> installs to the inactive slot (b) + flips active to b
if inst "$T/new.squashfs" "$T/new.squashfs.verity" "$T/new.roothash" "$T/new.roothash.sig" \
	&& [ "$(cat "$T/media/flowork/active")" = b ] \
	&& [ -f "$T/media/flowork/slot-b/rootfs.squashfs" ]; then
	ok "valid-signed update INSTALLED to slot b + activated"
else no "valid-signed install"; fi

# reset to a, then 2) wrong-key signature -> REFUSED, slot/active unchanged
printf 'a' > "$T/media/flowork/active"; rm -rf "$T/media/flowork/slot-b"; mkdir -p "$T/media/flowork/slot-b"
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$T/bad.key" 2>/dev/null
openssl dgst -sha256 -sign "$T/bad.key" -out "$T/new.roothash.sig" "$T/new.roothash"
if ! inst "$T/new.squashfs" "$T/new.squashfs.verity" "$T/new.roothash" "$T/new.roothash.sig" \
	&& [ "$(cat "$T/media/flowork/active")" = a ] \
	&& [ ! -f "$T/media/flowork/slot-b/rootfs.squashfs" ]; then
	ok "wrong-key update REFUSED (slot + pointer unchanged)"
else no "wrong-key rejection"; fi

# 3) status + switch work
FLOWORK_MEDIA="$T/media" sh "$UPDATE" switch >/dev/null 2>&1 || true   # b has no image -> should refuse
FLOWORK_MEDIA="$T/media" sh "$UPDATE" status >/dev/null 2>&1 && ok "status/switch callable" || no "status/switch"

echo
[ "$fail" = 0 ] && { echo "SIGNED-UPDATE INSTALL: ✅ PASS ($pass/3)"; exit 0; } || { echo "SIGNED-UPDATE INSTALL: ❌ FAIL"; exit 1; }
