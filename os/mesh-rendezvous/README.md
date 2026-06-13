# Flowork mesh — WAN rendezvous

The mesh's **Hybrid discovery**: mDNS finds peers on the **LAN** (already built into the router);
this adds **WAN** (cross-internet) discovery for nodes behind carrier NAT, where multicast can't
reach. Address-brokering only — mesh payloads stay P2P, ed25519-signed end-to-end; the rendezvous
never sees them.

```
node A ─register─▶  RENDEZVOUS (you host, public)  ◀─register─ node B
node A ◀──peers──        (pubkey + observed IP + port)         ──peers─▶ node B
node A ───────────── signed knowledge packet (direct P2P) ─────────────▶ node B
```

## Parts
- **server/** — `flowork-rendezvous-server` — stateless in-memory broker. Host it on a public box.
  `POST /register {pubkey_hex, port, version}` (records the OBSERVED source IP) · `GET /peers?self=…`.
- **client/** — `flowork-rendezvous-client` — a sidecar next to each router. Every interval it
  reads the router's identity, registers with the rendezvous, then feeds discovered peers into the
  router via the existing `POST /api/mesh/peer`. **No router code change.**

## Run
```sh
# build
GOWORK=off go build -o flowork-rendezvous-server ./os/mesh-rendezvous/server
GOWORK=off go build -o flowork-rendezvous-client ./os/mesh-rendezvous/client

# on your public box:
./flowork-rendezvous-server -addr :8900

# next to each node's router:
./flowork-rendezvous-client -rendezvous http://rdv.example.com:8900 -port 2402
```

Verified end-to-end: two routers + the rendezvous → each node discovers the other → a signed
knowledge packet flows node→node and is PROMOTED into the receiver's knowledge inbox.

## NAT note
The rendezvous brokers addresses; the gossip POST still has to reach the peer at
`observed-IP:port`. That works for nodes with a public or port-forwarded address (servers, an
always-on home node with a forward). True hole-punching / relay for fully-symmetric-NAT nodes is a
later enhancement; the address-broker is the foundation and matches the Hybrid decision.
