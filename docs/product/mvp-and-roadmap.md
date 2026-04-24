# TEAmate — MVP and Roadmap

This document translates transcript ideas into a concrete, buildable scope.

## Product north star
A meeting companion that can:
1) understand agenda and conversation state,
2) detect drift and help steer back,
3) produce structured outcomes (summary, action items, tickets),
4) retain context across a chain of meetings.

## MVP (hackathon / prototype)
### Goal
Prove that TEAmate can take a meeting transcript-like stream and reliably:
- detect drift relative to the agenda,
- summarize the meeting,
- extract action items,
- create (or draft) tickets.

### Inputs
Choose one (or support both):
- **Transcript file upload** (simplest, reliable)
- **Text-based live feed** (participants type; or a speech-to-text system outside TEAmate supplies text)

### MVP features
1) **Agenda + topic tracking**
   - User provides agenda (bullets).
   - System maintains a “current topic” and a drift score.

2) **Drift detection + nudge UI**
   - When drift score crosses a threshold → nudge.
   - Allow user feedback:
     - “Not drift” (false positive)
     - “Drift” (confirm)
   - Store feedback for later tuning.

3) **Post-meeting summary**
   - Key decisions
   - Discussion highlights
   - Open questions

4) **Action items extraction**
   - Action, owner (if present), due date (if present), and source quote.

5) **Ticket drafting/creation**
   - Default mode: draft + approve.
   - Optional mode: auto-create.
   - Integration target: start with **one** tool (e.g., Jira) or write to local JSON/CSV.

### Non-goals (MVP)
- Real-time multi-speaker voice segmentation and low-latency “interrupt” enforcement.
- Organizational performance scoring.
- Deep calendar/meeting provider integration.

## Roadmap (post-MVP)
### 1) Long-term meeting memory
- Meeting series concept: “part 1 / part 2” continuity.
- Queryable decisions and action items.
- Start-of-meeting recap generated from prior meeting outcomes.

### 2) More integrations
- Jira + others (Monday.com, Motion)
- Chat integration (e.g., Telegram) for nudges, recaps, and action-item follow-ups
- MCP-based connectors to simplify adding tools

### 3) Coaching features
- Meeting organizer feedback (what worked / what didn’t).
- Participant guidance framed as improvement suggestions (opt-in, careful tone).

### 4) Enforcement policies (optional)
If teams want it, explore escalating nudges:
- Gentle reminder → stronger reminder → facilitator prompt.

Important: keep human control. TEAmate should recommend actions; facilitators decide.
