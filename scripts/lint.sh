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
# Use 'go run' as a fail-safe to guarantee Go 1.26.1 compatibility
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"

# Lint all Go modules in the workspace
for d in server proto k8s/operator; do
    if [ -d "$d" ]; then
        echo "    Linting $d..."
        go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run --timeout 20m --fix --config server/.golangci.yml "./$d/..."
    fi
done
echo "    golangci-lint OK."

echo "==> Lint complete."
