#!/usr/bin/env bash
# install-desktop.sh — install the "Flowork OS Creator" launcher into your app menu (one click).
# Prebuilds the binary (so click-time needs no Go) and writes a .desktop with absolute paths.
set -e
SELF="$(cd "$(dirname "$0")" && pwd)"

echo "[install] building flowork-creator…"
( cd "$SELF" && GOWORK=off go build -o "$SELF/flowork-creator" . )
chmod +x "$SELF/flowork-creator-launch.sh"

APP="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
mkdir -p "$APP"
DESKTOP="$APP/flowork-os-creator.desktop"
cat > "$DESKTOP" <<EOF
[Desktop Entry]
Type=Application
Version=1.0
Name=Flowork OS Creator
GenericName=USB Creator
Comment=Build & flash a Flowork OS USB stick (Full / Default)
Exec=$SELF/flowork-creator-launch.sh
Icon=$SELF/icon.svg
Terminal=false
Categories=Utility;
Keywords=flowork;usb;flash;iso;creator;
StartupNotify=true
EOF
chmod +x "$DESKTOP"
update-desktop-database "$APP" 2>/dev/null || true

echo "[install] done → $DESKTOP"
echo "[install] Find 'Flowork OS Creator' in your application menu, or double-click the desktop file."
echo "[install] (Build/flash ask for your password via a graphical prompt — pkexec/polkit.)"
