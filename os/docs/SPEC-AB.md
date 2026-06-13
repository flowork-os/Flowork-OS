# Flowork OS — A/B Auto-Update (DEFINE)

> The last big hardening item (design Section 8). The security foundation is done — signed-update
> verify (P4a) + dm-verity-at-boot (P3b). This adds the **atomic A/B slot switch + rollback** so an
> update can never half-apply or brick the appliance. Pipeline stage: **DEFINE**.

## Goal
Two OS **slots** (A and B) live on writable media + an `active` pointer. The appliance boots the
active slot (verified by dm-verity). An update writes the new image to the **inactive** slot,
verifies it, then flips the pointer — atomic. If the new slot fails (bad image, or boots but never
becomes healthy), the appliance **rolls back** to the other slot automatically.

## Why
- **Atomic** → a power cut mid-update never corrupts the running system (the inactive slot is
  written, the pointer flip is the single atomic commit).
- **Rollback** → a bad release can't brick the box; it reverts to the last-good slot.
- **No new trust** → reuses P4a (only signed images accepted) + P3b (only verified roots boot).

## Layout (on the writable media)
```
flowork/
  active                       -> "a" or "b" (the single atomic commit point)
  slot-a/rootfs.squashfs(+.verity +.roothash [+.roothash.sig])
  slot-b/rootfs.squashfs(+.verity +.roothash [+.roothash.sig])
  slot-<x>.tries               -> unconfirmed boot attempts of slot x (rollback counter)
  slot-<x>.confirmed           -> written once slot x boots healthy (resets tries)
```

## Boot + rollback logic (stage-1 /init)
```
read active = x ; if slot-x unconfirmed: tries++ ; if tries > 3 -> flip active to other, reboot
verify+mount slot-x via dm-verity
  verity/mount FAILS  -> (A/B) flip active to other, reboot   [bad image auto-rollback]
                         (single slot) -> refuse to boot (poweroff)
expose active slot to the booted system (/run/flowork-slot + media at /flowork-media)
```
On a healthy boot, `flowork-health` writes `slot-x.confirmed` + clears `slot-x.tries`.

## Update (flowork-update, in the booted system; media at /flowork-media)
```
flowork-update install <squashfs> [<verity> <roothash> [<sig>]]
    -> write to the INACTIVE slot ; (P4a) verify signature if a key+sig is present
    -> clear that slot's tries/confirmed ; flip active ; ready to reboot
flowork-update switch | status
flowork-update run     -> download a signed slot from GitHub Releases, then install (P4a verify)
```

## Success criteria (binary — `build/verify-ab.sh`, QEMU)
| # | Criterion | How verified |
|---|-----------|--------------|
| AB.1 | Boots the active slot (verified) | active=a → serial "A/B mode: active slot = a" + `AGENT_UP`. |
| AB.2 | Switching the pointer boots the other slot | flip active=b → boots slot b + `AGENT_UP`. |
| AB.3 | A healthy boot is **confirmed** | serial "FLOWORK_AB slot=a confirmed"; `slot-a.confirmed` exists. |
| AB.4 | A bad slot (verity fail) **auto-rolls-back** | corrupt slot-b → boot → "rolling back to a" → slot a boots. |
| AB.5 | An unconfirmed slot rolls back after N tries | preset slot-b.tries=3, no confirm → boot → tries>3 → revert to a → slot a boots. |

## Non-goals (deferred → BACKLOG)
- GRUB-on-the-writable-USB packaging (the PoC boots `-kernel/-initrd`; real USB puts GRUB on the
  stick). The A/B logic is in `/init`, independent of how the kernel+initramfs are loaded.
- Live download wiring against a real GitHub Release with two signed slots (the `run` path exists;
  `install` is the verified mechanism).
