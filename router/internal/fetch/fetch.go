// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-05-30
// Reason: Audit pass — audit pass surface review.

// Web-fetch provider catalog.

package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/flowork-os/flowork_Router/internal/safeurl"
)

// Request is the URL-fetch shape. Mode is vendor-specific hint (e.g. "raw" /
// "markdown" / "screenshot") and may be ignored by simpler vendors.
type Request struct {
	URL     string
	Mode    string
	APIKey  string
	BaseURL string
	Extra   map[string]any
}

// Result is the vendor-neutral response. ContentType reflects the actual MIME
// of Body (text/markdown for reader services, text/html for raw fetch).
type Result struct {
	URL         string
	Title       string
	Body        []byte
	ContentType string
	StatusCode  int
}

// Fetcher is the vendor contract.
type Fetcher interface {
	Name() string
	Fetch(ctx context.Context, req Request) (Result, error)
}

var (
	regMu    sync.RWMutex
	registry = map[string]Fetcher{}
)

// Register adds a provider (idempotent — last writer wins).
func Register(p Fetcher) {
	if p == nil || p.Name() == "" {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	registry[p.Name()] = p
}

// Get returns the provider by name, or nil.
func Get(name string) Fetcher {
	regMu.RLock()
	defer regMu.RUnlock()
	return registry[name]
}

// List returns every registered vendor name.
func List() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// fetchHTTPClient is SSRF-hardened: redirects are re-validated against the
// public-only policy (a public URL must not 302 into a private/metadata
// address), and the dialer re-checks the actual resolved IP at connect time and
// pins it (closing the DNS-rebinding window where a host resolves public at
// validation and private at dial). Initial-URL validation still happens in the
// caller via safeurl.Validate; these are the second line of defence.
var fetchHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if _, err := safeurl.Validate(req.Context(), req.URL.String()); err != nil {
			return fmt.Errorf("redirect blocked: %w", err)
		}
		return nil
	},
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           safeDialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// safeDialContext resolves the target, rejects any non-public IP, and dials the
// validated IP literal directly (so no second resolution can swap in a private
// address). This is the anti-SSRF / anti-DNS-rebind dial path for web fetches.
// allowPrivateDial relaxes the dial-time public-IP enforcement. It exists ONLY
// so unit tests can reach loopback httptest servers; it is never set in
// production (default false → loopback/private/metadata are blocked).
var allowPrivateDial = false

func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	d := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
	if ip := net.ParseIP(host); ip != nil {
		if !allowPrivateDial && !safeurl.IsPublic(ip) {
			return nil, fmt.Errorf("%w: %s", safeurl.ErrBlocked, ip)
		}
		return d.DialContext(ctx, network, addr)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, a := range ips {
		if !allowPrivateDial && !safeurl.IsPublic(a.IP) {
			return nil, fmt.Errorf("%w: %s -> %s", safeurl.ErrBlocked, host, a.IP)
		}
	}
	// Pin the resolved IP (dial it directly) so no second resolution can swap in
	// a private address after validation — closes the DNS-rebind window.
	return d.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

// doHTTPRequest sends r and reads up to 8 MiB. Caller-supplied headers are
// already on the request.
func doHTTPRequest(r *http.Request) ([]byte, *http.Response, error) {
	resp, err := fetchHTTPClient.Do(r)
	if err != nil {
		return nil, nil, fmt.Errorf("upstream: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	return body, resp, nil
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
