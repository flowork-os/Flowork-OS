// tools.js — tab "Tool Plugins": plug-and-play TOOL-PACK gallery (clean glass-3D, full-width).
//
// Flowork = chatbot polos; TANGAN-nya diisi tool-pack. Tiap tool-pack = .fwpack (zip:
// plugin.json kind:tool + agents/<id>/{agent.wasm,manifest.json}) di-hot-load + RegisterDynamic.
// Tab ini: upload (install) · daftar ter-install · uninstall. Tool MCP (prefix "mcp_") dikelola
// di Connections (disaring di sini). Copy lewat i18n (en base + id) — nol hardcode.
// API: GET /api/tools/installed · POST /api/tools/install · POST /api/tools/uninstall?tool=
import { esc, escAttr, fetchJSON, loadStyle } from '../js/utils.js';
import { t } from '/js/i18n.js';

const L = new Proxy({}, { get: (_, k) => t('tools.' + String(k)) });
const fmt = (k, vars) => Object.entries(vars || {}).reduce((s, [n, v]) => s.replaceAll('{' + n + '}', v), L[k]);

const CSS = `
.tp-wrap { padding:18px 26px 40px; }
.tp-head { display:flex; align-items:center; gap:14px; margin-bottom:20px; }
.tp-glyph { font-size:1.8rem; filter:drop-shadow(0 0 12px var(--accent-glow)); }
.tp-title { margin:0; font-size:1.5rem; font-weight:800; line-height:1.05;
  background:linear-gradient(90deg,#c4b5fd,#67e8f9 58%,#6ee7b7); -webkit-background-clip:text; background-clip:text; color:transparent; }
.tp-subt { font-size:.82rem; color:var(--text-muted); margin-top:3px; max-width:90ch; line-height:1.45; }
.tp-stat { font-size:.74rem; color:var(--text-muted); margin-top:6px; display:flex; align-items:center; gap:7px; }
.tp-dot { width:7px; height:7px; border-radius:50%; background:#34d399; box-shadow:0 0 8px #34d399; animation:tpblink 1.8s ease-in-out infinite; }
@keyframes tpblink { 0%,100%{opacity:1} 50%{opacity:.3} }

.tp-card { position:relative; border-radius:16px; padding:18px 20px; margin-bottom:14px; transition:transform .2s, border-color .2s, box-shadow .2s;
  border:1px solid var(--glass-border);
  background:
    radial-gradient(circle at 18% 0%, color-mix(in srgb,var(--accent) 9%, transparent), transparent 55%),
    linear-gradient(165deg, rgba(255,255,255,.045), rgba(255,255,255,0) 52%), var(--bg-panel);
  box-shadow:0 10px 30px rgba(0,0,0,.34), inset 0 1px 0 rgba(255,255,255,.06); }
.tp-card:hover { border-color:var(--glass-border-hover); transform:translateY(-2px); box-shadow:0 18px 42px rgba(0,0,0,.44), inset 0 1px 0 rgba(255,255,255,.09); }
.tp-sec { font-size:.78rem; font-weight:700; letter-spacing:.03em; color:var(--accent); margin:0 0 12px; }
.tp-drop { border:1.5px dashed var(--glass-border-hover); border-radius:13px; padding:28px; text-align:center; cursor:pointer;
  color:var(--text-muted); font-size:.9rem; background:color-mix(in srgb,var(--bg-panel) 60%, transparent); transition:.2s; }
.tp-drop:hover, .tp-drop.over { background:color-mix(in srgb,var(--accent) 9%, transparent); border-color:var(--accent); color:var(--text-main); }
.tp-hint { font-size:.76rem; color:var(--text-muted); margin-top:10px; }
.tp-msg { font-size:.8rem; margin-top:10px; min-height:16px; }
.tp-row { display:flex; align-items:center; gap:14px; flex-wrap:wrap; }
.tp-row h3 { margin:0; font-size:1rem; color:var(--text-main); font-weight:700; display:flex; align-items:center; gap:9px; }
.tp-tag { font-size:.7rem; font-weight:700; letter-spacing:.4px; color:var(--accent); border-radius:999px; padding:2px 10px;
  border:1px solid color-mix(in srgb,var(--accent) 40%, transparent); background:color-mix(in srgb,var(--accent) 13%, transparent); }
.tp-id { font-size:.74rem; color:var(--text-muted); font-family:ui-monospace,monospace; }
.tp-grow { flex:1 1 auto; }
.tp-btn { font:inherit; font-size:.76rem; font-weight:700; padding:7px 15px; border-radius:10px; cursor:pointer; transition:filter .15s, transform .1s;
  color:var(--text-main); border:1px solid var(--glass-border-hover); background:color-mix(in srgb,var(--accent) 13%, transparent);
  box-shadow:0 4px 12px rgba(0,0,0,.2); }
.tp-btn:hover { filter:brightness(1.16); transform:translateY(-1px); }
.tp-btn.danger { color:#f87171; border-color:rgba(248,113,113,.42); background:rgba(248,113,113,.11); }
.tp-desc { margin-top:11px; font-size:.82rem; color:var(--text-muted); line-height:1.5; }
.tp-empty { font-size:.85rem; color:var(--text-muted); padding:22px; text-align:center; font-style:italic; }
.tp-card code { font-family:ui-monospace,monospace; background:rgba(2,6,18,.5); padding:1px 6px; border-radius:5px; font-size:.86em; }
`;

