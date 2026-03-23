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

echo "==> Running Buildifier..."
if command -v buildifier >/dev/null 2>&1; then
    find . -name "BUILD" -o -name "BUILD.bazel" -o -name "*.bzl" -not -path "./build/*" -exec buildifier {} +
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found."
fi

echo "==> Running Gazelle..."
if command -v gazelle >/dev/null 2>&1; then
    gazelle -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle not found."
fi

echo "==> Running golangci-lint..."
# Try several locations for the linter, prioritizing local build/env/bin and GOPATH
LINT_BIN=""
if [[ -x "build/env/bin/golangci-lint" ]]; then
    LINT_BIN="$(pwd)/build/env/bin/golangci-lint"
elif [[ -x "$(go env GOPATH 2>/dev/null)/bin/golangci-lint" ]]; then
    LINT_BIN="$(go env GOPATH)/bin/golangci-lint"
elif [[ -x "$HOME/go/bin/golangci-lint" ]]; then
    LINT_BIN="$HOME/go/bin/golangci-lint"
else
    LINT_BIN=$(command -v golangci-lint 2>/dev/null || true)
fi

if [[ -x "$LINT_BIN" ]]; then
    LINT_VERSION_OUT=$("$LINT_BIN" --version)
    echo "    Using linter at $LINT_BIN: $LINT_VERSION_OUT"

    # Check if linter is too old for Go 1.26
    if echo "$LINT_VERSION_OUT" | grep -q "go1.24"; then
        echo "    WARNING: Linter built with Go 1.24 might fail on Go 1.26 code."
        # If we are in the sandbox and have go 1.26, we can try to build it
        if go version | grep -q "1.26"; then
             echo "    Attempting to build a compatible linter..."
             GOTOOLCHAIN=go1.26.1 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
             LINT_BIN="$(go env GOPATH)/bin/golangci-lint"
             echo "    New linter version: $($LINT_BIN --version)"
        fi
    fi

    TARGETS=(
        "./server/..."
        "./proto/..."
        "./k8s/operator/..."
    )

    echo "    Linting targets..."
    # Ensure toolchain is set for the run
    export GOTOOLCHAIN=go1.26.1
    "$LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml "${TARGETS[@]}"
    echo "    golangci-lint OK."
else
    # Fallback to 'go run' if no binary is found - the ultimate fallback
    echo "    Linter binary not found, falling back to 'go run'..."
    export GOTOOLCHAIN=go1.26.1
    go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 run --timeout 20m --fix --config server/.golangci.yml ./server/... ./proto/... ./k8s/operator/...
    echo "    golangci-lint OK (via go run)."
fi

echo "==> Lint complete."
