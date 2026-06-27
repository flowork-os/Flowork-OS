//go:build (linux || darwin || windows) && !android

// browser_desktop_ext.go — TITIK EXTENSION (NON-FROZEN, BISA DIHAPUS) buat browser_desktop.go (frozen).
//
// ⚖️ ATURAN ABADI (owner Mr.Dev): file frozen TIDAK BOLEH dibuka buat nambah filtur. Switch-reader
// (headless/flags/idle) udah pindah ke browser_desktop_seam.go (BEKU, default aman) biar inti
// self-sufficient. Tuning lewat ENV (FLOWORK_BROWSER_*) — ga perlu edit kode.
//
// === CARA NAMBAH FILTUR BROWSER (tanpa buka file frozen) ===
//  1. TOOL BROWSER BARU (mis. browser_scroll): bikin FILE BARU `browser_<nama>.go` (build-tag sama)
//     dgn `func init(){ tools.Register(&browserXxxTool{}) }`. Go ngegabung init() sepaket → tool
//     ke-daftar TANPA edit browser_desktop.go.
//  2. TUNING LAUNCH/LIFECYCLE: pakai env switch FLOWORK_BROWSER_HEADLESS / _FLAGS / _IDLE_MIN.
//
// 📖 WAJIB BACA: FLowork_os/lock/browser.md sebelum ngutak-atik browser-control.
package builtins