export async function render(mainEl) {
  loadStyle('tp-style', CSS);
  mainEl.innerHTML = `
    <div class="tp-wrap">
      <div class="tp-head">
        <span class="tp-glyph">🔧</span>
        <div>
          <h2 class="tp-title">${esc(L.title)}</h2>
          <div class="tp-subt">${esc(L.sub)}</div>
          <div class="tp-stat"><span class="tp-dot"></span><span id="tp-count">0 ${esc(L.count_label)}</span> · ${esc(L.status_online)}</div>
        </div>
      </div>

      <div class="tp-card">
        <div class="tp-sec">${esc(L.install_h)}</div>
        <div class="tp-drop" id="tp-drop">${esc(L.install_drop)}</div>
        <input type="file" id="tp-file" accept=".fwpack,.zip" style="display:none">
        <div class="tp-hint">${esc(L.install_hint)}</div>
        <div class="tp-msg" id="tp-install-msg"></div>
      </div>

      <div id="tp-sidecar"></div>
      <div id="tp-list"></div>
    </div>`;

  const drop = mainEl.querySelector('#tp-drop');
  const file = mainEl.querySelector('#tp-file');
  drop.onclick = () => file.click();
  drop.ondragover = (e) => { e.preventDefault(); drop.classList.add('over'); };
  drop.ondragleave = () => drop.classList.remove('over');
  drop.ondrop = (e) => { e.preventDefault(); drop.classList.remove('over'); if (e.dataTransfer.files[0]) install(mainEl, e.dataTransfer.files[0]); };
  file.onchange = () => { if (file.files[0]) install(mainEl, file.files[0]); };

  await load(mainEl);
  await loadSidecar(mainEl);
}

// loadSidecar — SIDECAR TOOLS (native, folder self-contained, akses semua agent).
async function loadSidecar(mainEl) {
  const el = mainEl.querySelector('#tp-sidecar');
  if (!el) return;
  let data;
  try { data = await fetchJSON('/api/tools/sidecar', { method: 'POST' }); } catch (e) { el.innerHTML = ''; return; }
  const tools = (data && data.tools) || [];
  if (!tools.length) { el.innerHTML = ''; return; }
  el.innerHTML = `<div class="tp-card"><div class="tp-sec">⚙️ Sidecar Tools · ${tools.length}</div>`
    + `<div class="tp-hint">Native, self-contained (folder <code>tools/&lt;name&gt;/</code>, binary terpisah) — bisa diakses SEMUA agent.</div></div>`
    + tools.sort((a, b) => String(a.name).localeCompare(b.name)).map(sidecarCardHTML).join('');
}

function sidecarCardHTML(t) {
  return `<div class="tp-card">
    <div class="tp-row">
      <h3>⚙️ ${esc(t.name)}</h3>
      <span class="tp-tag">${t.capability ? esc(t.capability) : 'semua agent'}</span>
      <span class="tp-id">${fmt('params_label', { n: t.params || 0 })}</span>
      <span class="tp-grow"></span>
      <span class="tp-id" style="opacity:.55">sidecar</span>
    </div>
    ${t.description ? `<div class="tp-desc">${esc(t.description)}</div>` : ''}
  </div>`;
}

async function load(mainEl) {
  const list = mainEl.querySelector('#tp-list');
  let data;
  try { data = await fetchJSON('/api/tools/installed'); }
  catch (e) { list.innerHTML = `<div class="tp-card"><div class="tp-empty">${esc(String(e))}</div></div>`; return; }
  const items = ((data && data.installed) || []).filter((tl) => !String(tl.name || '').startsWith('mcp_'));
  mainEl.querySelector('#tp-count').textContent = `${items.length} ${L.count_label}`;
  if (!items.length) { list.innerHTML = `<div class="tp-card"><div class="tp-empty">${esc(L.empty)}</div></div>`; return; }
  list.innerHTML = items.map(cardHTML).join('');
  items.forEach((tl) => {
    const card = mainEl.querySelector(`[data-tool="${escAttr(tl.name)}"]`);
    if (!card) return;
    const b = card.querySelector('[data-act="uninstall"]');
    if (b) b.onclick = () => uninstall(mainEl, tl.name);
  });
}

function cardHTML(tl) {
  return `<div class="tp-card" data-tool="${escAttr(tl.name)}">
    <div class="tp-row">
      <h3>🔧 ${esc(tl.name)}</h3>
      ${tl.capability ? `<span class="tp-tag">${esc(tl.capability)}</span>` : ''}
      <span class="tp-id">${fmt('params_label', { n: tl.params || 0 })}</span>
      <span class="tp-grow"></span>
      <button class="tp-btn danger" data-act="uninstall">${esc(L.btn_uninstall)}</button>
    </div>
    ${tl.description ? `<div class="tp-desc">${esc(tl.description)}</div>` : ''}
  </div>`;
}

async function install(mainEl, f) {
  const msg = mainEl.querySelector('#tp-install-msg');
  msg.style.color = 'var(--text-muted)';
  msg.textContent = L.installing;
  const fd = new FormData();
  fd.append('file', f);
  try {
    const r = await fetch('/api/tools/install', { method: 'POST', body: fd });
    const j = await r.json();
    if (!r.ok || j.error) throw new Error(j.error || ('HTTP ' + r.status));
    msg.style.color = '#34d399';
    msg.textContent = fmt('install_ok', { tool: j.tool || '?' });
    await load(mainEl);
  } catch (e) {
    msg.style.color = '#f87171';
    msg.textContent = L.install_fail + e;
  }
}

async function uninstall(mainEl, name) {
  if (!confirm(fmt('uninstall_confirm', { tool: name }))) return;
  try { await fetchJSON('/api/tools/uninstall?tool=' + encodeURIComponent(name), { method: 'POST' }); await load(mainEl); }
  catch (e) { alert(L.uninstall_fail + e); }
}
