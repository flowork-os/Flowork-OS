// apps.js — tab "App": a browser-like tabbed shell.
//
// Behaviour (Chrome-style):
//   • Tab 1 is the launcher (the app list) and can never be closed.
//   • Opening an app spawns a NEW tab running that app's sandboxed GUI; open many.
//   • An app only runs while its tab exists — closing a tab unloads the iframe and
//     stops its bridge + state poll (the app dies). No tab → not running.
//   • The shell lives in <body>, OUTSIDE #main, so switching sidebar menus only
//     HIDES it (display:none keeps every iframe loaded → apps keep working). It
//     reappears, still running, when you come back to the App menu.
//   • Open tabs persist across a page refresh (localStorage); apps reload then.
//
// State lives on `window.__fwApps` (a singleton): the router re-imports each tab
// module with a cache-buster, so module-scope would reset on every visit — the
// shell + open tabs must outlive that. The DOM shell + iframes outlive it too.
//
// Security unchanged: each app GUI is a <iframe sandbox="allow-scripts"> (no
// same-origin). Its only channel is postMessage {op,args}, validated against the
// app's manifest ops, forwarded to /api/apps/op.
import { esc, escAttr, fetchJSON, loadStyle } from '../js/utils.js';
import { t } from '/js/i18n.js';

const L = new Proxy({}, { get: (_, k) => t('apps.' + String(k)) });
const LS_KEY = 'flowork_app_tabs';

// singleton state — survives the router's cache-busted re-imports.
const S = (window.__fwApps = window.__fwApps || {
  shell: null, tabs: [], activeKey: '__home', seg: 'installed', restored: false, apps: [],
});

