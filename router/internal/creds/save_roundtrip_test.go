package creds

import (
	"os"
	"testing"
)

// TestSaveClaudeRoundTrip proves a GUI-pasted Claude token is readable by the SAME Load() the
// dispatcher uses — the load-bearing guarantee for Android / USB appliance (no Claude Code present).
func TestSaveClaudeRoundTrip(t *testing.T) {
	tmp := t.TempDir() + "/creds.json"
	t.Setenv("FLOW_CREDS_PATH", tmp)

	if err := SaveClaude("sk-ant-oat-TESTTOKEN123", "rt-x", ""); err != nil {
		t.Fatalf("SaveClaude: %v", err)
	}
	c, err := Load()
	if err != nil {
		t.Fatalf("Load after SaveClaude: %v", err)
	}
	if got := c.ClaudeAiOauth.AccessToken; got != "sk-ant-oat-TESTTOKEN123" {
		t.Fatalf("token mismatch: %q", got)
	}
	if c.IsExpired() {
		t.Fatal("pasted token reads as expired — far-future expiry should prevent this")
	}
	if fi, _ := os.Stat(tmp); fi != nil && fi.Mode().Perm() != 0o600 {
		t.Fatalf("credential file mode = %o, want 0600", fi.Mode().Perm())
	}
	// Empty token must be rejected (no half-written credential file).
	if err := SaveClaude("   ", "", ""); err == nil {
		t.Fatal("SaveClaude accepted an empty token — should error")
	}
}
