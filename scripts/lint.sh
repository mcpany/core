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
# Use 'go run' as the primary and ONLY method in CI to guarantee Go 1.26.1 alignment.
# This eliminates all binary/cache/path issues.
export GOTOOLCHAIN=go1.26.1
export GOWORK=on
LINT_VERSION="v1.64.5"

TARGETS=(
    "./server/..."
    "./proto/..."
    "./k8s/operator/..."
)

echo "    Linting targets: ${TARGETS[*]}"
echo "    Executing: go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run ..."
go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run --timeout 20m --fix --config server/.golangci.yml "${TARGETS[@]}"
echo "    golangci-lint OK."

echo "==> Lint complete."
