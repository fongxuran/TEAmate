# T-007 — Meeting summary generation

## Goal
Generate a readable, structured meeting summary from transcript `Segment`s (T-004), with optional links back to agenda items and source segments for explainability.

This ticket focuses on **summary generation logic + data model**. UI wiring is handled later in T-011.

## Context
We already have:

- `model.MeetingInput` (agenda + transcript): `api/internal/model/transcript.go`
- deterministic segmentation into `model.Segment`: `api/internal/model/transcript_segmentation.go`
- a canonical fixture: `docs/fixtures/t-003/meeting_input.example.json`

T-007 produces a “human-friendly” summary that:

- is useful even in **no-LLM** mode
- is stable enough for a demo (deterministic fallback + caching/prompt pinning)
- can optionally embed T-008 results (decisions + action items) when available

### Non-goals (MVP)
- Perfect abstractive summarization quality.
- Cross-meeting memory.
- Realtime incremental summaries (can be a stretch later).

## Inputs

### Required
- `MeetingInput` (v1)
- `[]Segment` from `model.SegmentTranscript(transcript, opts)`

Recommended segmentation options for summary generation:
- `IncludeSpeakerLabels: true` (keeps attributions)
- a token/char cap that yields ~5–30 segments for typical meetings

### Optional (if available)
- decisions + action items extracted by T-008.

Important: T-007 should not *require* T-008 to function; it should gracefully omit those sections.

## Output data model (v1)
Define a JSON-serializable `MeetingSummary` model (recommended location: `api/internal/model/meeting_summary.go`).

### `MeetingSummary`
Fields (required unless marked optional):

- `schema_version`: string, must be `"v1"`
- `generated_at`: RFC3339 string
- `meeting_id`: string (optional; use `MeetingInput.transcript.meeting_id` when present)
- `meeting_name`: string (optional)
- `prompt_version`: string (required; pinned version identifier like `"summary_prompt.v1"`)
- `mode`: `"llm" | "fallback"`
- `agenda_hash`: string (recommended; see below)
- `segments_hash`: string (recommended; see below)

Content payload:

- `sections`: array of `SummarySection` (required; ordered)
- `agenda_item_summaries`: array of `AgendaItemSummary` (optional but recommended)

Linking / traceability:

- `source_segment_ids`: array of strings (optional; union of cited segments across sections)

### `SummarySection`
- `id`: string (stable slug; e.g. `"purpose"`, `"key-points"`)
- `title`: string
- `markdown`: string (human-readable)
- `source_segment_ids`: array of strings (optional; for explainability)

### `AgendaItemSummary`
- `agenda_item_id`: string
- `title`: string
- `bullets`: array of strings
- `source_segment_ids`: array of strings

### Summary sections (required ordering)
The produced summary should include (even if some sections are empty):

1) **Purpose / agenda recap**
2) **Key discussion points**
3) **Decisions** (may be empty if none detected / T-008 absent)
4) **Action items** (may be empty if none detected / T-008 absent)
5) **Open questions / risks**

## Determinism + caching

### Hashes
To make outputs reproducible for a demo:

- `agenda_hash`: stable hash of agenda items (IDs + titles + descriptions + keywords)
- `segments_hash`: stable hash of segment IDs + text

### Cache rules (recommended)
Persist the *final* `MeetingSummary` to disk keyed by:

$(meeting\_id, agenda\_hash, segments\_hash, prompt\_version)$

Proposed storage:
- default path: `data/summary/meeting_summary.v1.jsonl` (append-only) or `data/summary/` directory with one JSON per key
- optional env override: `SUMMARY_CACHE_PATH`

Cache behavior:
- if an entry exists for the key → return it (deterministic demo behavior)
- allow a “force regenerate” boolean to bypass cache (implementation detail)

## Summarization strategy

### Primary: LLM summarization (Sonnet 4.6 when available)

Requirements:
- Use `LLM_PROVIDER` to select mode (see `.env.example`). For MVP:
  - `LLM_PROVIDER=claude` uses Anthropic
  - fallback when provider is unset/unknown or required keys are missing
- Use `ANTHROPIC_API_KEY` when present.
- Use `ANTHROPIC_MODEL` (e.g. `sonnet-4.6`) when present; otherwise default to the repo’s chosen model.
- Keep parameters stable:
  - `temperature = 0`
  - prompt template version pinned via `prompt_version`
