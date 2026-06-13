# Architecture, How it works & Technology

## Architecture

One process, a few cooperating parts:

- **Universal endpoint** — serves OpenAI `/v1/chat/completions` (+ streaming), Anthropic
  `/v1/messages`, OpenAI `/v1/responses`, and Gemini `/v1beta/models` all at once.
- **Format translation** — transparent OpenAI ⇄ Anthropic ⇄ Gemini conversion (request, response, and
  streaming) via a dual-hop `source → openai → target` registry, including tool-call parity.
- **Vendor executors** — per-provider wire-format backends, so each upstream gets exactly the shape it
  expects.
- **Routing engine** — priority / round-robin / cost-optimal fallback, a cooldown/backoff table,
  combos (one alias → many models), and a cost-tier classifier that sends simple queries to cheaper or
  local models.
- **Subscription auth + cloaking** — drive subscriptions without API keys; subscription requests are
  cloaked to look like a genuine session, then restored in the response (auto-off for API-key
  providers).
- **The Brain** — a local SQLite FTS5 "Memory Palace" (RAG) plus a mistakes journal that powers the
  anti-hallucination antibody (see [brain-and-antibody.md](brain-and-antibody.md)).
- **The Mesh** — an optional P2P layer: each router has an ed25519 identity, discovers peers, and
  gossips knowledge — leaderless and offline-survivable, gated by a karma check.
- **Embedded web GUI** — the control panel ships inside the binary.

## How it works (one request's journey)

1. A client — your IDE, a CLI, or a Flowork agent — sends an OpenAI / Anthropic / Gemini request to
   `:2402`.
2. The router picks a provider from your fallback chain (subscription or local) and translates the
   request into that provider's format if needed.
3. If the brain is on, it can enrich the prompt — and inject anti-hallucination *antibodies* (past
   mistakes) before the model runs.
4. For a subscription provider, the request is cloaked; then it is dispatched.
5. If that provider errors or rate-limits, the router rolls to the next one automatically.
6. The response is translated back to the client's format; usage, tokens, latency, and cost are logged.

## Technology

- **Go 1.25**, one static binary — no Docker, no Python, no DB server. Linux / macOS / Windows; runs
  on a Raspberry Pi.
- **SQLite** for the brain (FTS5 full-text search) and local state.
- **ed25519** identities + a gossip engine for the optional P2P mesh.
- **Embedded web UI**, served from the binary.
- MIT-licensed; the codebase is audited file-by-file and locked.
