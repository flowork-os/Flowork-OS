# Flowork Router — Personal / Sovereign Build

The **sovereign** build of the Flowork LLM router — the one that ships inside
[**Flowork OS**](https://github.com/flowork-os/flowork), the Live-USB
sovereign appliance. Same router engine as upstream
[`flowork_Router`](https://github.com/flowork-os/Flowork-OS), tuned for one thing:
**your prompts and data never leave the box.**

```
Agent (:1987) ──LLM call──▶ Router (:2402) ──▶ Ollama (local Qwen, :11434)   [default: sovereign]
                                            └──▶ Cloud API (optional, opt-in, owner-supplied key)
```

## What's in here
| Path | What |
|------|------|
| `appliance/router-config.sovereign.seed.json` | Default config: **local Ollama only**, cloud opt-in, zero baked-in secrets. |
| `appliance/flowork-router.openrc` | The OpenRC service unit that runs the router on the appliance (`:2402`). |
| `build/build-router.sh` | Build the router as a **static** linux/amd64 binary (the appliance form). |
| `docs/SOVEREIGN.md` | Why local-first, why a separate binary, roadmap alignment, model choice. |

## Quick start
```sh
build/build-router.sh        # -> out/flowork-router (static, no libc/Docker/Python)
./out/flowork-router -addr 127.0.0.1:2402
```

## Relationship to the other repos
- **Canonical engine:** `flowork-os/flowork_Router` (full upstream router).
- **This repo:** the personal/sovereign profile + appliance service + build, consumed by the
  OS builder. The build pulls the engine source (sibling checkout or GitHub) and produces the
  static binary; the OS image runs it co-located with the agent.

> Sovereignty first. Collective evolution later, and only if you opt in. See `docs/SOVEREIGN.md`.
