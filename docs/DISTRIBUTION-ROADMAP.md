# Distribution & Auto-Update — Roadmap (living doc)

> Anchor for the ongoing discussion so nothing is lost. Updated as we decide. Flowork-OS is a
> PRIVATE repo, so this stays internal.

## North star
Owner codes in **one private place**; users get everything from **public channels**; installed
devices **update themselves** safely. Source location ≠ update channel.

```
  SOURCE + BUILDER (private)                 DISTRIBUTION (public)
  github.com/flowork-os/Flowork-OS  ──CI──►  agent binary  → users + auto-update
   (owner codes + builds here)        │      router binary → users + auto-update
                                      └────►  Default OS image (compressed) → users download + flash
```

## Decisions locked (this discussion)
1. **Source of truth = `Flowork-OS` monorepo, PRIVATE.** Owner codes here only. No dual source
   (avoids drift). The old per-component repos are NOT a second source.
2. **Auto-update pulls compiled binaries from public RELEASES**, per component (agent updates from
   its channel, router from its channel) — matches "2 binaries, independent updates".
3. **Builder vs Flasher split:**
   - **Builder** (the GUI `os/builder`, Full/Default) = PRIVATE, owner-only (needs source + docker).
   - **Flasher** = PUBLIC, user-facing: *download latest Default image → pick USB → flash*. NO build,
     no source. One-click. (= roadmap P4.1 "Flowork OS Creator".)
4. **Two image profiles:**
   - **Default** (public): clean, secrets → `change_this_token`, **NO bundled model**, smaller,
     DATA auto-grows on first boot. This is the public download.
   - **Full** (private, owner): bakes owner's `~/.flowork` (db + tokens + agents) + model. Offline-
     ready, ready-to-use. Never distributed.
5. **Image size:** the 6 GB is fixed allocation (mostly blank DATA), not the model (~380 MB).
   For Default: drop the model (default LLM = Claude cloak; local Qwen = optional), shrink the raw
   image to ~2–2.5 GB, auto-grow DATA partition on first boot (`sgdisk -e`).
6. **GitHub release assets cap at 2 GB/file** → ship images **compressed** (`*.img.zst`, ~0.9–1.3 GB).
   Flasher streams `zstd -d | dd`. (Even the full 6 GB image compresses to 1.3 GB, verified.)
7. **Auto-update SAFETY is already done:** additive migrations + `seed-new-by-id` content merge
   (data preserved, new agents auto-appear) + signed A/B install + rollback. See AUTO-UPDATE.md.

