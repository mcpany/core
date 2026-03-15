#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# This script is intended to be run via Bazel: bazel run //:lint
# It wraps the existing linting logic from the Makefile.

if [ -n "$BUILD_WORKSPACE_DIRECTORY" ]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT=$(bazel info workspace)
fi
cd "$PROJECT_ROOT"

echo "Running Gazelle..."
bazel run //:gazelle

echo "Running Buildifier..."
bazel run //:buildifier -- -r .

echo "Running golangci-lint..."
# We assume golangci-lint is already installed/available in the expected path or in PATH.
# If not, it will fail with a helpful message.
GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-build/env/bin/golangci-lint}"
if ! [ -x "$GOLANGCI_LINT_BIN" ]; then
    GOLANGCI_LINT_BIN=$(which golangci-lint 2>/dev/null || true)
fi

if [ -z "$GOLANGCI_LINT_BIN" ]; then
    echo "Error: golangci-lint not found. Please run 'make prepare' first."
    exit 1
fi

"$GOLANGCI_LINT_BIN" run --timeout 20m --fix ./server/cmd/... ./server/pkg/... ./server/tests/... ./server/examples/...

echo "Running pre-commit..."
if command -v pre-commit >/dev/null 2>&1; then
    pre-commit run --config server/.pre-commit-config.yaml --all-files
else
    echo "Warning: pre-commit not found. Skipping."
fi
