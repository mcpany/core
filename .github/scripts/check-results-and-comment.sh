#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Inputs via environment:
# NEEDS_JSON
# PR_NUMBER
# REPO
# RUN_ID
# (Optional) CI_PIPELINE_URL for Woodpecker

NEEDS_JSON="${NEEDS_JSON:-'{}'}"
PR_NUMBER="${PR_NUMBER:-$CI_COMMIT_PULL_REQUEST}"
REPO="${REPO:-$CI_REPO}"
RUN_ID="${RUN_ID:-$CI_BUILD_NUMBER}"

if [ -n "$CI_PIPELINE_URL" ]; then
  RUN_URL="$CI_PIPELINE_URL"
  CI_SYSTEM="Woodpecker CI"
else
  RUN_URL="https://github.com/${REPO}/actions/runs/${RUN_ID}"
  CI_SYSTEM="GitHub Actions"
fi

FAILED_JOBS=""

# Try python3 first (available on GitHub Actions ubuntu/oci runners)
if command -v python3 >/dev/null 2>&1; then
cat << 'PYEOF' > check_failures.py
import os, json, sys
try:
    needs = json.loads(os.environ.get("NEEDS_JSON", "{}"))
    failed_jobs = [k for k, v in needs.items() if isinstance(v, dict) and v.get("result") in ["failure", "cancelled"]]
    if failed_jobs:
        print(",".join(failed_jobs))
except Exception as e:
    print(f"Error parsing JSON: {e}", file=sys.stderr)
    sys.exit(1)
PYEOF
    FAILED_JOBS=$(python3 check_failures.py)
    rm -f check_failures.py

# Fallback to jq (available on Woodpecker iowoi/mcpany-runner-bot)
elif command -v jq >/dev/null 2>&1; then
    if ! echo "$NEEDS_JSON" | jq . >/dev/null 2>&1; then
        echo "Error parsing JSON with jq" >&2
        exit 1
    fi
    FAILED_JOBS=$(echo "$NEEDS_JSON" | jq -r 'to_entries | map(select(.value | type == "object" and (.value.result == "failure" or .value.result == "cancelled"))) | .[].key' | paste -sd, -)
else
    echo "Neither python3 nor jq is available for JSON parsing." >&2
    exit 1
fi

if [ -z "$FAILED_JOBS" ]; then
    echo "All jobs passed."
    exit 0
fi

echo "Failed jobs: $FAILED_JOBS"

if [ -z "$PR_NUMBER" ] || [ "$PR_NUMBER" == "null" ] || [ "$PR_NUMBER" == "false" ]; then
    echo "Not a PR, skipping comment."
    exit 1
fi

COMMENT_BODY_FILE="comment_body.md"

cat << MARKDOWN > "$COMMENT_BODY_FILE"
❌ **CI Checks Failed**

The following $CI_SYSTEM checks failed in this run:
MARKDOWN

IFS=',' read -ra JOBS <<< "$FAILED_JOBS"
for job in "${JOBS[@]}"; do
    if [ -n "$job" ]; then
        echo "- **$job: Failed**" >> "$COMMENT_BODY_FILE"
    fi
done

cat << MARKDOWN >> "$COMMENT_BODY_FILE"

---
Please investigate the failures above.
1. Check the [Action logs]($RUN_URL) for the specific error messages.
2. Analyze the root cause.
3. Apply fixes to the codebase.
MARKDOWN

echo "Posting comment to PR #$PR_NUMBER..."
gh issue comment "$PR_NUMBER" --body-file "$COMMENT_BODY_FILE" --repo "$REPO" || echo "Failed to post comment, gh cli might not be configured."

echo "::error::Workflow failed due to job dependencies."
exit 1
