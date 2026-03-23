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

echo "==> Running golangci-lint via go run (version locked)..."
# This is the only reliable way to ensure the linter is built with the correct Go version
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"

# We must use GOWORK=off to avoid the typechecking error about directory being outside module roots,
# then we run on each module individually.
export GOWORK=off

MODULES=("server" "proto" "k8s/operator")

for mod in "${MODULES[@]}"; do
    if [ -d "$mod" ]; then
        echo "    Linting module: $mod"
        (cd "$mod" && go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run --timeout 20m --fix --config ../server/.golangci.yml ./...)
    fi
done

echo "    golangci-lint OK."
echo "==> Lint complete."
