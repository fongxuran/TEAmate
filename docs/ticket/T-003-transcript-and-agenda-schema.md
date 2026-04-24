# T-003 — Transcript & agenda input format (schema)

## Goal
Define the canonical input formats so all downstream logic is deterministic.

## Requirements / tasks
- Define `AgendaItem`:
  - `id`, `title`, optional `description`, optional `keywords`
- Define `TranscriptTurn`:
  - `timestamp` (optional)
  - `speaker` (optional)
  - `text` (required)
- Define `Transcript`:
  - meeting id/name
  - list of turns
- Define a realtime input event (for the multi-device textbox MVP):
  - `RealtimeMessage`:
    - `client_id`
    - `timestamp`
    - `text_delta` (append-only for MVP) or `text` (full snapshot)
    - optional `author` (display name)
- Decide what MVP accepts:
  - **Realtime textbox stream** (primary): clients send messages over WebSocket; backend builds a canonical transcript
  - Plain text transcript (paste/upload) as fallback
  - JSON upload (structured; repeatable) as a stretch
- Document the schema + 1 example file.

## Acceptance criteria
- A documented schema in this ticket file.
- At least one example transcript input that can be used as a test fixture.

## Dependencies
- T-001
