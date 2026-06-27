// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// Cara kerja sistem: lihat os/.  ⚠️ FROZEN — jangan edit file ini.
// Nambah/ubah fitur TANPA buka frozen: pakai SEAM non-frozen + SWITCH
// (internal/fwswitch/registry.go). Pola lengkap: lock/frozen-core.md

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

func strconvItoa(n int) string { return strconv.Itoa(n) }

// loginMaxFailsBeforeLock / loginFailWindow — kebijakan lockout brute-force login. SWITCH
// FLOWORK_LOGIN_MAX_FAILS (default 5) + FLOWORK_LOGIN_FAIL_WINDOW (menit, default 60).
func loginMaxFailsBeforeLock() int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("FLOWORK_LOGIN_MAX_FAILS"))); err == nil && n > 0 {
		return n
	}
	return 5
}

func loginFailWindow() time.Duration {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("FLOWORK_LOGIN_FAIL_WINDOW"))); err == nil && n > 0 {
		return time.Duration(n) * time.Minute
	}
	return 60 * time.Minute
}

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

func loginCheckLock(ip string) (bool, int) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	e := loginLocks[ip]
	if e == nil {
		return false, 0
	}
	now := time.Now()

	if !e.lastFailAt.IsZero() && now.Sub(e.lastFailAt) > loginFailWindow() &&
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

func loginRecordFail(ip string) (bool, int) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	e := loginLocks[ip]
	if e == nil {
		e = &loginLockEntry{}
		loginLocks[ip] = e
	}

	now := time.Now()

	if !e.lockUntil.IsZero() && now.Before(e.lockUntil) {
		remaining := int(e.lockUntil.Sub(now).Seconds())
		if remaining < 1 {
			remaining = 1
		}
		return true, remaining
	}

	e.fails++
	e.lastFailAt = now
	if e.fails >= loginMaxFailsBeforeLock() {
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

func loginRecordSuccess(ip string) {
	loginLockMu.Lock()
	defer loginLockMu.Unlock()
	delete(loginLocks, ip)
}
