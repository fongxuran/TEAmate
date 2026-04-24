# T-006 — Drift feedback loop (“Not drift” correction)

## Goal
Allow human-in-the-loop correction to reduce false positives and refine boundaries.

## Requirements / tasks
- UI affordance per segment:
  - mark as `not_drift` (override)
  - mark as `drift` (confirm)
- Persist feedback locally:
  - store as JSON with (agenda, segment, label, timestamp)
- Apply feedback in scoring:
  - MVP approach: simple override rules + store feedback as labeled examples
  - If using Claude mode, optionally incorporate recent labeled examples as **few-shot** context in the drift-classification prompt
  - Stretch: use feedback to adjust thresholds per agenda item (offline mode) or to refine prompt templates (Claude mode)

## Acceptance criteria
- User can override drift label and see immediate UI update.
- Feedback is saved to disk and reloadable.

## Dependencies
- T-005
