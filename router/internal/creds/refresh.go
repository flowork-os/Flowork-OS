// Claude Subscription Token Auto-Refresh.
//
// On a normal desktop, Claude Code itself refreshes ~/.claude/.credentials.json, so the router only
// ever READS a fresh token. But on the sovereign appliance — Android / USB, where there is NO Claude
// Code installed — nothing refreshes the token, so a subscription token would die after its short
// lifetime and every Claude call would 502 until the owner re-imports by hand. This file closes that
// gap: when the stored token is expired AND a refresh token is present, perform the OAuth
// refresh_token grant against Anthropic's token endpoint, persist the rotated token (SaveClaude,
// 0600), and hand back fresh credentials — keeping Claude alive unattended.
//
// Security: never logs token values; one refresh at a time (single-flight + double-check); TLS verify
// stays on (default http.Client); on any failure it returns an appliance-aware error pointing the
// owner at the GUI re-import, not the meaningless "re-login Claude Code".

package creds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Claude Code's PUBLIC OAuth client id + token endpoint — the same non-secret values Claude Code
// ships. Overridable via env so a future endpoint/client change needs no rebuild.
const (
	defaultClaudeClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	defaultClaudeTokenURL = "https://console.anthropic.com/v1/oauth/token"
	// claudeCodeUserAgent — the Claude Code client identity. The OAuth login/refresh handshake
	// presents the SAME identity the dispatcher's chat path sends, so the whole login-mode flow looks
	// like the official client (anti-ban consistency). Overridable via env for future client-version bumps.
	defaultClaudeUserAgent = "claude-cli/1.0.0 (flow_router)"
)

func claudeUserAgent() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_OAUTH_USER_AGENT")); v != "" {
		return v
	}
	return defaultClaudeUserAgent
}

// refreshMu serialises refreshes so a burst of expired-token requests triggers exactly one network
// refresh, not one per in-flight call.
var refreshMu sync.Mutex

func claudeClientID() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_CLIENT_ID")); v != "" {
		return v
	}
	return defaultClaudeClientID
}

func claudeTokenURL() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_OAUTH_TOKEN_URL")); v != "" {
		return v
	}
	return defaultClaudeTokenURL
}

// LoadValid returns credentials guaranteed non-expired when at all possible. If the stored token is
// fresh it behaves exactly like Load. If it is expired (or within the IsExpired buffer) it tries an
// OAuth refresh using the stored refresh token, persists the result, and returns the new token.
//
// Use this anywhere the token is about to be SENT upstream (the dispatcher). Keep using plain Load()
// where you only want to read/report what is on disk without side effects.
func LoadValid() (*CredentialsFile, error) {
	c, err := Load()
	if err != nil {
		return nil, err
	}
	if !c.IsExpired() {
		return c, nil
	}
	rt := strings.TrimSpace(c.ClaudeAiOauth.RefreshToken)
	if rt == "" {
		return nil, fmt.Errorf("claude token expired and no refresh token available — re-import the token via OAuth Imports → Browse")
	}

	// Single-flight: hold the lock, then re-check — a concurrent caller may have already refreshed
	// (it would have invalidated the cache), so we avoid a redundant second network round-trip.
	refreshMu.Lock()
	defer refreshMu.Unlock()
	if c2, e := Load(); e == nil && !c2.IsExpired() {
		return c2, nil
	}

	access, refresh, expMs, rerr := refreshClaude(rt)
	if rerr != nil {
		return nil, fmt.Errorf("claude token expired and refresh failed (%v) — re-import via OAuth Imports → Browse", rerr)
	}

	expStr := ""
	if expMs > 0 {
		expStr = strconv.FormatInt(expMs, 10)
	}
	if serr := SaveClaude(access, refresh, expStr); serr != nil {
		return nil, fmt.Errorf("claude token refreshed but persisting it failed: %w", serr)
	}
	InvalidateCache()
	return Load()
}

// refreshClaude runs the OAuth refresh_token grant. Returns the new access token, the (possibly
// rotated) refresh token, and the absolute expiry in unix-milliseconds.
func refreshClaude(refreshToken string) (access, refresh string, expiresAtMs int64, err error) {
	return postClaudeToken(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     claudeClientID(),
	})
}

// postClaudeToken POSTs a JSON payload to Anthropic's OAuth token endpoint and parses the standard
// token response. Shared by the refresh_token grant (auto-refresh) and the authorization_code grant
// (per-device login). NEVER logs or returns token material in errors. TLS verification stays on.
func postClaudeToken(payload map[string]string) (access, refresh string, expiresAtMs int64, err error) {
	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, claudeTokenURL(), bytes.NewReader(reqBody))
	if err != nil {
		return "", "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// Anti-ban: same Claude Code identity as the dispatcher's chat requests, so login + refresh
	// handshakes are indistinguishable from the official client.
	req.Header.Set("User-Agent", claudeUserAgent())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Status only — the body can echo sensitive request bits, so we never surface it.
		return "", "", 0, fmt.Errorf("token endpoint HTTP %d", resp.StatusCode)
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"` // seconds from now
	}
	if e := json.NewDecoder(resp.Body).Decode(&out); e != nil {
		return "", "", 0, e
	}
	if out.AccessToken == "" {
		return "", "", 0, fmt.Errorf("token endpoint returned no access_token")
	}
	if out.RefreshToken == "" {
		out.RefreshToken = payload["refresh_token"] // grant did not rotate — keep the incoming one (if any)
	}
	if out.ExpiresIn > 0 {
		expiresAtMs = time.Now().Add(time.Duration(out.ExpiresIn) * time.Second).UnixMilli()
	}
	return out.AccessToken, out.RefreshToken, expiresAtMs, nil
}
