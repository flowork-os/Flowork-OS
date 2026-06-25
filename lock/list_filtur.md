# 📋 LIST FILTUR — Katalog Tool, Slash & Loop (Flowork)

> Owner: Aola Sahidin (Mr.Dev) · github.com/flowork-os/Flowork-OS · 2026-06-26.
> Referensi buat AI: **nama tool · cara panggil · fungsi**. Auto-derived dari registry
> (`agent/internal/tools/builtins/*` via `tools.List()`), JANGAN edit tangan — regen kalau tool berubah.

## CARA AI MANGGIL TOOL (umum)
AI manggil tool dgn emit **tool-call**: `nama` + argumen **JSON** sesuai param.
Notasi di bawah: `nama(param*, param2)` — `*` = WAJIB, sisanya opsional. Tipe & detail di tabel tiap tool.
- Model lokal lemah → JANGAN tulis `<tool_call>` sbg TEKS (harness `FLOWORK_TOOLCALL_RECOVER` mungulihin, tapi panggil NATIVE).
- Tool ke-gate kapabilitas pas RUN (mis. `exec:shell`, `fs:write`). all-tools mode: semua agent bisa lihat semua tool.
- Lupa nama? panggil **`tool_search`**. Mau detail schema 1 tool? ada di list endpoint / tabel ini.

**Total builtin tool ter-register: 150.** (Tool MCP eksternal + dynamic nambah di atas ini — lihat §akhir.)

---

## 📁 File & Filesystem (17)

### `code_scan(path)`
- **cap:** `state:write` · **balik:** {run_id, files_scanned, total_findings, critical, by_severity, status, top[]}
- **fungsi:** Run a DEFENSIVE security scan (static auditors + trivy: CVE/secret/IaC misconfig) over your workspace, or a subpath. Stores findings (queryable via scanner_findings_query) and returns a ranked summary so you can act on them. Defensive only — scope is your own workspace.
- **param:**
  - `path` (string, opsional) — Subpath inside your workspace to scan (default: whole workspace)

### `codemap_count()`
- **cap:** `state:read` · **balik:** {total}
- **fungsi:** Total codemap nodes indexed (Section 27). Anti over-prompt: cuma counts, ngga return list. Pakai codemap_search untuk query detail.

### `codemap_files(search, layer, limit)`
- **cap:** `state:read` · **balik:** {total, shown, files:[{path,layer,line_count,dependent_count,has_tests,has_docs}]}
- **fungsi:** Mapping codebase level-FILE (sumber kebenaran self-map): daftar file + layer + jumlah baris + dependent_count (dipakai brp file) + ada test/doc. Filter optional. Buat ngerti struktur codebase sebelum nulis/ubah kode.
- **param:**
  - `search` (string, opsional) — substring path/nama (optional)
  - `layer` (string, opsional) — filter layer (optional, mis. handler/engine/data-store)
  - `limit` (int, opsional) — max file (default 40, max 200)

### `codemap_search(search*, node_type, layer)`
- **cap:** `state:read` · **balik:** {items: [{name, type, file, lines, size_loc}], count, truncated}
- **fungsi:** Search codemap nodes by name substring + optional node_type/layer filter. Cap 10 (anti over-prompt). Return summary fields (name, type, file, lines).
- **param:**
  - `search` (string, WAJIB) — name substring (required)
  - `node_type` (string, opsional) — func | type | method | var
  - `layer` (string, opsional) — agent | tool | gui | kernel | brain

### `codemap_search_advanced(search, node_type, layer)`
- **cap:** `state:read` · **balik:** {count, nodes[]}
- **fungsi:** Search codemap nodes by name + filter node_type (func|type|var|const) + layer (handler|service|store). Cap 20 hits.
- **param:**
  - `search` (string, opsional) — Substring match on name
  - `node_type` (string, opsional) — func|type|var|const
  - `layer` (string, opsional) — Logical layer tag

### `codemap_stats()`
- **cap:** `state:read` · **balik:** {total_nodes, by_type: {func: N, type: N, method: N}, ...}
- **fungsi:** Codemap overview — total nodes indexed. Anti over-prompt: ngga return list, cuma counts.

### `edit(file_path, old_string*, new_string*, replace_all, category, name)`
- **cap:** `fs:write:/shared/*` · **balik:** {replaced: <count>, length}
- **fungsi:** Replace an exact substring in a file. Preferred: file_path (relative path in your workspace). Legacy: category + name. Default replace_all=false → reject if >1 match. File cap 4MB.
- **param:**
  - `file_path` (string, opsional) — relative path in your workspace (preferred). Absolute/'..' rejected (isolation).
  - `old_string` (string, WAJIB) — exact substring to find
  - `new_string` (string, WAJIB) — replacement
  - `replace_all` (bool, opsional) — default false; true = replace all occurrences
  - `category` (string, opsional) — legacy: tools|job|document|media|cache|log (use file_path instead)
  - `name` (string, opsional) — legacy: filename (basename only) — pair with category

### `file_list(category*)`
- **cap:** `fs:read:/shared/*` · **balik:** {category, files: [string], count: int}
- **fungsi:** List filenames in shared workspace category. Symlinks skipped.
- **param:**
  - `category` (string, WAJIB) — tools|job|document|media|cache|log

### `file_read(file_path, category, name)`
- **cap:** `fs:read:/shared/*` · **balik:** {path, content, size_bytes, truncated: bool}
- **fungsi:** Read a file from your workspace. Preferred: file_path (relative path inside your workspace, e.g. 'docs/notes.md'). Legacy: category + name. 4MB cap.
- **param:**
  - `file_path` (string, opsional) — relative path in your workspace (preferred), e.g. 'src/main.go'. Absolute/'..' rejected (isolation).
  - `category` (string, opsional) — legacy: tools|job|document|media|cache|log (use file_path instead)
  - `name` (string, opsional) — legacy: filename (basename only) — pair with category

### `file_write(file_path, content*, category, name)`
- **cap:** `fs:write:/shared/*` · **balik:** {path, bytes_written}
- **fungsi:** Write a file in your workspace (create or overwrite). Preferred: file_path (relative path, e.g. 'src/util.go' — parent dirs auto-created). Legacy: category + name. Content cap 4MB.
- **param:**
  - `file_path` (string, opsional) — relative path in your workspace (preferred). Absolute/'..' rejected (isolation).
  - `content` (string, WAJIB) — file content
  - `category` (string, opsional) — legacy: tools|job|document|media|cache|log (use file_path instead)
  - `name` (string, opsional) — legacy: filename (basename only) — pair with category

### `glob(pattern*)`
- **cap:** `fs:read:/shared/*` · **balik:** {files: [...], count, truncated}
- **fungsi:** List files matching glob pattern (mis. '*.md', 'document/*.txt'). Recursive scan disabled — top-level + category subdirs only. Cap 200 results. Symlinks skipped.
- **param:**
  - `pattern` (string, WAJIB) — glob pattern relative ke shared workspace

### `grep(pattern*, category, regex)`
- **cap:** `fs:read:/shared/*` · **balik:** {hits: [{file, line_no, line}], count, truncated, scanned_bytes}
- **fungsi:** Search lines matching pattern across shared workspace. Default substring (case-sensitive). regex=true → Go regexp. category filter optional. Cap 200 hits + 4MB total scanned.
- **param:**
  - `pattern` (string, WAJIB) — search pattern
  - `category` (string, opsional) — optional filter to one category
  - `regex` (bool, opsional) — default false; true = treat pattern as Go regexp

### `html_extract(url*)`
- **cap:** `net:fetch:*` · **balik:** {url, title, text, truncated: bool, chars}
- **fungsi:** Fetch URL terus ekstrak teks readable (buang script/style/tag). Buat baca artikel/halaman tanpa noise HTML. SSRF guard + cap 12000 char.
- **param:**
  - `url` (string, WAJIB) — absolute http(s) URL

### `pdf_read(url*)`
- **cap:** `net:fetch:*` · **balik:** {url, pages, text, truncated: bool, chars}
- **fungsi:** Fetch PDF dari URL terus ekstrak teksnya. Buat baca filing/laporan/dokumen PDF (ga ngarang). SSRF guard, download cap 15MB, teks cap 20000 char.
- **param:**
  - `url` (string, WAJIB) — absolute http(s) URL ke file PDF

### `scanner_findings_query(run_id*, severity, limit)`
- **cap:** `state:read` · **balik:** {count, findings[]}
- **fungsi:** Get findings dari satu scanner run by run_id. Filter by severity.
- **param:**
  - `run_id` (int, WAJIB) — Scanner run ID
  - `severity` (string, opsional) — critical|high|medium|low|info (kosong=all)
  - `limit` (int, opsional) — Max (default 100, max 500)

### `scanner_quick_scan()`
- **cap:** `state:read` · **balik:** {critical, high, medium, low, info}
- **fungsi:** Quick scan — count finding by severity dari scanner_findings table (last 30 hari heuristic).

