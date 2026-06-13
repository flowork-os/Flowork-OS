package creds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SaveClaude writes a pasted Claude OAuth token to the SAME credential file that Load() reads
// (FLOW_CREDS_PATH, else ~/.claude/.credentials.json). This is what makes a GUI-pasted Claude token
// actually work on a machine that has NO Claude Code installed — Android and the USB appliance —
// where there is no existing ~/.claude/.credentials.json to borrow. Without this, the Claude
// subscription provider fails auth ("creds: no such file") and the cloak has nothing to send.
//
// accessToken is required. expiresAt may be "" or a unix-millisecond string; if absent/unparseable a
// far-future expiry is written so a long-lived pasted token is not immediately treated as expired by
// IsExpired(). Existing fields in the file (org UUID, scopes) are preserved. File mode is 0600 (same
// as Claude Code writes); on the appliance the credential path lives on the LUKS-encrypted DATA
// partition, so the token is encrypted at rest there.
func SaveClaude(accessToken, refreshToken, expiresAt string) error {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return os.ErrInvalid
	}
	p := credentialsPath()

	var cf CredentialsFile
	if raw, err := os.ReadFile(p); err == nil { // preserve any pre-existing fields
		_ = json.Unmarshal(raw, &cf)
	}

	exp := int64(9999999999999) // ~year 2286 — a long-lived pasted token shouldn't read as expired
	if v, err := strconv.ParseInt(strings.TrimSpace(expiresAt), 10, 64); err == nil && v > 0 {
		exp = v
	}

	cf.ClaudeAiOauth.AccessToken = accessToken
	if rt := strings.TrimSpace(refreshToken); rt != "" {
		cf.ClaudeAiOauth.RefreshToken = rt
	}
	cf.ClaudeAiOauth.ExpiresAt = exp
	if cf.ClaudeAiOauth.SubscriptionType == "" {
		cf.ClaudeAiOauth.SubscriptionType = "max"
	}

	out, err := json.MarshalIndent(&cf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(p, out, 0o600); err != nil {
		return err
	}
	// A fresh token just landed on disk — drop the 30s read cache so the NEXT Load() (the dispatcher
	// authenticating the very next request, or a refresh persisting a rotated token) reflects it
	// immediately instead of serving a stale copy.
	InvalidateCache()
	return nil
}
