#!/usr/bin/env bash
# sign-release.sh — sign a release artifact with the Flowork OS update key, producing a
# detached <artifact>.sig that the appliance verifies before trusting an update.
# Run by the owner/CI at release time, with the offline private key.
#   sign-release.sh <artifact> [private-key]
set -euo pipefail
SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ART="${1:?usage: sign-release.sh <artifact> [private-key]}"
KEY="${2:-$(cd "$SELF_DIR/.." && pwd)/config/update.key}"
[ -f "$ART" ] || { echo "artifact not found: $ART" >&2; exit 1; }
[ -f "$KEY" ] || { echo "private key not found: $KEY (run make-update-keys.sh)" >&2; exit 1; }
openssl dgst -sha256 -sign "$KEY" -out "$ART.sig" "$ART"
echo "[sign] $ART.sig"
