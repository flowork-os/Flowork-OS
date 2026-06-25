// Package main — CETAKAN agent baru Flowork ("pasukan semut"), BOOTSTRAP TIPIS.
//
// Ini template yang di-COPY mk-agent.sh/spawn-agent.sh tiap lahirin semut baru (copy
// main.go + go.mod + manifest.json, sed module→id). Seluruh loop tool-calling + guard
// (ghost/flail/recovery) + #2C deferred-tools seam + time-bound/auto-continue +
// router-retry ADA di package SHARED `flowork-agentkit` (keystone roadmap AGENTKIT) →
// SEMUA semut baru OTOMATIS warisan guard (dulu template GAK punya flail-guard+seam).
//
// Persona/model/tool = RUNTIME (FLOWORK_AGENT_ID + FLOWORK_AGENT_CONFIG GUI + self-prompt
// render host) → file ini SAMA buat semua; yang beda cuma manifest.json + config DB.
// Build: GOWORK=off + tinygo (scripts/build-agent.sh) ATAU GOOS=wasip1 GOARCH=wasm go build.
package main

import "flowork-agentkit"

func main() { agentkit.Main() }
