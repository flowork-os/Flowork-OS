# Use Flowork Router with ANY agent

Flowork Router is a standalone, **OpenAI- and Anthropic-compatible** LLM gateway. Whatever agent
framework you already use — **Hermes, OpenClaw, Cursor, Claude Code, Continue, Aider, LangChain,
or your own scripts** — just point its base URL at the router and you get:

- 🔌 **Cloak** — route through *your own* Claude subscription (one sub, many tools).
- 🔀 **Smart routing** — by cost, intent, or priority; add any provider (OpenAI, Gemini, DeepSeek, local Qwen…).
- 🧠 **Brain + 🕸️ mesh** — optional shared memory and a P2P knowledge network.
- 📊 **One place** for usage, pricing, keys, and limits.

## Run it (30 seconds)
```sh
docker run -d -p 2402:2402 -v flowork-router-data:/data ghcr.io/flowork-os/flowork-router
# or grab the binary from Releases:  ./flowork-router -addr 0.0.0.0:2402
```
Open `http://localhost:2402`, add a provider (paste your Claude token), and create an API key
(`flr_…`) under **Settings → API Keys**.

## Point your agent at it
The router speaks the standard APIs, so it's a drop-in base URL:

| Your tool | Set this |
|---|---|
| **OpenAI SDK / most agents** | `OPENAI_BASE_URL=http://localhost:2402/v1` · `OPENAI_API_KEY=flr_…` |
| **Anthropic SDK / Claude Code** | `ANTHROPIC_BASE_URL=http://localhost:2402` · `ANTHROPIC_API_KEY=flr_…` |
| **Cursor / Continue** | Custom OpenAI base URL → `http://localhost:2402/v1`, key `flr_…` |
| **Hermes / OpenClaw / LangChain / Aider** | their OpenAI-compatible endpoint → `http://localhost:2402/v1` |
| **curl** | `curl http://localhost:2402/v1/chat/completions -H "Authorization: Bearer flr_…" -d '{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"hi"}]}'` |

That's it — keep your existing agent, gain the router. Pairs perfectly with the full **Flowork**
agent + OS, but works great on its own.
