// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-12
// Reason: First-run config seed — ExportSeed (redacted snapshot) + SeedFromBundleJSON
//   (import on a fresh DB only). Audited: all secrets blanked, never clobbers an
//   existing install. Verified build + store tests green.
//
// seed.go — first-run config seed (redacted export + import).
//
// Purpose: ship the OWNER's exact current config so a fresh download is
// pre-configured ("install = my setup, just paste your token"). ExportSeed
// produces a SyncBundle with EVERY secret blanked, safe to commit to a public
// repo; the binary imports it on first boot (see main.go) when the DB has no
// providers yet. Structure (provider names, base URLs, models, combos, pricing,
// dispatch settings) is preserved exactly — only secrets are stripped.
package store

import (
	"database/sql"
	"encoding/json"
	"strings"
)

// parseSyncBundle decodes a seed/export JSON into a SyncBundle.
func parseSyncBundle(raw []byte) (*SyncBundle, error) {
	var b SyncBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// SeedKeyPlaceholder is what an api_key provider's key becomes in the seed.
const SeedKeyPlaceholder = "PASTE_YOUR_KEY_HERE"

// ExportSeed = ExportConfig with all secrets blanked. Reproduces the owner's
// setup on a fresh install; the user only re-enters their own tokens.
func ExportSeed(d *sql.DB) *SyncBundle {
	b := ExportConfig(d)

	for i := range b.Providers {
		// Email is owner PII — never ship it.
		b.Providers[i].Email = ""
		data := b.Providers[i].Data
		if data == nil {
			continue
		}
		// apiKey is decrypted in-memory by ListProviders — blank it.
		if v, ok := data[CfgAPIKey].(string); ok && v != "" {
			if b.Providers[i].AuthType == AuthTypeAPIKey {
				data[CfgAPIKey] = SeedKeyPlaceholder
			} else {
				data[CfgAPIKey] = ""
			}
		}
		// Strip any auth-bearing custom headers.
		if hdr, ok := data[CfgHeaders].(map[string]any); ok {
			for k := range hdr {
				lk := strings.ToLower(k)
				if strings.Contains(lk, "authorization") || strings.Contains(lk, "api-key") ||
					strings.Contains(lk, "cookie") || strings.Contains(lk, "token") {
					hdr[k] = ""
				}
			}
		}
	}

	for i := range b.MediaProviders {
		if b.MediaProviders[i].APIKey != "" {
			b.MediaProviders[i].APIKey = SeedKeyPlaceholder
		}
	}

	// MCP env vars commonly hold tokens (e.g. GITHUB_TOKEN) — keep the keys as
	// placeholders so the user knows what to fill, drop the values.
	for i := range b.MCPServers {
		for k := range b.MCPServers[i].Env {
			b.MCPServers[i].Env[k] = SeedKeyPlaceholder
		}
	}

	// Proxy pool URLs embed user:pass credentials — never ship them.
	b.ProxyPools = nil

	return b
}

// SeedFromBundleJSON imports a redacted seed bundle on first run, but ONLY when
// the DB has no providers yet (fresh install) — it never clobbers an existing
// setup. Returns the per-section applied counts, or nil when skipped.
func SeedFromBundleJSON(d *sql.DB, raw []byte) map[string]int {
	if len(raw) == 0 {
		return nil
	}
	existing, _ := ListProviders(d)
	if len(existing) > 0 {
		return nil // already configured — leave it alone
	}
	b, err := parseSyncBundle(raw)
	if err != nil || b == nil || len(b.Providers) == 0 {
		return nil
	}
	return ImportConfig(d, b)
}
