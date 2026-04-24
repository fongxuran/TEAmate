# T-012 — Sample data + demo script

## Goal
Make the MVP demo repeatable.

## Requirements / tasks
- Create demo fixtures:
  - agenda example
  - transcript example(s)
- Create a short demo script:
  - what to click
  - what to say
  - expected outputs (screenshots optional)

Demo should include a realtime check:
- open the app on two devices (or two browser windows)
- type into the shared transcript textbox in one client
- verify the other client updates via WebSocket

Demo should also include a realtime drift prompt check:
- append a short off-agenda tangent that should trigger drift
- verify both clients receive a drift prompt (toast/banner/modal)
- click `Not drift` (or `Drift`) in one client
- verify the other client updates immediately (override applied)

## Acceptance criteria
- Anyone can reproduce the demo in <5 minutes locally.

## Demo fixtures
- Agenda (copy/paste): [docs/fixtures/t-012/agenda.example.txt](docs/fixtures/t-012/agenda.example.txt)
- Transcript (copy/paste): [docs/fixtures/t-012/transcript.example.txt](docs/fixtures/t-012/transcript.example.txt)
- Upload JSON (optional): [docs/fixtures/t-003/meeting_input.example.json](docs/fixtures/t-003/meeting_input.example.json)

## Demo script (5 minutes)
### 1) Start the app
- Run the local stack from the repo root:
  - `make up`
  - `make web-install`
  - `make web-dev`
- Open `http://localhost:3000` in two browser windows (A + B).

### 2) Realtime sync check
- In both windows, click **Connect** in the “Local MVP (T-011)” card.
- In window A, paste the agenda from [docs/fixtures/t-012/agenda.example.txt](docs/fixtures/t-012/agenda.example.txt).
  - Expected: the agenda text appears in window B within a second.
- In window A, paste the transcript from [docs/fixtures/t-012/transcript.example.txt](docs/fixtures/t-012/transcript.example.txt).
  - Expected: the transcript text appears in window B in real time.

### 3) Drift prompt check
- In window A, append this single line to the transcript:

```text
Sam: Quick tangent: the new espresso machine jammed again, and we spent 20 minutes on support chat.
```

- Expected: both windows show a **Drift alert** card.
- Click **Not drift** in window A.
  - Expected: the alert clears in both windows and the drift segment shows an override in the Drift view.

### 4) Analysis output check (optional)
- Click **Analyze now** in window A.
  - Expected: Summary/Decisions/Action items populate and Drift view lists segments.

### 5) Upload flow (optional)
- In either window, upload [docs/fixtures/t-003/meeting_input.example.json](docs/fixtures/t-003/meeting_input.example.json).
  - Expected: agenda + transcript fields auto-fill and sync to the other window.

## Dependencies
- T-011
