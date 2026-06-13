# Agent ↔ Router version compatibility (BUILD-E)

## The invariant: we always ship a MATCHED PAIR
In every Flowork delivery, the agent and router come from the **same build**, so they can never
drift out of sync:

- **Flowork OS (USB)** — both binaries live in one `rootfs.squashfs`. An A/B update swaps the
  WHOLE rootfs atomically → the pair is replaced together. **Zero skew, ever.**
- **Portable / desktop** — the launcher runs `bin/<os>/flowork-agent` + `flowork-router` from the
  same folder, shipped together in one release artifact. Self-update replaces both together.
- **Android** — both binaries ride in one APK.

So in normal operation, skew is structurally impossible — there is no path that updates one
component without the other.

## The safety net: `min_compatible` in the manifest
For defence-in-depth (a hand-swapped binary, a partial download, a future independent component
channel), every release `manifest.json` carries:

```json
"min_compatible": { "agent": "<version>", "router": "<version>" }
```

An updater MUST refuse to apply a component update that would leave the running peer below the
declared minimum. Because today we only ship matched pairs, the gate is always satisfied; it
exists so that the day we *do* allow independent component updates, the rule is already enforced
and documented.

## Rule for future changes
If you ever make the agent require a newer router (or vice-versa) — a changed wire contract,
a new shared table, a new endpoint the other side depends on — **bump `min_compatible`** in the
release for the side that introduced the requirement. Never assume the peer is current.
