# T-006 — Drift feedback loop (real-time prompt + “Not drift” correction)

## Goal
Allow human-in-the-loop correction to reduce false positives and refine boundaries.

In the MVP, this loop is **real time**: as users submit new transcript text to the backend, drift scoring runs and, when drift is detected, the backend prompts **all connected clients** to confirm/override.

This ticket is the **mechanism** for capturing human labels and applying them as overrides on top of T-005 drift scoring outputs.

## Context
From T-005, each segment will have a model output roughly shaped like:

- `best_agenda_item_id`
- `label` in {`on_track`, `maybe_drift`, `drift`}
- `confidence` (0–1)

T-006 adds a human override layer that:

1) updates the UI immediately (local + all connected clients)
2) persists to disk
3) is reloadable and re-applies deterministically
4) can optionally feed back into future scoring (few-shot / thresholds)

This ticket also defines the **prompting loop** for realtime drift:

- When the backend detects `maybe_drift`/`drift` on newly-submitted text, it emits a `drift_alert` event to connected clients.
- A connected user confirms/overrides via a feedback action.
- The backend persists that feedback and broadcasts the applied override so all clients stay consistent.

## MVP user flow

### A) Realtime prompting flow (primary)

- Users type into the shared transcript textbox and **submit/append** text to the backend.
- Backend (async) updates segmentation + drift scoring for the affected segment(s).
- If drift is detected (label is `maybe_drift` or `drift`, and meets thresholds), backend sends a **prompt** to all connected clients.
- Any connected user can respond:
  - **Confirm drift** (human says: this is drift)
  - **Mark not drift** (human says: this is not drift)
  - (Optional) **Reset** to remove human override
- UI updates immediately and the applied override is reflected across all clients.
- Feedback entry is appended to a local JSON log.
- On reload, the log is re-applied so overrides persist.

### B) Manual review flow (fallback)

- User views segments with their model drift label.
- For any segment, user can confirm/override/reset.
- UI updates immediately; feedback is persisted and reloadable.

## Requirements / tasks
- UI affordance per segment:
  - mark as `not_drift` (override)
  - mark as `drift` (confirm)
- Realtime prompting:
  - backend emits a `drift_alert` event to connected clients when drift is detected on newly submitted text
  - UI shows a lightweight prompt (toast/banner/modal) with one-click actions: `Drift` / `Not drift`
  - when any client submits feedback, the backend broadcasts `drift_feedback_applied` so all clients converge
- Persist feedback locally:
  - store as JSON with (agenda, segment, label, timestamp)
- Apply feedback in scoring:
  - MVP approach: simple override rules + store feedback as labeled examples
  - If using Claude mode, optionally incorporate recent labeled examples as **few-shot** context in the drift-classification prompt
  - Stretch: use feedback to adjust thresholds per agenda item (offline mode) or to refine prompt templates (Claude mode)

## Feedback data model (v1)

### Labels
Human labels are intentionally binary for MVP:

- `drift`: human confirms this segment is drift
- `not_drift`: human overrides a (suspected) drift to “not drift”

These are stored separately from the model’s multi-class output.

### `DriftFeedbackEntry`
JSON object fields (required unless marked optional):

- `schema_version`: string, must be `"v1"`
- `timestamp`: RFC3339 string
- `meeting_id`: string (recommended). Use `MeetingInput.transcript.meeting_id` when available.
- `agenda_item_id`: string (the agenda item the segment is being judged against; typically the model’s `best_agenda_item_id` at the time of feedback)
- `segment_id`: string (from T-004: e.g. `"seg-12-17"`)
- `label`: `"drift" | "not_drift"`

Optional but recommended for robustness/debugging:

- `agenda_hash`: string (hash of the agenda array, to detect “agenda changed” situations)
- `segment_start_turn_idx`: number
- `segment_end_turn_idx`: number
- `segment_text_excerpt`: string (short excerpt for human/debug; not required for applying overrides)
- `model_label_before`: `"on_track" | "maybe_drift" | "drift"`
- `model_confidence_before`: number
- `notes`: string (freeform, optional)

