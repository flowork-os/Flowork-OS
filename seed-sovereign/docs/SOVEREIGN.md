# Flowork Router — Sovereign Mode (the appliance build)

This repo is the **sovereign / personal** build of the Flowork router, the one that
ships inside **Flowork OS**. Same engine as the upstream router; different default posture.

## What "sovereign" means here
- **Local-first by default.** The shipped config (`appliance/router-config.sovereign.seed.json`)
  has exactly one active provider: a local **Ollama** endpoint at `127.0.0.1:11434`. Every
  request is served on the appliance. Nothing leaves the box.
- **Cloud is opt-in.** No cloud provider and **no API key** is baked in. The owner adds cloud
  providers at runtime via the control panel if (and only if) they choose to. This matches the
  main-repo rule: a fresh image ships empty of credentials.
- **Co-located, not remote.** The router runs as a service next to the agent (`:2402`), so the
  agent's LLM calls never cross the network boundary. (Design §3.)

## Why a separate binary/repo (not merged into the agent)
- **Different concern.** Agent = the AI citizens' runtime + body. Router = the LLM gateway +
  pricing/routing/provider brain. Keeping them apart keeps each maintainable and **independently
  updatable** (A/B updates can ship a router fix without rebuilding the whole OS).
- The router is still **bundled and started on the same appliance** — separation of code,
  co-location at runtime.

## Roadmap alignment
| Phase | Router role |
|-------|-------------|
| **P0** | Boots and answers on `:2402`. No model required yet (just proves the service is up). |
| **P2** | Ollama + a local **Qwen** GGUF model are added; this sovereign seed becomes live, the agent talks to local Qwen through the router → fully **offline** inference. |
| **P4** | Optional, opt-in: sync distilled "wisdom" to a central collective brain — sovereignty stays the default; collective evolution is a choice, not a leak. |

## Why Qwen as the default local model
Apache-2.0 (safe to ship commercially, unlike Llama's license), strong multilingual
(Indonesian + English), good agentic/tool-use and coding. `7B-Q4_K_M ≈ 4.5GB`, CPU-feasible;
a `3B` fallback covers weaker hardware.

## Build
```sh
build/build-router.sh        # -> out/flowork-router  (static linux/amd64)
```
Source resolves automatically: `$ROUTER_SRC` env → sibling `../flowork_Router` → GitHub clone.
