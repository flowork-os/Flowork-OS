package store

import "testing"

// TestUpsertProviderPreservesKeyOnBlankOrMasked locks in the read-mask + write-preserve contract:
// the read API never emits a plaintext key, so a save that round-trips a BLANK or MASKED key must
// keep the stored secret (not wipe it or store dots). A genuinely new key still replaces it.
func TestUpsertProviderPreservesKeyOnBlankOrMasked(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("FLOW_ROUTER_DATA", tmp)
	resetDBSingletonForTest()
	d, err := Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	const real = "sk-REAL-SECRET-123456"
	p := &ProviderConnection{
		Provider: "openai", AuthType: AuthTypeAPIKey, Name: "T", IsActive: true,
		Data: map[string]any{CfgAPIKey: real, CfgBaseURL: "https://x/v1", CfgModels: []any{"*"}},
	}
	if err := UpsertProvider(d, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	id := p.ID

	// 1) Blank key on update → keep the real key, still apply other field changes.
	if err := UpsertProvider(d, &ProviderConnection{ID: id, Provider: "openai", AuthType: AuthTypeAPIKey, Name: "T2",
		IsActive: true, Data: map[string]any{CfgBaseURL: "https://x/v1", CfgModels: []any{"*"}}}); err != nil {
		t.Fatalf("update-blank: %v", err)
	}
	got, _ := GetProvider(d, id)
	if got == nil || got.Data[CfgAPIKey] != real {
		t.Fatalf("blank update wiped key: %v", got.Data[CfgAPIKey])
	}
	if got.Name != "T2" {
		t.Fatalf("name not updated: %q", got.Name)
	}

	// 2) Masked key (contains •) on update → keep the real key.
	if err := UpsertProvider(d, &ProviderConnection{ID: id, Provider: "openai", AuthType: AuthTypeAPIKey, Name: "T3",
		IsActive: true, Data: map[string]any{CfgAPIKey: "sk-R••••••••3456", CfgBaseURL: "https://x/v1", CfgModels: []any{"*"}}}); err != nil {
		t.Fatalf("update-masked: %v", err)
	}
	got2, _ := GetProvider(d, id)
	if got2.Data[CfgAPIKey] != real {
		t.Fatalf("masked update wiped key: %v", got2.Data[CfgAPIKey])
	}

	// 3) A genuinely new key replaces the old one.
	if err := UpsertProvider(d, &ProviderConnection{ID: id, Provider: "openai", AuthType: AuthTypeAPIKey, Name: "T4",
		IsActive: true, Data: map[string]any{CfgAPIKey: "sk-NEW-KEY-999", CfgBaseURL: "https://x/v1", CfgModels: []any{"*"}}}); err != nil {
		t.Fatalf("update-new: %v", err)
	}
	got3, _ := GetProvider(d, id)
	if got3.Data[CfgAPIKey] != "sk-NEW-KEY-999" {
		t.Fatalf("new key not stored: %v", got3.Data[CfgAPIKey])
	}
}
