# T-001 — Decide MVP stack + LLM mode

## Goal
Lock down a buildable local MVP plan: tech stack + which “AI brain” we will use locally.

## Requirements / tasks
- Lock decisions (this ticket records the chosen direction):
  - **Frontend:** TypeScript + React via **Next.js** (local web UI)
  - **Backend:** **Go** API (this repo’s `api/` module)
  - **Realtime:** frontend ↔ backend via **WebSockets** to support multi-device simultaneous input
  - **LLM:** **Anthropic Claude Sonnet 4.6** (API key not ready yet)
    - Until `ANTHROPIC_API_KEY` is available, MVP must still run using a deterministic fallback (heuristics/stub/offline baseline).
  - **Ticket creation integration:** **Motion API** (we will be provided a Motion API key)
    - Local export remains available as a fallback.

## Decision (locked)
- Stack: **Next.js (React/TypeScript) + Go backend**
- LLM mode: **Sonnet 4.6 via Anthropic API when available**, otherwise fallback/no-key mode
- Ticket output: **Create in Motion (when configured)** + local export

## Repo structure outline
- `api/` — Go backend (REST + WebSocket)
- `web/` — Next.js frontend
- `docs/` — product + ticket specs
- `build/` — Docker/build assets (optional)

## Acceptance criteria
- A short written decision in this ticket file:
  - selected stack
  - selected LLM mode
  - selected export/integration approach
  - repo structure outline (folders)

## Notes
- Keeping LLM provider pluggable reduces lock-in.
- Realtime collaborative input does **not** require full OT/CRDT for MVP; an append-only shared “meeting feed” over WebSockets is acceptable.
