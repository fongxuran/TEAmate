# T-010 handoff — Notion connector (ticket drafts → Notion pages)

This handoff captures what was implemented for **T-010** (repurposed from Motion/Jira to **Notion**) and what to do next on a machine with working npm access.

## What’s implemented (backend ✅)

### Notion connector
- Added a generic connector interface in `api/internal/connector/task.go`:
  - `TaskCreator.CreateTask(ctx, draft) -> {id,url,dry_run}`
- Implemented Notion client in `api/internal/connector/notion/client.go`.
  - Calls Notion **Create a page**: `POST /v1/pages`
  - Required headers:
    - `Authorization: Bearer <NOTION_API_KEY>`
    - `Notion-Version: <NOTION_VERSION>` (default: `2026-03-11`)
  - Uses a configured **database parent** (`NOTION_DATABASE_ID`).
  - Writes the ticket draft `title` to a configurable title property (`NOTION_TITLE_PROPERTY`, default: `Name`).
  - Adds `description` + traceability as simple paragraph blocks.

### Safety / dry-run
- Dry-run is the default behavior:
  - env: `NOTION_DRY_RUN=true` by default
  - In dry-run the connector **does not perform any network calls** and returns `{dry_run: true}`.

### REST endpoints
- Added Notion integration REST handler:
  - `GET  /api/integrations/notion/status` → `{configured, dry_run, database_id?}`
  - `POST /api/integrations/notion/pages` with a `TicketDraft` JSON body → `{id,url,dry_run}`
  - Implementation: `api/internal/handler/rest/integrations/notion/handler.go`

### Server wiring
- Router now supports multiple REST “registrars” via `api/internal/handler/rest/registrar.go`.
- `api/cmd/serverd/main.go` now wires:
  - existing messages handler
  - Notion integration handler

### Tests
- Go unit tests were added and **pass**:
  - Notion connector tests: `api/internal/connector/notion/client_test.go`
  - Notion REST handler tests: `api/internal/handler/rest/integrations/notion/handler_test.go`
  - Verified with: `make api-test`

## What’s implemented (frontend code ✅, verification pending ⏳)

### Next.js proxy routes
- `GET  /api/notion/status` → proxies to `GET  http://<API_BASE_URL>/api/integrations/notion/status`
- `POST /api/notion/pages`  → proxies to `POST http://<API_BASE_URL>/api/integrations/notion/pages`

Files:
- `web/app/api/notion/status/route.ts`
- `web/app/api/notion/pages/route.ts`

### Demo UI panel
- Added a small demo panel to create a Notion page from a draft payload:
  - Component: `web/components/NotionDraftDemo.tsx`
  - Mounted on home page: `web/app/page.tsx`

## Config you’ll need on your machine

### Notion setup
1. Create a Notion integration (internal).
2. Give it “Insert content” capability.
3. Share your target database with the integration.
4. Set environment variables:

- `NOTION_API_KEY=<secret>`
- `NOTION_DATABASE_ID=<uuid>`
- `NOTION_TITLE_PROPERTY=Name` (or whatever your database uses for the title property)
- `NOTION_DRY_RUN=true` (keep true until you’re ready)

Optional:
- `NOTION_VERSION=2026-03-11`

### Registry / npm
This repo’s frontend registry is pinned in `web/.npmrc`.
- It’s currently set to the SPD Nexus proxy (same idea as `../dreadnought/.npmrc`).
- If your machine isn’t on that network, temporarily switch `web/.npmrc` back to `https://registry.npmjs.org/`.

## How to verify end-to-end

1. Backend tests:
   - `make api-test`

2. Start backend + DB:
   - `make up`

3. Frontend install + lint:
   - `make web-install`
   - `make web-test`

4. Start frontend:
   - `make web-dev`
   - Open `http://localhost:3000`
   - Use the **Notion integration** card:
     - Check status (configured/dry_run)
     - Click “Create in Notion”

Expected results:
- If `NOTION_DRY_RUN=true`: you should see `dry_run: true` and no Notion page is created.
- If `NOTION_DRY_RUN=false` + correct config: you should get a Notion `url` back.

## Files changed/added (quick map)

Backend:
- `api/internal/connector/task.go`
- `api/internal/connector/notion/client.go`
- `api/internal/connector/notion/client_test.go`
- `api/internal/handler/rest/registrar.go`
- `api/internal/handler/rest/integrations/notion/handler.go`
- `api/internal/handler/rest/integrations/notion/handler_test.go`
- `api/cmd/serverd/router.go`
- `api/cmd/serverd/main.go`

Frontend:
- `web/app/api/notion/status/route.ts`
- `web/app/api/notion/pages/route.ts`
- `web/components/NotionDraftDemo.tsx`
- `web/app/page.tsx`
- `web/.npmrc`

Docs:
- `docs/ticket/T-010-optional-jira-connector.md` (updated to Notion)
- `docs/ticket/T-009-ticket-draft-model-and-local-export.md`
- `docs/ticket/T-011-local-ui-upload-and-results.md`
- `.env.example`

## Follow-ups / nice-to-haves
- Rename the ticket file to match reality (it still has `jira` in the filename):
  - `docs/ticket/T-010-optional-jira-connector.md` → `docs/ticket/T-010-optional-notion-connector.md`
- Extend `TicketDraft` → Notion mapping to support real database properties (Status, Priority, Assignee) once we decide a canonical database schema.
- Wire “Create in Notion” to the *real* exported drafts UI (T-011) instead of the demo payload.
