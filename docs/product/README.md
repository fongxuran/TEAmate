# TEAmate — Product Description

## What it is
TEAmate is an **agentic meeting companion** that helps teams run meetings that actually move work forward.

Unlike “meeting notes” tools that only summarize after the fact, TEAmate is designed to be **proactive**:
- It detects when the conversation **drifts away from the agenda** and nudges the group back.
- It turns decisions and action items into **trackable work** (e.g., tickets) that fit into existing workflows.
- It builds a **long-term meeting memory** so teams can recall what was decided, what’s pending, and where context lives.

In short: **TEAmate improves meeting productivity by adding gentle, practical agency to the moments where meetings usually leak time and intent.**

## The problem we’re solving
Meetings often fail for reasons that are predictable:
- The agenda exists, but the discussion **wanders**.
- Action items are mentioned, but **not captured** or not converted into real work.
- Context is scattered across chat logs, calendars, docs, and people’s memory.
- Only a few people “drive” the meeting; facilitators often have **limited control** over participants.

TEAmate targets these failure modes with a system that can:
1) understand the meeting agenda and current topic,
2) detect drift and steer the group back,
3) produce structured outputs that plug into how teams already execute.

## Who it’s for
Primary users (from the transcripts):
- **PMs / meeting facilitators** who need help keeping meetings on-track and converting discussion into outcomes.
- **Non-engineering roles** (e.g., operations / delivery / stakeholder management) who benefit from agentic workflows but should not need to “use code” to get value.

Secondary users:
- **Participants** who want clearer recap, accountability, and continuity across recurring meetings.

## What TEAmate does (core capabilities)
### 1) Agenda drift detection (the “stay on track” agent)
- Detects when the discussion moves off-topic.
- Prompts a lightweight nudge like “Drift detected — return to agenda?”
- Supports human-in-the-loop correction (e.g., “this is not drift”) to improve future behavior.

> Design intent: the AI is a helper; the facilitator/PM remains the driver. The agent should feel supportive, not authoritarian.

### 2) Meeting summary + action items
- Produces a post-meeting summary.
- Extracts action items and assigns owners (where possible).

### 3) From action item → ticket (workflow integration)
- Converts action items into tickets/tasks (first target: **Motion**; other tools later).
- Supports a review/approval step (create-as-draft → confirm → submit) or an auto-create mode depending on team preference.
- Can also export to a local file when integrations aren’t available.

### 4) Long-term meeting memory (source of truth)
- Retains prior meeting context so the next meeting can start with:
  - what was previously decided,
  - what action items are still pending,
  - relevant recap for the current agenda.

This enables queries like:
- “When did we decide this, and what were the tradeoffs?”
- “What are the open action items from meeting part 1 / part 2?”

### 5) Coaching and feedback (roadmap)
A longer-term direction discussed in the transcripts is coaching:
- Provide feedback to the organizer on how the meeting went.
- Optionally provide individualized guidance (carefully framed so it’s helpful and not perceived as a performance review).
- Help facilitators navigate different stakeholder styles (“how to deal with this person next time”).

## What makes it different
- **Agency, not just reporting:** TEAmate doesn’t only summarize; it can nudge, steer, and execute follow-ups.
- **Workflow-first:** outputs become real work (tickets / tasks) rather than another static document.
- **Continuity across meetings:** persistent context turns recurring meetings into a coherent thread instead of isolated events.

## MVP focus (what we’d build first)
Based on the transcripts, the MVP should prove feasibility for:
1) **Drift detection on text** (agenda ↔ current topic ↔ drift) with simple nudges.
2) **Summary + action item extraction** from a transcript.
3) **Ticket creation** via a single easy integration (**Motion**) or local export.

MVP UX note:
- The MVP input is a **shared textbox** where multiple devices can join and type simultaneously (frontend ↔ backend over WebSockets).

MVP implementation note:
- Real-time voice processing is challenging due to segmentation/latency. A pragmatic MVP can operate on **text inputs** (pre-transcribed meeting, or a simple text-based “meeting feed”) while keeping the future vision of “joins the meeting and transcribes automatically.”

## Future roadmap (directional)
- Automatic meeting join + transcription (Teams/Zoom/etc.)
- Richer integrations (Monday.com, Motion, chat apps like Telegram, etc.)
- Stronger enforcement patterns (e.g., escalation policies) *only if the team wants it* and with clear safeguards
- Org-level analytics: recurring drift patterns, common blockers, meeting health trends

## Guardrails and risks to address
- **False positives / gaming the system:** drift detection must allow correction and should avoid over-policing.
- **Adoption inertia:** the product must be seamless with existing workflows; otherwise it becomes “another tool.”
- **Privacy and trust:** meeting content is sensitive. Teams need transparent controls on retention, access, and what is auto-shared or auto-created.

## Success looks like
- Fewer meetings that end with “so… what are the next steps?”
- Higher action-item completion rate and faster follow-through.
- Reduced time spent re-explaining context in recurring meetings.
