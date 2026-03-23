#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
    PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
fi
cd "$PROJECT_ROOT"

echo "==> Running golangci-lint..."
# Use 'go run' to definitively ensure the linter is built with the project's Go version (1.26.1).
# This is the most reliable way to avoid Go version mismatches in CI.
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"

TARGETS=(
    "./server/..."
    "./proto/..."
    "./k8s/operator/..."
)

echo "    Linting targets: ${TARGETS[*]}"
go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run --timeout 20m --fix --config server/.golangci.yml "${TARGETS[@]}"
echo "    golangci-lint OK."

echo "==> Lint complete."
