# 🌐 BROWSER — STABLE & LOCKED (jangan diutak-atik, owner 2026-06-28)

Browser udah BERES & STABIL → di-LOCK semua biar selalu normal. JANGAN edit tanpa unfreeze sadar owner.

## File beku (browser stack)
| File | Fungsi | Status |
|---|---|---|
| `internal/tools/builtins/browser_desktop.go` | core browser desktop + **capture (screenshot)** | 🔒 BEKU |
| `internal/tools/builtins/browser_desktop_seam.go` | seam/registry browser | 🔒 BEKU |
| `internal/tools/builtins/browser_desktop_ext.go` | ext browser (sengaja DI-LOCK biar browser stabil — bukan plug-in editable lagi) | 🔒 BEKU |
| `internal/tools/builtins/web.go` | webfetch + SSRF guard + header gaya-Claude | 🔒 BEKU |

## Catatan penting
- **Capture (screenshot)** ada di `browser_desktop.go` — bagian dari yg dikunci, jadi capture stabil.
- **WebFetch**: header browser realistis + Accept (contek Claude) biar ga keblokir anti-bot. UA override
  via env `FLOWORK_WEBFETCH_UA`. Situs JS-berat/Cloudflare → pakai **browse-surfer** (browser asli), bukan webfetch.
- Mau edit browser ke depan: `unlock.sh <file>` (FD colok) → edit → `lock.sh <file>`. Sadar + izin owner.
- Browser di-lock TOTAL (termasuk ext) = keputusan owner demi stabilitas; beda dari tool plug-and-play lain.