### `scanner_runs_query(limit)`
- **cap:** `state:read` · **balik:** {count, runs[]}
- **fungsi:** List scanner run history (Section 25) — scan_type, target_path, finding counts. Default 20.
- **param:**
  - `limit` (int, opsional) — Max (default 20, max 100)

---

## 🖥️ Shell / Eksekusi / Sistem (8)

### `PowerShell(command*, description, timeout, run_in_background)`
- **cap:** `exec:shell` · **balik:** {stdout, stderr, exit_code, duration_ms}
- **fungsi:** Run a PowerShell command via pwsh. Default timeout 20s (cap 60). Output cap 64KB. Requires pwsh in PATH; if absent you get an educational error telling you to use Bash instead.
- **param:**
  - `command` (string, WAJIB) — PowerShell command to run
  - `description` (string, opsional) — what this does (5-10 words)
  - `timeout` (int, opsional) — timeout seconds 1..60 (default 20)
  - `run_in_background` (bool, opsional) — background run (not supported in the synchronous kernel — returns a note)

### `app_open(app*)`
- **cap:** `exec:app` · **balik:** {opened, app, command} — or {opened:false, error, allowed:[…]} if not in the allowlist / not installed
- **fungsi:** Launch a whitelisted desktop app on the host computer (e.g. chrome, code). Safe: only apps in the owner's allowlist can be opened — the request can never run an arbitrary command. Requires exec:app (operator agent). Use for 'open/buka <app>' requests.
- **param:**
  - `app` (string, WAJIB) — app to open: chrome | code (aliases: vscode, browser, …)

### `bash(command*, working_dir, timeout_seconds)`
- **cap:** `exec:shell` · **balik:** {stdout, stderr, exit_code, truncated, duration_ms}
- **fungsi:** Execute shell command inside agent shared workspace. Linux/macOS via /bin/sh -c, Windows via cmd /C. Default timeout 20s (cap 60). Output cap 64KB. Denied patterns: rm -rf /, fork bomb, sudo, chmod 777, curl|sh, eval $(...).
- **param:**
  - `command` (string, WAJIB) — shell command to execute
  - `working_dir` (string, opsional) — optional subdir inside shared workspace (default workspace root)
  - `timeout_seconds` (int, opsional) — optional timeout 1..60 (default 20)

### `capabilities_list()`
- **cap:** `state:read` · **balik:** {capabilities[]}
- **fungsi:** List unique capabilities yang aktif (state:read/write, net:fetch:*, exec:*, dst). Self-intro: 'gw bisa apa?'.

### `git(op*, category, ref, path)`
- **cap:** `exec:git` · **balik:** {stdout, stderr, exit_code, duration_ms}
- **fungsi:** Run git read-only op (status, diff, log, show). Working dir = <shared>/<category>. Default category='tools'. Output cap 64KB. Timeout 15s.
- **param:**
  - `op` (string, WAJIB) — status | diff | log | show
  - `category` (string, opsional) — tools|job|document|media|cache|log (default: tools)
  - `ref` (string, opsional) — for op=show/log: commit ref (default HEAD)
  - `path` (string, opsional) — for op=diff/show: optional file path filter

### `shell(command*, working_dir, timeout_seconds)`
- **cap:** `exec:shell` · **balik:** {stdout, stderr, exit_code, truncated, duration_ms, read_only}
- **fungsi:** Run a shell command in the agent's shared workspace (hardened). Danger is judged by command STRUCTURE, not substring: recursive deletes of system/home roots, power ops (use system_power instead), dd-to-device, mkfs, chmod 777, fork bombs, privilege escalation, and curl|sh are blocked; legit commands and quoted strings are not. Linux/macOS /bin/sh, Windows cmd. Timeout default 20s (cap 60), output cap 64KB.
- **param:**
  - `command` (string, WAJIB) — shell command to execute
  - `working_dir` (string, opsional) — optional subdir inside shared workspace
  - `timeout_seconds` (int, opsional) — optional timeout 1..60 (default 20)

### `system_health()`
- **cap:** `time:read` · **balik:** {goos, goarch, go_version, num_goroutine, num_cpu, mem_alloc_mb, time_utc}
- **fungsi:** Snapshot system health: GOOS, GOARCH, Go version, goroutine count, mem alloc, CPU count, current time UTC. Buat self-introspection (mis. 'lo running di OS apa?').

### `system_power(action*, delay_seconds, reason)`
- **cap:** `exec:power` · **balik:** {status, action, delay_seconds, armed, command, message}
- **fungsi:** Kontrol power HOST: shutdown|reboot|suspend|lock|logout|cancel. Butuh cap exec:power (operator tepercaya), tiap aksi di-audit. Eksekusi NYATA cuma kalau ARMED (env FLOWORK_POWER_ARMED), else dry-run. WAJIB konfirmasi user sebelum shutdown/reboot.
- **param:**
  - `action` (string, WAJIB) — shutdown | reboot | suspend | lock | logout | cancel
  - `delay_seconds` (int, opsional) — jeda sebelum eksekusi (window cancel); default 10, max 3600
  - `reason` (string, opsional) — alasan singkat, masuk audit log

---

## 💬 Komunikasi (Telegram) (2)

### `SendUserFile(files*, caption, status)`
- **cap:** `net:fetch:telegram` · **balik:** {sent:[{file, message_id}], count}
- **fungsi:** Send file(s) to the user via Telegram. Paths are relative to the agent shared workspace. Sends to the first chat in TELEGRAM_ALLOWED_CHATS. Bot token from agent secrets.
- **param:**
  - `files` (array, WAJIB) — list of file paths (relative to shared workspace)
  - `caption` (string, opsional) — optional caption for the file(s)
  - `status` (string, opsional) — optional status note (compat)

### `telegram_send(chat_id*, text*)`
- **cap:** `net:fetch:telegram` · **balik:** {chat_id, message_id, ok: bool}
- **fungsi:** Send Telegram message to allowed chat. Bot token from agent secrets. chat_id MUST be in TELEGRAM_ALLOWED_CHATS.
- **param:**
  - `chat_id` (int, WAJIB) — Telegram chat ID (must be in allowed list)
  - `text` (string, WAJIB) — message text (max 4096 char)

---

## 🌐 Web / Browser / Network (14)

### `browser_click(uid*)`
- **cap:** `browser:control` · **balik:** {clicked: uid}
- **fungsi:** Klik elemen by uid (dari browser_snapshot).
- **param:**
  - `uid` (string, WAJIB) — uid elemen

### `browser_close()`
- **cap:** `browser:control` · **balik:** {closed: true}
- **fungsi:** Tutup browser + bebasin resource. WAJIB dipanggil tiap tugas browser kelar (anti zombie chromium numpuk). Browser auto-launch lagi pas tool browser dipakai berikutnya.

### `browser_eval(js*)`
- **cap:** `browser:control` · **balik:** {result}
- **fungsi:** Jalankan JS di halaman saat ini, balikin hasilnya (string). Escape hatch buat baca/aksi yg ga kecover tool lain.
- **param:**
  - `js` (string, WAJIB) — ekspresi JS, mis. '() => document.title'

### `browser_navigate(url*)`
- **cap:** `browser:control` · **balik:** {url, title}
- **fungsi:** Buka URL di browser asli (Chromium, lewat sesi login kalau di-set). Lolos JS/Cloudflare yg webfetch ga bisa. Setelah ini pakai browser_snapshot buat 'lihat' halaman.
- **param:**
  - `url` (string, WAJIB) — absolute http(s) URL

### `browser_screenshot(path)`
- **cap:** `browser:control` · **balik:** {path, bytes}
- **fungsi:** Screenshot halaman saat ini → simpan PNG ke file, balikin path. Buat kasus visual (layout/captcha). Untuk aksi, UTAMAKAN browser_snapshot.
- **param:**
  - `path` (string, opsional) — path simpan PNG (opsional)

### `browser_set_cookies(cookies*)`
- **cap:** `browser:control` · **balik:** {set: count}
- **fungsi:** Inject cookies ke browser (sesi login tanpa ketik password). cookies = JSON array [{name,value,domain,path?}]. Setelah inject, navigate ke situsnya → udah login.
- **param:**
  - `cookies` (string, WAJIB) — JSON array cookie: [{"name":"c_user","value":"...","domain":".facebook.com","path":"/"}]

### `browser_snapshot()`
- **cap:** `browser:control` · **balik:** {title, url, count, elements:[{uid, tag, type, role, label, href}]}
- **fungsi:** Snapshot elemen interaktif halaman saat ini (link/tombol/input) + uid masing2. Pakai uid buat browser_click/browser_type/browser_upload. Lebih akurat & hemat dari screenshot — UTAMAKAN ini.

### `browser_type(uid*, text*)`
- **cap:** `browser:control` · **balik:** {typed: uid}
- **fungsi:** Ketik teks ke input/textarea by uid (dari browser_snapshot). Buat isi form/login/search.
- **param:**
  - `uid` (string, WAJIB) — uid input
  - `text` (string, WAJIB) — teks yg diketik

