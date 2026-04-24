# T-007 — Meeting summary generation

## Goal
Generate a readable meeting summary from segments.

## Requirements / tasks
- Define summary sections:
  - purpose / agenda recap
  - key discussion points
  - decisions
  - action items (links to T-008 output)
  - open questions / risks
- Implement summarization strategy:
  - LLM prompt using chunked context
  - fallback heuristic summary if LLM unavailable
- Ensure the summary references agenda items when possible.

## Acceptance criteria
- Produces a summary for the sample transcript.
- Summary is deterministic enough for demo (prompt version pinned).

## Dependencies
- T-004
