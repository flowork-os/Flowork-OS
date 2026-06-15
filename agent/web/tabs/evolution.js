// evolution.js — R7 SELF-EVOLUTION control panel. Owner-approved 2026-06-15 (FASE 2).
// SAKLAR self-modify (OFF/STAGE/AUTO) + status gate berlapis + backlog usulan. KRUSIAL:
// owner pegang penuh. Auto-commit cuma jalan kalau mode=AUTO + karma matang + model cloud
// kuat (guard anti-LLM-lokal). Inline strings (no i18n domain) biar self-contained.

const esc = (s) => String(s == null ? '' : s).replace(/[&<>"']/g, (c) =>
  ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));

export async function render(container) {
  container.innerHTML = `
    <div style="padding:18px 22px;max-width:920px;color:#e2e8f0">
      <h2 style="margin:0 0 4px">🧬 Self-Evolution</h2>
      <p style="color:#94a3b8;margin:0 0 16px;font-size:0.88rem">
        Flowork bisa berevolusi sendiri — <b>tapi kamu pegang saklarnya</b>. Auto-commit hanya
        jalan kalau SEMUA gate aman (model cloud kuat, karma matang). Token Claude habis → mode
        lokal → auto-commit otomatis diblok.
      </p>
      <div id="ev-status" style="background:#0f172a;border:1px solid #1e293b;border-radius:10px;padding:14px;margin-bottom:14px">⏳ loading…</div>
      <div id="ev-modes" style="display:flex;gap:10px;margin-bottom:8px"></div>
      <div id="ev-modehint" style="color:#64748b;font-size:0.78rem;margin-bottom:20px"></div>
      <div style="display:flex;align-items:center;justify-content:space-between">
        <h3 style="margin:0">📋 Backlog usulan evolusi</h3>
        <button id="ev-reflect" style="background:#6366f1;color:#fff;border:0;border-radius:8px;padding:8px 14px;cursor:pointer">🔍 Refleksi sekarang</button>
      </div>
      <div id="ev-proposals" style="margin-top:12px">⏳…</div>
    </div>`;

  const statusEl = container.querySelector('#ev-status');
  const modesEl = container.querySelector('#ev-modes');
  const hintEl = container.querySelector('#ev-modehint');
  const propEl = container.querySelector('#ev-proposals');
  const reflectBtn = container.querySelector('#ev-reflect');

  const MODES = [
    { k: 'off', label: '🔴 OFF', hint: 'Self-modify mati total. Refleksi/usulan tetap jalan (read-only, aman).' },
    { k: 'stage', label: '🟡 STAGE', hint: 'Usul → sandbox → test, lalu di-STAGE buat kamu review & Apply manual.' },
    { k: 'auto', label: '🟢 AUTO', hint: 'Auto-commit perubahan yang lolos SEMUA gate. Hanya aktif kalau kamu arm + gate aman.' },
  ];

  async function loadConfig() {
    try {
      const d = await (await fetch('/api/evolve/config')).json();
      if (d.error) throw new Error(d.error);
      const k = d.karma || {}, m = d.model || {};
      const yn = (b) => (b ? '<span style="color:#4ade80">✓ ya</span>' : '<span style="color:#f87171">✗ tidak</span>');
      const allow = d.autocommit_allowed;
      statusEl.innerHTML = `
        <div style="display:flex;gap:24px;flex-wrap:wrap;font-size:0.85rem">
          <div>Mode aktif: <b style="font-size:1.05rem">${esc((d.mode || 'off').toUpperCase())}</b></div>
          <div>Karma matang: ${yn(k.ready)} <span style="color:#64748b">(${Math.round(k.reflect_ok || 0)}/${k.threshold || 20} sukses)</span></div>
          <div>Model cloud kuat: ${yn(m.strong)} <span style="color:#64748b">${esc(m.note || '')}</span></div>
        </div>
        <div style="margin-top:10px;padding:8px 12px;border-radius:8px;background:${allow ? '#052e16' : '#1e293b'};border:1px solid ${allow ? '#16a34a' : '#334155'}">
          Auto-commit: <b style="color:${allow ? '#4ade80' : '#fbbf24'}">${allow ? '🟢 AKTIF (semua gate lolos)' : '🔒 TERKUNCI'}</b>
          ${allow ? '' : '<span style="color:#94a3b8;font-size:0.8rem"> — perlu mode=AUTO + karma matang + model cloud kuat</span>'}
        </div>`;
      // mode buttons
      modesEl.innerHTML = '';
      MODES.forEach((mo) => {
        const active = (d.mode || 'off') === mo.k;
        const b = document.createElement('button');
        b.textContent = mo.label;
        b.style.cssText = `flex:1;padding:12px;border-radius:10px;cursor:pointer;font-size:0.95rem;border:2px solid ${active ? '#6366f1' : '#334155'};background:${active ? '#1e1b4b' : '#0f172a'};color:#e2e8f0`;
        b.addEventListener('click', () => setMode(mo.k));
        modesEl.appendChild(b);
      });
      hintEl.textContent = (MODES.find((x) => x.k === (d.mode || 'off')) || {}).hint || '';
    } catch (e) {
      statusEl.innerHTML = `<span style="color:#f87171">❌ ${esc(e.message)}</span>`;
    }
  }

  async function setMode(mode) {
    if (mode === 'auto' && !confirm('Aktifkan AUTO? Flowork bakal auto-commit perubahan yang lolos semua gate (model cloud kuat + karma matang). Lanjut?')) return;
    try {
      const r = await fetch('/api/evolve/config', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ mode }) });
      const d = await r.json();
      if (d.error) throw new Error(d.error);
      await loadConfig();
    } catch (e) { alert('Gagal set mode: ' + e.message); }
  }

  async function loadProposals() {
    try {
      const d = await (await fetch('/api/evolve/proposals?limit=30')).json();
      const items = d.items || [];
      if (!items.length) { propEl.innerHTML = '<div style="color:#64748b">Belum ada usulan. Klik "Refleksi sekarang".</div>'; return; }
      const riskColor = { low: '#4ade80', medium: '#fbbf24', high: '#f87171' };
      propEl.innerHTML = items.map((p) => `
        <div style="background:#0f172a;border:1px solid #1e293b;border-radius:8px;padding:10px 12px;margin-bottom:8px">
          <div style="display:flex;gap:8px;align-items:center;margin-bottom:4px">
            <span style="background:#1e293b;border-radius:4px;padding:1px 7px;font-size:0.72rem">${esc(p.kind || '?')}</span>
            <span style="color:${riskColor[p.risk] || '#94a3b8'};font-size:0.72rem">●${esc(p.risk || '?')}</span>
            <code style="color:#818cf8;font-size:0.76rem">${esc(p.target_file || '')}</code>
            <span style="margin-left:auto;color:#475569;font-size:0.7rem">${esc(p.status || '')}</span>
          </div>
          <div style="font-size:0.84rem;color:#cbd5e1">${esc(p.rationale || '')}</div>
        </div>`).join('');
    } catch (e) { propEl.innerHTML = `<span style="color:#f87171">❌ ${esc(e.message)}</span>`; }
  }

  reflectBtn.addEventListener('click', async () => {
    reflectBtn.disabled = true; const o = reflectBtn.textContent; reflectBtn.textContent = '⏳ refleksi…';
    try {
      const d = await (await fetch('/api/evolve/reflect', { method: 'POST' })).json();
      if (d.error) throw new Error(d.error);
      await loadProposals(); await loadConfig();
    } catch (e) { alert('Refleksi gagal: ' + e.message); }
    finally { reflectBtn.disabled = false; reflectBtn.textContent = o; }
  });

  await loadConfig();
  await loadProposals();
}