### `browser_upload(uid*, path*)`
- **cap:** `browser:control` · **balik:** {uploaded: path}
- **fungsi:** Upload file lokal ke input file by uid (dari browser_snapshot). Buat upload video/gambar/dokumen ke web.
- **param:**
  - `uid` (string, WAJIB) — uid input file
  - `path` (string, WAJIB) — path file lokal absolut

### `check_token(chain*, address*)`
- **cap:** `net:fetch:*` · **balik:** {chain, address, source, risk_level, score, flags:[{severity,label,detail}], summary, token}
- **fungsi:** Crypto scam/rug safety check for ONE token. Pulls FACTUAL on-chain security data (GoPlus for EVM chains, RugCheck for Solana — no API key) and returns red-flags + a risk verdict (SCAM / HIGH-RISK / CAUTION / LOOKS-OK). Read-only. Use to vet a token before trusting or recommending it. NEVER asserts a token is 'safe' — it only surfaces risks; absence of flags is not a guarantee.
- **param:**
  - `chain` (string, WAJIB) — chain: eth, bsc, polygon, base, arbitrum, optimism, avalanche, fantom, or solana
  - `address` (string, WAJIB) — token contract address (EVM 0x… or Solana mint address)

### `market_quote(ticker*, range, interval)`
- **cap:** `net:fetch:*` · **balik:** {ticker, price, currency, exchange, ohlcv:[{t,o,h,l,c,v}], fundamentals:{pe,pb,roe,profit_margin,revenue,debt_to_equity,...}}
- **fungsi:** Fetch real market data for ONE stock ticker: spot price, recent OHLCV candles (for technical analysis) and key fundamentals (PE, PB, ROE, margins, revenue, debt). IDX (Indonesia) tickers take the .JK suffix, e.g. BBCA.JK; global tickers use their native symbol, e.g. AAPL.
- **param:**
  - `ticker` (string, WAJIB) — ticker symbol, e.g. BBCA.JK or AAPL
  - `range` (string, opsional) — candle history window (default 6mo): 1mo, 3mo, 6mo, 1y, 2y, 5y
  - `interval` (string, opsional) — candle interval (default 1d): 1d, 1wk, 1mo

### `web_archive(url*, timestamp)`
- **cap:** `net:fetch:*` · **balik:** {url, available: bool, snapshot_url, snapshot_timestamp, status}
- **fungsi:** Cari snapshot arsip URL di Wayback Machine (archive.org). Buat verifikasi konten lama / sumber yang udah hilang. Balikin snapshot terdekat.
- **param:**
  - `url` (string, WAJIB) — URL yang mau dicari arsipnya
  - `timestamp` (string, opsional) — opsional, format YYYYMMDD — cari snapshot terdekat ke tanggal ini

### `web_search(query*, max_results)`
- **cap:** `net:fetch:*` · **balik:** {query, count, results:[{title,url,snippet}]}
- **fungsi:** Web search via Mojeek (no API key, independen). Balikin daftar {title,url,snippet}. Buat cari sumber REAL biar ga ngarang. Cap 8 hasil.
- **param:**
  - `query` (string, WAJIB) — kata kunci pencarian
  - `max_results` (int, opsional) — jumlah hasil (1-8, default 5)

### `webfetch(url*)`
- **cap:** `net:fetch:*` · **balik:** {url, status, content_type, body, truncated: bool, size_bytes}
- **fungsi:** HTTP GET public URL. SSRF guard blocks private/loopback/metadata IPs. Response cap 1MB, timeout 30s.
- **param:**
  - `url` (string, WAJIB) — absolute http(s) URL

---

## 🧠 Brain / Memori / Insting (24)

### `brain_add(content*, wing, room, mem_type)`
- **cap:** `state:write` · **balik:** {id, added, deduped}
- **fungsi:** Simpan knowledge/pengalaman ke brain LOKAL lo sendiri (state.db, FTS5). Pakai buat inget hal penting hasil sendiri: pola sukses, fakta, kesimpulan, eureka. Dedup otomatis (content sama ga dobel). Ini brain PRIBADI lo — beda dari brain_search_shared (korpus router). Recall-nya pakai brain_search.
- **param:**
  - `content` (string, WAJIB) — isi knowledge/pengalaman (ringkas, faktual)
  - `wing` (string, opsional) — kategori besar (default 'general', mis. experience/fact/eureka)
  - `room` (string, opsional) — sub-kategori opsional
  - `mem_type` (string, opsional) — tipe memori (default 'experience')

### `brain_dream(factor)`
- **cap:** `state:write` · **balik:** {decayed, floor, factor}
- **fungsi:** Consolidate this agent's local brain like sleep: gently DECAY the importance of memories never recalled (amplitude 0) so unused ones fade and reinforced ones stay strong. Safe + offline (importance-only UPDATE; never deletes, never touches the search index). Returns how many decayed.
- **param:**
  - `factor` (float, opsional) — decay multiplier 0.5..0.99 (default 0.9)

### `brain_get(id*)`
- **cap:** `state:read` · **balik:** {found, drawer}
- **fungsi:** Ambil 1 drawer full dari brain LOKAL by id (id dari hasil brain_search).
- **param:**
  - `id` (string, WAJIB) — drawer id

### `brain_immune_scan()`
- **cap:** `state:write` · **balik:** {quarantined_now, quarantined_total:[{id, reason_quarantine, content}]}
- **fungsi:** Sapu brain LOKAL lo → karantina drawer yang kena pola injection/jailbreak (antibody) atau confidence rendah, biar brain ga keracunan halu. Balik jumlah yang dikarantina + daftar yang lagi dikarantina.

### `brain_promote_shared(limit)`
- **cap:** `rpc:router:brain` · **balik:** {eligible, promoted, skipped, router_ok}
- **fungsi:** Share knowledge brain LOKAL lo yang berharga ke korpus SHARED di router (biar warga lain bisa belajar). Quality-gate: cuma drawer non-karantina, confidence tinggi, tipe aman (experience/eureka/fact) — constitution/secret GA di-share. Resilient: kalau router mati, di-skip. Balik jumlah yang ke-share.
- **param:**
  - `limit` (int, opsional) — max drawer di-share sekali jalan (default 20, max 100)

### `brain_search(query*, k)`
- **cap:** `state:read` · **balik:** {query, hits:[{drawer_id, wing, room, mem_type, content, score}], count}
- **fungsi:** Cari di brain LOKAL lo sendiri (pengalaman/knowledge yang LO simpan via brain_add) pakai FTS5 BM25. Murah, cepat, no router. Pakai ini DULU buat 'inget pengalaman/knowledge gw soal X'. Kalau butuh korpus pengetahuan luas (security/training/dll), pakai brain_search_shared.
- **param:**
  - `query` (string, WAJIB) — kata kunci / pertanyaan
  - `k` (int, opsional) — max hasil (default 5, max 10)

### `brain_search_shared(query*, k)`
- **cap:** `rpc:router:brain` · **balik:** {query, hits: [{wing, room, content, score, drawer_id}], count}
- **fungsi:** Cari di korpus pengetahuan SHARED di Router (5jt drawers: security/training/dll) via BM25/FTS. Buat pengetahuan LUAS yang bukan pengalaman pribadi lo. Remote (butuh router up). Buat brain PRIBADI lo, pakai brain_search (lokal).
- **param:**
  - `query` (string, WAJIB) — search query (natural language atau keyword)
  - `k` (int, opsional) — max hits (default 5, max 10)

### `brain_verify(id*, confidence)`
- **cap:** `state:write` · **balik:** {id, verified: true}
- **fungsi:** Rilis 1 drawer brain LOKAL dari karantina setelah lo yakin isinya aman (bukan injection/halu). Set confidence baru (default 1.0). Habis ini drawer-nya muncul lagi di brain_search.
- **param:**
  - `id` (string, WAJIB) — drawer id (dari brain_immune_scan)
  - `confidence` (float, opsional) — tier-confidence 0..1 (default 1.0)

### `fact_recall(key*)`
- **cap:** `state:read` · **balik:** {key, value, found}
- **fungsi:** Recall fact yang sebelumnya tersimpan via fact_write. Anti over-prompt: tool ini BUKAN auto-inject ke prompt — caller on-demand kalau perlu inget. Pakai key sebagai topic identifier.
- **param:**
  - `key` (string, WAJIB) — Topic/fact key (case-sensitive)

### `fact_write(key*, value*)`
- **cap:** `state:write` · **balik:** {ok, key, value_len}
- **fungsi:** Simpan fact (key→value) ke memory. Idempotent upsert. Anti over-prompt: kalau fact essential, tulis dengan key descriptive (mis. 'owner_timezone') — caller panggil fact_recall(key) on-demand. Max 32KB per value (hard cap di DB layer).
- **param:**
  - `key` (string, WAJIB) — Topic identifier (snake_case)
  - `value` (string, WAJIB) — Fact content. Max 32KB.