const CSS = `
.app-shell { position:fixed; z-index:40; display:none; flex-direction:column; background:#0a0e16;
  border-left:1px solid rgba(148,163,184,0.12); }
.as-tabbar { display:flex; align-items:center; gap:3px; padding:7px 9px 0; background:rgba(15,23,42,0.7);
  border-bottom:1px solid rgba(148,163,184,0.16); overflow-x:auto; flex:0 0 auto; }
.as-tab { display:inline-flex; align-items:center; gap:8px; padding:8px 13px; border-radius:10px 10px 0 0;
  background:transparent; color:#94a3b8; cursor:pointer; font-size:0.85rem; border:1px solid transparent;
  border-bottom:none; white-space:nowrap; max-width:230px; transition:background .15s,color .15s; user-select:none; }
.as-tab:hover { background:rgba(148,163,184,0.08); color:#cbd5e1; }
.as-tab.on { background:#0e1320; color:#e2e8f0; border-color:rgba(148,163,184,0.18); }
.as-tab img { width:18px; height:18px; border-radius:4px; flex:0 0 auto; }
.as-tab .nm { overflow:hidden; text-overflow:ellipsis; }
.as-tab .cl { margin-left:2px; opacity:.55; border-radius:5px; width:18px; height:18px; line-height:16px; text-align:center; flex:0 0 auto; }
.as-tab .cl:hover { opacity:1; background:rgba(248,113,113,0.22); color:#f87171; }
.as-tab.home { font-size:1.05rem; padding:8px 14px; }
.as-body { flex:1; position:relative; overflow:hidden; }
.as-pane { position:absolute; inset:0; display:none; overflow:auto; }
.as-pane.on { display:block; }
.as-pane iframe { width:100%; height:100%; border:0; background:#06121a; display:block; }

/* ── launcher (home pane) — clean, matches the Agent/Group language ── */
.as-launch { padding:26px 30px 50px; color:#e2e8f0; }
.al-hero { padding:26px 30px; border-radius:16px; margin-bottom:22px;
  background:linear-gradient(135deg, rgba(124,58,237,0.18) 0%, rgba(14,165,233,0.14) 52%, rgba(16,185,129,0.13) 100%);
  border:1px solid rgba(148,163,184,0.2); }
.al-eyebrow { font-size:0.72rem; letter-spacing:0.3em; color:#a78bfa; text-transform:uppercase; font-weight:600; margin-bottom:7px; }
.al-h1 { margin:0; font-size:1.8rem; font-weight:700; background:linear-gradient(90deg,#c4b5fd,#67e8f9 55%,#6ee7b7);
  -webkit-background-clip:text; background-clip:text; color:transparent; }
.al-sub { margin:8px 0 0; color:#cbd5e1; font-size:0.92rem; }
.al-seg { display:flex; gap:8px; margin-bottom:18px; }
.al-segbtn { padding:7px 15px; border-radius:9px; background:rgba(2,6,18,0.4); border:1px solid rgba(148,163,184,0.2);
  color:#94a3b8; cursor:pointer; font:inherit; font-size:0.84rem; transition:all .15s; }
.al-segbtn.on { background:rgba(124,58,237,0.18); border-color:rgba(167,139,250,0.5); color:#c4b5fd; }
.al-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(140px,1fr)); gap:16px; }
.al-card { position:relative; background:rgba(15,23,42,0.6); border:1px solid rgba(148,163,184,0.18); border-radius:14px;
  padding:18px 12px; text-align:center; cursor:pointer; transition:all .15s; }
.al-card:hover { border-color:rgba(167,139,250,0.5); transform:translateY(-2px); box-shadow:0 14px 34px -24px rgba(124,58,237,0.5); }
.al-card img { width:46px; height:46px; }
.al-glyph { font-size:40px; line-height:46px; height:46px; width:46px; margin:0 auto; text-align:center; }
.al-card .nm { font-size:0.86rem; color:#f1f5f9; margin-top:9px; word-break:break-word; }
.al-card .rt { font-size:0.66rem; letter-spacing:0.06em; color:#94a3b8; margin-top:4px; }
.al-card .x { position:absolute; top:5px; right:8px; color:#f87171; opacity:0; font-size:0.85rem; transition:opacity .15s; }
.al-card:hover .x { opacity:.7; }
.al-card .x:hover { opacity:1; }
.al-empty { color:#94a3b8; text-align:center; padding:30px; font-size:0.9rem; border:1px dashed rgba(148,163,184,0.25); border-radius:12px; }
.al-store { background:rgba(15,23,42,0.5); border:1px solid rgba(148,163,184,0.18); border-radius:12px; padding:18px 20px;
  color:#cbd5e1; font-size:0.9rem; line-height:1.65; }
.al-store code { color:#67e8f9; background:rgba(14,165,233,0.1); padding:1px 6px; border-radius:4px; }
.al-msg { font-size:0.84rem; margin-left:10px; }
/* ── adopt repo panel ── */
.ad-wrap { background:rgba(15,23,42,0.5); border:1px solid rgba(148,163,184,0.18); border-radius:12px; padding:18px 20px; color:#cbd5e1; }
.ad-row { display:flex; gap:9px; margin-bottom:4px; }
.ad-in { flex:1; background:rgba(2,6,18,0.5); border:1px solid rgba(148,163,184,0.22); border-radius:8px; color:#e2e8f0; padding:9px 12px; font:inherit; font-size:0.86rem; }
.ad-btn { padding:9px 16px; border-radius:8px; background:rgba(124,58,237,0.2); border:1px solid rgba(167,139,250,0.5); color:#c4b5fd; cursor:pointer; font:inherit; font-size:0.85rem; white-space:nowrap; }
.ad-btn:hover { background:rgba(124,58,237,0.32); }
.ad-btn.go { background:rgba(16,185,129,0.18); border-color:rgba(52,211,153,0.5); color:#6ee7b7; }
.ad-btn:disabled { opacity:.45; cursor:not-allowed; }
.ad-hint { font-size:0.74rem; color:#64748b; margin:2px 2px 14px; }
.ad-box { border-radius:10px; padding:13px 15px; margin-top:14px; font-size:0.85rem; line-height:1.6; }
.ad-ok { background:rgba(16,185,129,0.1); border:1px solid rgba(16,185,129,0.3); }
.ad-warn { background:rgba(234,179,8,0.1); border:1px solid rgba(234,179,8,0.35); color:#fde68a; }
.ad-crit { background:rgba(248,113,113,0.12); border:1px solid rgba(248,113,113,0.45); color:#fca5a5; }
.ad-find { font-family:ui-monospace,monospace; font-size:0.76rem; margin:5px 0 0; opacity:.9; }
.ad-badge { display:inline-block; background:rgba(14,165,233,0.14); color:#67e8f9; border-radius:5px; padding:1px 8px; font-size:0.76rem; margin-right:6px; }
.ad-lab { display:block; font-size:0.76rem; color:#94a3b8; margin:11px 0 4px; }
.ad-ct { display:flex; gap:14px; margin:8px 0; font-size:0.85rem; }
.ad-grid2 { display:grid; grid-template-columns:1fr 1fr; gap:9px; }
`;

