# T-010 — Optional: Notion connector (create pages from drafts)

## Goal
Allow TEAmate to create Notion pages from ticket drafts (optional integration).

## Requirements / tasks
- Add a connector interface:
  - `create_task(ticket_draft) -> task_id/url`
- Implement Notion API client:
  - auth via API key (`NOTION_API_KEY`)
  - database configuration (`NOTION_DATABASE_ID`)
  - title property name config (`NOTION_TITLE_PROPERTY`, default: `Name`)
- Safety:
  - default to **dry-run** (`NOTION_DRY_RUN=true`)
  - never create pages without explicit user action (a POST call / button click)

## Acceptance criteria
- When configured, user can click “Create in Notion” for a draft and receive a Notion page link.
- If not configured, app still works with local export.

## Dependencies
- T-009
