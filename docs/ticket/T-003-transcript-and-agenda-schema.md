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
- Decide what MVP accepts:
  - Plain text transcript (like the docs) + best-effort parsing
  - Or JSON upload (more structured)
  - Recommended: support both (plain text for ease, JSON for repeatability)
- Document the schema + 1 example file.

## Acceptance criteria
- A documented schema in this ticket file.
- At least one example transcript input that can be used as a test fixture.

## Dependencies
- T-001
