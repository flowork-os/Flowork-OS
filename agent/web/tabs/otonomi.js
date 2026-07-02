// Flowork OS — Dev: Aola Sahidin — github.com/flowork-os/Flowork-OS · floworkos.com
// tabs/otonomi.js — "Autonomy" panel: work board (worklog) + experience journal (journal).
// Read-only surface. Strings via i18n (i18n/<locale>/otonomi.json) — NO hardcoded copy.
// Data: GET /api/worklog + /api/journal (same-origin cookie auth). Dok: lock/worklog.md.

import { esc, fetchJSON } from '../js/utils.js';
import { t } from '/js/i18n.js';

// L.fooBar → t('otonomi.foo_bar') (pola sama tab lain, mis. commits.js).
const L = new Proxy({}, { get: (_, k) => t('otonomi.' + String(k).replace(/[A-Z]/g, (c) => '_' + c.toLowerCase())) });

async function safeGet(url) {
  try { return await fetchJSON(url); } catch (e) { return { _err: e && e.message ? e.message : String(e) }; }
}

function workBody(d) {
  if (!d || d._err) return `<div class="err">${esc(L.error)}: ${esc(d ? d._err : '')}</div>`;
  if (d.enabled === false) return `<div class="empty">${esc(L.workOff)}</div>`;
  const items = d.items || [];
  if (!items.length) return `<div class="empty">${esc(L.workEmpty)}</div>`;
  const rows = items.map((it) => `<tr>
      <td>${it.priority === 'high' ? `<span class="oto-b oto-b-hi">${esc(L.badgeHigh)}</span>` : '<span class="muted">·</span>'}</td>
      <td><b>${esc(it.agent)}</b></td>
      <td>${esc(it.label || it.id)}</td>
      <td>${esc(it.state)}${it.stale ? ` <span class="oto-b oto-b-st">${esc(L.badgeStuck)}</span>` : ''}</td>
      <td class="muted nowrap">${esc(it.updated)}</td></tr>`).join('');
  return `<table class="tt-table"><thead><tr>
      <th>${esc(L.colPriority)}</th><th>${esc(L.colAgent)}</th><th>${esc(L.colTask)}</th>
      <th>${esc(L.colState)}</th><th>${esc(L.colUpdated)}</th></tr></thead><tbody>${rows}</tbody></table>`;
}

function tile(emoji, n, label) {
  return `<div class="oto-stat"><div class="oto-stat-n">${n | 0}</div><div class="oto-stat-l">${emoji} ${esc(label)}</div></div>`;
}

function journalBody(d) {
  if (!d || d._err) return `<div class="err">${esc(L.error)}: ${esc(d ? d._err : '')}</div>`;
  if (d.enabled === false) return `<div class="empty">${esc(L.journalOff)}</div>`;
  const tot = d.totals || {};
  const tiles = `<div class="oto-stats">
      ${tile('📕', tot.mistakes, L.statMistakes)}${tile('💡', tot.eureka, L.statEureka)}
      ${tile('🛡️', tot.antibody, L.statAntibody)}${tile('🧩', tot.skills, L.statSkills)}
      ${tile('⚡', tot.instincts, L.statInstincts)}</div>`;
  const items = d.items || [];
  if (!items.length) return tiles + `<div class="empty">${esc(L.journalEmpty)}</div>`;
  const rows = items.map((s) => {
    const lessons = (s.lessons || []).map((l) => `<li>${esc(l.title)} <span class="muted">(${l.hits | 0}×)</span></li>`).join('') || '<li class="muted">—</li>';
    return `<tr><td><b>${esc(s.agent)}</b></td>
      <td class="muted nowrap">${s.mistakes | 0}/${s.eureka | 0}/${s.antibody | 0}/${s.skills | 0}/${s.instincts | 0}</td>
      <td><ul class="oto-lessons">${lessons}</ul></td></tr>`;
  }).join('');
  return tiles + `<table class="tt-table"><thead><tr><th>${esc(L.colAgent)}</th><th>${esc(L.colCounts)}</th><th>${esc(L.colLessons)}</th></tr></thead><tbody>${rows}</tbody></table>`;
}

// approvalBody — antrian aksi PENDING lintas-agent (F-B approval gate) + tombol
// approve/reject. Data: GET /api/agents/protector/approval/pending-all (agregat).
function approvalBody(d) {
  if (!d || d._err) return `<div class="err">${esc(L.error)}: ${esc(d ? d._err : '')}</div>`;
  const items = d.items || [];
  if (!items.length) return `<div class="empty">${esc(L.approvalEmpty)}</div>`;
  const rows = items.map((it) => `<tr>
      <td><b>${esc(it.agent)}</b></td>
      <td><code>${esc(it.tool_name)}</code></td>
      <td class="muted">${esc(it.reason || it.tool_name)}</td>
      <td class="muted nowrap">${esc(it.requested_at || '')}</td>
      <td class="nowrap">
        <button class="oto-appr oto-appr-ok" data-appr="approve" data-agent="${esc(it.agent)}" data-qid="${it.id | 0}">${esc(L.btnApprove)}</button>
        <button class="oto-appr oto-appr-no" data-appr="reject" data-agent="${esc(it.agent)}" data-qid="${it.id | 0}">${esc(L.btnReject)}</button>
      </td></tr>`).join('');
  return `<table class="tt-table"><thead><tr>
      <th>${esc(L.colAgent)}</th><th>${esc(L.colTool)}</th><th>${esc(L.colReason)}</th>
      <th>${esc(L.colRequested)}</th><th></th></tr></thead><tbody>${rows}</tbody></table>`;
}

