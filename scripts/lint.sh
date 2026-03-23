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
    # 1. Check build/env/bin (Highest priority for CI stability)
    if [[ -x "build/env/bin/$name" ]]; then
        echo "$(pwd)/build/env/bin/$name"
        return 0
    fi
    # 2. Check GOPATH/bin
    local gopath_bin
    gopath_bin=$(go env GOPATH 2>/dev/null)/bin/"$name"
    if [[ -x "$gopath_bin" ]]; then
        echo "$gopath_bin"
        return 0
    fi
    # 3. Check PATH
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
GOLANGCI_LINT_BIN="$(find_tool golangci-lint)"

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    # Verify Go version of the linter
    LINT_VERSION_OUT=$("$GOLANGCI_LINT_BIN" --version)
    echo "    Using linter: $LINT_VERSION_OUT"

    # We run on specific directories to avoid workspace boundary issues and "no Go files" errors.
    # Note: we run from root so it sees go.work.
    # We use space-separated list of paths.
    TARGETS="./server/... ./proto/... ./k8s/operator/... ./server/examples/upstream_service_demo/grpc/greeter_server/..."

    echo "    Linting targets..."
    "$GOLANGCI_LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml $TARGETS
    echo "    golangci-lint OK."
else
    echo "    Warning: golangci-lint not found."
fi

echo "==> Lint complete."
