# Flowork

> **One codebase → many bodies (desktop · USB-OS · Android), one mind (P2P mesh),**
> **powered by the user's own cloud LLM (Claude via cloak) with a local option.**

This is the unified Flowork monorepo. Everything builds from here.

## Layout

| Path | Module / role | Port |
|------|---------------|------|
| [agent/](agent/) | `flowork-gui` — the agent engine + unified control panel (web GUI) | `:1987` |
| [router/](router/) | `github.com/flowork-os/flowork_Router` — LLM router: cloak, routing, pricing/policy, brain, **P2P mesh** | `:2402` |
| [os/](os/) | Flowork OS — bootable USB appliance builder (Alpine + kiosk + A/B + verity + LUKS) | — |
| [seed-sovereign/](seed-sovereign/) | Sovereign router seed config (appliance default providers) | — |
| `go.work` | Go workspace tying the two modules together | — |

**Two binaries, two ports, one repo** — keeps each updatable independently while sharing one
source of truth, one CI, one history. The user sees ONE app: router management is rendered as
native tabs inside the agent panel.

## Targets (same binaries everywhere)
- **Desktop** — daily driver.
- **Flowork OS (USB)** — bootable sovereign appliance (`os/`).
- **Android** — 24/7 always-on node. The Go core cross-compiles to `linux/arm64` static
  (verified) and runs unmodified under the phone's Linux kernel.

## Build

```sh
# both modules, native:
go build ./agent/...   && go build ./router/...
# Android core (proven): static ARM64 binaries
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o flowork-agent-arm64  ./agent
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o flowork-router-arm64 ./router
# USB appliance:
os/build/build-flowork-os.sh
```

## Laws
- **Separate CODE from DATA.** Updates replace code; user data (`*.db`, `brain/`, agent
  workspaces) is sacred and never touched.
- **Never remove a router feature.** Merging only changes presentation; adding is allowed.
- **Stable · strong · secure** over speed. Verify every step before moving on.
