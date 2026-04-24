# T-002 — Repo scaffolding + dev tooling (local)

## Goal
Create a runnable local project skeleton with repeatable dev commands.

## Requirements / tasks
- Initialize the project structure (example for Python + Streamlit):
  - `app/` (UI)
  - `teammate/` (core library)
  - `tests/`
  - `docs/` (already exists)
- Add dependency management:
  - `pyproject.toml` using **uv** or **poetry**
- Add local dev commands (e.g., `make dev`, or documented `uv run ...`).
- Add `.env.example` for configuration:
  - `LLM_PROVIDER=ollama|claude|openai`
  - `OLLAMA_HOST=http://localhost:11434`
  - optional `ANTHROPIC_API_KEY=...`
  - optional `ANTHROPIC_MODEL=...`
  - optional `OPENAI_API_KEY=...`
- Add basic formatting/linting (optional but helpful):
  - `ruff` + `black` (or just `ruff format`)

## Acceptance criteria
- New developer can run the app locally following README steps.
- `git status` shows scaffold files committed.

## Dependencies
- T-001