### Storage format
Use **JSON Lines** (append-only) for MVP durability and simplicity:

- File: `data/feedback/drift_feedback.v1.jsonl`
- Each line is one `DriftFeedbackEntry` JSON object.

Example fixture:

- `docs/fixtures/t-006/drift_feedback.v1.example.jsonl`

Rationale: append-only writes are easy, and “latest entry wins” gives deterministic state reconstruction.

## Storage + merge rules

### Storage location
Persist to disk under repo-local data directory:

- Default path: `data/feedback/drift_feedback.v1.jsonl`
- Recommended: allow override via env var (implementation detail): `TEAmate_FEEDBACK_PATH`

The file SHOULD be treated as local state (typically gitignored).

### Merge key
When reconstructing effective overrides, reduce the JSONL log into a map using the key:

$(meeting\_id, agenda\_hash, agenda\_item\_id, segment\_id)$

If `agenda_hash` is not present, treat it as `""`.

### Conflict resolution
- If multiple entries exist for the same key, **the one with the latest `timestamp` wins**.
- If timestamps collide, prefer the last record in file order.

### Agenda/segment drift over time
- If `agenda_hash` mismatches current agenda, do not silently apply; surface a UI warning like “feedback is from a different agenda version”.
- If a `segment_id` no longer exists (due to segmentation option changes), ignore the override and optionally surface a “stale feedback” count.

## Applying feedback in scoring (MVP)

### Output layering
When scoring produces a model label, derive a “final label” for display/export:

- Keep `model_label` as produced by scorer (T-005).
- Compute `final_label` by applying overrides.

### Override rules
For a segment with a matching override entry:

- If human label is `drift` → set `final_label = drift`.
- If human label is `not_drift` → set `final_label = on_track`.

Also set `is_overridden = true` and (optional) `override_label = drift | not_drift` for UI.

Note: mapping `not_drift → on_track` is the simplest MVP. If you later want `not_drift` to mean “not drift but maybe off-topic”, introduce a richer human label set (out of scope for v1).

## Claude / LLM mode: few-shot feedback (optional)
When scoring a segment with an LLM, include a short “recent corrections” block to reduce repeat mistakes.

Suggested approach:

- Pull the last $k$ feedback entries matching the current `agenda_hash` (and optionally across meetings) where `segment_text_excerpt` exists.
- Add them as labeled examples:
  - `Example: <excerpt> → drift`
  - `Example: <excerpt> → not_drift`

Keep $k$ small (e.g. 5–20) to control prompt size.

## Stretch ideas (not required for MVP)
- **Per-agenda-item threshold tuning (offline mode)**
  - Track confusion counts for `maybe_drift` vs human overrides per `agenda_item_id`.
  - Adjust similarity thresholds per agenda item based on recent corrections.
- **Prompt refinement (LLM mode)**
  - If many `not_drift` overrides happen for a certain pattern (e.g. implementation detail tangents), add an explicit instruction to treat those as on-track.
- **Boundary refinement**
  - Add feedback type `split_segment_here` / `merge_with_next` to improve segmentation (would feed back into T-004).

## Acceptance criteria
- User can override drift label and see immediate UI update.
- When drift is detected during realtime input, all connected clients receive a prompt and converge after any one user responds.
- Feedback is saved to disk and reloadable.

## Verify (when implemented)
- Start the local UI and load a meeting.
- Apply `drift` and `not_drift` overrides on two segments.
- Refresh the page (or restart the service). Overrides should re-appear.
- Confirm the on-screen “final label” changes immediately while the “model label” remains visible (if UI supports it).
- Open two browser windows connected to the same meeting.
  - Append new transcript text that triggers drift.
  - Verify both windows receive the drift prompt.
  - Submit `not_drift` from one window and verify the other window updates.

## Dependencies
- T-005
