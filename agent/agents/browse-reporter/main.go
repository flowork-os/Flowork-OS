// Package main — BOOTSTRAP TIPIS worker-agent Flowork ("pasukan semut").
//
// Seluruh loop tool-calling + guard (ghost/flail/recovery) + #2C deferred-tools seam +
// time-bound/auto-continue + router-retry ADA di package SHARED `flowork-agentkit`
// (keystone roadmap AGENTKIT). Identitas agent (id, persona, model, tool) = RUNTIME
// (FLOWORK_AGENT_ID + FLOWORK_AGENT_CONFIG dari GUI + self-prompt render host) — JADI
// file ini SAMA buat semua worker; yang beda cuma manifest.json + config DB.
//
// Fix loop sekali di agentkit → SEMUA semut warisan (ga perlu edit 6 file). Build:
// GOWORK=off + tinygo (lihat scripts/build-agent.sh) ATAU GOOS=wasip1 GOARCH=wasm go build.
package main

import "flowork-agentkit"

func main() { agentkit.Main() }