## Open questions (decide next)
- **Q1 — Public channel location:** ✅ DECIDED = **`flowork-os/flowork`** (already public).
  All public release assets (agent binary, router binary, compressed Default image, Flasher,
  manifest.json) publish here. (`flowork-os/flowork` doesn't exist — owner had confused it with the
  private `Flowork-OS`. Can move to a dedicated public repo later; it's just a config.)
- **Q2 — Offline model for Default users:** ✅ DECIDED = **Ollama pulls on demand** (default = no
  model; user clicks "enable local AI" → downloaded online). Default LLM stays Claude cloak.
- **Q3 — Auto-update default:** ✅ DECIDED = **ON by default + an OFF toggle** in Settings
  (manifest.json polling is cheap, so default-on is fine; user can disable).

## Build backlog (once Q1–Q3 settled)
- [ ] **A · CI release pipeline** (in private monorepo): on tag/push → build agent + router (static)
      + Default image (model-less, slim) → sign → compress → publish RELEASE to the public channel.
- [ ] **B · Public Flasher** (one-click): reads latest release, downloads `*.img.zst`, lists USB
      (removable-only safety), `zstd -d | dd`, verifies. Ship as a small public binary + `.desktop`.
- [ ] **C · Slim Default image:** `make-distributable.sh public` → no model, raw ~2–2.5 GB, first-boot
      DATA auto-grow. Keep Full (private) = model + baked state.
- [ ] **D · Periodic auto-updater on devices:** OS OpenRC timer runs `flowork-update run`
      (check→download→verify→A/B-switch→reboot, rollback-safe); router/desktop background checker
      (`internal/updater`: LatestRelease→IsNewer→Swap→Restart). Point both at the public channel.
      Add a Settings on/off toggle (per Q3).
- [ ] **E · Compatibility:** agent↔router version skew handling (independent updates) — define a min
      compatible-version check so an updated agent won't break against an older router.

## Product architecture — ONE source, MANY products (decided 2026-06-13)
Owner's strategy: the **router is a standalone wedge** — it's OpenAI/Anthropic-compatible, so users
of OTHER agent frameworks (Hermes, OpenClaw, Cursor, Claude Code…) can point their `base_url` at
**our** router and get the cloak + routing + mesh. That captures their users. But maintaining
separate source repos (router/agent/os/android) = a headache, and fixes (mesh, etc.) only land in
the monorepo. Resolution:

```
Flowork-OS (private monorepo = THE source; fix once here)
   │ CI
   ├─▶ flowork-router  (binary + DOCKER image)  ← standalone wedge product
   ├─▶ flowork-agent   (binary)
   ├─▶ Flowork OS image (USB)
   └─▶ Android APK
```
- NO separate source repos to sync. One monorepo, CI publishes every product.
- Old `flowork_Router` / `Flowork_Agent` repos: **left untouched** (owner's explicit instruction).
- Public face = `flowork-os/flowork` (created 2026-06-13, public) — releases + landing.

## Build backlog (remaining)
- [ ] **Router standalone product**: `router/Dockerfile` + "use Flowork Router with any agent" doc
      + add to the release pipeline (docker image + binary). Realizes the wedge strategy.
- [ ] **CI auto-publish all products** from the monorepo (router/agent/os image/portable) to
      `flowork-os/flowork` releases on `git tag`.
- [ ] Android APK assembly (needs SDK machine / CI).

## Fixes — 2026-06-13 (post-flash feedback)
Owner-reported issues after flashing/using the Full stick, all fixed:
- **"Start opens DonutBrowser, not the OS"** — the portable Start launcher used `xdg-open`, which
  follows the *system default browser*; an anti-detect browser (DonutBrowser) had registered itself
  as that default, hijacking every launch. Fixed: Start now picks a real browser itself and opens
  the panel as a **chromeless app-mode window** (`chrome/edge/chromium --app=` on Win/Linux/Mac), so
  it reads like the OS appliance — and explicitly **skips any donut/adspower/dolphin/multilogin**
  binary. Never trusts the hijackable default again. (`os/portable/make-portable.sh`)
- **DonutBrowser purge** — the old anti-detect-browser path is gone product-wide. The router brain
  YouTube doctrines (`youtube_upload_batch`, `yt_pipeline`, `sandbox_phase1_youtube`, …) were
  rewritten onto the **sovereign YouTube Data API v3 (`videos.insert`)** path — no third-party
  browser, OAuth per channel, quota-aware. Capability preserved, dependency removed.
- **Full build shipped EMPTY settings/agents** — a hot `cp`/`tar` of the live SQLite DB could miss
  the WAL or tear; and a re-flash over an existing DATA volume skipped seeding (guard: only seed when
  no `flowork.db`). Fixed two ways: (1) bake a **consistent `sqlite3 .backup` snapshot** (WAL merged);
  (2) stamp the seed with `VERSION + DB-hash` and, on first boot, **restore owner state when a NEWER
  owner image is flashed over old DATA** (backs up the old DB first; plain reboots are untouched;
  PUBLIC images carry no seed so user data is never clobbered). Logic unit-tested (fresh/reboot/
  reflash). (`os/build/build-flowork-os.sh`, `os/rootfs-overlay/.../flowork-data-setup`)
- **Hardened the public sanitizer selftest** — removed a leftover real-token *fragment* needle,
  replaced with the fake input value. (`os/build/sanitize-public.py`)

## Status snapshot (done so far)
- **FIRST PUBLIC RELEASE v0.6.0** at **github.com/flowork-os/flowork** (public; 12 topics, killer
  README, release assets: usb-maker, portable.zip, OS image .zst, agent+router binaries, signed
  rootfs, manifest). [Image asset re-upload pending if an interrupted upload left it partial.]
- Monorepo `Flowork-OS` (private) — agent + router + os + builder; pushed.
- OS builds from monorepo + QEMU boot verified (kiosk renders, agent+router up, net, persist).
- Dual-mode builder + sanitizer (`change_this_token`) + GUI Creator + one-click `.desktop`.
- Default image (6 GB, clean) + Full image (6 GB, owner state baked) both built & verified.
- Desktop launcher; Android scaffold (Flowork-Mobile) de-risked.
