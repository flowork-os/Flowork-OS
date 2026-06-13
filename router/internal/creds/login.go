// Claude Per-Device OAuth Login (authorization-code + PKCE).
//
// To run Claude on a device that has NO Claude Code — the USB appliance, Android — WITHOUT borrowing
// (and fighting over) the desktop's shared refresh token, each device performs its OWN login: it
// gets an INDEPENDENT refresh token that it can rotate on its own, so no device ever invalidates
// another's credential. This is the "login sendiri per device" model.
//
// Flow (the same manual-code flow `claude login` uses on a headless box):
//  1. ClaudeAuthorizeURL(challenge,state) → owner opens it, signs in, authorizes.
//  2. claude.ai redirects to the code-callback page which DISPLAYS "code#state".
//  3. Owner pastes that back; ExchangeClaudeCode swaps it (with the PKCE verifier) for tokens.
//
// All endpoints/ids are the public Claude Code values, env-overridable so a future change needs no
// rebuild. Never logs token material.

package creds

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"os"
	"strings"
)

func claudeAuthURL() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_OAUTH_AUTH_URL")); v != "" {
		return v
	}
	return "https://claude.ai/oauth/authorize"
}

func claudeRedirectURI() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_OAUTH_REDIRECT_URI")); v != "" {
		return v
	}
	return "https://console.anthropic.com/oauth/code/callback"
}

func claudeScopes() string {
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_OAUTH_SCOPES")); v != "" {
		return v
	}
	return "org:create_api_key user:profile user:inference"
}

// PKCEPair returns a fresh code_verifier and its S256 code_challenge.
func PKCEPair() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// RandomState returns a high-entropy opaque state value for CSRF protection.
func RandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ClaudeAuthorizeURL builds the authorize URL the owner opens to log this device in independently.
func ClaudeAuthorizeURL(challenge, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", claudeClientID())
	q.Set("redirect_uri", claudeRedirectURI())
	q.Set("scope", claudeScopes())
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	return claudeAuthURL() + "?" + q.Encode()
}

// ExchangeClaudeCode swaps a pasted authorization code (+ the stored PKCE verifier) for a fresh,
// independent access/refresh token pair. Returns expiry in unix-milliseconds.
func ExchangeClaudeCode(code, verifier, state string) (access, refresh string, expiresAtMs int64, err error) {
	return postClaudeToken(map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"state":         state,
		"client_id":     claudeClientID(),
		"redirect_uri":  claudeRedirectURI(),
		"code_verifier": verifier,
	})
}
