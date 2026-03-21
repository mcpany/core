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
    command -v "$name" 2>/dev/null || true
}

echo "==> Running Buildifier..."
BUILDIFIER_BIN="$(find_tool buildifier)"
if [[ -n "$BUILDIFIER_BIN" ]]; then
    find . -name "BUILD" -o -name "BUILD.bazel" -o -name "*.bzl" -not -path "./build/*" -exec "$BUILDIFIER_BIN" {} +
    echo "    Buildifier OK."
else
    echo "    Warning: buildifier not found."
fi

echo "==> Running Gazelle..."
GAZELLE_BIN="$(find_tool gazelle)"
if [[ -n "$GAZELLE_BIN" ]]; then
    "$GAZELLE_BIN" -repo_root="$PROJECT_ROOT"
    echo "    Gazelle OK."
else
    echo "    Warning: gazelle not found."
fi

echo "==> Running golangci-lint..."
# Try to find a Go 1.26-built version first
GOLANGCI_LINT_BIN="$(go env GOPATH)/bin/golangci-lint"
if [[ ! -x "$GOLANGCI_LINT_BIN" ]]; then
    GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"
fi

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    # Filter modules to only those that exist and contain Go files
    MODULES=()
    for d in server proto k8s/operator server/examples/upstream_service_demo/grpc/greeter_server; do
        if [ -d "$d" ]; then
            # Use find to check for Go files, being careful about module boundaries if needed
            # For simplicity, we just check if any .go files exist in the dir tree
            if find "$d" -name "*.go" | grep -q .; then
                MODULES+=("./$d/...")
            fi
        fi
    done

    if [ ${#MODULES[@]} -gt 0 ]; then
        echo "    Linting ${MODULES[*]}..."
        "$GOLANGCI_LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml "${MODULES[@]}"
        echo "    golangci-lint OK."
    else
        echo "    Warning: No Go modules found to lint."
    fi
else
    echo "    Warning: golangci-lint not found."
fi

echo "==> Lint complete."
