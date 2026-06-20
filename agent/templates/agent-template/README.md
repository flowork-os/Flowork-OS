# 🧬 agent-template — Cetakan Warga Flowork

Template bersih buat bikin agent baru. **WAJIB baca [`AGENT_STANDARD.md`](AGENT_STANDARD.md)** — itu kontrak/alur/standar (DNA) tiap agent. Diambil & didistilasi dari roadmap, divalidasi live 2026-06-20.

## Isi folder
- **`AGENT_STANDARD.md`** — doktrin lengkap: 8 syarat wajib, autonomy (loop/sleep/bangun/anti-ghosting), GUI standar, alur pesan end-to-end, pantangan.
- **`manifest.json`** — skeleton manifest (ganti `TEMPLATE_AGENT_ID`/`display_name`/`description`; sesuaikan `capabilities_required`).

## Cara pakai (manual, sampai AI Studio diwire)
1. Copy folder ini → `agents/<id-baru>/`.
2. Edit `manifest.json` (id, display_name, description, caps).
3. Tulis `main.go` ikut pola **`agents/mr-flow/main.go`** (reference: DB-based, buildSystemPrompt 3-tier, tool-loop + ghost-guard, fetchHistory + anti-anchor). JANGAN pola .md.
4. Build: `GOWORK=off GOOS=wasip1 GOARCH=wasm go build -o <target>/agent.wasm .` (standard wasip1, ~3.5MB — BUKAN tinygo).
5. Boot → `ProvisionAgentDNA(id)` otomatis nyuntik DNA (konstitusi sacred + instinct + cognitive graph schema). Genome pipe (`brain_search_shared`/`instinct_recall`/`graph_recall`) auto ke-expose.
6. **QC WAJIB:** debug-chat harness (jalur Telegram) + GUI login+screenshot. No claim tanpa test.

## Status integrasi
- AI Studio (`coderTemplate()`) + Evolution = **wiring nanti** (owner: "AI Studio/Evolution nanti saja"). Saat itu: arahin `coderTemplate` ke template ini, dan pastiin agent produksi AI Studio lahir full-DNA (lihat AGENT_STANDARD §0).

> Reference agent yang udah lulus semua standar: **`agents/mr-flow/`**.
