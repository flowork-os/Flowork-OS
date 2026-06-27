#!/usr/bin/env bash
# build-agent.sh — compile satu agent source ke .wasm dan stage ke
# ~/.flowork/agents/<id>.fwagent/.
#
# Pakai:
#   ./scripts/build-agent.sh mr-flow
#
# Override:
#   FLOWORK_AGENTS_DIR  target staging dir (default ~/.flowork/agents)
#   TINYGO             path tinygo (auto-discover ~/.local/share/tinygo/bin)

set -euo pipefail

AGENT_ID="${1:-}"
if [ -z "$AGENT_ID" ]; then
  echo "Usage: $0 <agent-id>" >&2
  exit 1
fi

SRC_DIR="agents/$AGENT_ID"
if [ ! -d "$SRC_DIR" ]; then
  echo "Source not found: $SRC_DIR" >&2
  exit 1
fi
if [ ! -f "$SRC_DIR/manifest.json" ]; then
  echo "Missing manifest.json in $SRC_DIR" >&2
  exit 1
fi

# Resolve tinygo
TINYGO="${TINYGO:-}"
if [ -z "$TINYGO" ]; then
  for cand in "$HOME/.local/share/tinygo/bin/tinygo" "$(command -v tinygo || true)"; do
    if [ -x "$cand" ]; then TINYGO="$cand"; break; fi
  done
fi
if [ -z "$TINYGO" ] || [ ! -x "$TINYGO" ]; then
  echo "tinygo not found. Install from https://tinygo.org/getting-started/install/" >&2
  exit 1
fi

# Resolve target
TARGET_BASE="${FLOWORK_AGENTS_DIR:-$HOME/.flowork/agents}"
TARGET_DIR="$TARGET_BASE/$AGENT_ID.fwagent"

echo "→ Source:  $SRC_DIR"
echo "→ Target:  $TARGET_DIR"
echo "→ tinygo:  $TINYGO"
echo

mkdir -p "$TARGET_DIR"

# Build WASM. Prefer tinygo (wasm kecil), TAPI tinygo nempel Go ≤1.23 → di Go 1.25+ GAGAL.
# Jadi FALLBACK ke Go-standar wasip1 (selalu jalan, agak gede). Konsisten sama build-all-agents.sh.
# Biar GA "dev jalan, build agent baru gagal" pas toolchain beda.
built=0
if [ -x "$TINYGO" ] && ( cd "$SRC_DIR" && "$TINYGO" build -target=wasi -scheduler=none -opt=z -no-debug -o "$TARGET_DIR/agent.wasm" . ) 2>/dev/null; then
  echo "✓ compiled via tinygo ($(stat -c%s "$TARGET_DIR/agent.wasm") bytes)"
  built=1
fi
if [ "$built" = 0 ]; then
  echo "→ tinygo ga jalan (cek versi Go) → fallback Go-standar wasip1"
  if ( cd "$SRC_DIR" && GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o "$TARGET_DIR/agent.wasm" . ); then
    echo "✓ compiled via go wasip1 ($(stat -c%s "$TARGET_DIR/agent.wasm") bytes)"
  else
    echo "✗ build GAGAL (tinygo + go wasip1 dua-duanya)"; exit 1
  fi
fi

cp "$SRC_DIR/manifest.json" "$TARGET_DIR/manifest.json"
echo "✓ manifest copied"

for opt in persona.md skills ui i18n config.json; do
  if [ -e "$SRC_DIR/$opt" ]; then
    cp -r "$SRC_DIR/$opt" "$TARGET_DIR/"
    echo "✓ staged $opt"
  fi
done

echo
echo "── staged agent ──"
ls -la "$TARGET_DIR"
