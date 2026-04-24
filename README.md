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

For the frontend, copy the Next.js template:

- copy `web/.env.local.example` → `web/.env.local`

### 2) Start backend + DB (Docker)

Run Postgres + API using the provided compose file.

- `make up`

- API: `http://localhost:8080`
- Health: `http://localhost:8080/health`

The API defaults to HTTP BasicAuth (`admin` / `password`). You can disable it for local experiments by setting `API_AUTH_DISABLED=true`.

### 3) Start frontend

From `web/`, install deps and run the dev server:

- `make web-install`
- `make web-dev`

- UI: `http://localhost:3000`

The UI proxies requests via a Next.js route handler at `/api/messages`, which forwards to the Go API.

#### Note on `npx` / scaffolding

The `web/` Next.js scaffold is checked into the repo, so you **do not** need to run `npx create-next-app`.

If `make web-install` fails due to a custom/private npm registry configuration, update your npm registry settings to a registry that contains the required packages (e.g. the public npm registry) and retry the install.

## Dev commands

A Makefile is provided at the repo root:

- `make up` / `make down`
- `make api-test` / `make api-fmt` / `make api-tidy`
- `make web-install` / `make web-dev` / `make web-test`

## Notes

- T-002 scaffolding: this repo intentionally starts with a tiny “messages” API + UI wiring, so future tickets can build real drift/summary logic on top.
- T-011 UI: agenda input + shared transcript textbox over WebSockets is the next step.
