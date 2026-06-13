# Flowork OS — P3b Specification (DEFINE): Anti-Tamper (dm-verity)

> Part of **P3 — Hardening** (design Section 7, the moat). This delivers the **dm-verity
> integrity primitive**, buildable and verifiable now. Wiring it into BOOT (a read-only,
> verity-protected root that the kernel refuses to mount if tampered) needs the
> squashfs-on-disk boot path — deferred, see `docs/BACKLOG.md`.
> Pipeline stage: **DEFINE**.

## Goal
Prove that the read-only rootfs can be **integrity-protected**: every block is hashed into a
Merkle tree; a single changed byte is detected. Combined with P4a (signed root hash), this is the
"this OS can't be silently hijacked" guarantee that wealthy/enterprise clients pay for.

## Why dm-verity (design Section 7)
- Block-level Merkle tree → tamper is caught **at read time**, not just at boot.
- Same family as the agent's Guardian + kernel-freeze philosophy, one level lower (OS-locked).
- The cloud giants can't match it — their model is to *collect* your data, not seal it.

## Success criteria (binary, testable — `build/verify-p3b.sh`, host, root)
| # | Criterion | How verified |
|---|-----------|--------------|
| P3b.1 | A hash tree can be built for a rootfs image (root hash produced) | `veritysetup format` emits a root hash. |
| P3b.2 | An **intact** image VERIFIES against its root hash | `veritysetup verify` exits 0. |
| P3b.3 | A **tampered** image is REJECTED | flip a byte → verify exits non-zero. |
| P3b.4 | A **wrong root hash** is REJECTED | verify with a bogus hash → non-zero. |

## Tools
- `build/make-verity.sh <rootfs.squashfs> [--sign]` → `<img>.verity` (hash tree) + `<img>.roothash`.
  With `--sign` it also signs the root hash via `sign-release.sh` → the full chain of trust:
  **signed root hash (P4a) → dm-verity-verified squashfs (P3b)**.
- `build/verify-p3b.sh` → the host integrity test above.

## Non-goals (deferred → BACKLOG)
- **Boot integration**: mount the squashfs root through dm-verity and refuse to boot on mismatch.
  Needs the squashfs-on-disk boot path (the current PoC boots an in-RAM/tmpfs root). Same boot
  path that A/B auto-update (P4) needs.
- Secure Boot signed shim (trusted chain from firmware) → product.
- Real LUKS key custody (TPM/passphrase) → product.
