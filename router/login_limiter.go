// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-05-30
// Reason: Audit pass — audit pass surface review.

// Dashboard Login Rate Limiter (in-memory).

package main

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// clientIPForLock returns the host portion of r.RemoteAddr, falling back to the
// raw value when SplitHostPort fails (e.g. unix sockets or test fakes).
//
// Behind a trusted reverse proxy / the built-in tunnel the TCP peer is always
// the proxy (127.0.0.1), which would lump every remote client into one lockout
// bucket. When FLOW_ROUTER_TRUST_PROXY=1 we instead key on the left-most
// X-Forwarded-For entry. This is gated on explicit opt-in so an untrusted direct
// client cannot spoof XFF to evade (or DoS) the limiter.
func clientIPForLock(r *http.Request) string {
	if r == nil {
		return "unknown"
	}
	if os.Getenv("FLOW_ROUTER_TRUST_PROXY") == "1" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i >= 0 {
				xff = xff[:i]
			}
			if ip := strings.TrimSpace(xff); ip != "" {
				return ip
			}
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

// strconvItoa is a tiny local alias so handlers_auth.go does not need to import
// strconv just for the Retry-After header value.
func strconvItoa(n int) string { return strconv.Itoa(n) }

const (
	loginMaxFailsBeforeLock = 5
	loginFailWindow         = 60 * time.Minute
)

var loginLockSteps = []time.Duration{
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
}

type loginLockEntry struct {
	fails      int
	lockUntil  time.Time
	lockLevel  int
	lastFailAt time.Time
}

var (
	loginLockMu sync.Mutex
	loginLocks  = map[string]*loginLockEntry{}
)

// loginCheckLock returns (locked, retryAfterSeconds).
// Auto-prunes stale entries whose window elapsed and lock expired.
func loginCheckLock(ip string) (bool, int) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	e := loginLocks[ip]
	if e == nil {
		return false, 0
	}
	now := time.Now()
	// auto reset if window elapsed AND no active lock
	if !e.lastFailAt.IsZero() && now.Sub(e.lastFailAt) > loginFailWindow &&
		(e.lockUntil.IsZero() || !now.Before(e.lockUntil)) {
		delete(loginLocks, ip)
		return false, 0
	}
	if e.lockUntil.IsZero() || !now.Before(e.lockUntil) {
		return false, 0
	}
	remaining := int(e.lockUntil.Sub(now).Seconds())
	if remaining < 1 {
		remaining = 1
	}
	return true, remaining
}

// loginRecordFail increments the fail counter and, on threshold, sets the lock.
// Returns (locked, retryAfterSeconds) AFTER the increment, so the caller can
// emit a 429 + Retry-After when the threshold is just crossed.
// FIX #5: Check if already locked before incrementing to prevent race condition
func loginRecordFail(ip string) (bool, int) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	e := loginLocks[ip]
	if e == nil {
		e = &loginLockEntry{}
		loginLocks[ip] = e
	}

	now := time.Now()

	// ✓ NEW: Check if ALREADY LOCKED before incrementing
	// This prevents multiple concurrent threads from bypassing the lock
	if !e.lockUntil.IsZero() && now.Before(e.lockUntil) {
		remaining := int(e.lockUntil.Sub(now).Seconds())
		if remaining < 1 {
			remaining = 1
		}
		return true, remaining // RETURN IMMEDIATELY if already locked
	}

	// Otherwise increment fail counter
	e.fails++
	e.lastFailAt = now
	if e.fails >= loginMaxFailsBeforeLock {
		idx := e.lockLevel
		if idx >= len(loginLockSteps) {
			idx = len(loginLockSteps) - 1
		}
		e.lockUntil = now.Add(loginLockSteps[idx])
		e.lockLevel++
		e.fails = 0
		return true, int(loginLockSteps[idx].Seconds())
	}
	return false, 0
}

// loginRecordSuccess clears the IP's fail history on a successful login.
func loginRecordSuccess(ip string) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	delete(loginLocks, ip)
}
