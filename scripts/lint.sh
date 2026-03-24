#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

echo "==> Running lint automation..."

export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"
LINT_DIR="${PROJECT_ROOT}/build/env/bin"
mkdir -p "$LINT_DIR"

LINT_BIN="$LINT_DIR/golangci-lint"

if [[ ! -x "$LINT_BIN" ]] || ! "$LINT_BIN" --version | grep -q "${LINT_VERSION}" || ! "$LINT_BIN" --version | grep -q "go1.26"; then
    echo "    Installing golangci-lint ${LINT_VERSION} with Go 1.26.1..."
    GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi

echo "    Using linter: $($LINT_BIN --version)"

# Sync workspace
echo "    Syncing workspace..."
go work sync

CONFIG_PATH="${PROJECT_ROOT}/server/.golangci.yml"

# We lint modules that have Go files and are part of the workspace.
# We explicitly target directories with source code to avoid "no Go files" errors.

echo "    Linting..."

# server module
"$LINT_BIN" run --timeout 10m --fix --config "$CONFIG_PATH" ./server/...

# operator module
"$LINT_BIN" run --timeout 10m --fix --config "$CONFIG_PATH" ./k8s/operator/...

# greeter_server module (server package only)
if [ -d "server/examples/upstream_service_demo/grpc/greeter_server/server" ]; then
    "$LINT_BIN" run --timeout 5m --fix --config "$CONFIG_PATH" ./server/examples/upstream_service_demo/grpc/greeter_server/server/...
fi

echo "==> Lint complete."
