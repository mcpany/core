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
# We skip 'proto' because it only contains generated files (excluded by config) and tests (ignored by config).
echo "    Linting modules..."
"$LINT_BIN" run --timeout 20m --fix --config "$CONFIG_PATH" \
    "./server/..." \
    "./k8s/operator/..." \
    "./server/examples/upstream_service_demo/grpc/greeter_server/..."

echo "==> Lint complete."
