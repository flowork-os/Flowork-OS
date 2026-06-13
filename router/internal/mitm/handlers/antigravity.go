// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-05-30
// 2026-06-13 (owner-approved, release audit): fixed reroute body duplication — the hand-rolled
//   bytesReader re-emitted the slice prefix on every Read and never returned EOF, so the rerouted
//   request body was sent more than once and the dispatcher rejected it ("invalid character '{'
//   after top-level value"). Replaced with bytes.NewReader (stateful + Len() → correct
//   Content-Length). Verified E2E: TLS-intercepted call now byte-identical to a direct /v1 call.
// Reason: Audit pass — MITM proxy module.

// Per-IDE MITM handler: antigravity (Cloud Code Assist).
package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"
)

// maxRerouteBody — hard cap on request body size we will buffer before
// forwarding to the local router. Anything above this is rejected with 413.
// Sized for typical chat payloads (50 MiB covers very large prompt + tool
// histories) but stops a runaway client from exhausting RAM by streaming
// gigabytes into io.ReadAll.
const maxRerouteBody = 50 << 20

// rerouteClient — explicit timeout instead of relying on r.Context() alone.
// The local router runs on the loopback so request latency is bounded by
// upstream provider behaviour, not network distance; 10 minutes is enough
// for the slowest streaming completion and short enough to fail closed if
// the loopback dispatcher hangs.
var rerouteClient = &http.Client{Timeout: 10 * time.Minute}

func init() { Register(&antigravityHandler{}) }

type antigravityHandler struct{}

func (a *antigravityHandler) Name() string { return "antigravity" }

// Handle rewrites the antigravity-bound request to flow_router's own /v1
// dispatcher (so all the dispatcher logic — combos, fallback, executors,
// usage tracking — applies uniformly). The path translation strips
// /v1internal:<action> and maps to /v1/chat/completions.
func (a *antigravityHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rerouteToRouter(w, r, "/v1/chat/completions")
}

// rerouteToRouter is shared by all per-IDE handlers — write the original body
// through to the local router (which the MITM server runs alongside).
func rerouteToRouter(w http.ResponseWriter, r *http.Request, routerPath string) {
	defer r.Body.Close()
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRerouteBody))
	if err != nil {
		var mb *http.MaxBytesError
		if errors.As(err, &mb) {
			http.Error(w, "request body exceeds limit", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "read request body", http.StatusBadRequest)
		return
	}

	// Build a new request hitting flow_router on 127.0.0.1:2402.
	target := "http://127.0.0.1:2402" + routerPath
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, copyReader(body))
	if err != nil {
		http.Error(w, "build reroute request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Del("Host")
	resp, err := rerouteClient.Do(req)
	if err != nil {
		http.Error(w, "upstream unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func copyReader(b []byte) io.Reader {
	if len(b) == 0 {
		return http.NoBody
	}
	// bytes.NewReader is stateful (advances on each Read, returns io.EOF when
	// drained) AND exposes Len(), so http.NewRequestWithContext can detect the
	// body size and set Content-Length. The previous hand-rolled bytesReader
	// did neither — it re-copied the slice prefix on every Read and never hit
	// EOF, so the rerouted request body got duplicated and the dispatcher
	// rejected it ("invalid character '{' after top-level value"). Release-audit
	// bug fix 2026-06-13 (owner-approved).
	return bytes.NewReader(b)
}
