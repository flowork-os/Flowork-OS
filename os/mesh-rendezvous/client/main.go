// Flowork mesh rendezvous CLIENT — a sidecar next to the router. It registers this node with the
// WAN rendezvous and feeds discovered peers into the router via the EXISTING /api/mesh/peer API
// (no router code change). mDNS already handles LAN; this adds cross-internet discovery.
//
//	run:  flowork-rendezvous-client -rendezvous http://rdv.example.com:8900 -port 2402
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	routerURL  = flag.String("router", "http://127.0.0.1:2402", "local router base URL")
	rendezvous = flag.String("rendezvous", "", "rendezvous server base URL (required)")
	meshPort   = flag.Int("port", 2402, "this node's reachable mesh port (the router's -addr port)")
	interval   = flag.Duration("interval", 60*time.Second, "register/poll interval")
	once       = flag.Bool("once", false, "run a single cycle and exit (for testing)")
)

var cli = &http.Client{Timeout: 8 * time.Second}

func main() {
	flag.Parse()
	if *rendezvous == "" {
		fmt.Println("flowork-rendezvous-client: -rendezvous is required"); return
	}
	fmt.Printf("rendezvous client → %s (router %s, port %d)\n", *rendezvous, *routerURL, *meshPort)
	for {
		if err := cycle(); err != nil {
			fmt.Println("cycle:", err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}

func cycle() error {
	// 1. our identity (pubkey) from the local router.
	idBody, err := get(*routerURL + "/api/mesh/identity")
	if err != nil {
		return fmt.Errorf("router identity: %w", err)
	}
	var id struct {
		PubKey  string `json:"pubkey"`
		Version string `json:"version"`
	}
	if json.Unmarshal(idBody, &id) != nil || len(id.PubKey) != 64 {
		return fmt.Errorf("bad identity response")
	}

	// 2. register with the rendezvous (it observes our public IP from the request).
	if _, err := post(*rendezvous+"/register", map[string]any{
		"pubkey_hex": id.PubKey, "port": *meshPort, "version": id.Version,
	}); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// 3. fetch peers and feed each into the router's mesh peer table.
	pb, err := get(*rendezvous + "/peers?self=" + id.PubKey)
	if err != nil {
		return fmt.Errorf("peers: %w", err)
	}
	var pl struct {
		Peers []struct {
			PubKeyHex string `json:"pubkey_hex"`
			IP        string `json:"ip"`
			Port      int    `json:"port"`
			Version   string `json:"version"`
		} `json:"peers"`
	}
	_ = json.Unmarshal(pb, &pl)
	added := 0
	for _, p := range pl.Peers {
		if len(p.PubKeyHex) != 64 || p.IP == "" || p.Port <= 0 {
			continue
		}
		if _, err := post(*routerURL+"/api/mesh/peer", map[string]any{
			"pubkey_hex": p.PubKeyHex, "ip": p.IP, "port": p.Port,
			"hostname": "wan", "version": p.Version,
		}); err == nil {
			added++
		}
	}
	fmt.Printf("registered; %d WAN peer(s) fed to the router\n", added)
	return nil
}

func get(url string) ([]byte, error) {
	resp, err := cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func post(url string, body any) ([]byte, error) {
	b, _ := json.Marshal(body)
	resp, err := cli.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
