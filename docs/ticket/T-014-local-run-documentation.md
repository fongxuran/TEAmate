# T-014 — Local run documentation

## Goal
Make it easy to install and run TEAmate MVP locally.

## Requirements / tasks
- Add `README.md` (root or `docs/`) with:
  - prerequisites (Go version, Node.js version)
  - install dependencies
  - configure env vars
  - run the app
  - run tests
- Add troubleshooting section:
  - WebSocket connection issues (CORS/origin, proxy)
  - drift prompts not appearing (WebSocket connected, thresholds too strict, LLM mode latency)
  - missing `ANTHROPIC_API_KEY` (fallback mode expectations)
  - missing `MOTION_API_KEY` (local export still works)

## Acceptance criteria
- A new developer can run the MVP from scratch following docs.

## Dependencies
- T-011
