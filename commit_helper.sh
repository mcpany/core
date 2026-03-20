#!/bin/bash
git diff --name-only | grep -v "_test.go" | grep -v "cmd/mocks" | grep -v "tests/" | grep -v "examples/" | xargs git add
git add README.md
git commit -m "docs: comprehensive documentation overhaul

- Added structured docstrings (Summary, Parameters, Returns, Errors, Side Effects) to all public symbols avoiding empty calorie comments.
- Updated README to align with Gold Standard requirements."
