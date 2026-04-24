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

## Dependencies
- T-011
