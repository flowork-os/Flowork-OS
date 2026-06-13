# Flowork OS — P4a Specification (DEFINE): Signed-Update Verification

> Part of **P4 — Product / auto-update** (design Section 8). This is the security-critical
> CHECK→DOWNLOAD→**VERIFY** core, buildable and testable now. The atomic **A/B partition switch +
> reboot/rollback** is deferred (needs the disk-image boot path) — see `docs/BACKLOG.md`.
> Pipeline stage: **DEFINE**.

## Goal
The appliance can pull a newer Flowork OS rootfs from **our GitHub Releases** and **refuse any
image that is not signed by the owner's key**. A hijacked GitHub or a MITM cannot feed it a
poisoned image.

## Why (design Section 8)
- **Signed** → only an image signed with the owner's offline private key is accepted; the public
  key is baked into the appliance. Anti-MITM, anti-repo-hijack.
- Pulled from **GitHub Releases** → free, already our release channel, fits "fresh clone = same setup."

## Success criteria (binary, testable — `build/verify-p4a.sh`, host)
| # | Criterion | How verified |
|---|-----------|--------------|
| P4a.1 | A validly-signed artifact is **ACCEPTED** | `flowork-update verify` exits 0 on a good sig. |
| P4a.2 | A **tampered** artifact is **REJECTED** | flip a byte → verify exits non-zero. |
| P4a.3 | A signature from the **wrong key** is **REJECTED** | verify against another pubkey → non-zero. |
| P4a.4 | The public key + version are baked into the image; `flowork-update` + `openssl` present | staged rootfs has `/etc/flowork/update-pub.pem`, `/etc/flowork/version`, `/usr/local/bin/flowork-update`. |

## Scheme
- EC P-256 + SHA-256 detached signature (`openssl dgst`), streamed (no need to load the whole
  image into memory).
- `build/make-update-keys.sh` → `config/update.key` (private, **gitignored**, owner-offline) +
  `config/update-pub.pem` (public, committed + baked).
- `build/sign-release.sh <artifact>` → `<artifact>.sig` (release-time, owner).
- `flowork-update {verify|check|run}` (in the appliance): `run` = check GitHub → download
  `*.rootfs.squashfs` + `.sig` → verify → stage. On invalid signature it deletes the download
  and refuses.

## Non-goals (deferred → BACKLOG)
- Atomic **A/B partition switch + reboot + auto-rollback** → needs the squashfs-on-disk boot path
  (product). `run` stages the verified image and logs the next step.
- dm-verity of the running root → **P3b**.
- Secure Boot signed shim → product.