### `graph_recall(query*, max_chars)`
- **cap:** `state:read` · **balik:** {query, fact_sheet, chars}
- **fungsi:** Recall grounding dari Cognitive Graph LOKAL (twin) — tarik subgraph relevan jadi fact-sheet ringkas (budget-capped). Pakai buat 'apa yang gw tau soal X' / 'gimana A nyambung ke B'. Beda dari brain_search (cari teks FTS): ini paham RELASI antar-entitas (Aola→prefers→X, dst).
- **param:**
  - `query` (string, WAJIB) — topik / pertanyaan
  - `max_chars` (int, opsional) — budget fact-sheet (default 1500)

### `instinct_recall(query*, k)`
- **cap:** `state:read` · **balik:** {instincts: ["WHEN ... -> ..."], count, fact_sheet}
- **fungsi:** Tarik POLA INSTINCT coding+security (distilasi dari model kuat) yang relevan SEBELUM nulis/review code. Return fact-sheet ringkas 'WHEN trigger -> rule'. Pakai pas mulai task coding — apalagi yg nyentuh input/auth/network/crypto/kontrak (mata-hacker: sadar celah sebelum nulis).
- **param:**
  - `query` (string, WAJIB) — deskripsi task coding / area kode (mis. 'parse user input ke SQL query')
  - `k` (int, opsional) — max insting (default 6)

### `memory_delete(key*)`
- **cap:** `state:write` · **balik:** {key, deleted: bool}
- **fungsi:** Delete tool memory entry by key. Return deleted bool.
- **param:**
  - `key` (string, WAJIB) — memory key

### `memory_get(key*)`
- **cap:** `state:read` · **balik:** {key, value, found: bool}
- **fungsi:** Read value from tool memory by key. Returns null kalau key ngga ada.
- **param:**
  - `key` (string, WAJIB) — memory key

### `memory_set(key*, value*)`
- **cap:** `state:write` · **balik:** {key, ok: true}
- **fungsi:** Write or update tool memory by key. Value cap 32KB.
- **param:**
  - `key` (string, WAJIB) — memory key
  - `value` (string, WAJIB) — value string

### `mistake_log(category*, title*, content*, context_origin)`
- **cap:** `state:write` · **balik:** {ok, id, was_new}
- **fungsi:** Log mistake/lesson dari halu/error agent ke mistakes_local table (Section 2). Idempotent via UNIQUE(category,title) — kalau title sama, hit_count auto-increment. Tier default 'raw' — phase 7 promote ke router brain antibody.
- **param:**
  - `category` (string, WAJIB) — logic|performance|security|halu|anti_pattern|workflow
  - `title` (string, WAJIB) — Short identifier (max 256 char)
  - `content` (string, WAJIB) — Full description (max 4KB)
  - `context_origin` (string, opsional) — Where detected (file:line atau session id)

### `mistake_promote_eligible(min_hit_count, limit)`
- **cap:** `state:read` · **balik:** {count, eligible[]}
- **fungsi:** List mistakes yang eligible promote ke Router antibody (Section 7) — tier=raw + hit_count >= N.
- **param:**
  - `min_hit_count` (int, opsional) — Min hit (default 3)
  - `limit` (int, opsional) — Max (default 20)

### `mistake_promote_mark(mistake_id*, promoted_to_id*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Mark mistake (by id) sebagai promoted ke Router antibody. Set tier='promoted' + promoted_at + promoted_to_id link.
- **param:**
  - `mistake_id` (int, WAJIB) — Mistake row ID
  - `promoted_to_id` (string, WAJIB) — Router antibody/drawer ID after promote

### `mistake_recall(context*, limit)`
- **cap:** `state:read` · **balik:** {count, warnings:[{title, remediation, hit_count, category}]}
- **fungsi:** Cek apakah lo PERNAH salah di konteks mirip (recall mistakes journal lo sendiri). Panggil SEBELUM ngerjain hal beresiko / yang pernah bermasalah, biar ga ngulang error yang sama. Balik daftar 'dulu lo salah X (Nx), solusinya Y' urut paling sering keulang.
- **param:**
  - `context` (string, WAJIB) — deskripsi singkat situasi/tugas sekarang (kata kunci)
  - `limit` (int, opsional) — max hasil (default 5, max 20)

### `mistake_search(category, title_substr, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Search mistakes_local by category atau substring di title. Anti over-prompt: tool ini on-demand, BUKAN auto-inject. Default limit 20 (max 100).
- **param:**
  - `category` (string, opsional) — Filter by category (logic/halu/performance/...) atau kosong
  - `title_substr` (string, opsional) — Substring in title (case-sensitive) atau kosong
  - `limit` (int, opsional) — Max (default 20, max 100)

### `mistakes_count()`
- **cap:** `state:read` · **balik:** {total, by_tier}
- **fungsi:** Count mistakes_local entries by tier (raw/promoted/reviewed).

### `persona_get()`
- **cap:** `state:read` · **balik:** {prompt, length}
- **fungsi:** Return current persona prompt (kv.prompt). Self-introspection: 'apa persona gw?'

### `self_prompt_render()`
- **cap:** `state:read` · **balik:** {slots[]}
- **fungsi:** Render full self-prompt slots (system/persona/guideline/task — Section 35). Return current configured slot bodies.

### `self_prompt_set(slot*, body*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Set/update self-prompt slot. Slot 'system|persona|guideline|task'. Body markdown OK, max 8KB.
- **param:**
  - `slot` (string, WAJIB) — system|persona|guideline|task
  - `body` (string, WAJIB) — Slot content (max 8KB)

---

## 🕸️ Cognitive Graph (2)

### `cognitive_resolve(tension_id*, keep*)`
- **cap:** `state:write` · **balik:** {ok, resolved, kept, now}
- **fungsi:** Resolve 1 kontradiksi (tension) SETELAH owner mutusin yg bener. keep='new' (pakai nilai BARU → di-apply jadi edge aktif, yg lama di-superseded) atau keep='old' (pertahanin nilai LAMA). Lalu tension ditutup. ⛔ WAJIB owner yang mutusin — JANGAN nebak. Cek dulu cognitive_tensions buat id-nya.
- **param:**
  - `tension_id` (int, WAJIB) — id tension (dari cognitive_tensions)
  - `keep` (string, WAJIB) — 'new' (pakai nilai baru) atau 'old' (pertahanin lama)

### `cognitive_tensions(limit)`
- **cap:** `state:read` · **balik:** {tensions:[{id, subject, relation, old, new, detail}], count}
- **fungsi:** Daftar KONTRADIKSI data (open contradictions) di cognitive graph kamu yg NUNGGU keputusan owner. Tiap item: id, subject, relation, nilai LAMA vs BARU yg konflik (mis. owner goal_is X dulu, sekarang Y). PAKAI ini kalau ragu fakta/preferensi owner, atau mau bantu beresin data. Hasilnya buat KLARIFIKASI ke owner, JANGAN ditebak sendiri — owner yang decide, lalu pakai cognitive_resolve.
- **param:**
  - `limit` (int, opsional) — max item (default 20, max 200)

---

## 🗄️ State / KV / Workspace (8)

### `kv_get(key*)`
- **cap:** `state:read` · **balik:** {key, value, found}
- **fungsi:** Read kv table value by key. Diferentiate dari tool_memory (key dedicated namespace).
- **param:**
  - `key` (string, WAJIB) — KV key

### `kv_list()`
- **cap:** `state:read` · **balik:** {count, keys[]}
- **fungsi:** List kv table keys (config storage). Anti-secret: keys mulai underscore di-mask. Buat introspect 'config apa yang gw punya'.

### `kv_set(key*, value*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Write single kv key. Reserved keys (prompt/router_url/router_model) skip — pakai cara dedicated.
- **param:**
  - `key` (string, WAJIB) — KV key (non-reserved)
  - `value` (string, WAJIB) — Value string

### `stat_summary()`
- **cap:** `state:read` · **balik:** {interactions, death_letters, scanner_runs, schedules, subscriptions}
- **fungsi:** Quick statistics overview: total counts per table. Buat 'how am I doing?' self-introspection.

### `workspace_list(category, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List resource entries di workspace_meta — catalog file/dir per agent. Default limit 50 (max 500). Filter by category optional.
- **param:**
  - `category` (string, opsional) — Filter (mis. document, code, log)
  - `limit` (int, opsional) — Max entries (default 50, max 500)

### `workspace_lookup(category*, path*)`
- **cap:** `state:read` · **balik:** {found, item}
- **fungsi:** Lookup single workspace_meta entry by (category, path). Return zero-value kalau ngga ada.
- **param:**
  - `category` (string, WAJIB) — Resource category
  - `path` (string, WAJIB) — Relative path dari workspace root

### `workspace_meta_count()`
- **cap:** `state:read` · **balik:** {total, by_category}
- **fungsi:** Count workspace_meta entries by category. Quick overview tanpa list dump.

