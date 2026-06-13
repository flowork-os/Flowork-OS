# Use it with Flowork Agents

Flow Router works with *any* OpenAI-compatible client. But it shines as the **brain** for
**[Flowork Agent](https://github.com/flowork-os/Flowork-OS)** — the sovereign agent OS, its purpose-built companion.

Point a Flowork agent's router endpoint at `http://127.0.0.1:2402/v1` and it automatically gets:

- **Model-agnostic routing** — swap models without touching the agent.
- **The anti-hallucination antibody** — the agent's own confirmed mistakes injected before each answer.
- **Sovereignty** — run fully local, offline, no third-party key.
- **Resilience** — fallback chains so the agent never stalls on a rate-limit.

**One brain (this router) + many bodies (your agents) = a full sovereign AI stack.**
