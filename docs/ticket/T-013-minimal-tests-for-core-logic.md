# T-013 — Minimal tests for core logic

## Goal
Reduce regressions while iterating quickly.

## Requirements / tasks
- Add unit tests for:
  - transcript parsing / segmentation determinism (T-004)
  - drift scoring returns expected shape (T-005)
  - action item extraction schema validation (T-008)
- Add smoke test:
  - end-to-end run on a fixture transcript without UI.

## Acceptance criteria
- Tests run locally with a single command.
- Core logic changes don’t silently break schemas.

## Dependencies
- T-004, T-005, T-008
