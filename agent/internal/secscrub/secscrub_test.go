package secscrub

import (
	"strings"
	"testing"
)

// Fixture rahasia PALSU dibangun runtime (concat) biar literal pola secret
// (anthropic/github/aws) TIDAK muncul utuh di source → scanner ga false-alarm,
// tapi string runtime-nya tetap valid buat nguji redaksi.
var (
	fakeAnt  = "sk-" + "ant-api03-abcdefghijklmnop"
	fakeAntT = "sk-" + "ant-api03-abcdefghijklmnopWXYZ"
	fakeAnt2 = "sk-" + "ant-bocordisinipanjang123"
	fakeGHP  = "ghp" + "_ABCDEFGHIJKLMNOPQRSTUV123456"
	fakeAWS  = "AKIA" + "IOSFODNN7EXAMPLE"
)

func TestRedactString_Tokens(t *testing.T) {
	cases := map[string]string{ // input → substring yang HARUS ilang
		"pakai " + fakeAnt + " buat call":         fakeAnt,
		"token gw " + fakeGHP + " jangan bocor":   fakeGHP,
		"aws " + fakeAWS + " dipakai":             fakeAWS,
		"hdr Authorization: Bearer abcdef1234567890XYZT": "Bearer abcdef1234567890XYZT",
		`config: {"api_key": "supersecret12345"}`:        "supersecret12345",
		"password=RahasiaBanget99 di log":                "RahasiaBanget99",
	}
	for in, gone := range cases {
		out := RedactString(in)
		if strings.Contains(out, gone) {
			t.Errorf("masih bocor %q di %q", gone, out)
		}
		if !strings.Contains(out, "REDACTED") {
			t.Errorf("ga ada marker REDACTED di %q", out)
		}
	}
	// Teks normal utuh.
	plain := "halo bro, tolong build ./... terus laporin hasilnya"
	if RedactString(plain) != plain {
		t.Error("teks normal ga boleh berubah")
	}
}

func TestRedactString_ClipKeepsTail(t *testing.T) {
	out := RedactString(fakeAntT)
	if !strings.Contains(out, "WXYZ") {
		t.Errorf("4 char terakhir harus disisain buat korelasi: %q", out)
	}
}

func TestRedact_RecursiveMapSlice(t *testing.T) {
	in := map[string]any{
		"caller": "owner",
		"token":  "nilai-token-panjang-sekali",
		"nested": map[string]any{
			"password": "jangansampekeliatan",
			"note":     "pakai " + fakeGHP + " ya",
			"list":     []any{"aman", fakeAnt2},
		},
	}
	out := Redact(in).(map[string]any)
	if s := out["token"].(string); !strings.Contains(s, "REDACTED") {
		t.Errorf("key 'token' harus ke-redact: %q", s)
	}
	nested := out["nested"].(map[string]any)
	if s := nested["password"].(string); !strings.Contains(s, "REDACTED") {
		t.Errorf("nested password harus ke-redact: %q", s)
	}
	if s := nested["note"].(string); strings.Contains(s, fakeGHP) {
		t.Errorf("token di string nested harus ke-redact: %q", s)
	}
	if s := nested["list"].([]any)[1].(string); strings.Contains(s, fakeAnt2) {
		t.Errorf("token di slice harus ke-redact: %q", s)
	}
	if nested["list"].([]any)[0].(string) != "aman" {
		t.Error("elemen aman ga boleh berubah")
	}
	// Non-destruktif: input asli utuh.
	if in["token"].(string) != "nilai-token-panjang-sekali" {
		t.Error("Redact harus COPY, bukan mutasi input")
	}
}

func TestRedactMap_NilAndDepth(t *testing.T) {
	if RedactMap(nil) != nil {
		t.Error("nil in → nil out")
	}
	deep := map[string]any{}
	cur := deep
	for i := 0; i < 30; i++ { // > maxDepth → ga boleh panic/hang
		next := map[string]any{}
		cur["d"] = next
		cur = next
	}
	cur["token"] = "x"
	_ = RedactMap(deep)
}
