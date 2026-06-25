// toolcall_recover_ext.go — GROWTH-POINT (NON-frozen). HARNESS angkat beban (model lokal lemah).
//
// AKAR: gemma ~26B kadang emit `<tool_call>{...}</tool_call>` sebagai TEKS di message.content
// (BUKAN native tool_calls). Dua varian rusak yg kebukti live:
//   (1) JSON valid tapi key "parameters" bukan "arguments": {"name":"brain_immune_scan","parameters":{}}
//   (2) JSON INVALID — key tanpa quote: {name: "capabilities_list", arguments: {}}
// Dua-duanya bikin chat-template --jinja GA parse → tag BOCOR ke user (Telegram). Agent
// (mr-flow/agentkit) liat ToolCalls kosong → return content mentah → bocor.
// FIX di choke-point ROUTER (1 tempat, semua agent kebantu, no rebuild): deteksi tag di content →
// parse LENIENT (strict-JSON dulu, lalu regex buat unquoted) → native tool_calls + strip content.
//
// SWITCH: ENV FLOWORK_TOOLCALL_RECOVER=0 → matiin (default ON). No-op kalau ToolCalls udah ada
// ATAU content ga ada tag → byte-identik perilaku lama. Fails-safe (parse gagal → cuma strip, anti-bocor).
package router

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	toolCallTagRe = regexp.MustCompile(`(?s)<tool_call>\s*(\{.*?\})\s*</tool_call>`)
	leniNameRe    = regexp.MustCompile(`(?:"|')?name(?:"|')?\s*:\s*(?:"|')?([a-zA-Z0-9_.\-]+)(?:"|')?`)
	leniArgsRe    = regexp.MustCompile(`(?s)(?:"|')?(?:parameters|arguments)(?:"|')?\s*:\s*(\{.*\})`)
)

func toolcallRecoverEnabled() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv("FLOWORK_TOOLCALL_RECOVER"))) != "0"
}

// recoverTextToolCalls — kalau model emit <tool_call> sbg TEKS (native ToolCalls kosong), parse
// jadi native tool_calls + bersihin content. Mutate resp in-place. Dipanggil dispatcher sebelum
// return (non-stream). Best-effort, fails-safe.
func recoverTextToolCalls(resp *OpenAIResponse) {
	if resp == nil || !toolcallRecoverEnabled() {
		return
	}
	for i := range resp.Choices {
		msg := &resp.Choices[i].Message
		if hasNativeToolCalls(msg.ToolCalls) || !strings.Contains(msg.Content, "<tool_call>") {
			continue
		}
		type fnObj struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}
		type tcall struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function fnObj  `json:"function"`
		}
		var calls []tcall
		for j, m := range toolCallTagRe.FindAllStringSubmatch(msg.Content, -1) {
			name, args := parseToolCallInner(m[1])
			if name == "" {
				continue
			}
			calls = append(calls, tcall{ID: "call_recover_" + strconv.Itoa(j), Type: "function",
				Function: fnObj{Name: name, Arguments: args}})
		}
		// strip SEMUA tag (parsed/malformed) dari content → anti-bocor.
		msg.Content = stripToolCallTags(msg.Content)
		if len(calls) > 0 {
			if b, err := json.Marshal(calls); err == nil {
				msg.ToolCalls = b
				resp.Choices[i].FinishReason = "tool_calls"
				log.Printf("flow_router toolcall-recover: %d <tool_call> teks → native tool_calls (anti-bocor)", len(calls))
			}
		}
	}
}

// parseToolCallInner — ekstrak (name, argsJSON) dari isi 1 tag. Strict-JSON dulu (handle
// "parameters"/"arguments"); kalau gagal (JSON invalid, key tanpa quote) → lenient regex.
func parseToolCallInner(inner string) (name, args string) {
	var raw struct {
		Name       string          `json:"name"`
		Arguments  json.RawMessage `json:"arguments"`
		Parameters json.RawMessage `json:"parameters"`
	}
	if json.Unmarshal([]byte(inner), &raw) == nil && strings.TrimSpace(raw.Name) != "" {
		a := raw.Arguments
		if len(a) == 0 {
			a = raw.Parameters
		}
		if len(a) == 0 {
			a = json.RawMessage("{}")
		}
		return raw.Name, string(a)
	}
	// lenient — model emit JSON invalid (unquoted key) tapi maksudnya jelas.
	nm := leniNameRe.FindStringSubmatch(inner)
	if nm == nil {
		return "", ""
	}
	args = "{}"
	if am := leniArgsRe.FindStringSubmatch(inner); am != nil && json.Valid([]byte(am[1])) {
		args = am[1]
	}
	return nm[1], args
}

// hasNativeToolCalls — true kalau ToolCalls JSON = array non-kosong.
func hasNativeToolCalls(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null" && s != "[]"
}

func stripToolCallTags(s string) string {
	s = toolCallTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "<tool_call>", "")
	s = strings.ReplaceAll(s, "</tool_call>", "")
	return strings.TrimSpace(s)
}
