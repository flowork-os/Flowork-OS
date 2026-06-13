# Flowork OS — P3a Specification (DEFINE): App Sandbox

> Phase **P3 — App platform** (the sandbox half; hardening = P3b). Builds on P0+P1.
> Ollama/local-AI (P2) is deferred to last by owner decision; the roadmap otherwise proceeds in order.
> Pipeline stage: **DEFINE**.

## Goal
The appliance can run **apps** (Python and Go — the two supported languages) each **confined to
its own room** by `bubblewrap` (bwrap): an app cannot read the owner's state (`/root/.flowork`),
cannot touch other apps, and cannot roam the host filesystem. Same doctrine as the AI citizens:
*sandboxed, confined, no exfil.*

## Why bubblewrap, Python+Go only (from the design)
- **Go** → a static binary, zero runtime. **Python** → one CPython in the base + a per-app venv.
  Two languages only → no container engine needed; a namespace sandbox is enough.
- **bwrap** (not podman) → a tiny, daemonless, image-less namespace sandbox. Just enough to
  jail a process's filesystem/network view. Keeps the base small.

## Success criteria (binary, testable)
| # | Criterion | How verified |
|---|-----------|--------------|
| P3a.1 | `bubblewrap` + `python3` are present in the image | self-test: `bwrap --version`, `python3 --version` |
| P3a.2 | A **Python** sample app runs sandboxed and produces output | self-test serial: `SANDBOX py app -> <result>` |
| P3a.3 | A **Go** sample app (static binary) runs sandboxed and produces output | self-test serial: `SANDBOX go app -> <result>` |
| P3a.4 | A sandboxed app **cannot read** `/root/.flowork` (owner state) | self-test serial: app's read attempt is denied → `SANDBOX isolation: DENIED (good)` |
| P3a.5 | A sandboxed app **cannot see** another app's directory | self-test: cross-app path not present inside the jail |
| P3a.6 | Normal boots are unaffected (self-test only runs when asked) | self-test gated by kernel cmdline `flowork.selftest=1` |

## Non-goals (deferred)
- Wiring the agent's `apps` subsystem to launch via `flowork-app-run` → agent-repo work, separate
  from this OS-level capability. P3a delivers the OS sandbox primitive + proof.
- dm-verity / signed images → **P3b**.
- Per-app network policy beyond on/off, seccomp profiles, user-namespace UID remap polish → later.
- Local LLM (Ollama/Qwen) → **last**, per owner.

## Deliverables
- Image gains `bubblewrap` + `python3`.
- `/usr/local/bin/flowork-app-run` — run a command under a confined bwrap profile.
- `/usr/share/flowork/sample-apps/{py-hello,go-hello}` — one Python, one Go sample.
- `/usr/local/bin/flowork-sandbox-check` + `flowork-sandbox-check` service (gated by
  `flowork.selftest=1`) — proves P3a.1–P3a.5 and prints results to the serial console.
- `build/verify-p3a.sh` — boot with the self-test flag, assert the sandbox results.
- STATUS.md updated.
