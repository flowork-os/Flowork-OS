# Getting Started

## What Flow Router is

Flow Router is a **sovereign LLM gateway**: one OpenAI-compatible endpoint —
`http://127.0.0.1:2402/v1` — that every tool and agent can point at, and it routes the call to
whatever provider you choose (a subscription you already pay for, or a fully local model). It speaks
OpenAI, Anthropic, and Gemini formats at once, falls back automatically when a provider rate-limits,
trims tokens, and can run a P2P mesh so your stack survives offline. **One Go binary — no Docker, no
Python, no database.**

Same job as the popular gateways — route every provider through one endpoint — but a single binary
that also ships anti-ban cloaking, a token-saver, a shared brain, and a sovereign mesh.

## Why a router

- **One endpoint, every model** — point any OpenAI-compatible client (an IDE, a CLI, a Flowork agent)
  at one URL and reach any provider.
- **Use what you already pay for** — drive workloads through an existing subscription, no extra API key.
- **Never stop** — priority → round-robin → cost-optimal fallback chains with a cooldown/backoff
  table; one rate-limit just rolls to the next provider.
- **Cut tokens** — a token-saver trims a large share off agent loops.
- **Survive anything** — an optional P2P mesh replicates knowledge host-to-host, internet-optional.
- **Zero ops** — one binary, no runtime, no DB; runs on a Raspberry Pi.

## Install

```
git clone https://github.com/flowork-os/Flowork-OS.git
cd flowork_Router
./start.sh
```

`start.sh` builds the binary on first run (needs **Go 1.25+**; it auto-detects Go from common install
dirs) and serves the control panel + API on `http://127.0.0.1:2402`. Stop with `./stop.sh`.

Then point any tool at the endpoint:

```
OpenAI base URL : http://127.0.0.1:2402/v1
Anthropic base  : http://127.0.0.1:2402
Gemini base     : http://127.0.0.1:2402
```

- Runs on **Linux, macOS, and Windows**. No Docker, no Python, no database server.
- First run: open the panel → connect a provider (*Providers*) → test it (*Chat*).
