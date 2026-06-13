# 🧠 Local models — `models/`

This folder is where you drop **local model weights** (GGUF files) for Flow Router's
built-in `llama-server` (llama.cpp) runtime. The weights themselves are **multi-GB**
and are **git-ignored** — only this README ships in the repo. Download a recommended
model below, drop it here, and point the router at it.

> You don't strictly need any local model: Flow Router happily drives your
> **Claude/OpenAI/Gemini/…** providers. Local models are for **offline / zero-cost /
> sovereign** operation (and they run on a Raspberry Pi). This is opt-in.

---

## Two ways to run a local model

### A) Ollama (easiest) — recommended start
[Ollama](https://ollama.com) downloads, quantizes and serves the model for you, and
Flow Router has a native **`ollama-local`** provider (passthrough to `127.0.0.1:11434`).

```bash
# 1. install ollama (https://ollama.com), then pull a model:
ollama pull qwen3:8b

# 2. in the Flow Router dashboard → Providers → add an "ollama-local" provider
#    (base URL http://127.0.0.1:11434). Done — route to it like any model.
```

### B) Drop a GGUF here + the built-in `llama-server` runtime (advanced)
For full sovereignty (no Ollama daemon), put a `.gguf` in this folder and let the
router spawn `llama-server` (llama.cpp, port 8088):

```bash
# 1. get llama.cpp's llama-server on your PATH (or set its path in Settings)
# 2. download a GGUF (see table) into this folder, e.g. models/qwen3-8b-q4_k_m.gguf
# 3. dashboard → LocalAI runtime → register { model_name, gguf_path } and Start
```

**Quantization:** prefer **`Q4_K_M`** — ~75% smaller with almost no quality loss
(a 7–8B model needs ~4–5 GB RAM instead of ~16 GB). Use `Q5_K_M`/`Q6_K` if you have
RAM to spare, `Q3_K_M`/`Q2_K` only on very tight hardware.

---

## ✅ Recommended chat models (2026)

Pick by your RAM/VRAM budget. All are open-weight and run great as `Q4_K_M` GGUF via
Ollama or llama.cpp. Sizes are the approximate `Q4_K_M` download/RAM footprint.

| Model | ~Size (Q4_K_M) | Why / best for | Get it |
|---|---|---|---|
| **Qwen3 8B** ⭐ | ~5 GB | Best all-round local default in 2026 — strong reasoning **and** tool-calling, leads its class on code | `ollama pull qwen3:8b` · [HF GGUF](https://huggingface.co/Qwen) |
| **Llama 3.3 8B** | ~5 GB | Best general-purpose balance, very stable instruction-following | `ollama pull llama3.3` · [HF GGUF](https://huggingface.co/meta-llama) |
| **Qwen3 4B** | ~2.5 GB | Excellent quality-per-GB; great middle ground for 8 GB machines | `ollama pull qwen3:4b` |
| **Qwen3 1.7B** | ~1.2 GB | **Raspberry Pi / low-RAM** — usable agentic loops on a tiny box | `ollama pull qwen3:1.7b` |
| **Gemma 3 4B** | ~3 GB | Strong small model, good multilingual | `ollama pull gemma3:4b` |
| **gpt-oss 20B** | ~12 GB | OpenAI open-weight, **128K context** — if you have the RAM | `ollama pull gpt-oss:20b` |
| **Qwen3-Coder** | ~5 GB+ | **Coding/agentic** workloads (Claude Code / Cursor style) | `ollama pull qwen3-coder` |

**Rule of thumb:** ≤8 GB RAM → a 1.7B–4B model · 16 GB → 7–8B `Q4_K_M` · 32 GB+ →
20B-class or higher quant. For agent/tool-use workloads, prefer **Qwen3** (best
tool-calling of the small open models).

---

## 🧬 Embedding model (for the brain's semantic layer)

Flow Router's prebuilt brain stores its vectors as **`bge-m3`** (1024-dim). Keyword
retrieval (FTS5/BM25) needs **no** model, but to use **semantic** retrieval you need
an embedding endpoint — and it should **match the brain's vectors**:

| Model | Size | Notes |
|---|---|---|
| **bge-m3** ⭐ | ~1.2 GB | **Matches the prebuilt brain (1024-dim).** Use this if you downloaded the Drive brain. Dense + sparse + multi-vector, 100+ languages, MIT. | `ollama pull bge-m3` |
| **nomic-embed-text** | ~274 MB | Lightest good option, 8192-token context — fine for a brain you build yourself | `ollama pull nomic-embed-text` |
| **qwen3-embedding 4B/8B** | ~3–5 GB | Highest accuracy if you have a GPU | `ollama pull qwen3-embedding:4b` |

> ⚠️ If you mix embedding models, **re-embed** — query and stored vectors must come
> from the same model (and dimension) or semantic search returns garbage. The shipped
> brain is `bge-m3`/1024-dim.

See the prebuilt 5M-drawer brain in [`../brain/README.md`](../brain/README.md).

---

### Sources
- [The Best Open-Source / Open-Weight LLMs to Run Locally in 2026 — Hugging Face](https://huggingface.co/blog/daya-shankar/open-source-llm-models-to-run-locally)
- [Best Local LLM Models 2026 — SitePoint](https://www.sitepoint.com/best-local-llm-models-2026/)
- [Running LLMs Locally in 2026: Ollama, llama.cpp — daily.dev](https://daily.dev/blog/running-llms-locally-ollama-llama-cpp-self-hosted-ai-developers/)
- [Best Ollama Embedding Models 2026 — morphllm](https://www.morphllm.com/ollama-embedding-models)
- [The Best Open-Source Embedding Models in 2026 — BentoML](https://www.bentoml.com/blog/a-guide-to-open-source-embedding-models)
