#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Change to the root directory of the workspace
cd "$(dirname "$0")/.."

echo "Starting reproducible UI state capture..."

# Run the screenshot generation suite using bazel
./bazelisk run //ui:playwright_tests_generate_docs_screenshots_spec_ts

echo "Copying screenshots to docs/screenshots/..."
mkdir -p docs/screenshots
cp -r ui/docs/screenshots/* docs/screenshots/

echo "UI state capture complete. Screenshots saved to docs/screenshots/."