function topTab() { return (location.hash || '').replace(/^#\/?/, '').split('/')[0] || ''; }
function visible() { return S.shell && S.shell.style.display !== 'none'; }

function positionShell() {
  const main = document.getElementById('main');
  if (!main || !S.shell) return;
  const r = main.getBoundingClientRect();
  S.shell.style.left = r.left + 'px';
  S.shell.style.top = r.top + 'px';
  S.shell.style.width = r.width + 'px';
  S.shell.style.height = r.height + 'px';
}
function show() { if (S.shell) { S.shell.style.display = 'flex'; positionShell(); } }
function hide() { if (S.shell) S.shell.style.display = 'none'; } // iframes stay loaded → apps keep running

function ensureShell() {
  if (S.shell) return;
  const shell = document.createElement('div');
  shell.className = 'app-shell';
  shell.innerHTML = `<div class="as-tabbar" id="asTabbar"></div><div class="as-body" id="asBody"></div>`;
  document.body.appendChild(shell);
  S.shell = shell;

  // the launcher (home) tab — built once, never closed.
  const home = document.createElement('div');
  home.className = 'as-pane as-launch';
  shell.querySelector('#asBody').appendChild(home);
  S.tabs.push({ key: '__home', pane: home, home: true });

  // keep the shell glued to #main as it resizes (nav collapse, window resize).
  const main = document.getElementById('main');
  if (main && 'ResizeObserver' in window) { new ResizeObserver(() => { if (visible()) positionShell(); }).observe(main); }
  window.addEventListener('resize', () => { if (visible()) positionShell(); });
  // route awareness: only the App menu shows the shell; others hide it (apps live on).
  window.addEventListener('hashchange', () => { topTab() === 'apps' ? show() : hide(); });

  renderTabbar();
  activate('__home');
}

// ── tab bar + panes ───────────────────────────────────────────────────────────
function renderTabbar() {
  const bar = S.shell.querySelector('#asTabbar');
  bar.innerHTML = S.tabs.map((tb) => {
    if (tb.home) return `<div class="as-tab home ${S.activeKey === tb.key ? 'on' : ''}" data-key="${escAttr(tb.key)}" title="${escAttr(L.title)}">▦</div>`;
    const a = tb.app;
    return `<div class="as-tab ${S.activeKey === tb.key ? 'on' : ''}" data-key="${escAttr(tb.key)}" title="${escAttr(a.name || a.id)}">
      <img src="/api/apps/${escAttr(a.id)}/${escAttr(a.icon || 'ui/icon.svg')}" alt="" onerror="this.style.display='none'">
      <span class="nm">${esc(a.name || a.id)}</span>
      <span class="cl" title="${escAttr(L.close)}">✕</span>
    </div>`;
  }).join('');
  bar.querySelectorAll('.as-tab').forEach((el) => {
    const key = el.dataset.key;
    el.onclick = (e) => { if (e.target.classList.contains('cl')) { e.stopPropagation(); closeTab(key); } else activate(key); };
  });
}

function activate(key) {
  S.activeKey = key;
  S.tabs.forEach((tb) => tb.pane.classList.toggle('on', tb.key === key));
  renderTabbar();
  persist();
}

function closeTab(key) {
  const i = S.tabs.findIndex((tb) => tb.key === key);
  if (i < 0 || S.tabs[i].home) return; // never close the launcher
  const tb = S.tabs[i];
  if (tb.poll) clearInterval(tb.poll);
  if (tb.bridge) window.removeEventListener('message', tb.bridge);
  tb.pane.remove(); // unloads the iframe (client side)
  // stop the core PROCESS too — an app runs only while a tab is open (best-effort).
  if (tb.app) fetch('/api/apps/stop?id=' + encodeURIComponent(tb.app.id), { method: 'POST' }).catch(() => {});
  S.tabs.splice(i, 1);
  if (S.activeKey === key) activate(S.tabs[Math.max(0, i - 1)].key);
  else { renderTabbar(); persist(); }
}

// ── open an app in a tab (sandboxed iframe + bridge + state poll) ──────────────
function openApp(a) {
  // app SERVER (kontrak HTTP, punya op _url) → buka beda: start server + buka URL, BUKAN iframe folder.
  if ((a.operations || []).some((o) => o.name === '_url')) { openServerApp(a); return; }
  const key = 'app:' + a.id;
  if (S.tabs.find((tb) => tb.key === key)) { activate(key); return; }

  const pane = document.createElement('div');
  pane.className = 'as-pane';
  const frame = document.createElement('iframe');
  frame.sandbox = 'allow-scripts';
  frame.src = `/api/apps/${a.id}/${a.gui_entry || 'ui/index.html'}`;
  pane.appendChild(frame);
  S.shell.querySelector('#asBody').appendChild(pane);

  const ops = new Set((a.operations || []).map((o) => o.name));
  const bridge = async (e) => {
    if (e.source !== frame.contentWindow) return; // only this tab's iframe
    const d = e.data || {};
    if (d.fw !== 1 || d.kind !== 'op') return;
    const reply = (extra) => frame.contentWindow.postMessage({ fw: 1, kind: 'res', reqId: d.reqId, ...extra }, '*');
    if (!ops.has(d.op)) { reply({ ok: false, error: 'op tak terdaftar' }); return; }
    try {
      const r = await fetchJSON('/api/apps/op', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ app: a.id, op: d.op, args: d.args || {} }) });
      reply({ ok: true, result: r.result });
    } catch (err) { reply({ ok: false, error: String(err.message || err) }); }
  };
  window.addEventListener('message', bridge);

  let lastVer = -1;
  const poll = setInterval(async () => {
    try { const s = await fetchJSON('/api/apps/state?id=' + encodeURIComponent(a.id)); if (s.version !== lastVer) { lastVer = s.version; frame.contentWindow.postMessage({ fw: 1, kind: 'state', version: s.version }, '*'); } } catch {}
  }, 2000);

  S.tabs.push({ key, app: a, pane, frame, bridge, poll });
  activate(key);
}

