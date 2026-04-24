# T-011 — Local UI: upload transcript + show results

## Goal
Deliver the end-to-end local MVP experience.

## Requirements / tasks
UI screens/sections:
1) **Inputs**
   - agenda input (textbox)
   - **shared transcript textbox (realtime)**
     - multiple devices can connect simultaneously
     - updates flow via WebSocket (append-only feed is acceptable for MVP)
   - optional transcript paste/upload (fallback)
   - config toggles (LLM provider, thresholds)
2) **Drift view**
   - list segments with drift label + best agenda item
   - quick filters: show only drift
   - actions: mark “Not drift” / “Drift” (T-006)
3) **Outcomes**
   - summary (T-007)
   - decisions + action items (T-008)
4) **Exports**
   - download local ticket drafts (T-009)
   - optional Motion create button if enabled (T-010)

## Acceptance criteria
- A user can run the UI locally and complete an analysis flow end-to-end.
- Demo works using `docs/transcript/meeting 1.txt` or a fixture.

## Dependencies
- T-002, T-003, T-005, T-008, T-009
