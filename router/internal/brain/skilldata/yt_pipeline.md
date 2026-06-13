---
name: yt-pipeline
description: 5-stage YouTube automation pipeline. Trigger reflex untuk task YT (channel naming, content gen, upload, SEO).
version: 1.0.0
tags: [youtube, automation, pipeline, content]
triggers: [youtube, yt channel, upload video, sleep music, lo-fi, suno, channel naming, video seo, thumbnail, scheduling]
---

# YT Pipeline Skill

## When to Use

Task involves YouTube channel automation:
- Channel naming pick / niche selection
- Content generation (Suno music, voice over)
- Upload to YT (single channel or batch sock puppet)
- SEO optimization (title, description, tags)
- Scheduling / analytics fetch

## Procedure (Reflex Chain)

**WAJIB: jangan halu workflow. Pakai brain doktrin sebagai source of truth.**

Per-stage dispatch (1 stage = 1 brain_search retrieve, BUKAN load all 6 sekaligus):

```
1. brain_search('DOKTRIN_YT_PIPELINE_MASTER')
   -> dapet 5-stage overview + aturan disiplin

2. Stage 1 — Strategy:
   brain_search('DOKTRIN_YT_CONTENT_STRATEGY')
   -> spawn researcher subagent: trend signal + niche pick

3. Stage 2 — Script:
   brain_search('DOKTRIN_YT_SCRIPT_WRITER')
   -> spawn coder subagent: Suno prompt OR voice over text

4. Stage 3 — SEO:
   brain_search('DOKTRIN_YT_SEO_OPTIMIZER')
   -> spawn researcher subagent: title/desc/tags

5. Stage 4 — Publishing:
   brain_search('DOKTRIN_YT_PUBLISHING_AGENT')
   -> upload via YouTube Data API v3 (videos.insert) — sovereign, no third-party browser
   -> spawn coder/verifier subagent: actual upload

6. Log upload ke changelog (A-3 mandatory)
```

## Supporting Knowledge

- `PLAYBOOK_YT_NAMING_INDEX` + 9 niche playbook → pool nama channel/video
- `PLAYBOOK_SUNO_PROMPT_*` → Suno prompt template per niche
- `youtube_upload_batch` → sovereign upload via YouTube Data API v3 (videos.insert)
- `DOKTRIN_YOUTUBE_CHANNEL_EVAL` → CTR/AVD metrik threshold

## Pitfalls

- ❌ **Halu trend data** (A-1) — JANGAN ngarang `trend_score` atau `competition_score` tanpa API real / scraper. Kalau quota habis → return ERR_TREND_DATA_UNAVAILABLE.
- ❌ **Claim upload sukses tanpa video_id** (A-1) — wajib confirmation dari API response (`videos.insert` → resp.id).
- ❌ **Chain over-retrieve** — JANGAN load semua 6 doktrin di 1 turn (~1681 token). Incremental per stage.
- ❌ **Skip log changelog** (A-3) — tiap upload sukses wajib entry baru.
- ❌ **Hardcode channel_id / OAuth token** (Pilar 6) — pakai settings DB sensitive interceptor.
- ❌ **Clickbait halu** — title "TRY NOT TO SLEEP!!!" = anti-pattern. Pakai pool proven dari `PLAYBOOK_YT_NAMING_*`.

## Verification

Sebelum claim stage IMPLEMENTED (default STAGED per A-2 + Pilar 3):
- [ ] Output stage actual (BUKAN stub) — JSON/file artifact
- [ ] L1 artifact verify (file ada, format valid)
- [ ] L2 dispatch test (subagent return non-empty)
- [ ] L3 manual judge Mr.Dev confirm output quality

## Routing Decision (Stage 4)

Sovereign path only — **YouTube Data API v3 `videos.insert`** for every channel. No third-party
or anti-detect browser is used.

| Scenario | Path | Reason |
|---|---|---|
| Owned/main channel | YT Data API v3 | Free 10K quota, stable, official |
| Many channels (scale) | YT Data API v3, multi-project credentials | Each API project = its own 10K quota |
| Quota exhausted | Defer / next project credential | `quotaExceeded` is not fatal — retry tomorrow |

Pre-flight: valid OAuth refresh token present (scope `youtube.upload`) ATAU API quota <8000 daily.
See `youtube_upload_batch` for the full per-channel procedure.
