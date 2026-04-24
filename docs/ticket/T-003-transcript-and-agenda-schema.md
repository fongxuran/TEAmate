# T-003 — Transcript & agenda input format (schema)

## Goal
Define the canonical input formats so all downstream logic is deterministic.

This ticket defines **v1** canonical schemas for:

- agenda
- transcript (turn-based)
- realtime WebSocket events (append-only for MVP)

It also defines what the MVP accepts and provides a fixture you can use in tests.

## Conventions (v1)

- **Encoding**: JSON (UTF-8).
- **Timestamps**: RFC3339 strings (e.g. `"2026-04-24T09:30:00Z"`).
- **IDs**: treat as opaque strings.
- **Optional fields**: may be omitted or set to `null` (MVP should treat both as missing).
- **Ordering**:
  - `agenda` order is meaningful (the meeting’s intended flow).
  - `transcript.turns` order is chronological.
  - `RealtimeMessage` ordering is by `(timestamp, client_id)` when timestamps collide.

## Canonical meeting input (recommended wrapper)

Downstream tickets (segmentation, drift scoring, extraction) all need **agenda + transcript**.
To keep inputs repeatable, we define a single wrapper payload:

### `MeetingInput`

Required:

- `schema_version`: string (MUST be `"v1"` for this spec)
- `agenda`: `AgendaItem[]`
- `transcript`: `Transcript`

Optional:

- `source`: string (where this came from: `"realtime" | "paste" | "upload"` etc.)

```json
{
  "schema_version": "v1",
  "source": "paste",
  "agenda": [/* AgendaItem[] */],
  "transcript": {/* Transcript */}
}
```

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
- Define server→client realtime events needed by downstream UX:
  - `DriftAlert` (prompt connected users when drift is detected)
  - `DriftFeedbackApplied` (broadcast convergence when any client responds)
- Decide what MVP accepts:
  - **Realtime textbox stream** (primary): clients send messages over WebSocket; backend builds a canonical transcript
  - Plain text transcript (paste/upload) as fallback
  - JSON upload (structured; repeatable) as a stretch
- Document the schema + 1 example file.

## Schema definitions (v1)

### `AgendaItem`

- `id` (string, required): stable agenda identifier, unique within a meeting.
- `title` (string, required): short label (shown in UI).
- `description` (string, optional): longer detail.
- `keywords` (string[], optional): helpful tags for matching; free-form.

```json
{
  "id": "a-1",
  "title": "Project status",
  "description": "Where we are today; blockers",
  "keywords": ["status", "blockers", "milestones"]
}
```

### `TranscriptTurn`

- `timestamp` (string RFC3339, optional): time within meeting; can be omitted for pasted transcripts.
- `speaker` (string, optional): display name or label (e.g. `"Alex"`, `"Speaker 1"`).
- `text` (string, required): the spoken content for this turn. Must be non-empty after trimming.

```json
{
  "timestamp": "2026-04-24T09:31:12Z",
  "speaker": "Alex",
  "text": "Let’s start with the release timeline."
}
```

### `Transcript`

- `meeting_id` (string, optional): stable identifier if known.
- `meeting_name` (string, optional): human label.
- `turns` (`TranscriptTurn[]`, required): in chronological order.

Constraint (application-level): at least one of `meeting_id` or `meeting_name` SHOULD be present.

```json
{
  "meeting_id": "m-2026-04-24",
  "meeting_name": "Weekly product sync",
  "turns": [/* TranscriptTurn[] */]
}
```

### `RealtimeMessage` (WebSocket event)

Primary MVP mode is **append-only** (clients send deltas):

- `type` (string, required): must be `"realtime_message"`.
- `client_id` (string, required): unique per browser/device instance.
- `timestamp` (string RFC3339, required): event time.
- `text_delta` (string, required for MVP): append-only text.
- `author` (string, optional): display name.

Optional/Stretch mode (snapshot):

- `text` (string): full content snapshot (useful for reconciliation / recovery).

Rules:

- Exactly one of `text_delta` or `text` MUST be provided.
- For MVP, servers MAY reject `text` snapshots and only accept `text_delta`.

```json
{
  "type": "realtime_message",
  "client_id": "c-8d7b1d",
  "timestamp": "2026-04-24T09:30:05Z",
  "author": "Sam",
  "text_delta": "Hi all — starting now.\n"
}
```

### `DriftAlert` (WebSocket event)

Server emits this event to **prompt connected users** when drift is detected on newly-submitted text.

- `type` (string, required): must be `"drift_alert"`.
- `timestamp` (string RFC3339, required)
- `meeting_id` (string, optional but recommended): same semantic as `Transcript.meeting_id`.
- `segment_id` (string, required)
- `agenda_item_id` (string, required): agenda item the segment is being judged against (typically `best_agenda_item_id`).
- `model_label` (string, required): `"on_track" | "maybe_drift" | "drift"`.
- `confidence` (number 0–1, required)

Optional but useful for UX/debugging:

- `rationale` (string)
- `segment_text_excerpt` (string)

```json
{
  "type": "drift_alert",
  "timestamp": "2026-04-24T09:34:12Z",
  "meeting_id": "m-2026-04-24",
  "segment_id": "seg-12-17",
  "agenda_item_id": "a-2",
  "model_label": "drift",
  "confidence": 0.86,
  "segment_text_excerpt": "...anyway, about the pricing model...",
  "rationale": "This discussion is about pricing; agenda item a-2 is release planning."
}
```

### `DriftFeedbackApplied` (WebSocket event)

When any client submits feedback (T-006), the server broadcasts an applied override so **all clients converge**.

- `type` (string, required): must be `"drift_feedback_applied"`.
- `timestamp` (string RFC3339, required)
- `meeting_id` (string, optional but recommended)
- `agenda_item_id` (string, required)
- `segment_id` (string, required)
- `label` (string, required): `"drift" | "not_drift"`

```json
{
  "type": "drift_feedback_applied",
  "timestamp": "2026-04-24T09:34:25Z",
  "meeting_id": "m-2026-04-24",
  "agenda_item_id": "a-2",
  "segment_id": "seg-12-17",
  "label": "not_drift"
}
```

## What the MVP accepts

### Primary: realtime textbox stream

- Transport: WebSocket (TBD in later ticket).
- Payload: `RealtimeMessage` events.
- Backend responsibility: build a canonical `Transcript` deterministically from the event stream.

### Fallback: plain text transcript (paste/upload)

MVP accepts a single UTF-8 text blob.

Suggested line format (best-effort parsing; not strictly required):

- `HH:MM:SS Speaker: text`
- or `Speaker: text`
- or just `text`

Server should preserve the original as `TranscriptTurn.text` if parsing fails.

### Stretch: JSON upload

- Accept `MeetingInput` JSON for repeatability and testing.

## Example fixture

- `docs/fixtures/t-003/meeting_input.example.json`

This file is intended to be reused as a deterministic fixture for T-004+ unit tests.

## Acceptance criteria
- A documented schema in this ticket file.
- At least one example transcript input that can be used as a test fixture.

## Dependencies
- T-001
