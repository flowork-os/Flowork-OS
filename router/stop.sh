#!/usr/bin/env bash
# Flow Router — one-click stop (Linux / macOS)
# Sends SIGTERM to whatever listens on FLOW_ROUTER_PORT (default 2402).
PORT="${FLOW_ROUTER_PORT:-2402}"
echo "🛣️  Flow Router — stopping :$PORT…"
PID=$(ss -ltnp 2>/dev/null | grep ":$PORT " | grep -oE 'pid=[0-9]+' | head -1 | cut -d= -f2)
if [ -z "$PID" ]; then
  # macOS / fallback via lsof
  PID=$(lsof -ti tcp:"$PORT" 2>/dev/null | head -1)
fi
if [ -n "$PID" ]; then
  kill "$PID" 2>/dev/null && echo "✓ sent SIGTERM to pid $PID" || echo "kill failed"
else
  echo "(nothing listening on :$PORT)"
fi

# One-click teardown: unload model lokal dari VRAM (free GPU). Best-effort.
LOCAL_MODEL="${FLOWORK_LOCAL_MODEL:-qwen-flowork}"
if command -v ollama >/dev/null 2>&1; then
  if ollama stop "$LOCAL_MODEL" >/dev/null 2>&1; then
    echo "✓ unloaded local model '$LOCAL_MODEL' (VRAM freed)"
  else
    # Fallback: keep_alive 0 lewat API (ollama lama tanpa 'stop').
    curl -fsS -m5 http://127.0.0.1:11434/api/generate \
      -d "{\"model\":\"$LOCAL_MODEL\",\"keep_alive\":0}" >/dev/null 2>&1 || true
  fi
fi
