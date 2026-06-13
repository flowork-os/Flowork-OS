# Flowork USB Maker (public flasher)

The user-facing tool: download the latest Flowork OS release and write it to a USB stick.
**No build, no source** — it pulls the prebuilt, signed, compressed image from the public
release channel (`flowork-os/flowork`), verifies its checksum, and flashes it.

```sh
GOWORK=off go build -o flowork-usb-maker ./os/flasher
sudo ./flowork-usb-maker          # opens http://127.0.0.1:8800 in your browser
#   -repo  owner/name   public release repo (default flowork-os/flowork)
#   -api   URL          GitHub API base (override for a mirror / testing)
#   -no-browser
```

In the browser: it shows the latest version, you pick a USB (only removable/USB disks are
ever listed — your system disk is never offered), and click **Download & Flash**. It streams
`curl image.zst | sha256 -c | zstd -dc | dd`, so it never needs gigabytes of free disk.

After flashing: **boot** the stick for the Flowork OS appliance, or **plug it into a running
Windows/macOS/Linux** and run the launchers on the `FLOWORK` partition (no reboot).

> Builder (Full/Default, owner-only, private) vs **Flasher (this, public, users)** — see
> ../docs/DISTRIBUTION-ROADMAP.md.
