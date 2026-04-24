# T-009 — Ticket draft model + local export

## Goal
Convert extracted action items into ticket drafts and export locally.

## Requirements / tasks
- Define `TicketDraft` schema:
  - `title`, `body/description`, `labels/tags`, `owner`, `priority` (optional)
  - `source_action_item_id`
- Implement export formats:
  - JSON (primary)
  - Markdown (human-friendly)
  - CSV (optional)
- Ensure each ticket draft includes traceability back to meeting + segment IDs.

## Acceptance criteria
- Export generates files under a local output directory (e.g., `out/`).
- Export contains all extracted action items as drafts.

## Notes
- Drafts should be shaped so they can be mapped cleanly into Motion tasks (T-010) when `MOTION_API_KEY` is configured.

## Dependencies
- T-008
