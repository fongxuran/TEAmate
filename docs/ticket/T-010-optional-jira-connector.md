# T-010 — Optional: Motion connector (create tasks from drafts)

## Goal
Allow TEAmate to create Motion tasks from ticket drafts (optional feature flag).

## Requirements / tasks
- Add a connector interface:
  - `create_task(ticket_draft) -> task_id/url`
- Implement Motion API client:
  - auth via API key (`MOTION_API_KEY`)
  - workspace/project/list configuration (IDs in env or config file)
- Safety:
  - default to **dry-run** or **draft mode**
  - never auto-create without explicit user action

## Acceptance criteria
- When configured, user can click “Create in Motion” for a draft and receive a Motion task link.
- If not configured, app still works with local export.

## Dependencies
- T-009
