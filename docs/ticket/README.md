# TEAmate — MVP Ticket Backlog (runs locally)

This folder contains a setup → MVP task breakdown as **tickets**.

## MVP scope (what we’re building)
A local-running prototype that lets a user:
1) provide an **agenda** and a **meeting transcript** (text),
2) compute **agenda drift signals** and show nudges/flags,
3) generate **summary + decisions + action items**,
4) convert action items into **draft tickets** (local export; optional Jira connector).

### Explicit non-goals
- No CI/CD pipeline.
- No cloud hosting requirements (the app runs locally).
- No real-time multi-speaker audio segmentation.

> Note: some MVP modes may call an external LLM API (e.g., Claude) for drift detection. That still runs **locally** (no hosting), but requires an API key.

## Recommended MVP architecture (can be revised in T-001)
- **UI:** Streamlit (fast local demo) *or* minimal web UI.
- **Core logic:** Python module for parsing/segmenting transcript, drift scoring, extraction.
- **LLM provider:** pluggable (local Ollama or Claude/OpenAI API, depending on mode).

## Ticket conventions
- IDs: `T-###`
- Each ticket has:
  - goal
  - requirements / tasks
  - acceptance criteria
  - dependencies

## Tickets (ordered)
| ID | Title | Depends on |
|---:|---|---|
| T-001 | Decide MVP stack + LLM mode | — |
| T-002 | Repo scaffolding + dev tooling (local) | T-001 |
| T-003 | Transcript & agenda input format (schema) | T-001 |
| T-004 | Transcript segmentation (turns → chunks) | T-003 |
| T-005 | Drift scoring v1 (LLM classifier + fallback embeddings) | T-004 |
| T-006 | Drift feedback loop (“Not drift” correction) | T-005 |
| T-007 | Meeting summary generation | T-004 |
| T-008 | Decisions + action items extraction | T-004 |
| T-009 | Ticket draft model + local export (JSON/CSV/MD) | T-008 |
| T-010 | Optional: Jira connector (create draft tickets) | T-009 |
| T-011 | Local UI: upload transcript + show results | T-002, T-003, T-005, T-008, T-009 |
| T-012 | Sample data + demo script | T-011 |
| T-013 | Minimal tests for core logic | T-004, T-005, T-008 |
| T-014 | Local run documentation | T-011 |

---

## Definition of Done (MVP)
- Running locally on macOS:
  - user can upload/paste transcript + agenda
  - drift flags are shown on segments
  - summary + action items are produced
  - action items export as ticket drafts
- Clear README instructions for setup and running.
