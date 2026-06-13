package creds

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestExchangeClaudeCode proves the per-device login exchange: an authorization code + PKCE verifier
// are POSTed as an authorization_code grant and the returned token pair is parsed. Uses a local mock
// (never the real Anthropic endpoint, so it creates/rotates nothing).
func TestExchangeClaudeCode(t *testing.T) {
	var got map[string]string
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		gotUA = r.Header.Get("User-Agent")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "sk-ant-DEVICE-A",
			"refresh_token": "rt-DEVICE-A",
			"expires_in":    3600,
		})
	}))
	defer srv.Close()
	t.Setenv("ANTHROPIC_OAUTH_TOKEN_URL", srv.URL)
	t.Setenv("ANTHROPIC_CLIENT_ID", "client-x")

	access, refresh, expMs, err := ExchangeClaudeCode("the-code", "the-verifier", "the-state")
	if err != nil {
		t.Fatalf("ExchangeClaudeCode: %v", err)
	}
	if access != "sk-ant-DEVICE-A" || refresh != "rt-DEVICE-A" || expMs <= 0 {
		t.Fatalf("bad exchange result: access=%q refresh=%q exp=%d", access, refresh, expMs)
	}
	if got["grant_type"] != "authorization_code" || got["code"] != "the-code" ||
		got["code_verifier"] != "the-verifier" || got["client_id"] != "client-x" {
		t.Fatalf("wrong token request: %+v", got)
	}
	// Anti-ban: the per-device login exchange must carry the Claude Code identity (User-Agent).
	if !strings.Contains(gotUA, "claude-cli") {
		t.Fatalf("login exchange missing Claude Code User-Agent (anti-ban): %q", gotUA)
	}
}

// TestPKCEAndAuthorizeURL — the verifier/challenge pair is well-formed and the authorize URL carries
// the S256 challenge + state + the configured scopes.
func TestPKCEAndAuthorizeURL(t *testing.T) {
	v, ch, err := PKCEPair()
	if err != nil || v == "" || ch == "" || v == ch {
		t.Fatalf("bad PKCE pair: v=%q ch=%q err=%v", v, ch, err)
	}
	u := ClaudeAuthorizeURL(ch, "STATE123")
	for _, want := range []string{"code_challenge=" + ch, "code_challenge_method=S256", "state=STATE123", "response_type=code"} {
		if !strings.Contains(u, want) {
			t.Fatalf("authorize URL missing %q: %s", want, u)
		}
	}
}
