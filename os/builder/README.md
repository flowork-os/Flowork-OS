# Flowork OS Creator (GUI)

A small local web GUI to build + flash a bootable Flowork OS USB stick. Pick the flashdisk, pick a
profile, click Build then Flash.

## Profiles
- **Full** — bakes the owner's whole live state (database + settings + **tokens**) so the stick
  boots ready-to-use, no re-setup. Data source defaults to `~/.flowork` (where `flowork.db` =
  settings + tokens lives); you can point it at another folder. **Contains secrets — keep private.**
  (= `make-distributable.sh dev`)
- **Default** — clean: NO database/settings ship; every secret becomes `change_this_token`.
  Safe to share. (= `make-distributable.sh public`)

## One-click (recommended)
```sh
os/builder/install-desktop.sh     # builds the binary + adds a menu launcher
```
Then launch **“Flowork OS Creator”** from your application menu (or double-click the desktop
entry). It opens the GUI in your browser and asks for your password once via a graphical prompt
(pkexec) — build + flash need root. Needs: `pkexec`, `xdg-open` (both standard on Linux desktops).

## Manual run
```sh
GOWORK=off go build -o flowork-creator ./os/builder
sudo ./flowork-creator            # opens http://127.0.0.1:8799 in your browser
#   -os <folder>          override the Flowork OS folder (default: auto-detected)
#   -data-default <dir>   default Full-profile data source (default: ~/.flowork)
#   -no-browser           don't auto-open a browser
#   -addr <ip:port>
```

## Safety
- Only **removable / USB** whole disks are ever offered as flash targets — your system disk is
  never listed, and the target is re-checked right before writing.
- Build needs docker (rootfs) + root (USB image); flash needs root (`dd`).

Under the hood it just drives `os/build/make-distributable.sh` + `dd`, streaming progress live.
