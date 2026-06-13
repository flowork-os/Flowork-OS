// Flowork mesh rendezvous SERVER — the WAN half of the Hybrid discovery (mDNS handles LAN).
// Nodes behind carrier NAT can't find each other via multicast; they register here and read the
// peer list. The server only brokers ADDRESSES (pubkey + observed IP + port) — it never sees or
// relays mesh payloads (those stay P2P, ed25519-signed end-to-end). Stateless-ish, in-memory, TTL.
//
//	run:  flowork-rendezvous-server -addr :8900
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	addr = flag.String("addr", ":8900", "listen address")
	ttl  = flag.Duration("ttl", 10*time.Minute, "drop a peer not seen within this window")
)

type peer struct {
	PubKeyHex string `json:"pubkey_hex"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Version   string `json:"version"`
	lastSeen  time.Time
}

var (
	mu    sync.Mutex
	peers = map[string]*peer{}
)

func main() {
	flag.Parse()
	http.HandleFunc("/register", register)
	http.HandleFunc("/peers", list)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	fmt.Printf("flowork rendezvous on %s (ttl=%s)\n", *addr, *ttl)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// POST /register {pubkey_hex, port, version}. The IP is taken from the OBSERVED source address
// (the node's public/NAT address), not from the body — so a node never has to know its own
// external IP.
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		PubKeyHex string `json:"pubkey_hex"`
		Port      int    `json:"port"`
		Version   string `json:"version"`
	}
	if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.PubKeyHex) != 64 || body.Port <= 0 {
		http.Error(w, "need pubkey_hex(64) + port", 400)
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" { // behind a reverse proxy
		if host, _, e := net.SplitHostPort(fwd); e == nil {
			ip = host
		} else {
			ip = fwd
		}
	}
	mu.Lock()
	peers[body.PubKeyHex] = &peer{PubKeyHex: body.PubKeyHex, IP: ip, Port: body.Port, Version: body.Version, lastSeen: time.Now()}
	mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "observed_ip": ip})
}

// GET /peers?self=<pubkey> → live peers (excluding self), evicting stale ones.
func list(w http.ResponseWriter, r *http.Request) {
	self := r.URL.Query().Get("self")
	out := []peer{}
	mu.Lock()
	for k, p := range peers {
		if time.Since(p.lastSeen) > *ttl {
			delete(peers, k)
			continue
		}
		if k == self {
			continue
		}
		out = append(out, *p)
	}
	mu.Unlock()
	writeJSON(w, map[string]any{"peers": out})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
