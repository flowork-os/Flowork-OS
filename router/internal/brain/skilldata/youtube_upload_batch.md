---
name: youtube_upload_batch
description: Upload video batch ke multiple YouTube channel via YouTube Data API v3 (sovereign, no third-party browser). OAuth per channel + quota-aware pacing.
metadata:
  domain: youtube
  created_at: 2026-05-18
  updated_at: 2026-06-13
  trigger_count: 0
  is_catalyst: false
---

# YouTube Upload Batch via YouTube Data API v3 (sovereign)

> Sovereign path only. Flowork does NOT drive any third-party / anti-detect browser. Uploads go
> through the official **YouTube Data API v3** (`videos.insert`) with one OAuth credential per
> channel. No external browser binary, no UI-scripting, no fingerprint games.

## When to Use
Task: upload N video ke N channel YouTube (1 video per channel atau bulk).

## Trigger Pattern
- User request mengandung: "upload N video", "publish ke channel", "batch upload"
- N channel >= 2

## Pre-requisites (per channel)
- OAuth 2.0 client + refresh token with scope `https://www.googleapis.com/auth/youtube.upload`.
- Stored in the settings DB via the sensitive interceptor (Pilar 6) — NEVER hardcoded.
- Quota budget: `videos.insert` costs ~1600 units; default daily quota 10,000 → ~6 uploads/day
  per project. Spread channels across separate API projects/credentials to scale.

## Steps

1. **Pre-check** — confirm each target channel has a valid, non-expired OAuth refresh token in the
   settings DB. Refresh access tokens up front; abort early on any auth failure (don't half-run).
2. **Load target channels** dari config (atau user spec).
3. **Loop per channel:**
   ```
   FOR channel in target_channels:
       - access_token = oauth_refresh(channel.refresh_token)
       - resp = POST https://www.googleapis.com/upload/youtube/v3/videos
                ?part=snippet,status&uploadType=resumable
                Authorization: Bearer <access_token>
                body(snippet={title, description, tags, categoryId},
                     status={privacyStatus: "private"|"unlisted"|"public"})
                + resumable media upload of the video file
       - REQUIRE resp.id (video_id) — itu satu-satunya bukti sukses (A-1, no halu)
       - log audit ke task_events (channel, video_id, privacy)
       - if quotaExceeded: stop this credential, continue next project OR defer to tomorrow
       - delay 5-15s antar upload (politeness, hindari burst 429)
   END FOR
   ```

## Quota & Reliability Checklist
- Resumable upload (`uploadType=resumable`) untuk file besar — bisa retry chunk tanpa ulang.
- Exponential backoff pada 5xx / 429 / `userRateLimitExceeded`.
- `quotaExceeded` (403) bukan error fatal → defer, jangan retry-loop bakar quota.
- Multi-project credential rotation untuk scale (tiap project 10K quota terpisah).

## Output Format
- Total uploads: N/N success
- Per-channel: channel_id, video_id, video_url (`https://youtu.be/<video_id>`), privacy
- Failure: log ke mistakes_journal tier raw (dengan reason: auth/quota/network)

## References
- `mixing.md` workflow audio
- `yt.md` Bagian 9 workflow production
- YouTube Data API v3 — `videos.insert` (official docs)

## Side Effects
- Audit log: task_events per upload
- Mistakes journal: kalau ada error
- No browser state, no cookies, no screenshots (API-only path)
