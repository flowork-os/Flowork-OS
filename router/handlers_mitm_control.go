// === LOCKED FILE ===
// Status: STABLE — DO NOT MODIFY without owner approval.
// Owner: Aola Sahidin (Mr.Dev)
// Repo: https://github.com/flowork-os/Flowork-OS
// Locked at: 2026-06-13
// Reason: Release audit — wires the missing MITM interceptor Start/Stop control (E2E tested).
//
// MITM Proxy lifecycle control (Start / Stop the TLS interceptor).
//
// 2026-06-13 (owner-approved, release audit): the MITM proxy module shipped with CA + DNS + status
// endpoints and a fully unit-tested TLS interception engine (internal/mitm), but NOTHING in the
// production path ever called Manager.Start — so the GUI could install a CA and hijack DNS yet no
// listener was bound, leaving target hosts pointing at 127.0.0.1:<nothing> (connection refused
// instead of interception). This file wires the missing Start/Stop control so the feature is
// actually usable end-to-end, without touching the locked engine/handler files.
//
// SAFETY: starting is explicit (button/endpoint), never automatic. DNS hijack is opt-in per request
// (default OFF) so Start can bind the listener without rewriting /etc/hosts unless the user asks.
// The default listen addr is 127.0.0.1:443 because DNS-hijacked hosts resolve to 127.0.0.1 and IDE
// clients use the implicit HTTPS port 443 — binding :443 needs root/admin (documented). Override via
// the request `addr` field or FLOW_ROUTER_MITM_ADDR for non-privileged testing.

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/flowork-os/flowork_Router/internal/mitm"
)

var (
	mitmMgrMu sync.Mutex
	mitmMgr   *mitm.Manager
)

// mitmListenAddr resolves the interceptor listen address: request override →
// FLOW_ROUTER_MITM_ADDR env → default 127.0.0.1:443.
func mitmListenAddr(reqAddr string) string {
	if reqAddr != "" {
		return reqAddr
	}
	if v := os.Getenv("FLOW_ROUTER_MITM_ADDR"); v != "" {
		return v
	}
	return "127.0.0.1:443"
}

// mitmStartHandler — POST { addr?: string, hijackDNS?: bool }. Boots the TLS
// interceptor. Idempotent: a no-op (with ok:true) when already running.
func mitmStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Addr      string `json:"addr"`
		HijackDNS bool   `json:"hijackDNS"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	mitmMgrMu.Lock()
	defer mitmMgrMu.Unlock()
	if mitmMgr != nil && mitm.IsRunning() {
		writeJSON(w, http.StatusOK, map[string]any{"started": true, "already": true, "pid": mitm.ReadPidFile()})
		return
	}
	cm, err := mitm.NewCertManager(mitm.DataDir())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"started": false, "error": "cert manager: " + err.Error()})
		return
	}
	// DNS hijack is opt-in: pass the target hosts only when explicitly requested,
	// otherwise NewManager gets no hosts and Start skips the /etc/hosts rewrite.
	var hosts []string
	if body.HijackDNS {
		hosts = mitm.TargetHosts
	}
	mgr := mitm.NewManager(mitmListenAddr(body.Addr), cm, hosts)
	if err := mgr.Start(nil); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"started": false, "error": err.Error(), "addr": mitmListenAddr(body.Addr)})
		return
	}
	mitmMgr = mgr
	writeJSON(w, http.StatusOK, map[string]any{"started": true, "addr": mitmListenAddr(body.Addr), "pid": mitm.ReadPidFile(), "dnsHijacked": body.HijackDNS})
}

// mitmStopHandler — POST. Drains the interceptor + clears DNS + pidfile.
func mitmStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mitmMgrMu.Lock()
	defer mitmMgrMu.Unlock()
	if mitmMgr == nil {
		writeJSON(w, http.StatusOK, map[string]any{"stopped": true, "already": true})
		return
	}
	err := mitmMgr.Stop()
	mitmMgr = nil
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"stopped": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stopped": true})
}

// stopMITMOnShutdown — best-effort drain on router shutdown so a crash/exit
// doesn't leave a bound :443 listener or a stale DNS hijack behind.
func stopMITMOnShutdown() {
	mitmMgrMu.Lock()
	defer mitmMgrMu.Unlock()
	if mitmMgr != nil {
		_ = mitmMgr.Stop()
		mitmMgr = nil
	}
}
