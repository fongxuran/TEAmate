---
description: Orchestrated PR review (selective + inline). Usage: /review-orchestrated <PR_NUMBER>
---

# Orchestrated Code Review (Selective + Inline)

## Setup
- PR: $ARGUMENTS
- Session ID: !`REVIEW_SESSION_ID="session-$(date +%Y%m%d-%H%M%S)" && echo "$REVIEW_SESSION_ID"`
- Session Dir: !`echo ".claude/reviews/$ARGUMENTS/$REVIEW_SESSION_ID"`
- Repo: !`gh repo view --json nameWithOwner -q .nameWithOwner`
- Commit SHA: !`gh pr view $ARGUMENTS --json headRefOid -q .headRefOid`

## Run
Call **orchestrator-review-agent** with:
- PR Number: $ARGUMENTS
- Repository: [Repo above]
- Commit SHA: [Commit SHA above]
- Session ID: [Session ID above]
- Session Directory: [Session Dir above]

## Expected behavior
- Orchestrator auto-stashes local changes, checks out PR head once, and leaves stash untouched.
- Orchestrator writes canonical patch file to `[Session Dir]/pr.patch` and passes it to all subagents.
- Consolidated findings deduplicate overlapping reviewer issues aggressively.
- `[nitpick]` findings are optional (`priority=P3`, `is_required=false`) and are not required to fix.

## Required outputs
- `[Session Dir]/consolidated-review.json`
- `[Session Dir]/review-post-body.md`
- `[Session Dir]/github-review-payload.json` with inline `comments[]` when findings are anchorable

## Submit (user action)
`.github/scripts/submit-pending-review.sh <OWNER/REPO> <PR_NUMBER> <SESSION_DIR>/github-review-payload.json`