async function refreshApproval(mainEl) {
  const box = mainEl.querySelector('#otoApproval');
  if (box) box.innerHTML = approvalBody(await safeGet('/api/agents/protector/approval/pending-all'));
}

function wireApproval(mainEl) {
  const box = mainEl.querySelector('#otoApproval');
  if (!box) return;
  box.addEventListener('click', async (e) => {
    const btn = e.target.closest('[data-appr]');
    if (!btn) return;
    btn.disabled = true;
    const url = `/api/agents/protector/${btn.dataset.appr}_pending?id=${encodeURIComponent(btn.dataset.agent)}&queue_id=${encodeURIComponent(btn.dataset.qid)}`;
    try {
      await fetchJSON(url, { method: 'POST' });
      await refreshApproval(mainEl);
    } catch (err) { btn.disabled = false; alert('❌ ' + (err && err.message ? err.message : err)); }
  });
}

export async function render(mainEl) {
  mainEl.innerHTML = `
    <h2>${esc(L.title)}</h2>
    <div class="sub">${esc(L.sub)}</div>
    <div class="card mt-4"><div class="ch">${esc(L.approvalHeader)}</div>
      <div class="cb" id="otoApproval"><div class="empty">${esc(L.loading)}</div></div></div>
    <div class="card mt-4"><div class="ch">${esc(L.workHeader)}</div>
      <div class="cb" id="otoWork"><div class="empty">${esc(L.loading)}</div></div></div>
    <div class="card mt-4"><div class="ch">${esc(L.journalHeader)}</div>
      <div class="cb" id="otoJourn"><div class="empty">${esc(L.loading)}</div></div></div>
    <div class="sub mt-4">${esc(L.note)}</div>
    <style>
      .oto-stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(108px,1fr));gap:12px;margin-bottom:14px}
      .oto-stat{position:relative;background:radial-gradient(circle at 20% 0%,rgba(139,92,246,.10),transparent 60%),linear-gradient(165deg,rgba(255,255,255,.05),rgba(255,255,255,0) 55%),rgba(15,17,26,.55);border:1px solid var(--glass-border);border-radius:14px;padding:16px 12px;text-align:center;box-shadow:0 6px 18px rgba(0,0,0,.28),inset 0 1px 0 rgba(255,255,255,.06);transition:transform .2s ease,border-color .2s ease}
      .oto-stat:hover{transform:translateY(-3px);border-color:var(--glass-border-hover,var(--accent))}
      .oto-stat-n{font-size:1.8rem;font-weight:800;line-height:1;color:var(--text-main);text-shadow:0 2px 8px rgba(139,92,246,.25)}
      .oto-stat-l{font-size:.78rem;color:var(--text-muted);margin-top:6px}
      .oto-b{padding:1px 8px;border-radius:999px;font-size:.7rem;font-weight:700;letter-spacing:.3px}
      .oto-b-hi{background:rgba(239,68,68,.18);color:#f87171;border:1px solid rgba(239,68,68,.45)}
      .oto-b-st{background:rgba(245,158,11,.16);color:#fbbf24;border:1px solid rgba(245,158,11,.45)}
      .oto-lessons{margin:0;padding-left:1.1em} .oto-lessons li{margin:1px 0}
      .nowrap{white-space:nowrap}
      .oto-appr{padding:3px 12px;border-radius:8px;font-size:.74rem;font-weight:700;cursor:pointer;margin-left:6px;border:1px solid transparent}
      .oto-appr-ok{background:rgba(16,185,129,.16);color:#6ee7b7;border-color:rgba(16,185,129,.4)}
      .oto-appr-ok:hover{background:rgba(16,185,129,.28)}
      .oto-appr-no{background:rgba(239,68,68,.14);color:#f87171;border-color:rgba(239,68,68,.4)}
      .oto-appr-no:hover{background:rgba(239,68,68,.26)}
      .oto-appr:disabled{opacity:.5;cursor:wait}
    </style>
  `;
  const [ap, wl, jr] = await Promise.all([
    safeGet('/api/agents/protector/approval/pending-all'),
    safeGet('/api/worklog'), safeGet('/api/journal'),
  ]);
  const a = mainEl.querySelector('#otoApproval'); if (a) a.innerHTML = approvalBody(ap);
  wireApproval(mainEl);
  const w = mainEl.querySelector('#otoWork'); if (w) w.innerHTML = workBody(wl);
  const j = mainEl.querySelector('#otoJourn'); if (j) j.innerHTML = journalBody(jr);
}
