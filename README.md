# TEAmate

Monorepo for TEAmate MVP.

- `api/` — Go backend (REST now; WebSocket planned)
- `web/` — Next.js frontend (local UI)
- `docs/` — product + ticket specs
- `build/` — docker assets

## Prerequisites

- Go 1.21+
- Node.js 20+
- Docker Desktop (for local Postgres)

## Local quickstart

### 1) Environment

Config templates live in `.env.example`.

For the API + shared config, copy the template:

- copy `.env.example` → `.env`

For the frontend, copy the Next.js template:

- copy `web/.env.local.example` → `web/.env.local`

If you do not have API keys yet, leave `ANTHROPIC_API_KEY` and `MOTION_API_KEY` empty; the app will still run with limited functionality.

### 2) Install dependencies

- `make web-install`
- `cd api && go mod download`

### 3) Start backend + DB (Docker)

Run Postgres + API using the provided compose file.

- `make up`

- API: `http://localhost:8080`
- Health: `http://localhost:8080/health`

The API defaults to HTTP BasicAuth (`admin` / `password`). You can disable it for local experiments by setting `API_AUTH_DISABLED=true`.

### 4) Start frontend

From `web/`, install deps and run the dev server:

- `make web-dev`

- UI: `http://localhost:3000`

The UI proxies requests via a Next.js route handler at `/api/messages`, which forwards to the Go API.

### 5) Run the app locally without Docker (optional)

If you prefer to run the API locally while using Docker for Postgres:

- `make db-up`
- set `DATABASE_URL` in `.env` to point at the Postgres container
- `make api-run`

#### Note on `npx` / scaffolding

The `web/` Next.js scaffold is checked into the repo, so you **do not** need to run `npx create-next-app`.

If `make web-install` fails due to npm registry issues, check the registry configuration.

- This repo pins the frontend registry via `web/.npmrc`.
- In SPD environments, the intended setting is the Nexus proxy: `https://nexus.in.spdigital.sg/repository/npm-all` (mirrors the reference setup in `../dreadnought/.npmrc`).

## Dev commands

A Makefile is provided at the repo root:

- `make up` / `make down`
- `make api-test` / `make api-fmt` / `make api-tidy`
- `make web-install` / `make web-dev` / `make web-test`

## Tests

- `make api-test`
- `make web-test`

## Troubleshooting

- WebSocket connection issues: confirm `NEXT_PUBLIC_WS_URL` and `CORS_ALLOWED_ORIGINS` match `http://localhost:3000`, and ensure the UI is proxying to the API base URL.
- Drift prompts not appearing: verify the WebSocket is connected, check any drift threshold settings, and allow extra time if running in a remote LLM mode.
- Missing `ANTHROPIC_API_KEY`: the app falls back to non-LLM behavior or cached outputs; summaries and drift prompts may be unavailable.
- Missing `MOTION_API_KEY`: Motion sync is disabled, but local export still works.

## Notes

- T-002 scaffolding: this repo intentionally starts with a tiny “messages” API + UI wiring, so future tickets can build real drift/summary logic on top.
- T-011 UI: agenda input + shared transcript textbox over WebSockets is the next step.
