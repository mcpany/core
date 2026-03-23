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
# Use 'go run' to definitively ensure the linter is built with the project's Go version (1.26.1).
# This prevents the "Go language version used to build golangci-lint is lower than the targeted Go version" error.
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"

TARGETS=(
    "./server/..."
    "./proto/..."
    "./k8s/operator/..."
    "./server/examples/upstream_service_demo/grpc/greeter_server/..."
)

echo "    Linting targets: ${TARGETS[*]}"
go run github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION} run --timeout 20m --fix --config server/.golangci.yml "${TARGETS[@]}"
echo "    golangci-lint OK."

echo "==> Lint complete."
