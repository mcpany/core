---
description: Rebase the mentioned branch and test
---

Objective: Rebase the [target branch or current branch] onto the latest main, resolve any conflicts, verify code quality, and push to remote.

Execution Steps:

Sync & Rebase:
git fetch origin main and rebase the target branch onto origin/main.
Conflict Resolution: If merge conflicts arise, analyze the code to resolve them logically. If a resolution is ambiguous, stop and ask for clarification.
Verification (Quality Gate):
Run make lint and resolve any linting/formatting errors.
Run make test and ensure all unit tests pass.
CI Alignment: Check the .github/workflows directory to identify and run additional validation commands (e.g., make e2e, make k8s-e2e) typically used in GitHub Actions.
Completion:
Only if all previous steps pass (100% success), push the rebased changes to the remote branch using git push origin [branch-name] --force-with-lease.
Constraints & Requirements:

No Partial Pushes: Do not push unless both linting and tests pass completely.
Backtracking: If a test fails after rebase, investigate if it's a regression caused by the rebase and fix it before proceeding.
Safety: Always use --force-with-lease instead of a blind --force.
