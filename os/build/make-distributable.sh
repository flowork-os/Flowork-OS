#!/usr/bin/env bash
# make-distributable.sh — build a Flowork OS USB image in one of TWO data profiles:
#
#   public   →  for sharing. NO user database/settings ship; every secret in the baked
#               config becomes the placeholder `change_this_token`. The recipient flashes it,
#               boots, and fills in their OWN tokens. Safe to hand out / put on a download page.
#
#   dev      →  for the OWNER. The full live state (flowork.db + config + real tokens) is baked
#               in, so the stick boots ready-to-use. CONTAINS SECRETS — never distribute it.
#
# Wraps build-flowork-os.sh + make-usb-image.sh; the only difference between modes is which
# seed/state goes in. Needs whatever those scripts need (docker, and root for the USB image).
#
#   usage: make-distributable.sh <public|dev> [SIZE]
#   env:   FLOWORK_HOME (dev source of live state, default ~/.flowork)
set -euo pipefail
MODE="${1:-}"; SIZE="${2:-6G}"
case "$MODE" in public|dev|simple) ;; *) echo "usage: $0 <public|dev|simple> [SIZE]" >&2; exit 2;; esac

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SELF_DIR/.." && pwd)"
OUT="$REPO/out"; mkdir -p "$OUT"
TAG="flowork-os-$(cat "$REPO/VERSION" 2>/dev/null || echo 0.0.0)"
SOV_REAL="${FLOWORK_SOVEREIGN_SEED:-$REPO/../seed-sovereign/appliance/router-config.sovereign.seed.json}"

echo "==> Flowork distributable: MODE=$MODE"

if [ "$MODE" = public ]; then
	# 1) sanitized seed — every secret -> change_this_token.
	PUB="$OUT/public-seed"; mkdir -p "$PUB"
	if [ -f "$SOV_REAL" ]; then
		python3 "$SELF_DIR/sanitize-public.py" "$SOV_REAL" "$PUB/router-config.sovereign.seed.json"
		export FLOWORK_SOVEREIGN_SEED="$PUB/router-config.sovereign.seed.json"
	fi
	# 2) no owner state → clean DB on first boot.
	unset FLOWORK_STATE_SEED || true
	# 3) SLIM: no bundled model (Claude cloak default; Ollama pulls on demand) + small image
	#    (DATA auto-grows to fill the real USB on first boot). Tiny download = low friction.
	export FLOWORK_NO_MODEL=1
	export FLOWORK_AB_SIZE="${FLOWORK_AB_SIZE:-1600M}"
	USB_SIZE="${PUBLIC_SIZE:-3000M}"   # +launcher partition (~320M) over the bare 2.6G
	echo "    public: secrets -> change_this_token, NO model, slim image ($USB_SIZE), DATA auto-grows on boot"
elif [ "$MODE" = simple ]; then
	# simple: bake the owner's SETTINGS/credentials (flowork.db) so the stick boots configured,
	# but NO local LLM/model — it relies on the cloud router (ideal for a slow PC). Slim image.
	export FLOWORK_SOVEREIGN_SEED="$SOV_REAL"
	export FLOWORK_STATE_SEED="${FLOWORK_STATE_SEED:-${FLOWORK_HOME:-$HOME/.flowork}}"
	export FLOWORK_NO_MODEL=1
	export FLOWORK_AB_SIZE="${FLOWORK_AB_SIZE:-1600M}"
	USB_SIZE="${SIMPLE_SIZE:-3000M}"
	[ -f "$FLOWORK_STATE_SEED/flowork.db" ] \
		&& echo "    simple: settings/tokens baked from $FLOWORK_STATE_SEED, NO local model (cloud router), slim ($USB_SIZE)" \
		|| echo "    simple: no flowork.db at $FLOWORK_STATE_SEED — boots to a clean DB"
else
	# dev: real seed + bake the owner's live state.
	export FLOWORK_SOVEREIGN_SEED="$SOV_REAL"
	# Data source for the FULL profile: explicit FLOWORK_STATE_SEED (e.g. from the GUI) wins,
	# else FLOWORK_HOME, else the live ~/.flowork (where flowork.db = settings+tokens lives).
	export FLOWORK_STATE_SEED="${FLOWORK_STATE_SEED:-${FLOWORK_HOME:-$HOME/.flowork}}"
	[ -f "$FLOWORK_STATE_SEED/flowork.db" ] \
		&& echo "    dev: baking owner state from $FLOWORK_STATE_SEED (CONTAINS SECRETS — keep private)" \
		|| echo "    dev: no flowork.db at $FLOWORK_STATE_SEED — image will boot to a clean DB"
	USB_SIZE="$SIZE"   # dev keeps the model + a roomy image
fi

# 3) build rootfs/binaries/squashfs, then assemble the USB image.
echo "==> build-flowork-os.sh"
"$SELF_DIR/build-flowork-os.sh"
echo "==> make-usb-image.sh (${USB_SIZE:-$SIZE})"
"$SELF_DIR/make-usb-image.sh" "$OUT" "$TAG" "${MODEL:-qwen2.5:0.5b}" "${USB_SIZE:-$SIZE}"

# 4) name the artifact by mode so public/dev images never get mixed up.
IMG="$OUT/$TAG.usb.img"; FINAL="$OUT/$TAG-$MODE.usb.img"
[ -f "$IMG" ] && mv -f "$IMG" "$FINAL" && echo "==> DONE: $FINAL"
if [ "$MODE" = dev ]; then
	echo "!!  DEV image contains your real tokens/DB — do NOT share it."
else
	echo "==> PUBLIC image is safe to share. Recipients replace change_this_token with their own keys."
fi