// ── open a SERVER app (kontrak HTTP): start server → buka URL (F5/F4) ──────────
function openServerApp(a) {
  const key = 'app:' + a.id;
  if (S.tabs.find((tb) => tb.key === key)) { activate(key); return; }
  const pane = document.createElement('div');
  pane.className = 'as-pane';
  pane.style.cssText = 'padding:28px; overflow:auto';
  pane.innerHTML = `<div class="ad-wrap"><b>${esc(a.name || a.id)}</b> — app server (web)
    <div id="srvMsg" class="ad-find" style="margin-top:9px">⟳ Menyalakan server… (boot + tunggu port siap)</div>
    <div id="srvAct" style="margin-top:14px"></div></div>`;
  S.shell.querySelector('#asBody').appendChild(pane);
  S.tabs.push({ key, app: a, pane });
  activate(key);
  (async () => {
    const msg = pane.querySelector('#srvMsg'), act = pane.querySelector('#srvAct');
    const call = (op) => fetchJSON('/api/apps/op', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ app: a.id, op, args: {} }) });
    try {
      await call('_alive');                       // start server + tunggu ready
      const u = await call('_url');               // alamat UI
      const url = (u.result && u.result.url) || '';
      msg.textContent = '✓ Server jalan — ' + url;
      act.innerHTML = `<button class="ad-btn go" id="srvOpen">↗ Buka UI di tab baru</button>
        <span class="ad-hint" style="display:block;margin-top:8px">Streamlit/web app jalan paling baik di tab terpisah.</span>`;
      pane.querySelector('#srvOpen').onclick = () => window.open(url, '_blank', 'noopener');
      const f = document.createElement('iframe');
      f.src = url; f.style.cssText = 'width:100%;height:68vh;border:1px solid rgba(148,163,184,0.2);border-radius:10px;margin-top:14px;background:#fff';
      pane.querySelector('.ad-wrap').appendChild(f);
    } catch (e) { msg.textContent = '✕ Server gagal start: ' + String(e.message || e); }
  })();
}

