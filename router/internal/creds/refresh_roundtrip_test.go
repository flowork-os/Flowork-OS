package creds

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// writeCredsFile drops a minimal credentials.json with the given token material + expiry (unix-ms).
func writeCredsFile(t *testing.T, path, access, refresh string, expMs int64) {
	t.Helper()
	var cf CredentialsFile
	cf.ClaudeAiOauth.AccessToken = access
	cf.ClaudeAiOauth.RefreshToken = refresh
	cf.ClaudeAiOauth.ExpiresAt = expMs
	b, _ := json.Marshal(&cf)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestLoadValidRefreshesExpired proves the appliance-critical path: an EXPIRED subscription token
// with a refresh token is transparently refreshed (OAuth refresh_token grant), the rotated token is
// persisted, and the caller gets a fresh, non-expired credential — all WITHOUT touching the real
// Anthropic endpoint (a local httptest mock stands in).
func TestLoadValidRefreshesExpired(t *testing.T) {
	var gotGrant, gotRefresh, gotClient, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotGrant, gotRefresh, gotClient = body["grant_type"], body["refresh_token"], body["client_id"]
		gotUA = r.Header.Get("User-Agent")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "sk-ant-NEW-ACCESS",
			"refresh_token": "rt-NEW",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()

	tmp := t.TempDir() + "/creds.json"
	t.Setenv("FLOW_CREDS_PATH", tmp)
	t.Setenv("ANTHROPIC_OAUTH_TOKEN_URL", srv.URL)
	t.Setenv("ANTHROPIC_CLIENT_ID", "test-client")

	writeCredsFile(t, tmp, "sk-ant-OLD", "rt-OLD", time.Now().Add(-time.Hour).UnixMilli())
	InvalidateCache()

	c, err := LoadValid()
	if err != nil {
		t.Fatalf("LoadValid: %v", err)
	}
	if c.ClaudeAiOauth.AccessToken != "sk-ant-NEW-ACCESS" {
		t.Fatalf("access token not refreshed: got %q", c.ClaudeAiOauth.AccessToken)
	}
	if c.IsExpired() {
		t.Fatal("refreshed token still reads as expired")
	}
	if gotGrant != "refresh_token" || gotRefresh != "rt-OLD" || gotClient != "test-client" {
		t.Fatalf("wrong request to token endpoint: grant=%q refresh=%q client=%q", gotGrant, gotRefresh, gotClient)
	}
	// Anti-ban: the refresh handshake must carry the Claude Code identity (User-Agent).
	if !strings.Contains(gotUA, "claude-cli") {
		t.Fatalf("refresh missing Claude Code User-Agent (anti-ban): %q", gotUA)
	}

	// Rotated refresh token must be persisted to disk for the next restart.
	InvalidateCache()
	c2, err := Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if c2.ClaudeAiOauth.RefreshToken != "rt-NEW" {
		t.Fatalf("rotated refresh token not persisted: got %q", c2.ClaudeAiOauth.RefreshToken)
	}
}

// TestLoadValidNoRefreshTokenErrors — an expired token with NO refresh token must fail with a clear,
// appliance-aware message (re-import via GUI), never silently send an expired bearer.
func TestLoadValidNoRefreshTokenErrors(t *testing.T) {
	tmp := t.TempDir() + "/creds.json"
	t.Setenv("FLOW_CREDS_PATH", tmp)
	writeCredsFile(t, tmp, "sk-ant-OLD", "", time.Now().Add(-time.Hour).UnixMilli())
	InvalidateCache()

	if _, err := LoadValid(); err == nil {
		t.Fatal("expected an error for an expired token with no refresh token")
	}
}

// TestLoadValidFreshIsNoOp — a non-expired token passes straight through, no network call.
func TestLoadValidFreshIsNoOp(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true }))
	defer srv.Close()

	tmp := t.TempDir() + "/creds.json"
	t.Setenv("FLOW_CREDS_PATH", tmp)
	t.Setenv("ANTHROPIC_OAUTH_TOKEN_URL", srv.URL)
	writeCredsFile(t, tmp, "sk-ant-FRESH", "rt-x", time.Now().Add(time.Hour).UnixMilli())
	InvalidateCache()

	c, err := LoadValid()
	if err != nil {
		t.Fatalf("LoadValid: %v", err)
	}
	if c.ClaudeAiOauth.AccessToken != "sk-ant-FRESH" {
		t.Fatalf("unexpected token: %q", c.ClaudeAiOauth.AccessToken)
	}
	if hit {
		t.Fatal("a fresh token must NOT trigger a refresh network call")
	}
}
