#!/usr/bin/env bash
# install-watchdog.sh — R9 SELF-HEAL: pasang/perbarui systemd-user unit yang nge-supervise
# watchdog.sh (Restart=always). Idempotent — aman dijalanin berkali2 / tiap mesin baru.
# Owner-approved 2026-06-15 (FASE 2 autonomi). Ganti docktor lama (binary external ilang).
#
# Pakai: ./os/selfheal/install-watchdog.sh        (pasang + start)
#        FLOWORK_NO_WATCHDOG=1 ... (watchdog.sh sendiri yang nurutin opt-out saat runtime)

set -eu
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"   # os/selfheal/ → FLowork_os root
WD="$ROOT/os/selfheal/watchdog.sh"
UNIT_DIR="$HOME/.config/systemd/user"
UNIT="$UNIT_DIR/flowork-docktor.service"

[ -x "$WD" ] || chmod +x "$WD" 2>/dev/null || true

if ! command -v systemctl >/dev/null 2>&1; then
  echo "systemctl gak ada (bukan systemd). Watchdog bisa dijalanin manual: nohup $WD &"
  exit 0
fi

mkdir -p "$UNIT_DIR"
cat > "$UNIT" <<UNITEOF
[Unit]
Description=Flowork Docktor (process watchdog for Flowork stack)
After=default.target

[Service]
Type=simple
ExecStart=$WD
Restart=always
RestartSec=10
WorkingDirectory=$ROOT

[Install]
WantedBy=default.target
UNITEOF

systemctl --user daemon-reload
systemctl --user enable flowork-docktor >/dev/null 2>&1 || true
systemctl --user restart flowork-docktor

# LINGER — biar docktor (+ stack Flowork) auto-start saat BOOT, BUKAN cuma pas login.
# Owner ide #2: "pas PC nyala dia HARUS auto nyala". Tanpa ini, systemd-user unit nunggu login.
# Idempotent + best-effort (butuh polkit/root; ga fatal kalau gagal).
loginctl enable-linger "$USER" >/dev/null 2>&1 || sudo loginctl enable-linger "$USER" >/dev/null 2>&1 || \
  echo "  (warn: enable-linger gagal — stack nyala pas login, bukan boot. Jalanin: loginctl enable-linger $USER)"

echo "✅ flowork-docktor terpasang → $WD (linger=$(loginctl show-user "$USER" --property=Linger 2>/dev/null | cut -d= -f2))"
systemctl --user is-active flowork-docktor