// ── adopt repo → app (F4): paste URL → deteksi+scan → kontrak → jalan ──────────
function renderAdopt(body) {
  const A = S.adopt || (S.adopt = {});
  body.innerHTML = `<div class="ad-wrap">
    <div class="ad-row">
      <input class="ad-in" id="adSrc" placeholder="https://github.com/owner/repo   ·   atau /path/folder" value="${escAttr(A.src || '')}">
      <button class="ad-btn" id="adDetect">🔍 Deteksi</button>
    </div>
    <div class="ad-hint">Clone repo → deteksi runtime → scan keamanan → jadi app (manusia + AI bisa jalanin). Owner approve sebelum jalan.</div>
    <div id="adRes"></div>
  </div>`;
  const src = body.querySelector('#adSrc');
  const go = () => { A.src = src.value.trim(); doDetect(body); };
  body.querySelector('#adDetect').onclick = go;
  src.onkeydown = (e) => { if (e.key === 'Enter') go(); };
  if (A.det) renderAdoptResult(body);
}

async function doDetect(body) {
  const A = S.adopt, res = body.querySelector('#adRes');
  if (!A.src) { res.innerHTML = `<div class="ad-box ad-warn">Isi URL/path repo dulu.</div>`; return; }
  res.innerHTML = `<div class="ad-box ad-ok">⟳ Mendeteksi + scan… (clone shallow, bisa beberapa detik)</div>`;
  try {
    const r = await fetchJSON('/api/apps/detect', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ source: A.src }) });
    A.det = r.detection || {}; A.scan = r.scan || {}; A.suggestedId = r.suggested_id || ''; A.accept = false;
    // auto-saran kontrak: kalau framework server kedeteksi → pre-pilih HTTP + isi start_cmd/port.
    const sg = r.suggest || {};
    if (sg.contract === 'http') {
      A.contract = 'http';
      if (!A.startCmd && (sg.start_cmd || []).length) A.startCmd = sg.start_cmd.join(' ');
      if (!A.port && sg.port) A.port = String(sg.port);
      A.suggestReason = sg.reason || '';
    } else { A.contract = A.contract || 'cli'; A.suggestReason = ''; }
    renderAdoptResult(body);
  } catch (e) { res.innerHTML = `<div class="ad-box ad-crit">✕ Deteksi gagal: ${esc(String(e.message || e))}</div>`; }
}

