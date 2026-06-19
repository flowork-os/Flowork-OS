package routerclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbedText_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2,0.3],"index":0}],"model":"bge-m3"}`))
	}))
	defer srv.Close()

	vec, err := New(srv.URL).EmbedText(context.Background(), "", "verify before trust")
	if err != nil {
		t.Fatalf("EmbedText: %v", err)
	}
	if len(vec) != 3 || vec[0] != 0.1 {
		t.Fatalf("vec = %v, want [0.1 0.2 0.3]", vec)
	}
}

func TestEmbedText_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":{"message":"no embedding model configured"}}`))
	}))
	defer srv.Close()

	if _, err := New(srv.URL).EmbedText(context.Background(), "", "x"); err == nil {
		t.Fatal("expected error on 400 status")
	}
}

func TestEmbedText_EmptyInput(t *testing.T) {
	if _, err := New("http://127.0.0.1:0").EmbedText(context.Background(), "", "  "); err == nil {
		t.Fatal("expected error on empty input")
	}
}
