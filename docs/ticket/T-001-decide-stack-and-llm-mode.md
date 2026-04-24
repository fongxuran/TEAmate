# T-001 — Decide MVP stack + LLM mode

## Goal
Lock down a buildable local MVP plan: tech stack + which “AI brain” we will use locally.

## Requirements / tasks
- Choose one primary implementation path (recommended):
  - **Option A (recommended): Python + Streamlit UI**
  - Option B: Python + FastAPI API + simple web UI
  - Option C: Node/TS + minimal UI
- Choose LLM execution mode:
  - **Local-first:** Ollama (e.g., `llama3.1`, `qwen2.5`) + local embeddings
  - Hybrid: Ollama for summaries + local embeddings
  - **Claude API mode:** Anthropic Claude for drift detection/classification (runs locally but requires `ANTHROPIC_API_KEY`)
  - Optional: OpenAI API fallback (still runs locally but calls cloud)
- Decide the “first integration” for ticket outputs:
  - **Local export only (JSON/CSV/MD)** for MVP
  - Optional: Jira REST API connector behind a feature flag

## Acceptance criteria
- A short written decision in this ticket file:
  - selected stack
  - selected LLM mode
  - selected export/integration approach
  - repo structure outline (folders)

## Notes
- Keeping LLM provider pluggable reduces lock-in.
- Streamlit is usually the fastest demo for hackathon-style MVP.
