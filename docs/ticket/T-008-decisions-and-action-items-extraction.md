# T-008 — Decisions + action items extraction

## Goal
Extract structured outcomes that can be converted into work.

## Requirements / tasks
- Define `ActionItem` schema:
  - `title`
  - `description`
  - `owner` (optional)
  - `due_date` (optional)
  - `source_segment_ids`
  - `confidence`
- Define `Decision` schema:
  - `statement`
  - `source_segment_ids`
  - `confidence`
- Implement extraction:
  - LLM-based extraction per chunk + merge/dedupe (Sonnet 4.6 when available)
  - basic dedupe (same title / high similarity)

## Acceptance criteria
- For sample transcript, outputs at least:
  - 1+ decision (if present)
  - 3+ action items (or as available)
- Each action item links to the source segments.

## Dependencies
- T-004
