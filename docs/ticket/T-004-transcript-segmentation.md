# T-004 — Transcript segmentation (turns → chunks)

## Goal
Convert transcript turns into analysis chunks suitable for drift scoring and extraction.

## Requirements / tasks
- Implement segmentation rules:
  - group turns into chunks of ~N tokens or ~M seconds (MVP: token/character based)
  - keep mapping to source turns (for explainability)
- Output `Segment` objects:
  - `segment_id`
  - `start_turn_idx`, `end_turn_idx`
  - `text`
  - optional `speaker_distribution`
- Provide a simple renderer that can show segment boundaries.

## Acceptance criteria
- Given a transcript fixture, segmentation output is deterministic.
- Each segment has a stable ID and contains non-empty text.

## Dependencies
- T-003
