#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Change to the root directory of the workspace
cd "$(dirname "$0")/.."

echo "Starting reproducible UI state capture..."

# Run the screenshot generation suite using bazel
./bazelisk run //ui:playwright_tests_generate_docs_screenshots_spec_ts

echo "UI state capture complete. Screenshots saved to ui/docs/screenshots/."
