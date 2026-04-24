# T-010 — Optional: Jira connector (create draft tickets)

## Goal
Allow TEAmate to create Jira issues from ticket drafts (optional feature flag).

## Requirements / tasks
- Add a connector interface:
  - `create_issue(ticket_draft) -> issue_key/url`
- Implement Jira REST API client:
  - auth via API token
  - project key configuration
- Safety:
  - default to **dry-run** or **draft mode**
  - never auto-create without explicit user action

## Acceptance criteria
- When configured, user can click “Create in Jira” for a draft and receive a Jira key.
- If not configured, app still works with local export.

## Dependencies
- T-009
