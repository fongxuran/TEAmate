# T-005 — Drift scoring v1 (LLM classifier + fallback embeddings)

## Goal
Compute a drift score per segment relative to the agenda.

## Requirements / tasks
- Primary (current approach): LLM-based drift classification
  - Prompt an LLM with:
    - agenda items
    - current segment text
    - optional prior segment context
  - Output per segment:
    - `best_agenda_item_id`
    - `label` in {`on_track`, `maybe_drift`, `drift`}
    - `rationale` (short, for explainability)
    - `confidence` (0-1)
  - **Claude mode requires** `ANTHROPIC_API_KEY`.
- Fallback (offline mode): embedding similarity baseline
  - Create embeddings for agenda items + each segment
  - Compute similarity and assign:
    - `best_agenda_item_id`
    - `similarity`
    - `drift_score` (calibrated from similarity)
- Provide thresholds:
  - `on_track`, `maybe_drift`, `drift`
- Keep everything runnable locally:
  - Claude mode: local app + external API call
  - Offline mode: local embedding model (e.g., `sentence-transformers`) or Ollama embeddings if available

## Acceptance criteria
- For a sample transcript + agenda, the system outputs drift labels per segment.
- Drift scoring supports:
  - Claude mode (requires API key)
  - Offline mode (no network dependency)

## Dependencies
- T-004
