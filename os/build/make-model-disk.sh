#!/usr/bin/env bash
# make-model-disk.sh — build a writable ext4 model-store disk (label FLOWORKMODELS) holding
# one Ollama model, for OFFLINE local inference on the appliance. Extracts the model
# (manifest + referenced blobs) from a host Ollama store. mkfs needs no root (mkfs.ext4 -d);
# reading the system store (/usr/share/ollama) may need sudo.
#   args: OUT_DIR TAG [MODEL] [SRC_STORE] [SIZE]
set -euo pipefail
OUT="$1"; TAG="$2"; MODEL="${3:-qwen2.5:0.5b}"
SRC="${4:-/usr/share/ollama/.ollama/models}"; SIZE="${5:-1G}"
command -v mkfs.ext4 >/dev/null || { echo "mkfs.ext4 required" >&2; exit 1; }

# model "name:tag" -> manifest path registry.ollama.ai/library/<name>/<tag>
name="${MODEL%%:*}"; tag="${MODEL#*:}"; [ "$tag" = "$MODEL" ] && tag=latest
MFREL="registry.ollama.ai/library/$name/$tag"
MF="$SRC/manifests/$MFREL"
SUDO=""; [ -r "$MF" ] || SUDO="sudo"
$SUDO test -f "$MF" || { echo "model $MODEL not in store ($MF) — 'ollama pull $MODEL' first" >&2; exit 1; }

STAGE="$(dirname "$OUT")/.work/model-stage"; rm -rf "$STAGE"
mkdir -p "$STAGE/models/manifests/$(dirname "$MFREL")" "$STAGE/models/blobs"
$SUDO cp "$MF" "$STAGE/models/manifests/$MFREL"
for d in $($SUDO grep -ohE 'sha256:[a-f0-9]+' "$MF" | sort -u); do
	$SUDO cp "$SRC/blobs/sha256-${d#sha256:}" "$STAGE/models/blobs/sha256-${d#sha256:}"
done
$SUDO chown -R "$(id -u):$(id -g)" "$STAGE" 2>/dev/null || true

IMG="$OUT/$TAG.model-disk.img"
rm -f "$IMG"; truncate -s "$SIZE" "$IMG"
mkfs.ext4 -F -q -L FLOWORKMODELS -d "$STAGE" "$IMG"
echo "model disk -> $IMG ($SIZE, label FLOWORKMODELS, model=$MODEL, store=$(du -shL "$STAGE"|cut -f1))"
