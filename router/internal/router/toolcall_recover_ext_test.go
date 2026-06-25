package router

import (
	"strings"
	"testing"
)

// varian-1: JSON valid tapi key "parameters" (screenshot 1) → recovered + content bersih.
func TestRecover_ParametersKey(t *testing.T) {
	r := &OpenAIResponse{Choices: []OpenAIChoice{{Message: OpenAIMessage{
		Content: `siap:\n<tool_call>{"name":"brain_immune_scan","parameters":{}}</tool_call>`}}}}
	recoverTextToolCalls(r)
	if !hasNativeToolCalls(r.Choices[0].Message.ToolCalls) {
		t.Fatalf("expected recovered tool_calls, got %s", r.Choices[0].Message.ToolCalls)
	}
	if strings.Contains(r.Choices[0].Message.Content, "tool_call") {
		t.Fatalf("content masih bocor tag: %q", r.Choices[0].Message.Content)
	}
	if !strings.Contains(string(r.Choices[0].Message.ToolCalls), "brain_immune_scan") {
		t.Fatalf("name hilang: %s", r.Choices[0].Message.ToolCalls)
	}
}

// varian-2: JSON INVALID (unquoted key) (screenshot 2) → lenient recover.
func TestRecover_UnquotedKey(t *testing.T) {
	r := &OpenAIResponse{Choices: []OpenAIChoice{{Message: OpenAIMessage{
		Content: `oke:\n<tool_call>{name: "capabilities_list", arguments: {}}</tool_call>`}}}}
	recoverTextToolCalls(r)
	if !hasNativeToolCalls(r.Choices[0].Message.ToolCalls) {
		t.Fatalf("expected lenient recover, got content=%q tc=%s", r.Choices[0].Message.Content, r.Choices[0].Message.ToolCalls)
	}
	if !strings.Contains(string(r.Choices[0].Message.ToolCalls), "capabilities_list") {
		t.Fatalf("name hilang: %s", r.Choices[0].Message.ToolCalls)
	}
}

// native tool_calls udah ada → JANGAN diutak-atik (no-op).
func TestRecover_NativeUntouched(t *testing.T) {
	orig := `[{"id":"x","type":"function","function":{"name":"foo","arguments":"{}"}}]`
	r := &OpenAIResponse{Choices: []OpenAIChoice{{Message: OpenAIMessage{Content: "hi", ToolCalls: []byte(orig)}}}}
	recoverTextToolCalls(r)
	if string(r.Choices[0].Message.ToolCalls) != orig || r.Choices[0].Message.Content != "hi" {
		t.Fatalf("native harus utuh, got tc=%s content=%q", r.Choices[0].Message.ToolCalls, r.Choices[0].Message.Content)
	}
}

// teks biasa tanpa tag → no-op (byte-identik).
func TestRecover_PlainTextNoop(t *testing.T) {
	r := &OpenAIResponse{Choices: []OpenAIChoice{{Message: OpenAIMessage{Content: "halo bro santai aja"}}}}
	recoverTextToolCalls(r)
	if hasNativeToolCalls(r.Choices[0].Message.ToolCalls) || r.Choices[0].Message.Content != "halo bro santai aja" {
		t.Fatalf("plain text harus utuh, got tc=%s content=%q", r.Choices[0].Message.ToolCalls, r.Choices[0].Message.Content)
	}
}
