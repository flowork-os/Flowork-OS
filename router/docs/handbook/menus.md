# Menus

Every item in the left sidebar, by group. Stuck on one? Jump to its section.

## Core
- **💎 Endpoint** — connection details for external tools: the OpenAI-compatible URL to copy into
  Cursor, Codex, Claude Code, or any client.
- **💬 Chat** — a quick test bench: send a prompt through the router to any configured model (with
  streaming) to confirm your setup.

## General
- **🔌 Providers** — connect any LLM provider from preset cards (search + filter); this is where your
  models come from.
- **🧩 Combos** — group several models into one alias that auto-routes between them (priority /
  round-robin / random / cost-optimal), with per-model fallback.
- **📊 Usage** — token consumption and cost, per provider, per day.
- **⏱️ Quota Tracker** — token consumption and cost per provider over today / last 7 / last 30 days.
- **⌨️ CLI Tools** — detect installed AI CLIs and point them at Flow Router in one click.
- **📥 OAuth Imports** — auto-detect credential files from CLI tools you are already logged into, so
  you can reuse them.

## Infrastructure
- **🚇 Tunnel** — expose Flow Router to the internet via Cloudflare or Tailscale for secure remote
  access; a watchdog probes health every 60s.
- **🗂️ Models** — manage model aliases, disabled models, and custom models.
- **💲 Pricing** — rate cards per model (USD per 1M tokens), used for cost estimates and cost-optimal
  combos.
- **🏷️ Tags** — create colored labels and attach them to providers (in a provider's Advanced
  settings).
- **🔄 Translator** — convert a request between OpenAI ⇄ Anthropic ⇄ Gemini formats, or send it live
  and see the response.
- **🧩 MCP Servers** — register Model Context Protocol servers and list their tools live.

## System
- **🎨 Media Providers** — generate images, audio, embeddings and more: Embedding (text → vector for
  RAG), Text-to-Image, Text-to-Speech, Speech-to-Text, and Web Fetch & Search.
- **🌐 Proxy Pools** — HTTP/SOCKS5 outbound proxies with rotation, for privacy, geo-bypass, and
  multi-account distribution.
- **🔑 API Keys** — Flow Router's own keys for client auth (format `flr_…`), so your tools
  authenticate to the router.
- **🧠 Skills** — reusable prompt templates with variables; clients invoke them by skill name plus
  variables.
- **🧬 Brain** — the server-side knowledge brain: overview and search of the Memory Palace (see
  [brain-and-antibody.md](brain-and-antibody.md)).
- **🕵️ MITM Proxy** — a local HTTPS interceptor for AI-coding IDEs (Antigravity, Copilot, Cursor,
  Kiro); generates a per-machine root cert.
- **🕸️ Mesh & Policy** — the operations dashboard for the mesh stack, LLM orchestration, pricing, and
  policy guardrails; shows your mesh identity.
- **📜 Console Log** — a live request log: recent dispatches with provider, tokens, latency, and cost
  (optional body capture).
- **⚙️ Settings** — router configuration (auto-saved): dispatch behaviour, default model, and more.