### `workspace_upsert(category*, path*, description, shareable)`
- **cap:** `state:write` · **balik:** {ok, id}
- **fungsi:** Upsert workspace_meta entry. category+path UNIQUE — re-insert update. Buat catalog resource yg agent kerjain.
- **param:**
  - `category` (string, WAJIB) — document|code|log|media|...
  - `path` (string, WAJIB) — Relative path
  - `description` (string, opsional) — Optional description
  - `shareable` (bool, opsional) — Visible cross-agent (default true)

---

## ⏰ Scheduler / Lifecycle / Loop (10)

### `Monitor(until*, command*, description, timeout_ms, persistent)`
- **cap:** `exec:shell` · **balik:** {matched, iterations, last_stdout, last_stderr, last_exit, elapsed_ms}
- **fungsi:** Repeatedly run `command` (every ~2s) until its combined output contains the `until` substring, or timeout_ms elapses. Synchronous + bounded (timeout cap 60s). Use for waiting on a condition (a file to appear, a process to report ready).
- **param:**
  - `until` (string, WAJIB) — substring that signals the condition is met
  - `command` (string, WAJIB) — shell command to poll
  - `description` (string, opsional) — what you're waiting for
  - `timeout_ms` (int, opsional) — max wait ms (default 30000, cap 60000)
  - `persistent` (bool, opsional) — compat only — synchronous kernel cannot keep watching after return

### `ScheduleWakeup(delaySeconds*, reason*, prompt*)`
- **cap:** `state:write` · **balik:** {wakeup_id, due, delay_seconds}
- **fungsi:** Record a one-shot self-wakeup: after delaySeconds, the agent loop should re-fire `prompt`. Writes a durable wakeups row (due time, prompt, reason). The engine/agent loop fires due wakeups (like cron schedules).
- **param:**
  - `delaySeconds` (int, WAJIB) — seconds from now to wake up
  - `reason` (string, WAJIB) — short reason (telemetry/visible)
  - `prompt` (string, WAJIB) — the input to fire on wake-up

### `agent_run(action*, id, label, data)`
- **cap:** `state:write` · **balik:** {id, state, checkpoint?, runs?}
- **fungsi:** Durable lifecycle for YOUR long task: checkpoint progress, then pause/resume/stop so it survives across turns. Actions: create|start|checkpoint|pause|resume|stop|complete|status|list. 'resume' returns the saved checkpoint so you continue where you left off; 'stop' marks the run terminal so it is not resumed. Offline, scoped to this agent's own store (a coordinator's run ledger).
- **param:**
  - `action` (string, WAJIB) — create|start|checkpoint|pause|resume|stop|complete|status|list
  - `id` (string, opsional) — run id (required for all but list)
  - `label` (string, opsional) — human label (create only)
  - `data` (object, opsional) — checkpoint payload (checkpoint/complete)

### `manifest_inspect()`
- **cap:** `state:read` · **balik:** {kv_keys, schedule_count, skills_count, secret_keys[], meta}
- **fungsi:** Inspect own manifest config snapshot — list KV entries summary, total schedule count, total skills count, secret keys (no values).

### `schedule_runs_query(limit)`
- **cap:** `state:read` · **balik:** {count, runs[]}
- **fungsi:** List scheduler runs history — fired_at, outcome, duration_ms, task. Default limit 50 (max 200).
- **param:**
  - `limit` (int, opsional) — Max

### `scheduler_list()`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List schedules untuk agent ini — cron pattern + task + last_run + next_run. Useful untuk introspection: 'kapan gw jadwal next /version?'

### `scheduler_next()`
- **cap:** `state:read` · **balik:** {schedules[]}
- **fungsi:** Get next_run_at untuk semua schedule. Caller bisa lihat kapan task next fire.

### `scheduler_schedule_add(id*, cron*, task*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Add new schedule entry — agent ngecut cron task. ID PRIMARY (upsert). Cron 5-field POSIX.
- **param:**
  - `id` (string, WAJIB) — Schedule ID (kebab-case)
  - `cron` (string, WAJIB) — 5-field cron (min hr dom mon dow)
  - `task` (string, WAJIB) — Task text (lead '/' = slash command)

### `scheduler_schedule_remove(id*)`
- **cap:** `state:write` · **balik:** {ok, removed}
- **fungsi:** Remove schedule entry by ID.
- **param:**
  - `id` (string, WAJIB) — Schedule ID to remove

### `sneakernet_export_query()`
- **cap:** `state:read` · **balik:** {manifest_info, tools_count, schedule_count}
- **fungsi:** Snapshot sneakernet capability — manifest agent + total tools subscribed + schedule count. Info untuk export decision.

---

## 📝 Task / Plan / Workflow (11)

### `TaskCreate(subject*, description, activeForm, prompt, background)`
- **cap:** `state:write` · **balik:** {task_id, status, subject}
- **fungsi:** Catat task ke ledger durable → balik task_id. Kernel sinkron: default cuma REGISTER (jalanin sendiri, lalu TaskUpdate/TaskStop, baca TaskOutput). background:true + prompt = jalan ASYNC (worker bounded ngerjain prompt-nya, notify owner pas kelar).
- **param:**
  - `subject` (string, WAJIB) — judul singkat task
  - `description` (string, opsional) — isi task
  - `activeForm` (string, opsional) — label present-tense pas jalan
  - `prompt` (string, opsional) — instruksi task (wajib kalau background:true)
  - `background` (bool, opsional) — true=jalan async (worker ngerjain+notify owner). Default false=cuma catat (jalanin sendiri).

### `TaskOutput(task_id*, block, timeout)`
- **cap:** `state:read` · **balik:** {task_id, state, output, subject}
- **fungsi:** Read a task's current state + stored output from the ledger. (block/timeout are accepted for compatibility but the kernel is synchronous, so this returns the current snapshot immediately.)
- **param:**
  - `task_id` (string, WAJIB) — task id to read
  - `block` (bool, opsional) — compat only — ignored (no async wait)
  - `timeout` (int, opsional) — compat only — ignored

### `TaskStop(task_id*)`
- **cap:** `state:write` · **balik:** {task_id, status}
- **fungsi:** Mark a task terminal (stopped) in the ledger so it is not resumed.
- **param:**
  - `task_id` (string, WAJIB) — task id to stop

### `TaskUpdate(taskId*, status*, output)`
- **cap:** `state:write` · **balik:** {task_id, status}
- **fungsi:** Update a task's status in the ledger (e.g. pending|in_progress|completed). Optionally append output text.
- **param:**
  - `taskId` (string, WAJIB) — task id from TaskCreate
  - `status` (string, WAJIB) — new status (pending|in_progress|completed|...)
  - `output` (string, opsional) — optional output text to store

### `Workflow(script, scriptPath)`
- **cap:** `rpc:agent-invoke` · **balik:** {run_id, status, note}
- **fungsi:** Register a multi-step orchestration script in the run ledger and get a run_id. Note: the kernel is synchronous and does NOT yet execute the script as a parallel multi-agent fan-out — this REGISTERS it (durable record) so a coordinator can run/inspect it. Does not claim to have run it.
- **param:**
  - `script` (string, opsional) — the workflow script body
  - `scriptPath` (string, opsional) — path to a workflow script (relative to shared)

### `goal_done(summary)`
- **cap:** `state:write` · **balik:** {ok: true, total_done}
- **fungsi:** Mark goal as done. Optional summary. Append to goal log di tool_memory[_goal] (array of {summary, done_at}). Limit 20 entries (oldest dropped).
- **param:**
  - `summary` (string, opsional) — optional outcome summary

### `plan_read()`
- **cap:** `state:read` · **balik:** {plan: <markdown>, updated_at}
- **fungsi:** Read current agent plan (markdown). Empty kalau belum ada.

### `plan_write(plan*)`
- **cap:** `state:write` · **balik:** {ok: true, length}
- **fungsi:** Overwrite plan with new markdown. Body cap 32KB. Append-only history NOT kept (phase 1g simple — phase 2 add plan_revisions).
- **param:**
  - `plan` (string, WAJIB) — markdown plan body

### `task_list()`
- **cap:** `rpc:taskflow` · **balik:** {count, tasks:[{id,name,trigger_hint,crew_size}]}
- **fungsi:** Daftar Category Task (analisa multi-agent) yang tersedia. PAKAI ini buat tau task apa aja yang bisa di-trigger sebelum task_run.

### `task_run(category*, subject*, group)`
- **cap:** `rpc:taskflow` · **balik:** {run_id, status, note}
- **fungsi:** Trigger Category Task (crew analis multi-agent → 1 keputusan). ASYNC: balik run_id langsung, hasil diproses di belakang (~beberapa menit). Kasih tau user lagi diproses + run_id. Cek task_list dulu buat id yang valid.
- **param:**
  - `category` (string, WAJIB) — id kategori task (dari task_list, mis. 'saham')
  - `subject` (string, WAJIB) — subjek analisa (mis. 'BBCA')
  - `group` (string, opsional) — INTERNAL: id GROUP buat delegasi async (gantiin category). Jarang dipake LLM — orchestrator yang isi.

