# Flowork OS — P2 Specification (DEFINE): Local AI (Ollama + Qwen)

> Phase **P2 — Local AI** (the sovereign brain). Deferred to last by owner; now being built.
> The appliance runs an LLM **on-device, offline** — prompts and data never leave the box.
> Pipeline stage: **DEFINE**.

## Goal
Ollama + a local **Qwen** GGUF model run on the appliance, bound to loopback only. Inference
works with **no network**. The sovereign router (`:2402`) routes LLM calls to local Ollama
(`:11434`) — co-located, zero data egress.

## Why
- **Sovereignty = the core pitch.** A local model means prompts/data stay on the appliance.
- **Offline.** The appliance is useful with no internet (USB in a field laptop, an air-gapped box).
- **Apache-2.0 Qwen** (design choice): commercial-safe, strong multilingual (ID+EN) + tool-use.

## The musl problem (and the fix)
Ollama is a **glibc/cgo** binary (its llama.cpp runner is glibc too); the appliance is **musl**
(Alpine). Fix: lay a **glibc sidecar** — the glibc loader + libs at the standard multiarch paths
(`/lib64/ld-linux-x86-64.so.2`, `/lib/x86_64-linux-gnu`, libstdc++/libgcc). These don't collide
with musl's loader (`/lib/ld-musl-x86_64.so.1`), so glibc and musl binaries coexist. Proven:
`ollama serve` + the `llama-server` runner + offline inference all work this way.

## Storage (where things live)
- **Ollama runtime** (binary + CPU runner libs, GPU libs stripped → ~64 MB) + **glibc sidecar**
  (~42 MB) ride in the **squashfs** (verified by dm-verity, part of the OS).
- **The model** rides on a **writable disk** (label `FLOWORKMODELS`) — large + user-data-like, so
  it is NOT in the squashfs (no per-A/B-slot duplication) and persists.

## Success criteria (binary — `build/verify-ollama.sh`, QEMU, no network)
| # | Criterion | How verified |
|---|-----------|--------------|
| P2.1 | Ollama serves on `127.0.0.1:11434` on the appliance | serial: "OLLAMA serve starting"; `/api/tags` answers. |
| P2.2 | The local model is present + listed | self-test: "first model = qwen2.5:0.5b". |
| P2.3 | **Offline** inference returns a completion | VM has **no NIC**; `/api/generate` returns a non-empty response → "OLLAMA_AI RESULT: OK". |
| P2.4 | Loopback-only (sovereign) | Ollama binds `127.0.0.1` only; no external exposure. |
| P2.5 | Coexists with the rest (verified boot, sandbox, persistence) | the same image still passes squashroot/verity boot. |

## Non-goals (deferred → BACKLOG)
- **Router → Ollama** routing end-to-end (the router's sovereign seed points at `:11434`; wiring
  the appliance router to embed that seed + match model names is the integration step — **P2b**).
- The big model (Qwen 7B Q4 ~4.5 GB) — the PoC verifies with **qwen2.5:0.5b** (~380 MB) for speed;
  the mechanism is model-size-independent. Real stick: 16 GB with Qwen 7B (see BACKLOG/sizing).
- GPU acceleration (CPU-only runner here) — product/hardware stage.
