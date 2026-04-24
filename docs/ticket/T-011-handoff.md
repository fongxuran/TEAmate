# T-011 handoff (local UI: upload + realtime + results)

This is a handoff note for continuing **T-011** work on a machine that can install Node deps via Nexus.

## What’s already implemented

### Backend (Go)

A minimal, deterministic “local MVP” backend was added to support:

- **Shared realtime session** over WebSocket at **`/ws`** (NOT under `/api` so browser WS does not need Basic Auth headers)
- REST endpoints under `/api` (will be protected by Basic Auth if enabled)
- Deterministic analysis pipeline (no LLM dependency) producing:
  - segmentation
  - drift scoring (Jaccard similarity of keywords)
  - drift alerts when new drift segments appear
  - decisions/action items extraction (simple prefixed-line heuristic)
  - ticket drafts derived from action items

New/changed backend files:

- `api/internal/analysis/analysis.go`
  - agenda parsing from textbox
  - transcript parsing from plaintext
  - deterministic `Analyze()` -> `analysis.Result`
- `api/internal/realtime/hub.go`
  - WS server and event protocol
- `api/internal/realtime/session.go`
  - per-session state (agenda, transcript, config, feedback overrides)
- `api/internal/handler/rest/mvp/handler.go`
  - `/api/meeting/state` (GET)
  - `/api/meeting/analyze` (POST)
  - `/api/meeting/reset` (POST)
  - `/api/exports/ticket-drafts` (POST)
- `api/cmd/serverd/router.go`
  - mounts `GET /ws` (hub)
  - mounts MVP REST handler under `/api`
- `api/cmd/serverd/main.go`
  - instantiates hub + MVP handler
- `api/internal/export/ticket_drafts.go`
  - added in-memory render helpers: `RenderTicketDraftsMarkdown`, `RenderTicketDraftsCSV`
- `api/internal/connector/notion/client.go`
  - fixed default config handling in `NewClient` so unit tests pass consistently

Verification:

- `make api-test` passes.

### Frontend (Next.js)

A single-page “Local MVP” UI component was added with:

- inputs:
  - agenda textbox (shared)
  - shared transcript textbox (realtime WS)
  - upload transcript (`.txt` or `.json` meeting input)
  - load sample `docs/transcript/meeting 1.txt`
  - config toggles: drift threshold, segment limits
- drift view:
  - segment list + best agenda match + score
  - drift alert banner (from `drift_alert` WS events)
  - buttons to apply feedback (broadcasts to all clients)
  - filter “drift only”
- outcomes:
  - summary, decisions, action items (from deterministic extraction)
- exports:
  - downloads `ticket_drafts.json`, `ticket_drafts.md`, `ticket_drafts.csv`

New/changed frontend files:

- `web/components/MeetingMvp.tsx` (main UI)
- `web/app/page.tsx` (renders `MeetingMvp` above `MessageDemo`)
- `web/app/api/_upstream.ts` (shared proxy helper for server-side route handlers)
- `web/app/api/meeting/state/route.ts`
- `web/app/api/meeting/analyze/route.ts`
- `web/app/api/meeting/reset/route.ts`
- `web/app/api/exports/ticket-drafts/route.ts`
- `web/app/api/samples/meeting-1/route.ts` (reads `../docs/transcript/meeting 1.txt`)
- `web/.npmrc` updated to match `../dreadnought/.npmrc`:
  - `registry=https://nexus.in.spdigital.sg/repository/npm-all`

Notes:

- On this machine, web deps were not installed (npm access blocked), so TypeScript language service may show missing-module errors until `make web-install` succeeds.

## Realtime event protocol (current)

Client -> server (JSON text frames):

- `set_agenda` `{ agenda_text }`
- `set_config` `{ drift_threshold, segment_max_tokens, segment_max_chars }`
- `realtime_message` `{ message: { text_delta? | text?, author? } }`
- `drift_feedback` `{ segment_id, is_drift }`
- `reset` `{}`

Server -> client:

- `sync` `{ session_id, agenda_text, transcript_text, agenda, config, analysis }`
- `agenda_updated` (same payload as `sync`)
- `transcript_updated` `{ client_id, timestamp, text_delta? | text?, author? }`
- `analysis_updated` full `analysis.Result`
- `drift_alert` `{ segment_id, best_agenda_title?, best_score, text_preview }`
- `drift_feedback_applied` `{ segment_id, is_drift }`
- `reset_applied` (same payload as `sync`)

## How to run on a machine with npm access

### 1) Backend

- Start DB + API: `make up`
- Or run API locally (requires env loaded):
  - ensure `DATABASE_URL` exists in your shell
  - `make api-run`

### 2) Frontend

- Install deps: `make web-install`
  - `web/.npmrc` is already set to the Nexus registry used by `dreadnought`.
- Run dev server: `make web-dev`

Then open: http://localhost:3000

### 3) Demo flow (acceptance criteria)

- Open the UI in **two browser windows**.
- Click **Connect** in both.
- Set agenda lines (e.g. "Status", "Decisions", "Action items").
- Load sample transcript: **Load sample (meeting 1)**.
- Drift alerts should show when a new segment is scored as drift.
- Click **Drift** / **Not drift** and verify both windows update.
- Click **Download ticket drafts**.

## Known gaps / follow-ups

1) **Web install/lint not verified** on this machine (no npm access). After install, run:
   - `make web-test` (currently lint)
   - `make web-build`

2) **LLM provider toggle** is not implemented (only drift threshold + segmentation limits). If T-005/T-006 expects provider selection, extend config:
   - add a `provider` field to WS config payload
   - wire to real LLM scoring (or keep deterministic as fallback)

3) **Persistent storage**: realtime session state is in-memory only. That’s fine for local MVP, but if you need persistence, add a small repo layer (or write to disk) per session.

4) **Messages demo** remains separate; you may optionally refactor `web/app/api/messages/route.ts` to reuse `web/app/api/_upstream.ts`.

## Environment/config notes

- Root `.env` is gitignored; it was updated to remove a hardcoded Notion API key value and to add backend defaults (`DATABASE_URL`, auth, CORS).
- The Next.js page uses `NEXT_PUBLIC_WS_URL` (default `ws://localhost:8080/ws`).