- Prefer a **JSON-only** response that matches `MeetingSummary` (no prose outside JSON).

Chunking approach (recommended for long transcripts):

1) **Map / notes pass**: summarize each segment (or small batch of segments) into a compact `SegmentNote` JSON:
   - bullet points
   - suggested agenda item(s) referenced
   - extracted candidate decisions / action items / open questions
   - `source_segment_ids`
2) **Reduce pass**: merge `[]SegmentNote` into final `MeetingSummary` JSON.

Agenda referencing:
- If T-005 drift scoring is available later, you can use `best_agenda_item_id` hints.
- For now, the LLM prompt should be given the agenda list and asked to cite `agenda_item_id` whenever it confidently maps content.

### Fallback: deterministic heuristic summary (no-key / offline)

Fallback must be **fully deterministic** and should produce reasonable output for the fixture transcript.

Suggested deterministic heuristics:

- **Purpose / agenda recap**:
  - Use `MeetingInput.transcript.meeting_name` when present.
  - Scan the first 1–2 segments for phrases like `"goal"`, `"today"`, `"agenda"`; otherwise synthesize: “Discussed agenda items and next steps.”
- **Agenda item mapping** (to support agenda references without LLM):
  - Build a token set from each agenda item’s `title`, `description`, and `keywords`.
  - For each segment, compute a simple overlap score (case-insensitive exact word match).
  - Assign the best agenda item when score ≥ threshold (e.g. 1–2 keyword hits); else “unmapped”.
- **Key discussion points**:
  - For each agenda item (in order), include 1–3 bullets derived from the first sentence (or first ~140 chars) of the top-matching segments.
  - Deduplicate bullets by exact string match.
- **Decisions / action items**:
  - If T-008 results are provided, render them directly.
  - Otherwise, do a minimal pattern-based extraction from segment text:
    - decisions: lines containing `"Decision:"` (case-insensitive)
    - action items: lines containing `"Action item"` or `"Action items:"`
- **Open questions / risks**:
  - open questions: sentences containing `?`
  - risks: sentences containing `"risk"`, `"blocker"`, `"concern"`

## Requirements / tasks

### 1) Define the summary model
- Add `MeetingSummary` (and nested types) to `api/internal/model/`.
- Ensure the JSON is versioned (`schema_version: v1`).

### 2) Implement summary generation
- Add a summarizer module (suggested package: `api/internal/summary`).
- Entry point (suggested):
  - `Generate(in model.MeetingInput, segs []model.Segment, opts Options) (model.MeetingSummary, error)`
- Support two modes:
  - `llm` mode when `ANTHROPIC_API_KEY` is present
  - `fallback` mode otherwise

### 3) Pin prompt version + cache results
- Introduce a `prompt_version` constant.
- Store and reuse cached summary outputs by key (see determinism section).

### 4) Ensure agenda references
- Populate `agenda_item_summaries` when mapping confidence is adequate.
- Include `source_segment_ids` for each section where possible.

## Acceptance criteria
- For `docs/fixtures/t-003/meeting_input.example.json` segmented with fixed options (recommended to match existing segmentation tests):
  - `SegmentOptions{MaxTokens: 25, IncludeSpeakerLabels: true}`

  summary generation produces:
  - all required sections in order
  - at least 3 “key discussion point” bullets
  - at least 1 decision and 1 action item **in fallback mode** for the fixture (it contains `Decision:` and `Action items:` lines)
- Output includes:
  - `schema_version = v1`
  - `prompt_version` set
  - `mode` set correctly
  - `source_segment_ids` where applicable
- Re-running summary generation with the same inputs returns the same output (fallback determinism; LLM mode via cache).

## Verify (when implemented)
- Run: `make api-test`
- Add a deterministic fixture-based unit test similar to `TestSegmentTranscript_DeterministicFixture`:
  - load `docs/fixtures/t-003/meeting_input.example.json`
  - compute segments with fixed `SegmentOptions`
  - run summary generation in fallback mode
  - assert key fields/sections and that the output is deterministic
- Optional: compare the produced JSON shape to `docs/fixtures/t-007/meeting_summary.v1.example.json` (example only; not necessarily byte-for-byte).

## Dependencies
- T-004

## Related
- T-008 (optional integration): decisions + action items can be rendered into sections 3–4 when those structured outputs exist.
