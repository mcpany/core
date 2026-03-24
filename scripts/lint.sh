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

echo "==> Running lint automation..."

# Use Go 1.26.1 as it matches the project's go.work and go.mod files.
# If we are in the cimg/go:1.26.1 image, this should be fine.
export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"

# If we are in CircleCI or a similar environment, we might want to use a specific cache dir.
LINT_DIR="${PROJECT_ROOT}/build/env/bin"
mkdir -p "$LINT_DIR"

LINT_BIN="$LINT_DIR/golangci-lint"

# Re-install if version or toolchain changed
LINT_VER_CHECK=$("$LINT_BIN" --version 2>/dev/null | grep "${LINT_VERSION}" | grep "go1.26") || true
if [[ -z "$LINT_VER_CHECK" ]]; then
    echo "    Installing golangci-lint ${LINT_VERSION} built with Go 1.26.1 to $LINT_DIR..."
    GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi

echo "    Using linter: $($LINT_BIN --version)"

# To avoid "outside module roots" or "no Go files" errors, we run the linter
# using the workspace, but we point it at the module directories.
# We also ensure the config path is absolute to avoid any relative path issues.
CONFIG_PATH="$(pwd)/server/.golangci.yml"

echo "    Linting modules from root with workspace enabled..."
"$LINT_BIN" run --timeout 20m --fix --config "$CONFIG_PATH" \
    ./server/... \
    ./proto/... \
    ./k8s/operator/... \
    ./server/examples/upstream_service_demo/grpc/greeter_server/...

echo "==> Lint complete."