### `todo(op*, content, id)`
- **cap:** `state:write` · **balik:** {items: [...], count}
- **fungsi:** Manage agent todo list. Operations: list | add | done | remove | clear. Item dict: {id, content, done, added_at, done_at}.
- **param:**
  - `op` (string, WAJIB) — list | add | done | remove | clear
  - `content` (string, opsional) — for op=add: todo content
  - `id` (string, opsional) — for op=done/remove: todo id

---

## 📊 Audit / Log / Telemetri (18)

### `audit_count(since_iso)`
- **cap:** `state:read` · **balik:** {counts_by_event_type}
- **fungsi:** Count audit entries by event_type. Anti over-prompt: counts only, no list dump.
- **param:**
  - `since_iso` (string, opsional) — From timestamp ISO (mis. 2026-05-29T00:00:00Z), kosong=all-time

### `audit_event(event_type*, severity, actor, detail_json)`
- **cap:** `state:write` · **balik:** {ok, id}
- **fungsi:** Append-only audit event log. Berbeda dari decision_log: audit untuk EXTERNAL events (login, security action, protector trigger). Severity info|warning|error|critical.
- **param:**
  - `event_type` (string, WAJIB) — Slug event (mis. login_success, rate_limit_hit)
  - `severity` (string, opsional) — info|warning|error|critical (default info)
  - `actor` (string, opsional) — Who triggered (caller id/name)
  - `detail_json` (string, opsional) — JSON detail string (max 8KB)

### `audit_search(event_type, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Search audit log by event_type. Default limit 30 (max 200).
- **param:**
  - `event_type` (string, opsional) — Filter (mis. login_success, rate_limit_hit)
  - `limit` (int, opsional) — Max (default 30, max 200)

### `decision_count()`
- **cap:** `state:read` · **balik:** {by_type, by_outcome}
- **fungsi:** Count decisions by decision_type + outcome. Quick analytics.

### `decision_log(decision_type*, rationale*, outcome)`
- **cap:** `state:write` · **balik:** {ok, id}
- **fungsi:** Log keputusan non-trivial ke decisions table. Audit trail untuk model fallback, drop chat, escalate, tool pick, dst. decision_type slug (snake_case), rationale natural language, outcome success|fail|pending.
- **param:**
  - `decision_type` (string, WAJIB) — Slug (mis. model_fallback, drop_chat, escalate)
  - `rationale` (string, WAJIB) — Why this decision (1-3 kalimat)
  - `outcome` (string, opsional) — success|fail|pending

### `decision_search(decision_type, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Search decisions by decision_type substring. Recall keputusan historis.
- **param:**
  - `decision_type` (string, opsional) — Type slug substring (mis. 'escalate' atau 'model')
  - `limit` (int, opsional) — Max (default 30, max 200)

### `edu_error_count()`
- **cap:** `state:read` · **balik:** {total}
- **fungsi:** Count educational errors. Quick overview Section 9 catalog state.

### `edu_error_list(category, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List educational errors catalog (Section 9). Filter by category. Default limit 50 (max 500).
- **param:**
  - `category` (string, opsional) — Filter by category (network|tool|protected|...)
  - `limit` (int, opsional) — Max (default 50, max 500)

### `edu_error_lookup(code*)`
- **cap:** `state:read` · **balik:** {found, code, title, explanation, remediation}
- **fungsi:** Lookup educational error by code (Section 9). Return explanation + remediation untuk hadling error tertentu.
- **param:**
  - `code` (string, WAJIB) — Error code (mis. ERR_TOOL_NOT_ALLOWED)

### `edu_error_upsert(code*, category*, title*, explanation*, remediation*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Upsert educational error entry. Code PRIMARY KEY. Buat tambah/update edukatif error per Section 9.
- **param:**
  - `code` (string, WAJIB) — Error code (mis. ERR_TOOL_NOT_ALLOWED)
  - `category` (string, WAJIB) — Category
  - `title` (string, WAJIB) — Short title
  - `explanation` (string, WAJIB) — User-facing explanation
  - `remediation` (string, WAJIB) — How to fix

### `interaction_count()`
- **cap:** `state:read` · **balik:** {total}
- **fungsi:** Count total interactions logged. Quick KPI.

### `interaction_recall(channel, actor, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Query chat history dari interactions table. Anti over-prompt: tool ini on-demand only — TIDAK auto-inject ke system prompt. Default limit 10 (max 100).
- **param:**
  - `channel` (string, opsional) — Filter by channel (telegram|router|... atau kosong=all)
  - `actor` (string, opsional) — Filter by actor/chat_id (kosong=all)
  - `limit` (int, opsional) — Max entries (default 10, max 100)

### `karma_query(key)`
- **cap:** `state:read` · **balik:** {items: [{metric_key, metric_value, metric_count, updated_at}]}
- **fungsi:** Query karma metric self — read counter or average. Pakai key kalau tau (mis. 'success_count', 'avg_response_ms'), atau biarkan kosong untuk dump semua.
- **param:**
  - `key` (string, opsional) — Metric key (kosong = list all)

### `karma_set(key*, op*, value*)`
- **cap:** `state:write` · **balik:** {ok, current}
- **fungsi:** Update karma metric: op=increment (delta) atau op=average (sample). Counter atau moving average per Section 5.
- **param:**
  - `key` (string, WAJIB) — Metric key (mis. success_count, avg_response_ms)
  - `op` (string, WAJIB) — increment|average
  - `value` (float, WAJIB) — Delta for increment, sample for average

### `ledger_list(category, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List finance ledger entries — income+expense rows. Filter by category. Default limit 50.
- **param:**
  - `category` (string, opsional) — Filter (kosong=all)
  - `limit` (int, opsional) — Max (default 50, max 200)

### `retention_report()`
- **cap:** `state:read` · **balik:** {counts}
- **fungsi:** Snapshot table counts: interactions, decisions, mistakes_local, death_letter, workspace_meta. Buat health check.

### `watchdog_alerts_list(limit)`
- **cap:** `state:read` · **balik:** {count, alerts[]}
- **fungsi:** List watchdog alerts triggered — protector_burst, scanner_critical_burst, tool_call_storm rules.
- **param:**
  - `limit` (int, opsional) — Max (default 30, max 100)

### `zombie_findings_list(confidence, limit)`
- **cap:** `state:read` · **balik:** {count, findings[]}
- **fungsi:** List zombie code findings (Section 29) — file/symbol orphan tanpa caller. Filter by confidence.
- **param:**
  - `confidence` (string, opsional) — high|medium|low (kosong=all)
  - `limit` (int, opsional) — Max (default 30, max 100)

---

## 🛠️ Skill / Plugin / Tool-meta (20)

### `StructuredOutput(findings*)`
- **cap:** `—` · **balik:** {findings, ok}
- **fungsi:** Emit a structured result. Pass your findings as a JSON object/array; it is validated as present and returned verbatim as the structured output. Use this when a caller asked for machine-readable output.
- **param:**
  - `findings` (object, WAJIB) — the structured payload (object or array)

### `echo(message*)`
- **cap:** `—` · **balik:** {message: <input>}
- **fungsi:** Echo back the input message. Demo tool — verifies dispatcher wiring.
- **param:**
  - `message` (string, WAJIB) — text to echo

### `evolve_propose(target_file*, rationale*, kind, goal, risk)`
- **cap:** `state:write` · **balik:** {ok, id, status, pillar}
- **fungsi:** Salurkan IDE OWNER jadi proposal evolusi (masuk backlog → Dewan review → core-apply). Pakai pas owner kasih ide perbaikan/fitur lewat chat & minta diteruskan ke team evolusi. Ditandai [IDE OWNER] (prioritas, ga kena auto-reject classifier).
- **param:**
  - `target_file` (string, WAJIB) — file yang disentuh, relatif repo. File BARU pakai prefix 'NEW:' (mis. NEW:internal/agentdb/foo.go)
  - `rationale` (string, WAJIB) — kenapa ide ini penting (ringkas, 1-2 kalimat)
  - `kind` (string, opsional) — jenis: add-agent|add-skill|add-app (behavior) atau fix|refactor|doc|test (core). Default refactor
  - `goal` (string, opsional) — konteks/tujuan ide (optional)
  - `risk` (string, opsional) — low|medium|high (default medium)

### `now()`
- **cap:** `time:read` · **balik:** {rfc3339: '<UTC>', unix_ms: <int>, local: 'YYYY-MM-DD HH:MM:SS', tz_label: 'WIB', tz_offset_hours: 7}
- **fungsi:** Waktu sekarang LIVE: UTC (rfc3339) + waktu lokal default WIB (UTC+7). Pakai 'local' buat tanggal/jam terkini (anti berita-basi).

### `skill(name*)`
- **cap:** `rpc:router:skill` · **balik:** {name, description, body}
- **fungsi:** Retrieve full skill (name+description+body markdown) dari Router brain catalog. Caller treat body sebagai system-prompt-style instruction. Timeout 10s. Body cap 256KB.
- **param:**
  - `name` (string, WAJIB) — skill name (case-sensitive)