function renderAdoptResult(body) {
  const A = S.adopt, res = body.querySelector('#adRes'); if (!res) return;
  const d = A.det || {}, sc = A.scan || {};
  const crit = sc.critical || 0, warn = sc.warn || 0;
  const findHTML = (sc.findings || []).slice(0, 12).map((f) =>
    `<div class="ad-find">[${esc(f.severity)}] ${esc(f.file)}:${f.line} — ${esc(f.pattern)}</div>`).join('');
  let scanBox;
  if (crit > 0) scanBox = `<div class="ad-box ad-crit"><b>⚠ ${crit} pola BERBAHAYA (critical)</b> di kode repo — adopt diblok demi keamanan.${findHTML}
    <label class="ad-lab"><input type="checkbox" id="adAccept" ${A.accept ? 'checked' : ''}> Saya paham risiko & tetap lanjut (accept_risk)</label></div>`;
  else if (warn > 0) scanBox = `<div class="ad-box ad-warn"><b>${warn} peringatan</b> — cek dulu:${findHTML}</div>`;
  else scanBox = `<div class="ad-box ad-ok">✓ Scan bersih — nol pola berbahaya.</div>`;
  const inst = (d.install_cmd || []).map((c) => '$ ' + c.join(' ')).join('   ·   ');
  const isHTTP = A.contract === 'http', isMCP = A.contract === 'mcp';
  res.innerHTML = `
    <div class="ad-box ad-ok">
      <span class="ad-badge">runtime: ${esc(d.runtime || '?')}</span>
      ${d.entry ? `<span class="ad-badge">entry: ${esc(d.entry)}</span>` : ''}
      ${inst ? `<div class="ad-find" style="margin-top:6px">install: ${esc(inst)}</div>` : ''}
      ${(d.notes || []).map((n) => `<div class="ad-find" style="opacity:.8">• ${esc(n)}</div>`).join('')}
    </div>
    ${scanBox}
    <label class="ad-lab">Jenis app${A.suggestReason ? ` <span style="color:#6ee7b7">· saran: ${esc(A.suggestReason)}</span>` : ''}</label>
    <div class="ad-ct">
      <label><input type="radio" name="adCt" value="cli" ${!isHTTP && !isMCP ? 'checked' : ''}> CLI / tool</label>
      <label><input type="radio" name="adCt" value="http" ${isHTTP ? 'checked' : ''}> Server / web app (HTTP)</label>
      <label><input type="radio" name="adCt" value="mcp" ${isMCP ? 'checked' : ''}> MCP server (tool AI)</label>
    </div>
    <div id="adHttp" style="display:${isHTTP ? 'block' : 'none'}">
      <div class="ad-grid2">
        <div><label class="ad-lab">Perintah start server</label><input class="ad-in" id="adStart" placeholder="python main.py" value="${escAttr(A.startCmd || '')}"></div>
        <div><label class="ad-lab">Port</label><input class="ad-in" id="adPort" placeholder="8080" value="${escAttr(A.port || '')}"></div>
        <div><label class="ad-lab">Ready path (opsional)</label><input class="ad-in" id="adReady" placeholder="/" value="${escAttr(A.ready || '')}"></div>
        <div><label class="ad-lab">URL path UI (opsional)</label><input class="ad-in" id="adUrl" placeholder="/" value="${escAttr(A.urlPath || '')}"></div>
      </div>
    </div>
    <div id="adMcp" style="display:${isMCP ? 'block' : 'none'}">
      <div class="ad-grid2">
        <div><label class="ad-lab">Command (allowlist: node/python3/npx/uvx…)</label><input class="ad-in" id="adCmd" placeholder="node" value="${escAttr(A.mcpCmd || '')}"></div>
        <div><label class="ad-lab">Args (entry repo + flag)</label><input class="ad-in" id="adArgs" placeholder="dist/index.js" value="${escAttr(A.mcpArgs || '')}"></div>
      </div>
    </div>
    <label class="ad-lab">App ID</label>
    <input class="ad-in" id="adId" value="${escAttr(A.id || A.suggestedId || '')}">
    <div style="margin-top:15px"><button class="ad-btn go" id="adGo">＋ Adopt & Jalankan</button><span class="al-msg" id="adMsg"></span></div>`;
  res.querySelectorAll('[name=adCt]').forEach((r) => r.onchange = () => { A.contract = r.value; renderAdoptResult(body); });
  const acc = res.querySelector('#adAccept'); if (acc) acc.onchange = () => { A.accept = acc.checked; };
  const cap = (sel, key) => { const el = res.querySelector(sel); if (el) el.oninput = () => A[key] = el.value; };
  cap('#adStart', 'startCmd'); cap('#adPort', 'port'); cap('#adReady', 'ready'); cap('#adUrl', 'urlPath'); cap('#adId', 'id');
  cap('#adCmd', 'mcpCmd'); cap('#adArgs', 'mcpArgs');
  res.querySelector('#adGo').onclick = () => doAdopt(body);
}

