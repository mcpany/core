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

find_tool() {
    local name="$1"
    if [[ -x "$(pwd)/build/env/bin/$name" ]]; then
        echo "$(pwd)/build/env/bin/$name"
    elif [[ -x "$(go env GOPATH 2>/dev/null)/bin/$name" ]]; then
        echo "$(go env GOPATH)/bin/$name"
    else
        command -v "$name" 2>/dev/null || true
    fi
}

echo "==> Running golangci-lint..."
LINT_BIN=$(find_tool golangci-lint)

if [[ -x "$LINT_BIN" ]]; then
    LINT_VER_OUT=$($LINT_BIN --version)
    echo "    Using linter: $LINT_VER_OUT"

    if echo "$LINT_VER_OUT" | grep -q "go1.24"; then
        echo "    ERROR: Linter built with Go 1.24 detected. Project needs Go 1.26+."
        # No exit for session
    else
        MODULES=("./server/..." "./proto/..." "./k8s/operator/...")
        echo "    Linting ${MODULES[*]}..."
        export GOTOOLCHAIN=go1.26.1
        "$LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml "${MODULES[@]}"
        echo "    golangci-lint OK."
    fi
else
    echo "    Error: golangci-lint not found."
fi

echo "==> Lint complete."