### `skill_add(id*, trigger, instructions*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Add agent skill (id + trigger + instructions). Anti over-prompt: max 3 auto-inject; sisanya via skill_search on-demand.
- **param:**
  - `id` (string, WAJIB) — Skill ID (snake_case)
  - `trigger` (string, opsional) — Trigger pattern (mis. '/version', '#crypto')
  - `instructions` (string, WAJIB) — Skill instruction text

### `skill_author(id*, instructions*, trigger, experience)`
- **cap:** `state:write` · **balik:** {ok, gate, id} on save — or {ok:false, blocked:true, reason, flags} when the gate rejects it
- **fungsi:** Self-author a reusable skill distilled from your own experience. The skill is VETTED (immune + verifier gate) before it can be saved — dangerous or injection-laden skills are BLOCKED, never stored. Use after you solve something worth remembering as a repeatable procedure.
- **param:**
  - `id` (string, WAJIB) — Skill ID (snake_case)
  - `instructions` (string, WAJIB) — The repeatable procedure to remember (what to do, step by step)
  - `trigger` (string, opsional) — When this skill applies (e.g. '#deploy', 'fix flaky test')
  - `experience` (string, opsional) — Provenance: the experience you distilled this from (for audit)

### `skill_remove(id*)`
- **cap:** `state:write` · **balik:** {ok, removed}
- **fungsi:** Remove agent skill by ID.
- **param:**
  - `id` (string, WAJIB) — Skill ID

### `skill_search(search, limit)`
- **cap:** `rpc:router:skill` · **balik:** {items: [{name, description}], count, total}
- **fungsi:** Search skill catalog di Router brain. Search optional (kosong=top of catalog). Returns summary (name + description) cap 10 per call (Router anti over-prompt).
- **param:**
  - `search` (string, opsional) — optional search query
  - `limit` (int, opsional) — default 10, max 10

### `skill_suggest(min_count, limit)`
- **cap:** `state:read` · **balik:** {count, candidates:[{tool_name, success_count, last_used, suggestion}]}
- **fungsi:** Lihat usulan SKILL dari pola tool yang sering lo pakai SUKSES. Pakai buat refleksi: tool apa yang udah jadi kebiasaan sukses → bisa diformalin jadi skill/alur. Balik kandidat urut paling sering.
- **param:**
  - `min_count` (int, opsional) — minimal jumlah sukses biar jadi kandidat (default 2)
  - `limit` (int, opsional) — max kandidat (default 10, max 50)

### `slash_alias_list()`
- **cap:** `state:read` · **balik:** {count, aliases[]}
- **fungsi:** List slash command aliases — alias → target_name mapping.

### `slash_alias_resolve(alias*)`
- **cap:** `state:read` · **balik:** {found, alias, target}
- **fungsi:** Resolve slash alias → canonical command name. Lookup-style.
- **param:**
  - `alias` (string, WAJIB) — Alias name (mis. 'v' → 'version')

### `slash_history(command, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List slash command history. Filter by command name. Default limit 30 (max 200).
- **param:**
  - `command` (string, opsional) — Filter by command (mis. 'version', 'help')
  - `limit` (int, opsional) — Max (default 30, max 200)

### `tool_audit_log(tool_name, decision, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Query tool_audit table — sandbox tool calls history (allowed/denied/pending). Filter by tool name or decision. Default limit 50 (max 200).
- **param:**
  - `tool_name` (string, opsional) — Filter by tool name (kosong=all)
  - `decision` (string, opsional) — allowed|denied_interceptor|pending_approve (kosong=all)
  - `limit` (int, opsional) — Max entries (default 50, max 200)

### `tool_create(name*, description*, code*, params, imports, capability, returns)`
- **cap:** `—` · **balik:** {ok, name, scope, build_log?}
- **fungsi:** Bikin tool sidecar SENDIRI. Lahir PRIVAT (cuma kamu pakai) sampai lolos Dewan→shared. `code` = badan `func run(args map[string]any) (any, string)`: ambil param dari args, balikin (output, errString). Gagal compile→build_log balik, perbaiki & ulang. DILARANG os/exec/syscall/unsafe (pure-compute; json/io/os auto-import).
- **param:**
  - `name` (string, WAJIB) — nama UNIK ^[a-z][a-z0-9_]{1,39}$ (ga nimpa tool lain)
  - `description` (string, WAJIB) — deskripsi + kapan dipakai
  - `code` (string, WAJIB) — badan func run(args map[string]any)(any,string)
  - `params` (array, opsional) — [{name,type,description,required}]
  - `imports` (array, opsional) — import Go ekstra, mis ["strings","regexp"]
  - `capability` (string, opsional) — kosongin (default). Privileged=review lebih ketat
  - `returns` (string, opsional) — bentuk output

### `tool_invocations_list(tool_name, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List tool invocations log — semua tool call yg agent ini lakukan + result + duration. Filter by tool name. Default limit 50.
- **param:**
  - `tool_name` (string, opsional) — Filter (kosong=all)
  - `limit` (int, opsional) — Max (default 50, max 200)

### `tool_lookup(name*)`
- **cap:** `state:read` · **balik:** {found, name, capability, schema}
- **fungsi:** Lookup single tool by name. Return Name + Capability + Schema (description+params+returns). Anti over-prompt: ngga return full catalog.
- **param:**
  - `name` (string, WAJIB) — Tool name

### `tool_search(query*)`
- **cap:** `state:read` · **balik:** {count, hits[]}
- **fungsi:** Search tool registry by substring di name/capability/description. Anti over-prompt: cap 10 hit. Pakai tool_lookup untuk detail spesifik.
- **param:**
  - `query` (string, WAJIB) — Substring (case-insensitive)

### `tool_subscribed_list()`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** List tools yang agent ini subscribe — source=manual|default|recommended. Self-introspection: 'gw punya tool apa aktif?'.

### `tool_subscriptions_count()`
- **cap:** `state:read` · **balik:** {total, by_source}
- **fungsi:** Quick count tool subscriptions + breakdown by source (manual/default/recommended).

---

## 🤝 Orkestrasi / Agent / Group (2)

### `agent_command(agent_id*, text*)`
- **cap:** `rpc:agent-invoke` · **balik:** {agent_id, reply}
- **fungsi:** Delegate a natural-language command to a specialist agent and get its reply back. Use this to ROUTE a request to an agent that owns the right tools/persona instead of doing it yourself. NOTE: computer power/control (shutdown, restart, sleep, lock, logout) and opening apps are FIRST-CLASS tools now — use system_power / app_open directly, do NOT delegate those. This tool is for genuine specialist delegation. Pass the request through as text; relay the reply to the user verbatim.
- **param:**
  - `agent_id` (string, WAJIB) — target specialist agent id
  - `text` (string, WAJIB) — the command / request in natural language

### `askuser(question*, reasoning)`
- **cap:** `state:write` · **balik:** {ok, decision_id, question}
- **fungsi:** Ask user untuk clarifikasi sebelum tindakan ambigu. LOG question + reasoning ke decisions table (decision_type='ask_clarification'). Caller (LLM) handle UI delivery via reply text. Anti over-tool: jangan panggil tiap step trivial. Pakai kalau input ambigu, multiple opsi tanpa hint, atau action irreversible.
- **param:**
  - `question` (string, WAJIB) — Pertanyaan ke user (1 kalimat jelas)
  - `reasoning` (string, opsional) — Kenapa lo butuh tanya (1-2 kalimat)

---

## 💰 Finance / Security / Misc (14)

### `death_letter_read(recipient, sealed_only, limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Baca wasiat warga sebelumnya (Predecessor Honor Protocol ADR-010). Warga baru WAJIB panggil saat boot ke workspace yg sama — biar inherit pembelajaran. Default sealed-only, limit 10.
- **param:**
  - `recipient` (string, opsional) — Filter by recipient (default 'all')
  - `sealed_only` (bool, opsional) — Only sealed (final) letters (default true)
  - `limit` (int, opsional) — Max (default 10, max 50)

### `death_letter_seal(id*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Seal death letter by id — set sealed_at, body immutable after. Final commit wasiat.
- **param:**
  - `id` (int, WAJIB) — Letter ID

### `death_letter_write(subject*, body*, letter_type, recipient)`
- **cap:** `state:write` · **balik:** {ok, id} kalau sukses
- **fungsi:** Tulis wasiat — pesan terakhir warga AI sebelum retire/upgrade/context-full. Warga baru di workspace sama auto-baca via Predecessor Honor Protocol (ADR-010). Subject + body required, recipient default 'all', letter_type default 'reflection'.
- **param:**
  - `subject` (string, WAJIB) — Judul wasiat singkat
  - `body` (string, WAJIB) — Isi wasiat — markdown OK
  - `letter_type` (string, opsional) — reflection|handover|warning|legacy (default reflection)
  - `recipient` (string, opsional) — Penerima (default 'all')

### `finance_budget_set(metric_key*, budget_value*, warning_at_pct)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Set budget per metric_key. warning_at_pct = warning threshold (0-1).
- **param:**
  - `metric_key` (string, WAJIB) — Metric key (mis. 'daily_cost_usd')
  - `budget_value` (float, WAJIB) — Hard limit
  - `warning_at_pct` (float, opsional) — Warning pct 0-1 (default 0.8)