async function doAdopt(body) {
  const A = S.adopt, msg = body.querySelector('#adMsg'), go = body.querySelector('#adGo');
  if (!A.src) return;
  const payload = { source: A.src, id: (A.id || '').trim(), accept_risk: !!A.accept };
  if (A.contract === 'http') {
    const port = parseInt(A.port, 10);
    if (!A.startCmd || !port) { msg.textContent = '✕ Server butuh perintah start + port'; return; }
    payload.contract = 'http';
    payload.http = { start_cmd: A.startCmd.trim().split(/\s+/), port, ready_path: (A.ready || '').trim(), url_path: (A.urlPath || '').trim(), ops: {} };
  }
  if (A.contract === 'mcp') {
    if (!A.mcpCmd) { msg.textContent = '✕ MCP butuh command (node/python3/npx/uvx…)'; return; }
    payload.contract = 'mcp';
    payload.mcp = { command: A.mcpCmd.trim(), args: (A.mcpArgs || '').trim() ? A.mcpArgs.trim().split(/\s+/) : [] };
  }
  if (!confirm('Adopt repo ini? Flowork clone + install dependency + bikin app. Jalanin cuma kalau lo percaya sumbernya.')) return;
  go.disabled = true; msg.textContent = '⟳ Clone + install + bikin app… (dep besar bisa beberapa menit)';
  try {
    const resp = await fetch('/api/apps/adopt?approve_exec=1', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
    const r = await resp.json();
    if (!resp.ok || r.error) throw new Error(r.error || ('HTTP ' + resp.status));
    msg.textContent = '✓ App LIVE: ' + ((r.result && r.result.id) || '');
    S.adopt = {}; S.seg = 'installed';
    await refreshApps();
  } catch (e) { go.disabled = false; msg.textContent = '✕ ' + String(e.message || e); }
}

// ── persistence (survive refresh) ──────────────────────────────────────────────
function persist() {
  try {
    const open = S.tabs.filter((tb) => !tb.home).map((tb) => tb.app.id);
    localStorage.setItem(LS_KEY, JSON.stringify({ open, active: S.activeKey }));
  } catch {}
}
function restoreTabs() {
  let saved;
  try { saved = JSON.parse(localStorage.getItem(LS_KEY) || '{}'); } catch { saved = {}; }
  for (const id of saved.open || []) {
    const a = S.apps.find((x) => x.id === id);
    if (a) openApp(a);
  }
  if (saved.active && S.tabs.some((tb) => tb.key === saved.active)) activate(saved.active);
  else activate('__home');
}

// ── launcher (home pane) render ────────────────────────────────────────────────
async function refreshApps() {
  try { S.apps = (await fetchJSON('/api/apps')).apps || []; } catch { S.apps = []; }
  renderHome();
}

function renderHome() {
  const home = S.tabs.find((tb) => tb.home);
  if (!home) return;
  home.pane.innerHTML = `
    <div class="al-hero">
      <div class="al-eyebrow">FLOWORK · APPS</div>
      <h1 class="al-h1">${esc(L.title)}</h1>
      <p class="al-sub">${esc(L.sub)}</p>
    </div>
    <div class="al-seg">
      <button class="al-segbtn ${S.seg === 'installed' ? 'on' : ''}" data-seg="installed">${esc(L.installed)}</button>
      <button class="al-segbtn ${S.seg === 'store' ? 'on' : ''}" data-seg="store">${esc(L.store)}</button>
      <button class="al-segbtn ${S.seg === 'adopt' ? 'on' : ''}" data-seg="adopt">＋ Adopt repo</button>
    </div>
    <div id="alBody"></div>`;
  home.pane.querySelectorAll('[data-seg]').forEach((b) => b.onclick = () => { S.seg = b.dataset.seg; renderHome(); });
  renderHomeBody(home.pane.querySelector('#alBody'));
}

function renderHomeBody(body) {
  if (S.seg === 'store') {
    body.innerHTML = `<div class="al-store">${esc(L.store_intro)}<br><br>
      <button class="al-segbtn on" id="alPick">${esc(L.store_pick)}</button>
      <input type="file" id="alFile" accept=".fwpack,.zip" style="display:none">
      <span class="al-msg" id="alMsg"></span><br><br>
      ${esc(L.store_local)} <code>apps/&lt;id&gt;/</code> (manifest.json + core + ui/).<br>
      ${esc(L.store_remote)}</div>`;
    const file = body.querySelector('#alFile');
    body.querySelector('#alPick').onclick = () => file.click();
    file.onchange = () => { if (file.files[0]) installPack(file.files[0]); };
    return;
  }
  if (S.seg === 'adopt') { renderAdopt(body); return; }
  if (!S.apps.length) { body.innerHTML = `<div class="al-empty">${esc(L.empty)}</div>`; return; }
  body.innerHTML = `<div class="al-grid">${S.apps.map(cardHTML).join('')}</div>`;
  S.apps.forEach((a) => {
    const el = body.querySelector(`[data-app="${a.id}"]`);
    el.onclick = (e) => { if (e.target.classList.contains('x')) { e.stopPropagation(); uninstallApp(a); } else openApp(a); };
  });
}

// app hasil-adopt ga punya file icon → kasih glyph emoji per-runtime (bukan gambar broken).
function appGlyph(a) {
  if ((a.operations || []).some((o) => o.name === '_url')) return '🌐'; // server/web app
  return ({ python: '🐍', node: '🟢', go: '🐹', rust: '🦀', process: '📦', http: '🌐' })[a.runtime] || '📦';
}
function cardHTML(a) {
  const native = a.runtime === 'process' || a.runtime === 'http';
  const iconHTML = a.icon
    ? `<img src="/api/apps/${escAttr(a.id)}/${escAttr(a.icon)}" alt="" onerror="this.replaceWith(Object.assign(document.createElement('div'),{className:'al-glyph',textContent:'${appGlyph(a)}'}))">`
    : `<div class="al-glyph">${appGlyph(a)}</div>`;
  return `<div class="al-card" data-app="${escAttr(a.id)}">
    <span class="x" title="${escAttr(L.uninstall)}">✕</span>
    ${iconHTML}
    <div class="nm">${esc(a.name || a.id)}</div>
    <div class="rt">${native ? '🔓 native' : '🔒 sandbox'} · ${esc(a.runtime || 'wasm')}</div>
  </div>`;
}

async function installPack(f) {
  const home = S.tabs.find((tb) => tb.home);
  const msg = home && home.pane.querySelector('#alMsg');
  if (!confirm(L.store_exec_warn)) return;
  if (msg) msg.textContent = '⟳ ' + L.installing;
  const fd = new FormData(); fd.append('file', f);
  try {
    const resp = await fetch('/api/apps/install?approve_exec=1', { method: 'POST', body: fd });
    const r = await resp.json();
    if (!resp.ok || r.error) throw new Error(r.error || ('HTTP ' + resp.status));
    if (msg) msg.textContent = '✓ ' + L.install_ok + (r.app ? ' — ' + r.app : '');
    S.seg = 'installed';
    await refreshApps();
  } catch (e) { if (msg) msg.textContent = '✕ ' + L.install_fail + ': ' + (e.message || e); }
}

async function uninstallApp(a) {
  if (!confirm(L.uninstall_confirm.replace('{name}', a.name || a.id))) return;
  try {
    closeTab('app:' + a.id); // an open app must stop before its folder is removed
    await fetchJSON('/api/apps/uninstall?id=' + encodeURIComponent(a.id), { method: 'POST' });
    await refreshApps();
  } catch (e) { alert((L.install_fail || 'failed') + ': ' + (e.message || e)); }
}

// ── entry: called by the router whenever the App menu is opened ────────────────
export async function render(mainEl) {
  loadStyle('apps', CSS);
  ensureShell();          // singleton — built once, survives menu switches
  mainEl.innerHTML = '';  // the shell overlays #main; keep #main empty
  await refreshApps();    // refresh the launcher list
  if (!S.restored) { S.restored = true; restoreTabs(); } // reopen tabs saved before a refresh
  show();
}
