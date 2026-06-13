// Claude per-device OAuth login HTTP handlers.
//
// Lets a device with NO Claude Code (USB appliance / Android) log into Claude on its OWN — getting an
// independent refresh token instead of sharing (and rotating) the desktop's. Two steps: /start
// returns the authorize URL + stashes the PKCE verifier; /complete takes the pasted code and swaps it
// for tokens (persisted via creds.SaveClaude, which the dispatcher reads). New routes, no edit to the
// locked oauth handler.

package main

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flowork-os/flowork_Router/internal/creds"
	"github.com/flowork-os/flowork_Router/internal/store"
)

const claudeLoginPending = "claude:login-pending"

// claudeLoginStartHandler — POST /api/claude-login/start. Generates PKCE + state, stores them, and
// returns the authorize URL the owner opens to log THIS device into Claude independently.
func claudeLoginStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	verifier, challenge, err := creds.PKCEPair()
	if err != nil {
		http.Error(w, "pkce: "+err.Error(), http.StatusInternalServerError)
		return
	}
	state, err := creds.RandomState()
	if err != nil {
		http.Error(w, "state: "+err.Error(), http.StatusInternalServerError)
		return
	}
	d, _ := store.Open()
	_ = store.UpsertOAuthToken(d, &store.OAuthTokenRecord{
		Provider:  claudeLoginPending,
		TokenType: "pkce-pending",
		Extra: map[string]any{
			"verifier":  verifier,
			"state":     state,
			"expiresAt": time.Now().Add(10 * time.Minute).Format(time.RFC3339),
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"authUrl": creds.ClaudeAuthorizeURL(challenge, state),
		"state":   state,
		"note":    "Open authUrl, sign in, authorize, then paste the shown code (code#state) into Complete.",
	})
}

// claudeLoginCompleteHandler — POST /api/claude-login/complete { code }. Accepts the pasted code
// (which may arrive as "code#state"), validates state, exchanges it for an independent token pair,
// and persists it (SaveClaude → dispatcher; also recorded in the KV store).
func claudeLoginCompleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "parse: "+err.Error(), http.StatusBadRequest)
		return
	}
	code := strings.TrimSpace(body.Code)
	state := strings.TrimSpace(body.State)
	// The callback page shows "code#state" — accept that whole string pasted into the code field.
	if i := strings.Index(code, "#"); i >= 0 {
		if state == "" {
			state = code[i+1:]
		}
		code = code[:i]
	}
	if code == "" {
		http.Error(w, "authorization code required", http.StatusBadRequest)
		return
	}

	d, _ := store.Open()
	pending, _ := store.GetOAuthToken(d, claudeLoginPending)
	if pending == nil {
		http.Error(w, "no pending Claude login — click Start first", http.StatusBadRequest)
		return
	}
	extra, _ := pending.Extra.(map[string]any)
	verifier, _ := extra["verifier"].(string)
	storedState, _ := extra["state"].(string)
	if verifier == "" || storedState == "" {
		http.Error(w, "pending login malformed — click Start again", http.StatusBadRequest)
		return
	}
	// CSRF: when the pasted blob carried a state, it MUST match (constant-time). PKCE verifier binding
	// is the primary protection; the state check is belt-and-suspenders.
	if state != "" && subtle.ConstantTimeCompare([]byte(state), []byte(storedState)) != 1 {
		http.Error(w, "state mismatch — restart the login", http.StatusBadRequest)
		return
	}

	access, refresh, expMs, err := creds.ExchangeClaudeCode(code, verifier, storedState)
	if err != nil {
		http.Error(w, "exchange failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	expStr := ""
	if expMs > 0 {
		expStr = strconv.FormatInt(expMs, 10)
	}
	if err := creds.SaveClaude(access, refresh, expStr); err != nil {
		http.Error(w, "save: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = store.UpsertOAuthToken(d, &store.OAuthTokenRecord{
		Provider: "claude", AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", Scope: "device-login",
	})
	_ = store.DeleteOAuthToken(d, claudeLoginPending)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "provider": "claude", "loggedIn": true})
}
