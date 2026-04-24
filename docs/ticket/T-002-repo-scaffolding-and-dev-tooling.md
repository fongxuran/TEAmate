# T-002 — Repo scaffolding + dev tooling (local)

## Goal
Create a runnable local project skeleton with repeatable dev commands.

## Requirements / tasks
- Confirm / finalize the monorepo structure:
  - `api/` (Go backend; already present)
  - `web/` (Next.js frontend; to be added)
  - `docs/` (specs)
  - `build/` (Docker/build assets; optional)
- Add dependency management / tooling:
  - Go: `go mod tidy`, `go test ./...`, `go fmt ./...`
  - Web: `npm`/`pnpm` standard Next.js scripts (`dev`, `build`, `test`)
- Add local dev commands (Makefile or documented commands):
  - run API server
  - run Next.js dev server
  - run unit tests (API + web)
- Ensure `.env.example` covers configuration for:
  - Anthropic (Sonnet 4.6): `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`
  - Motion: `MOTION_API_KEY` (and any required workspace IDs)
  - Web ↔ API wiring: `NEXT_PUBLIC_API_BASE_URL`, `NEXT_PUBLIC_WS_URL`
- Add basic formatting/linting (optional but helpful):
  - Go: `gofmt` (and optionally `golangci-lint`)
  - Web: `eslint` + `prettier`

## Acceptance criteria
- New developer can run the app locally following README steps.
- `git status` shows scaffold files committed.

## Dependencies
- T-001
