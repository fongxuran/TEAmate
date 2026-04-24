#!/usr/bin/env bash
set -euo pipefail

# Submit a PENDING PR review using gh cli + GitHub REST API
# Usage:
#   ./submit-pending-review.sh <owner/repo> <pr_number> <payload_json_path>
#
# Notes:
# - PENDING is created by omitting "event" in the payload.
# - Requires: gh auth login

REPO="${1:?owner/repo required}"
PR="${2:?pr number required}"
PAYLOAD="${3:?payload json path required}"

if [[ ! -f "$PAYLOAD" ]]; then
  echo "Payload file not found: $PAYLOAD" >&2
  exit 1
fi

echo "Submitting PENDING review to $REPO PR #$PR using payload: $PAYLOAD"
gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  "/repos/${REPO}/pulls/${PR}/reviews" \
  --input "$PAYLOAD"

echo "Done. Review created as PENDING. You can now convert/submit later (COMMENT/REQUEST_CHANGES) after updates."
