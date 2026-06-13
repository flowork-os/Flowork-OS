# Dual-mode USB builder

Build a Flowork OS stick in one of two data profiles. Same OS, same code — only *what data
rides along* differs.

```sh
os/build/make-distributable.sh public   # for sharing  — no secrets, clean DB
os/build/make-distributable.sh dev       # for the owner — full state, real tokens
```

## `public` — safe to share
- Every secret in the baked config is replaced with the literal `change_this_token`
  (`sanitize-public.py`): provider API keys, tokens, cookies, bot tokens, PEM keys, JWTs, and
  the old `PASTE_YOUR_KEY_HERE` placeholder. Non-secret values (model names, URLs, `tokenSource`)
  are kept.
- **No** user database/settings ship — the stick boots to a clean DB. The recipient flashes,
  boots, opens Settings → API Keys, and replaces `change_this_token` with their own keys.
- This is the image you put on a download page.

## `dev` — ready-to-use, keep private
- The owner's live state (`flowork.db` + config + **real tokens**) from `FLOWORK_HOME`
  (default `~/.flowork`) is baked into the image and seeded into `/root/.flowork` on first boot
  (only if the live DB is empty — never overwrites later user data).
- Boots straight into the owner's working setup. **Contains secrets — never distribute.**

## How it works
`make-distributable.sh` wraps the normal pipeline (`build-flowork-os.sh` → `make-usb-image.sh`):
- public → `FLOWORK_SOVEREIGN_SEED` points at the sanitized seed; no `FLOWORK_STATE_SEED`.
- dev → real seed; `FLOWORK_STATE_SEED=$FLOWORK_HOME` → `build-flowork-os.sh` stages `flowork.db`
  into `/usr/share/flowork/state-seed`, which `flowork-data-setup` seeds on first boot.
The artifact is named by mode: `…-public.usb.img` / `…-dev.usb.img`, so the two never get mixed up.

## Verify
```sh
python3 os/build/sanitize-public.py --selftest          # the secret scrubber's own test
python3 os/build/sanitize-public.py in.json out.json    # sanitize any config by hand
```

> Needs docker (rootfs build) and root (USB image), same as the underlying scripts. The sanitize
> + mode-selection logic is unit-verified; the full image build runs on a build machine.