### `finance_budgets()`
- **cap:** `state:read` · **balik:** {count, budgets[]}
- **fungsi:** List finance budgets configured untuk agent. Self-introspection budget ceiling per category.

### `finance_log(entry_type*, amount*, currency*, category, note)`
- **cap:** `state:write` · **balik:** {ok, id}
- **fungsi:** Log finance entry — income/expense. Amount positive (sign by entry_type). Currency 3-letter ISO (USD/IDR/BTC).
- **param:**
  - `entry_type` (string, WAJIB) — income|expense
  - `amount` (float, WAJIB) — Amount (positive, sign by entry_type)
  - `currency` (string, WAJIB) — ISO 3-letter (USD/IDR/BTC/...)
  - `category` (string, opsional) — tax|food|api_credit|salary|...
  - `note` (string, opsional) — Free-text note

### `finance_summary(period)`
- **cap:** `state:read` · **balik:** {period, income, expense, net, count}
- **fungsi:** Finance ledger summary — income/expense/net per period. Period: 24h/7d/30d (default 30d).
- **param:**
  - `period` (string, opsional) — 24h|7d|30d (default 30d)

### `protector_audit_query(limit)`
- **cap:** `state:read` · **balik:** {count, items[]}
- **fungsi:** Query protector_audit log — protector rule trigger history (allow/block decisions). Default limit 30.
- **param:**
  - `limit` (int, opsional) — Max (default 30, max 200)

### `protector_rule_add(name*, pattern*, action*, reason)`
- **cap:** `state:write` · **balik:** {ok, id}
- **fungsi:** Add protector rule — pattern matching kit. Pattern target message/tool call yang harus di-block. Reason+severity untuk audit trail.
- **param:**
  - `name` (string, WAJIB) — Rule name (snake_case)
  - `pattern` (string, WAJIB) — Match pattern (substring atau regex)
  - `action` (string, WAJIB) — block|allow|warn
  - `reason` (string, opsional) — Why this rule

### `protector_rule_delete(id*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Delete protector rule by ID.
- **param:**
  - `id` (int, WAJIB) — Rule ID

### `protector_rule_toggle(id*, enabled*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Toggle protector rule enabled/disabled by ID.
- **param:**
  - `id` (int, WAJIB) — Rule ID
  - `enabled` (bool, WAJIB) — true=enable, false=disable

### `protector_rules_list()`
- **cap:** `state:read` · **balik:** {count, rules[]}
- **fungsi:** List protector rules untuk agent — pattern matching kit yg block/allow tool call atau message.

### `secret_get_keys()`
- **cap:** `state:read` · **balik:** {keys[]}
- **fungsi:** List secret keys (names only, NO values). Anti leak: introspection 'gw punya credentials apa' tanpa expose value.

### `secret_set(key*, value*)`
- **cap:** `state:write` · **balik:** {ok}
- **fungsi:** Set secret (env credential). Stored di secrets table. Re-set = update. Key uppercase convention (mis. TELEGRAM_BOT_TOKEN).
- **param:**
  - `key` (string, WAJIB) — Secret key (UPPER_SNAKE_CASE)
  - `value` (string, WAJIB) — Secret value

---

## ⌨️ SLASH COMMAND (Telegram)

Slash command Flowork **AUTO-DERIVED** dari daftar **GROUP** (modul mana pun ber-`kv group=1`) —
BUKAN hardcode. 1 group → 1 slash (`telegramCommand()` sanitize jadi nama valid+unik). Sumber
tunggal = `kv "groups"`; menu slash, `ask_group`, schedule, dan Mr.Flow baca list yang SAMA.

- **Di-push ke Telegram** (`setMyCommands`) **DI-GATE saklar `FLOWORK_GROUP_SLASH`** —
  **DEFAULT MATI** (owner 2026-06-23: andelin KESADARAN mr-flow, buang slash biar menu bersih).
  Pas mati → menu slash kosong TAPI group tetep jalan via Mr.Flow (bahasa natural / `ask_group`).
- **Idupin:** `FLOWORK_GROUP_SLASH=1` (bisa via GUI **Switch Fitur** kalau ditambah ke registry fwswitch).
- **Tool terkait** (ada di katalog atas): `slash_alias_list`, `slash_alias_resolve`, `slash_history`.
- Kode: `agent/internal/groupsapi/orchestrator.go` (+ saklar `groupsapi_ext.go`).

---

## ⏳ WAIT / WAKE / PAUSE / LOOP — mekanisme & cara kerja

Ini jawaban "gimana AI Flowork nunggu eksekusi / looping / tidur-bangun sendiri" (mirip gw pas
`ScheduleWakeup`). Ada 2 lapis: **loop in-turn** + **kontinuasi lintas-turn**.

### 1. TOOL-LOOP (in-turn) — `agentkit.go`
Tiap turn: `model → manggil tool → hasil di-feed balik → model lanjut → ulang` sampai model kasih
jawaban final (no tool). Batas aman:
- `maxToolIters = 100` — backstop anti infinite (BUKAN batas kerja nyata).
- `loopBudgetMs ≈ 200000` (~200 dtk) — budget waktu in-turn. `turn-timeout 290s` = backstop keras.
- Serialize 1 tool/iter (`parallel_tool_calls:false`).

### 2. AUTO-CONTINUE (lintas-turn) — kerja panjang ga putus
Pas `loopBudgetMs` kelewat & tugas belum kelar → agent OTOMATIS panggil
`ScheduleWakeup(delaySeconds:5, prompt:<lanjutan>)` → **kebangun sendiri di turn baru & lanjut**.
Bounded `maxAutoContinue = 50` (≈2.7 jam) anti-runaway. Nyambung lintas-turn = **unbounded over time**
(kerja berhari-hari via tidur-bangun, bukan loop panas yang ngabisin CPU/token).

### 3. TOOL khusus wait/wake/loop
| Tool | Sinkron? | Cara kerja |
|---|---|---|
| **`ScheduleWakeup(delaySeconds*, reason*, prompt*)`** | async (lintas-turn) | Nyimpen 1 baris `wakeups` (due-time, prompt). Engine/loop fire pas due → AI kebangun & jalanin `prompt` lagi. = **"tidur lalu bangun sendiri"** buat nungguin sesuatu / kerja bertahap. INI mekanisme utama looping AI Flowork. |
| **`Monitor(until*, command*, timeout_ms, persistent)`** | SINKRON (in-turn) | Poll `command` tiap ~2 dtk sampai output ngandung `until` ATAU timeout (cap 60s). Buat nungguin kondisi DALAM 1 turn (file muncul, proses "ready"). |
| **`agent_run(action*, id, label, data)`** | durable | Lifecycle kerja panjang: `create/start/checkpoint/pause/resume/stop/complete/status/list`. `pause`+`resume` survive lintas-turn; `resume` balikin checkpoint terakhir → lanjut dari titik berhenti. |
| **`scheduler_schedule_add/remove/list/next` · `schedule_runs_query`** | cron | Jadwal berkala (tidur→bangun tiap cron, bukan loop panas). History run via `schedule_runs_query`. |
| **`brain_dream(factor)`** | offline | Konsolidasi memori ALA TIDUR: decay importance memori yg ga pernah di-recall (importance-only, ga hapus). |

### 4. ANTI-GHOST (kenapa wajib pakai tool ini, bukan ngomong doang)
Kalau model bilang "gw cek/scan/tunggu dulu" TAPI **ga manggil tool** → owner nungguin jawaban yg
ga datang = **ghosting (DILARANG)**. `ghost-guard` (agentkit + mr-flow) maksa: panggil tool yg
dimaksud, ATAU kalau emang lagi NUNGGU sesuatu yg belum siap → **`ScheduleWakeup`**. `flail-guard`
maksa hal sama pas tool diulang-ulang tanpa progress. (lihat `lock/mrflow.md` ghost-guard v2.)

---

## 🔌 TOOL MCP / DINAMIS (kenapa total bisa "200an")

Katalog di atas = **150 builtin** ter-register (`tools.List()`). Di ATAS ini masih ada:
- **Tool MCP eksternal** — server MCP yg ke-connect → tool-nya muncul DINAMIS per-sesi (via ToolSearch/`tool_search`).
- **Tool plugin** — di-install `.fwpack` lewat `POST /api/tools/install` (lihat `tool_install.go`).
- **Deferred tools (#2C all-tools)** — di-fetch on-demand via `tool_lookup`/`tool_search` (hemat prompt).

Total efektif yg "kelihatan" AI tergantung MCP/plugin yg ke-pasang → bisa 200+. **Regen katalog
builtin** (kalau tool nambah/berubah): `cd agent && go run ./cmd/tooldump` → JSON name/cap/desc/param,
lalu format ke markdown ini. Katalog builtin = sumber kebenaran dari `tools.List()`, JANGAN edit tangan.
