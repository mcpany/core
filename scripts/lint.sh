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
    # Explicitly check GOPATH first
    local gp
    gp=$(go env GOPATH 2>/dev/null)
    if [[ -x "$gp/bin/$name" ]]; then
        echo "$gp/bin/$name"
        return 0
    fi
    # Then check build/env/bin
    if [[ -x "build/env/bin/$name" ]]; then
        echo "$(pwd)/build/env/bin/$name"
        return 0
    fi
    # Finally check PATH
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
# If linter is already in PATH and is the right version, use it.
# Otherwise try to find it.
LINT_BIN=$(find_tool golangci-lint)

if [[ -x "$LINT_BIN" ]]; then
    echo "    Using linter: $($LINT_BIN --version)"

    TARGETS="./server/... ./proto/... ./k8s/operator/... ./server/examples/upstream_service_demo/grpc/greeter_server/..."

    echo "    Linting targets..."
    "$LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml $TARGETS
    echo "    golangci-lint OK."
else
    echo "    Warning: golangci-lint not found."
fi

echo "==> Lint complete."
