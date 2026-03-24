#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Support running from any directory
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

echo "==> Running lint automation..."

export GOTOOLCHAIN=go1.26.1
LINT_VERSION="v1.64.5"
LINT_DIR="${PROJECT_ROOT}/build/env/bin"
mkdir -p "$LINT_DIR"

LINT_BIN="$LINT_DIR/golangci-lint"

# Install golangci-lint if not present
if [[ ! -x "$LINT_BIN" ]] || ! "$LINT_BIN" --version | grep -q "${LINT_VERSION}" || ! "$LINT_BIN" --version | grep -q "go1.26"; then
    echo "    Installing golangci-lint ${LINT_VERSION} with Go 1.26.1..."
    GOBIN="$LINT_DIR" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi

echo "    Using linter: $($LINT_BIN --version)"

# Sync workspace
echo "    Syncing workspace..."
go work sync

# Absolute path to config
CONFIG_PATH="${PROJECT_ROOT}/server/.golangci.yml"

# Explicitly lint the important modules to avoid issues with empty directories or patterns.
echo "    Linting modules..."

echo "    -> Linting server..."
"$LINT_BIN" run --timeout 10m --fix --config "$CONFIG_PATH" ./server/...

echo "    -> Linting k8s/operator..."
"$LINT_BIN" run --timeout 10m --fix --config "$CONFIG_PATH" ./k8s/operator/...

echo "    -> Linting proto..."
# We use the config but keep in mind that .pb.go files are excluded by the config itself.
"$LINT_BIN" run --timeout 5m --fix --config "$CONFIG_PATH" ./proto/...

echo "    -> Linting greeter_server..."
if [ -d "server/examples/upstream_service_demo/grpc/greeter_server/server" ]; then
    "$LINT_BIN" run --timeout 5m --fix --config "$CONFIG_PATH" ./server/examples/upstream_service_demo/grpc/greeter_server/server/...
fi

echo "==> Lint complete."
