#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "Starting UI test reproducibility script..."

# Ensure we are in the repo root
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "Running Playwright screenshot generation via Bazel..."
# We use the playwright target specifically designed to generate doc screenshots
bazelisk test //ui:playwright_tests_generate_docs_screenshots_spec_ts --test_output=all --test_timeout=1200

echo "Copying generated screenshots to docs/screenshots/..."
mkdir -p docs/screenshots
# Assuming successful test run outputs to ui/docs/screenshots based on config
cp -r ui/docs/screenshots/*.png docs/screenshots/

echo "Screenshots successfully generated and copied to docs/screenshots/"
