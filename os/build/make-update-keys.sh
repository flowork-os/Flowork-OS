#!/usr/bin/env bash
# make-update-keys.sh — generate the Flowork OS update-signing keypair (EC P-256).
#   config/update-pub.pem  -> PUBLIC key, committed + baked into the image (verifies updates)
#   config/update.key      -> PRIVATE signing key, GITIGNORED, kept offline by the owner
# Real deployments use the owner's offline key; this is the PoC/dev keypair.
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CFG="$(cd "$SELF_DIR/.." && pwd)/config"
mkdir -p "$CFG"
if [ -f "$CFG/update.key" ] && [ "${1:-}" != "--force" ]; then
	echo "[keys] $CFG/update.key already exists (use --force to regenerate)"; exit 0
fi
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$CFG/update.key"
chmod 0400 "$CFG/update.key"
openssl pkey -in "$CFG/update.key" -pubout -out "$CFG/update-pub.pem"
echo "[keys] private: $CFG/update.key (gitignored — keep offline)"
echo "[keys] public : $CFG/update-pub.pem (committed + baked into the image)"
