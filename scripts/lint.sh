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

export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"
LINT_DIR="${PROJECT_ROOT}/build/env/bin"
mkdir -p "$LINT_DIR"

LINT_BIN="$LINT_DIR/golangci-lint"

# Sync workspace
echo "    Syncing workspace..."
go work sync

# Re-install if version or toolchain changed
if [[ ! -x "$LINT_BIN" ]] || ! "$LINT_BIN" --version | grep -q "${LINT_VERSION}" || ! "$LINT_BIN" --version | grep -q "go1.26"; then
    echo "    Installing golangci-lint ${LINT_VERSION} built with Go 1.26.1 to $LINT_DIR..."
    GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi

echo "    Using linter: $($LINT_BIN --version)"

# Absolute path to config
CONFIG_PATH="${PROJECT_ROOT}/server/.golangci.yml"

echo "    Linting modules from root with workspace enabled..."
# Running from root with workspace should work.
# We target all modules explicitly.
"$LINT_BIN" run --timeout 20m --fix --config "$CONFIG_PATH" \
    ./server/... \
    ./proto/... \
    ./k8s/operator/... \
    ./server/examples/upstream_service_demo/grpc/greeter_server/...

echo "==> Lint complete."
