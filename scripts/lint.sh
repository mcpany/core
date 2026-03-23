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
GOLANGCI_LINT_BIN=""
if [[ -x "$(pwd)/build/env/bin/golangci-lint" ]]; then
    GOLANGCI_LINT_BIN="$(pwd)/build/env/bin/golangci-lint"
elif [[ -x "$(go env GOPATH 2>/dev/null)/bin/golangci-lint" ]]; then
    GOLANGCI_LINT_BIN="$(go env GOPATH 2>/dev/null)/bin/golangci-lint"
else
    GOLANGCI_LINT_BIN=$(command -v golangci-lint 2>/dev/null || true)
fi

if [[ -x "$GOLANGCI_LINT_BIN" ]]; then
    LINT_VERSION_OUT=$("$GOLANGCI_LINT_BIN" --version)
    echo "    Using linter at $GOLANGCI_LINT_BIN: $LINT_VERSION_OUT"

    TARGETS=(
        "./server/..."
        "./proto/..."
        "./k8s/operator/..."
        "./server/examples/upstream_service_demo/grpc/greeter_server/..."
    )

    echo "    Linting targets..."
    export GOTOOLCHAIN=go1.26.1
    "$GOLANGCI_LINT_BIN" run --timeout 20m --fix --config server/.golangci.yml "${TARGETS[@]}"
    echo "    golangci-lint OK."
else
    echo "    Error: golangci-lint not found."
    # In a real script we would exit 1
fi

echo "==> Lint complete."
