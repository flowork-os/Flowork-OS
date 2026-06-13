# Flowork OS — P0 PoC Specification (DEFINE)

> Phase: **P0 — Proof of Concept (QEMU VM, ephemeral)**
> Pipeline stage: **DEFINE**
> Status authority: this document defines DONE for P0. Do not advance to P1 until every
> success criterion below is met and recorded in `docs/STATUS.md`.

## 1. Goal (one sentence)
Boot a minimal Alpine-based image inside a QEMU VM that auto-starts the Flowork **agent**
(`:1987`) and **router** (`:2402`) and presents the Flowork control panel full-screen via a
**chromium kiosk**, with **zero persistence** (everything in RAM — ephemeral).

## 2. Why P0 is in a VM, not a USB stick (carried from the design blueprint)
- Iterate in seconds (snapshot/restart) instead of minutes (re-flash).
- Zero risk to the owner's physical PC (no chance of bricking boot media).
- The boot flow can be exercised repeatedly and deterministically.
Physical USB comes only after the VM path is proven.

## 3. Success criteria (binary, testable)
A build passes P0 iff, from a single command (`build/build-flowork-os.sh` then
`build/run-qemu.sh`), all of the following hold:

| # | Criterion | How it is verified |
|---|-----------|--------------------|
| S1 | Image builds reproducibly from source with no host `apk`/root (docker only) | `build-flowork-os.sh` exits 0; artifacts present |
| S2 | Agent + router are **fully static** linux/amd64 binaries (no libc dependency) | `file` reports "statically linked" |
| S3 | The VM boots to userspace (kernel → init) without manual intervention | serial console reaches the boot-health stage |
| S4 | Agent answers HTTP on `:1987` (control panel HTML served) | in-guest `curl localhost:1987` returns 200 + HTML; written to serial |
| S5 | Router answers HTTP on `:2402` | in-guest `curl localhost:2402` returns a response; written to serial |
| S6 | Chromium kiosk renders the Flowork UI full-screen | QEMU framebuffer screendump shows the UI (visual check) |
| S7 | No writes leak to a host disk; state lives in tmpfs only | image boots with no attached writable disk |

S1–S5 and S7 are the **substance** of P0 (a sovereign OS that boots straight into Flowork)
and are fully machine-verifiable headless. S6 (kiosk pixels) is verified by capturing the
VM framebuffer. If S6 proves environment-flaky, it is recorded honestly in STATUS.md with its
exact blocker; it must not silently pass.

## 4. Non-goals (explicitly deferred — do NOT build in P0)
- Persistence / LUKS DATA partition → **P1**.
- Local LLM (Ollama + GGUF model) → **P2**. Router boots but needs no model for P0.
- App sandbox (CPython + bubblewrap), dm-verity, signed images → **P3**.
- A/B auto-update, physical USB, GPU, hardware bundle → **P4**.
- Secrets/API keys in the image. The image ships **empty of credentials**; the owner pastes
  keys at runtime via Settings (same rule as the main repos).

## 5. Constraints
- Languages in the OS: shell + the existing Go binaries. No new runtime added in P0.
- The agent's web GUI is compiled in (`//go:embed web`) → the binary is self-contained.
- Agent state base is `$HOME/.flowork/` (override `FLOWORK_HOME` / `FLOWORK_AGENTS_DIR`).
  In the appliance, `HOME=/root` and agent definitions are pre-seeded read-only.
- All code, scripts, comments, and docs in **English**. (Owner chat is Indonesian; product is English.)

## 6. Deliverables
- `build/build-flowork-os.sh` — one-command image build (docker Alpine → rootfs → kernel +
  initramfs → ephemeral bootable artifact + ISO).
- `build/run-qemu.sh` — boot the artifact in QEMU and run the health checks above.
- `rootfs-overlay/` — init services for agent, router, kiosk, plus boot health probe.
- `docs/STATUS.md` — the verification record (which criteria passed, with evidence).
